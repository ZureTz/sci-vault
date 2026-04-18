"""Search query embedding cache — Redis access for repeated query lookups."""

import base64
import hashlib
import logging

import numpy as np
import redis

log = logging.getLogger(__name__)

_TTL = 24 * 60 * 60  # seconds


def _key(query: str) -> str:
    digest = hashlib.sha256(query.encode("utf-8")).hexdigest()
    return f"search:embed:{digest}"


class SearchCache:
    """Caches query embeddings keyed by the query text.

    Embeddings are stored as base64-encoded float32 bytes because the shared
    Redis client uses ``decode_responses=True`` (string mode).
    """

    def __init__(self, client: redis.Redis) -> None:
        self._redis = client

    def get_embedding(self, query: str) -> np.ndarray | None:
        try:
            raw = self._redis.get(_key(query))
        except Exception:
            log.warning("Redis get failed for query embedding")
            return None
        if not isinstance(raw, (str, bytes)):
            return None
        try:
            data = base64.b64decode(raw)
            return np.frombuffer(data, dtype=np.float32)
        except Exception:
            log.warning("failed to decode cached embedding; ignoring")
            return None

    def set_embedding(self, query: str, embedding: np.ndarray) -> None:
        try:
            encoded = base64.b64encode(
                embedding.astype(np.float32, copy=False).tobytes()
            ).decode("ascii")
            self._redis.set(_key(query), encoded, ex=_TTL)
        except Exception:
            log.warning("Redis set failed for query embedding")
