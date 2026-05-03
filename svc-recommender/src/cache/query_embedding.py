"""Query embedding cache — Redis access for repeated query lookups.

Shared by SemanticSearch and RecommendForUser: a (text, task_type) → embedding
mapping is deterministic for a fixed model, so any flow that needs to embed
text benefits from the same cache.
"""

import base64
import logging
from typing import cast

import numpy as np
import redis

from utils.query_embedding_key import query_redis_key

log = logging.getLogger(__name__)

_TTL = 24 * 60 * 60  # seconds


class QueryEmbeddingCache:
    """Caches embeddings keyed by (text, task_type).

    Embeddings produced under different Gemini task types live in different
    vector spaces and must not be conflated, so task_type is part of the key.

    Embeddings are stored as base64-encoded float32 bytes because the shared
    Redis client uses ``decode_responses=True`` (string mode).
    """

    def __init__(self, client: redis.Redis) -> None:
        self._redis = client

    def get_embedding(self, text: str, task_type: str) -> np.ndarray | None:
        return self.get_embeddings([text], task_type)[0]

    def get_embeddings(
        self, texts: list[str], task_type: str
    ) -> list[np.ndarray | None]:
        """Batch variant of `get_embedding`. Single MGET round-trip; returned
        list is aligned with `texts` (None for misses or decode failures)."""
        if not texts:
            return []
        keys = [query_redis_key(t, task_type) for t in texts]
        try:
            # redis-py types mget() as ResponseT (sync list or async awaitable);
            # the sync client always returns a list. Cast to satisfy the checker.
            raws = cast(list, self._redis.mget(keys))
        except Exception:
            log.warning("Redis mget failed for query embeddings")
            return [None] * len(texts)
        return [_decode(r) for r in raws]

    def set_embedding(self, text: str, task_type: str, embedding: np.ndarray) -> None:
        self.set_embeddings([(text, embedding)], task_type)

    def set_embeddings(
        self, items: list[tuple[str, np.ndarray]], task_type: str
    ) -> None:
        """Batch variant of `set_embedding`. Pipelined SET+EXPIRE so the whole
        write is one round-trip."""
        if not items:
            return
        try:
            with self._redis.pipeline(transaction=False) as pipe:
                for text, embedding in items:
                    encoded = base64.b64encode(
                        embedding.astype(np.float32, copy=False).tobytes()
                    ).decode("ascii")
                    pipe.set(query_redis_key(text, task_type), encoded, ex=_TTL)
                pipe.execute()
        except Exception:
            log.warning("Redis pipeline set failed for query embeddings")


def _decode(raw: object) -> np.ndarray | None:
    if not isinstance(raw, (str, bytes)):
        return None
    try:
        data = base64.b64decode(raw)
        return np.frombuffer(data, dtype=np.float32)
    except Exception:
        log.warning("failed to decode cached embedding; ignoring")
        return None
