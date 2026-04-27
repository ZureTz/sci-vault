"""Query-embedding repository — persistent cache of (text, task_type) → embedding."""

import logging

import numpy as np
from psycopg_pool import ConnectionPool

from utils.query_embedding_key import query_hash

log = logging.getLogger(__name__)


class QueryEmbeddingRepository:
    """Read/write the persistent query_embeddings table.

    Schema is owned by svc-gateway (model.QueryEmbedding); this repo is the
    only consumer. The table is the durable backstop behind the Redis
    QueryEmbeddingCache: once we've paid Gemini for a (text, task_type) pair
    we never want to pay again, even after the Redis TTL expires.

    Primary key is (query_hash, task_type) because Gemini's RETRIEVAL_QUERY
    and RETRIEVAL_DOCUMENT produce vectors in asymmetric, incompatible
    spaces — the same string under two task types is two distinct embeddings.
    """

    def __init__(self, pool: ConnectionPool) -> None:
        self._pool = pool

    def get(self, text: str, task_type: str) -> np.ndarray | None:
        """Singleton variant — wraps get_many. On hit refreshes last_used_at
        for LRU; failures are best-effort."""
        return self.get_many([text], task_type).get(query_hash(text))

    def get_many(self, texts: list[str], task_type: str) -> dict[bytes, np.ndarray]:
        """Batch fetch. Returns {sha256(text): embedding} for the subset of
        texts that have a row under task_type. Single SELECT round-trip plus
        one UPDATE for the LRU bump."""
        if not texts:
            return {}
        digests = [query_hash(t) for t in texts]
        try:
            with self._pool.connection() as conn:
                rows = conn.execute(
                    "SELECT query_hash, embedding FROM query_embeddings"
                    " WHERE query_hash = ANY(%s) AND task_type = %s",
                    (digests, task_type),
                ).fetchall()
                if not rows:
                    return {}
                hit_hashes = [bytes(row[0]) for row in rows]
                # Best-effort LRU bump for the hit set in one statement.
                try:
                    conn.execute(
                        "UPDATE query_embeddings SET last_used_at = NOW()"
                        " WHERE query_hash = ANY(%s) AND task_type = %s",
                        (hit_hashes, task_type),
                    )
                except Exception:
                    log.warning("query_embeddings: last_used_at bump failed")
            return {bytes(row[0]): row[1] for row in rows if row[1] is not None}
        except Exception:
            log.warning("query_embeddings: get_many failed")
            return {}

    def set(self, text: str, task_type: str, embedding: np.ndarray) -> None:
        """Singleton variant — wraps set_many."""
        self.set_many([(text, embedding)], task_type)

    def set_many(self, items: list[tuple[str, np.ndarray]], task_type: str) -> None:
        """Persist (text, task_type, embedding) tuples. One round-trip via
        executemany; ON CONFLICT just bumps last_used_at so concurrent
        writers race harmlessly."""
        if not items:
            return
        rows = [
            (query_hash(text), task_type, text, embedding) for text, embedding in items
        ]
        try:
            with self._pool.connection() as conn:
                with conn.cursor() as cur:
                    cur.executemany(
                        "INSERT INTO query_embeddings"
                        "  (query_hash, task_type, query, embedding,"
                        "   created_at, last_used_at)"
                        " VALUES (%s, %s, %s, %s, NOW(), NOW())"
                        " ON CONFLICT (query_hash, task_type)"
                        " DO UPDATE SET last_used_at = NOW()",
                        rows,
                    )
        except Exception:
            log.warning("query_embeddings: set_many failed")
