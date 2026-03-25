"""Document servicer – implements RecommenderService.EnrichDocument."""

import concurrent.futures
import logging

import grpc
import psycopg
import redis

from recommender import recommender_pb2

log = logging.getLogger(__name__)

# Enrich status values – must match the constants defined in the Go service.
_STATUS_PROCESSING = "processing"
_STATUS_DONE = "done"
_STATUS_FAILED = "failed"

_ENRICH_STATUS_TTL = 24 * 60 * 60  # 24 hours in seconds

_enrichment_executor = concurrent.futures.ThreadPoolExecutor(
    max_workers=4, thread_name_prefix="enrich"
)


def _enrich_status_key(doc_id: int) -> str:
    return f"doc:enrich:{doc_id}"


class DocumentServicer:
    """Extends HealthServicer with document enrichment RPC."""

    def __init__(self, redis_client: redis.Redis, db_dsn: str) -> None:
        self._redis = redis_client
        self._db_dsn = db_dsn

    def EnrichDocument(
        self,
        request: recommender_pb2.EnrichDocumentRequest,
        context: grpc.ServicerContext,
    ) -> recommender_pb2.EnrichDocumentResponse:
        """Accept the enrichment job immediately and process in a background thread."""
        doc_id: int = request.doc_id
        file_key: str = request.file_key

        log.info("EnrichDocument accepted: doc_id=%d file_key=%s", doc_id, file_key)

        _enrichment_executor.submit(self._run_enrichment, doc_id, file_key)

        return recommender_pb2.EnrichDocumentResponse(accepted=True)

    def _set_status(self, doc_id: int, status: str) -> None:
        """Write enrichment status to both Redis and the documents table."""
        try:
            self._redis.set(_enrich_status_key(doc_id), status, ex=_ENRICH_STATUS_TTL)
        except Exception:
            log.warning("failed to set Redis enrich status: doc_id=%d status=%s", doc_id, status)

        try:
            with psycopg.connect(self._db_dsn) as conn:
                conn.execute(
                    "UPDATE documents SET enrich_status = %s WHERE id = %s",
                    (status, doc_id),
                )
        except Exception:
            log.warning("failed to update DB enrich status: doc_id=%d status=%s", doc_id, status)

    def _run_enrichment(self, doc_id: int, file_key: str) -> None:
        """Background worker: runs LLM/embedding pipeline and writes results to DB."""
        try:
            self._set_status(doc_id, _STATUS_PROCESSING)
            log.info("enrichment started: doc_id=%d", doc_id)

            # TODO: download PDF from S3
            # TODO: extract text
            # TODO: call LLM to generate authors, summary, tags
            # TODO: call embedding model to generate 1536-dim vector
            # TODO: write authors, summary, tags, embedding back to the documents table

            self._set_status(doc_id, _STATUS_DONE)
            log.info("enrichment done: doc_id=%d", doc_id)

        except Exception:
            log.exception("enrichment failed: doc_id=%d", doc_id)
            self._set_status(doc_id, _STATUS_FAILED)
