"""Document servicer — implements RecommenderService.EnrichDocument."""

import concurrent.futures
import logging
import time
from typing import Optional

import grpc
import numpy as np

from genai.document import DocumentGenAI, DocumentMetadata
from pb.recommender.v1 import recommender_pb2
from cache.enrichment import EnrichmentStatusCache
from repository.document import DocumentRepository
from storage.document import DocumentStorage

log = logging.getLogger(__name__)

_ENRICH_MAX_ATTEMPTS = 3
_ENRICH_RETRY_BASE_DELAY = 2.0  # seconds; doubles each attempt

_enrichment_executor = concurrent.futures.ThreadPoolExecutor(
    max_workers=4, thread_name_prefix="enrich"
)


class DocumentServicer:
    """Implements the EnrichDocument RPC; Python owns all enrich-status writes."""

    def __init__(
        self,
        enrich_cache: EnrichmentStatusCache,
        doc_repo: DocumentRepository,
        doc_storage: DocumentStorage,
        genai: DocumentGenAI,
    ) -> None:
        self._cache = enrich_cache
        self._doc_repo = doc_repo
        self._storage = doc_storage
        self._genai = genai

    def EnrichDocument(
        self,
        request: recommender_pb2.EnrichDocumentRequest,
        context: grpc.ServicerContext,
    ) -> recommender_pb2.EnrichDocumentResponse:
        """Set Redis -> pending, queue background job, return ACK immediately."""
        doc_id: int = request.doc_id
        file_key: str = request.file_key

        # Mark as pending before the task even enters the thread pool.
        self._cache.set_pending(doc_id)

        _enrichment_executor.submit(self._run_enrichment, doc_id, file_key)

        log.info("EnrichDocument queued: doc_id=%d file_key=%s", doc_id, file_key)
        return recommender_pb2.EnrichDocumentResponse(accepted=True)

    # ------------------------------------------------------------------ #
    # Background worker                                                    #
    # ------------------------------------------------------------------ #

    def _run_enrichment(self, doc_id: int, file_key: str) -> None:
        """Enrichment pipeline with retries. On success writes DB done; on failure leaves DB unchanged."""
        self._cache.set_processing(doc_id)
        log.info("enrichment started: doc_id=%d", doc_id)

        pdf_bytes = self._storage.download_pdf(file_key)
        log.info("downloaded PDF: doc_id=%d size=%d bytes", doc_id, len(pdf_bytes))

        # Cached results — each step is only called if not yet succeeded.
        metadata: Optional[DocumentMetadata] = None
        embedding: Optional[np.ndarray] = None

        last_exc: Exception | None = None
        for attempt in range(1, _ENRICH_MAX_ATTEMPTS + 1):
            try:
                if metadata is None:
                    metadata = self._genai.extract_metadata(pdf_bytes)
                    log.info(
                        "metadata extracted: doc_id=%d title=%s authors=%s tags=%s",
                        doc_id,
                        metadata.title,
                        metadata.authors,
                        metadata.tags,
                    )

                if embedding is None:
                    embedding = self._genai.compute_embedding(metadata.summary)

                self._doc_repo.write_enrichment_and_done(
                    doc_id=doc_id,
                    title=metadata.title,
                    authors=metadata.authors,
                    summary=metadata.summary,
                    tags=metadata.tags,
                    year=metadata.year,
                    doi=metadata.doi,
                    embedding=embedding,
                )
                self._cache.set_done(doc_id)
                log.info("enrichment done: doc_id=%d", doc_id)
                return

            except Exception as exc:
                last_exc = exc
                if attempt < _ENRICH_MAX_ATTEMPTS:
                    delay = _ENRICH_RETRY_BASE_DELAY * (2 ** (attempt - 1))
                    log.warning(
                        "enrichment attempt %d/%d failed: doc_id=%d, retrying in %.1fs: %s",
                        attempt,
                        _ENRICH_MAX_ATTEMPTS,
                        doc_id,
                        delay,
                        exc,
                    )
                    time.sleep(delay)

        log.exception(
            "enrichment failed after %d attempts: doc_id=%d",
            _ENRICH_MAX_ATTEMPTS,
            doc_id,
            exc_info=last_exc,
        )
        # DB stays not_started; Redis records the failure for polling.
        self._cache.set_failed(doc_id)
