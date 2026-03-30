"""Document servicer - implements RecommenderService.EnrichDocument."""

import concurrent.futures
import logging
import time
from typing import Optional

import grpc
import google.genai as genai
import numpy as np
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
    authors: list[str] = Field(
        description="List of author full names (e.g. ['John Doe', 'Jane Smith'])."
    )
    summary: str = Field(
        description=(
            "High-density core summary in academic English, strictly 150-300 words. "
            "Must cover research background, core methodology, and key results/conclusions. "
            "Maximize semantic information density for downstream vector embedding. "
            "Exclude trivial experimental setups, formulas, and reference noise."
        )
    )
    tags: list[str] = Field(
        description="5-10 highly relevant technical keywords or topic labels in English."
    )
    year: Optional[int] = Field(
        None,
        description="Publication year as an integer (e.g. 2024). Null if not explicitly found.",
    )
    doi: Optional[str] = Field(
        None,
        description="Official DOI string exactly as found (e.g. '10.1145/1234.5678'). Null if not explicitly found.",
    )


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
        embedding: Optional[np.ndarray] = None

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

                if embedding is None:
                    embedding = self._compute_embedding(metadata.summary)

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
                    """
                    You are an expert academic paper analyst. Your task is to extract highly condensed metadata from the provided academic document. 
                    1. "authors": A list of strings containing author full names (e.g., ["John Doe", "Jane Smith"]). 
                    2. "summary": A high-density core summary written in academic English. 
                        - Structure: It MUST cover the research background, core methodology, and key results/conclusions.
                        - Length constraint: Strictly between 150 and 300 words (approximately 7 to 15 sentences).
                        - Maximize semantic information density for downstream Vector Embedding. Exclude trivial experimental setups, formulas, and reference noise.
                    3. "tags": A list of 5 to 10 highly relevant technical keywords or topic labels in English.
                    4. "year": Publication year as an integer (e.g., 2024). Return null if not explicitly found.
                    5. "doi": The official DOI string exactly as found (e.g., "10.1145/1234.5678"). Return null if not explicitly found.
                    """
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

    def _compute_embedding(self, summary_text: str) -> np.ndarray:
        """Call embedding model to compute a 1536-dim vector for the summary."""

        response = self._genai.models.embed_content(
            model="gemini-embedding-001",
            contents=summary_text,
            config=types.EmbedContentConfig(
                task_type="RETRIEVAL_DOCUMENT",
                output_dimensionality=1536,
            ),
        )
        if not response.embeddings:
            raise ValueError("embedding model returned empty response")

        [embedding_obj] = response.embeddings
        return np.array(embedding_obj.values, dtype=np.float32)
