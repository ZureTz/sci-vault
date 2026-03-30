"""Enrichment status cache — Redis access for document enrich state."""

import logging

import redis

log = logging.getLogger(__name__)

_STATUS_PENDING = "pending"
_STATUS_PROCESSING = "processing"
_STATUS_DONE = "done"
_STATUS_FAILED = "failed"

_TTL = 24 * 60 * 60  # seconds


def _key(doc_id: int) -> str:
    return f"doc:enrich:{doc_id}"


class EnrichmentStatusCache:
    """Wraps Redis to manage per-document enrichment status.

    Status values and key schema are internal implementation details;
    callers express intent through typed methods.
    """

    def __init__(self, client: redis.Redis) -> None:
        self._redis = client

    def set_pending(self, doc_id: int) -> None:
        self._set(doc_id, _STATUS_PENDING)

    def set_processing(self, doc_id: int) -> None:
        self._set(doc_id, _STATUS_PROCESSING)

    def set_done(self, doc_id: int) -> None:
        self._set(doc_id, _STATUS_DONE)

    def set_failed(self, doc_id: int) -> None:
        self._set(doc_id, _STATUS_FAILED)

    def _set(self, doc_id: int, status: str) -> None:
        try:
            self._redis.set(_key(doc_id), status, ex=_TTL)
        except Exception:
            log.warning(
                "Redis status update failed: doc_id=%d status=%s", doc_id, status
            )
