"""Three-tier text-embedding lookup: Redis → Postgres → Gemini.

Two flows use this:
- SemanticSearch resolves typed queries with task_type=RETRIEVAL_QUERY so the
  result matches the asymmetric RETRIEVAL_DOCUMENT space the corpus lives in.
- RecommendForUser folds historical search strings into a profile centroid
  alongside liked/viewed *document* embeddings, so it resolves them with
  task_type=RETRIEVAL_DOCUMENT to keep everything in one space.

Caching is per (text, task_type) — the same string under two task types is
two distinct vectors and must not collide in either tier.

Single-text and bulk APIs are both exposed; bulk is what callers should
prefer when they have a list (e.g. RecommendForUser's recent_queries),
since each tier collapses to one round-trip regardless of input size.
"""

import logging

import numpy as np

from cache.query_embedding import QueryEmbeddingCache
from genai.query_embedder import QueryEmbedder
from repository.query_embedding import QueryEmbeddingRepository
from utils.query_embedding_key import query_hash

log = logging.getLogger(__name__)


class QueryEmbeddingResolver:
    """Resolves (text, task_type) to an embedding via the three-tier chain."""

    def __init__(
        self,
        cache: QueryEmbeddingCache,
        repo: QueryEmbeddingRepository,
        embedder: QueryEmbedder,
    ) -> None:
        self._cache = cache
        self._repo = repo
        self._embedder = embedder

    def resolve(self, text: str, task_type: str) -> np.ndarray:
        return self.resolve_many([text], task_type)[0]

    def resolve_many(self, texts: list[str], task_type: str) -> list[np.ndarray]:
        """Bulk variant — at most one round-trip per tier, regardless of how
        many texts are passed in. Order of the returned list matches `texts`
        so positional consumers (e.g. recency-weighted accumulators) stay
        correct.
        """
        if not texts:
            return []

        result: list[np.ndarray | None] = [None] * len(texts)

        # Tier 1: Redis MGET — one round-trip.
        cache_hits = self._cache.get_embeddings(texts, task_type)
        for i, vec in enumerate(cache_hits):
            if vec is not None:
                result[i] = vec
        redis_hit_count = sum(1 for v in result if v is not None)

        # Tier 2: Postgres for whatever Redis missed — one round-trip.
        misses_after_redis = [i for i, v in enumerate(result) if v is None]
        postgres_hits: list[tuple[str, np.ndarray]] = []
        if misses_after_redis:
            miss_texts = [texts[i] for i in misses_after_redis]
            db_map = self._repo.get_many(miss_texts, task_type)
            for i in misses_after_redis:
                vec = db_map.get(query_hash(texts[i]))
                if vec is not None:
                    result[i] = vec
                    postgres_hits.append((texts[i], vec))
        # Warm Redis for Postgres hits in one pipelined batch.
        if postgres_hits:
            self._cache.set_embeddings(postgres_hits, task_type)

        # Tier 3: Gemini for whatever's still missing — one batched API call.
        misses_after_postgres = [i for i, v in enumerate(result) if v is None]
        if misses_after_postgres:
            miss_texts = [texts[i] for i in misses_after_postgres]
            log.info(
                "query embedding: %d miss (%s), calling gemini",
                len(miss_texts),
                task_type,
            )
            embedded = self._embedder.embed_many(miss_texts, task_type)
            new_items: list[tuple[str, np.ndarray]] = []
            for idx, vec in zip(misses_after_postgres, embedded):
                result[idx] = vec
                new_items.append((texts[idx], vec))
            # Persist new embeddings to both stores in batched writes.
            self._repo.set_many(new_items, task_type)
            self._cache.set_embeddings(new_items, task_type)

        log.info(
            "query embedding (%s): redis=%d postgres=%d gemini=%d",
            task_type,
            redis_hit_count,
            len(postgres_hits),
            len(misses_after_postgres),
        )

        # By construction every slot is filled (Gemini covers any remaining
        # misses); narrow the type for the return.
        return [vec for vec in result if vec is not None]
