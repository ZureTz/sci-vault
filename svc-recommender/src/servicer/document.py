"""Document servicer - implements RecommenderService.EnrichDocument."""

import concurrent.futures
import logging
from typing import TYPE_CHECKING

import grpc
import psycopg
import redis

from pb.recommender.v1 import recommender_pb2

if TYPE_CHECKING:
    from mypy_boto3_s3 import S3Client

log = logging.getLogger(__name__)

# Redis-only transient states (fine-grained, 24 h TTL).
_STATUS_PENDING = "pending"
_STATUS_PROCESSING = "processing"
_STATUS_DONE = "done"
_STATUS_FAILED = "failed"

# DB persistent states (source of truth, two values only).
_DB_STATUS_DONE = "done"

_ENRICH_STATUS_TTL = 24 * 60 * 60  # seconds

_enrichment_executor = concurrent.futures.ThreadPoolExecutor(
    max_workers=4, thread_name_prefix="enrich"
)


def _enrich_status_key(doc_id: int) -> str:
    return f"doc:enrich:{doc_id}"


class DocumentServicer:
    """Implements the EnrichDocument RPC; Python owns all enrich-status writes."""

    def __init__(
        self,
        redis_client: redis.Redis,
        db_dsn: str,
        s3_client: "S3Client",
        s3_private_bucket: str,
    ) -> None:
        self._redis = redis_client
        self._db_dsn = db_dsn
        self._s3 = s3_client
        self._s3_bucket = s3_private_bucket

    def EnrichDocument(
        self,
        request: recommender_pb2.EnrichDocumentRequest,
        context: grpc.ServicerContext,
    ) -> recommender_pb2.EnrichDocumentResponse:
        """Set Redis -> pending, queue background job, return ACK immediately."""
        doc_id: int = request.doc_id
        file_key: str = request.file_key

        # Mark as pending before the task even enters the thread pool.
        self._set_redis(doc_id, _STATUS_PENDING)

        _enrichment_executor.submit(self._run_enrichment, doc_id, file_key)

        log.info("EnrichDocument queued: doc_id=%d file_key=%s", doc_id, file_key)
        return recommender_pb2.EnrichDocumentResponse(accepted=True)

    # ------------------------------------------------------------------ #
    # Background worker                                                    #
    # ------------------------------------------------------------------ #

    def _run_enrichment(self, doc_id: int, file_key: str) -> None:
        """Enrichment pipeline. On success writes DB done; on failure leaves DB unchanged."""
        self._set_redis(doc_id, _STATUS_PROCESSING)
        log.info("enrichment started: doc_id=%d", doc_id)
        try:
            pdf_bytes = self._download_pdf(file_key)
            log.info("downloaded PDF: doc_id=%d size=%d bytes", doc_id, len(pdf_bytes))

            # TODO: extract text / images from pdf_bytes
            # TODO: call LLM -> authors, summary, tags
            # TODO: call embedding model -> 1536-dim vector
            # TODO: write authors, summary, tags, embedding to documents table

            self._set_db_done(doc_id)
            self._set_redis(doc_id, _STATUS_DONE)
            log.info("enrichment done: doc_id=%d", doc_id)

        except Exception:
            log.exception("enrichment failed: doc_id=%d", doc_id)
            # DB stays not_started; Redis records the failure for polling.
            self._set_redis(doc_id, _STATUS_FAILED)

    # ------------------------------------------------------------------ #
    # Helpers                                                              #
    # ------------------------------------------------------------------ #

    def _download_pdf(self, file_key: str) -> bytes:
        """Download a PDF from the private S3 (RustFS) bucket and return raw bytes."""
        response = self._s3.get_object(Bucket=self._s3_bucket, Key=file_key)
        return response["Body"].read()

    def _set_redis(self, doc_id: int, status: str) -> None:
        try:
            self._redis.set(_enrich_status_key(doc_id), status, ex=_ENRICH_STATUS_TTL)
        except Exception:
            log.warning(
                "Redis status update failed: doc_id=%d status=%s", doc_id, status
            )

    def _set_db_done(self, doc_id: int) -> None:
        try:
            with psycopg.connect(self._db_dsn) as conn:
                conn.execute(
                    "UPDATE documents SET enrich_status = %s WHERE id = %s",
                    (_DB_STATUS_DONE, doc_id),
                )
        except Exception:
            log.warning("DB done update failed: doc_id=%d", doc_id)
            raise  # re-raise so _run_enrichment marks Redis as failed
