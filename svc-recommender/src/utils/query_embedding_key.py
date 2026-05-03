"""Shared hashing for query-embedding cache keys.

The same hash needs to appear in three places:
- Postgres (`bytea` primary key): raw 32 bytes — half the storage of hex and
  the natural fit for a binary column.
- Redis (string key): hex form, namespaced by task_type so the namespace is
  inspectable via ``KEYS query:embed:RETRIEVAL_QUERY:*``.
- The resolver, when cross-referencing a Postgres row back to its source text.

Centralising it here ensures all three tiers compute the digest the same way,
so a row written by one path is always findable by another.
"""

import hashlib


def query_hash(text: str) -> bytes:
    """Raw 32-byte sha256 digest of `text` (UTF-8 encoded)."""
    return hashlib.sha256(text.encode("utf-8")).digest()


def query_redis_key(text: str, task_type: str) -> str:
    """Redis cache key for the (text, task_type) embedding."""
    return f"query:embed:{task_type}:{query_hash(text).hex()}"
