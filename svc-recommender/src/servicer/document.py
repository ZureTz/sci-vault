"""Document servicer - implements RecommenderService.EnrichDocument."""

import concurrent.futures
import logging
import time
from typing import Optional

import grpc
import google.genai as genai
from pydantic import BaseModel, Field

from google.genai import types
from infrastructure.genai import DEFAULT_MODEL
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


class DocumentMetadata(BaseModel):
    year: Optional[int] = Field(
        None, description="Publication year (e.g. 2024). Null if not found."
    )
    doi: Optional[str] = Field(
        None, description="DOI string (e.g. '10.1145/...'). Null if not found."
    )
    authors: list[str] = Field(description="List of author full names.")
    summary: str = Field(description="Concise 3-5 sentence summary of the paper.")
    tags: list[str] = Field(description="5-10 relevant keywords or topic labels.")


class DocumentServicer:
    """Implements the EnrichDocument RPC; Python owns all enrich-status writes."""

    def __init__(
        self,
        enrich_cache: EnrichmentStatusCache,
        doc_repo: DocumentRepository,
        doc_storage: DocumentStorage,
        genai_client: genai.Client,
    ) -> None:
        self._cache = enrich_cache
        self._doc_repo = doc_repo
        self._storage = doc_storage
        self._genai = genai_client

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
        embedding: Optional[list[float]] = None

        last_exc: Exception | None = None
        for attempt in range(1, _ENRICH_MAX_ATTEMPTS + 1):
            try:
                if metadata is None:
                    metadata = self._extract_metadata(pdf_bytes)
                    log.info(
                        "metadata extracted: doc_id=%d authors=%s tags=%s",
                        doc_id,
                        metadata.authors,
                        metadata.tags,
                    )

                # TODO: call embedding model using summary -> 1536-dim vector
                # if embedding is None:
                #     embedding = self._compute_embedding(metadata.summary)

                self._doc_repo.write_enrichment_and_done(
                    doc_id=doc_id,
                    authors=list(metadata.authors),
                    summary=metadata.summary,
                    tags=list(metadata.tags),
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

    # ------------------------------------------------------------------ #
    # Helpers                                                              #
    # ------------------------------------------------------------------ #

    def _extract_metadata(self, pdf_bytes: bytes) -> DocumentMetadata:
        """Call LLM with the PDF directly to extract authors, summary, and tags."""
        response = self._genai.models.generate_content(
            model=DEFAULT_MODEL,
            contents=[
                types.Part.from_bytes(data=pdf_bytes, mime_type="application/pdf"),
                (
                    "You are an academic paper analyst. "
                    "Extract the following from this paper:\n"
                    "- authors: list of author full names\n"
                    "- summary: a concise 3-5 sentence summary of the paper\n"
                    "- tags: 5-10 relevant keywords or topic labels\n"
                    "- year: publication year as an integer (null if not found)\n"
                    "- doi: DOI string e.g. '10.1145/...' (null if not found)"
                ),
            ],
            config={
                "response_mime_type": "application/json",
                "response_json_schema": DocumentMetadata.model_json_schema(),
            },
        )
        if not response.text:
            raise ValueError("LLM returned empty response")
        return DocumentMetadata.model_validate_json(response.text)
