"""Recommend repository — pgvector cosine similarity against a source doc's embedding."""

import logging

import numpy as np
from psycopg import sql
from psycopg_pool import ConnectionPool

from repository import (
    DOC_VISIBILITY_LAB,
    DOC_VISIBILITY_PRIVATE,
    ENRICH_STATUS_DONE,
)
from repository.types import ScoredDocument, ScoredUser

log = logging.getLogger(__name__)

_DEFAULT_LIMIT = 5
_MAX_LIMIT = 20
_MIN_SIMILARITY = 0.6

_DEFAULT_PERSONALIZED_LIMIT = 20
_MAX_PERSONALIZED_LIMIT = 50
# Personalized feed uses a looser threshold than similar-docs because the
# centroid of mixed signals is, by construction, less peaked than a single
# document embedding; demanding 0.6 cosine sim would empty most feeds.
_MIN_PERSONALIZED_SIMILARITY = 0.4

_DEFAULT_COLLABORATOR_LIMIT = 10
_MAX_COLLABORATOR_LIMIT = 50
# Same reasoning as personalized feed — comparing two centroids of noisy
# user signals will rarely cross 0.6.
_MIN_COLLABORATOR_SIMILARITY = 0.4

_BASE_WHERE = sql.SQL("d.deleted_at IS NULL AND d.enrich_status = {}").format(
    sql.Literal(ENRICH_STATUS_DONE)
)

_ACCESS_PRIVATE_ONLY = sql.SQL(
    "AND d.uploaded_by_user_id = %(user_id)s AND d.visibility = {}"
).format(sql.Literal(DOC_VISIBILITY_PRIVATE))

_ACCESS_PRIVATE_OR_LAB = sql.SQL(
    "AND ("
    "  (d.uploaded_by_user_id = %(user_id)s AND d.visibility = {private})"
    "  OR (d.lab_id = %(lab_id)s AND d.visibility = {lab})"
    ")"
).format(
    private=sql.Literal(DOC_VISIBILITY_PRIVATE), lab=sql.Literal(DOC_VISIBILITY_LAB)
)


class RecommendRepository:
    """Fetches the source doc's embedding and finds nearest neighbours in
    the vector space, honouring the same access-control shape as search."""

    def __init__(self, pool: ConnectionPool) -> None:
        self._pool = pool

    @staticmethod
    def _access(lab_id: int) -> sql.Composable:
        return _ACCESS_PRIVATE_OR_LAB if lab_id > 0 else _ACCESS_PRIVATE_ONLY

    @staticmethod
    def _clamp_limit(limit: int) -> int:
        return min(max(limit, 1), _MAX_LIMIT) if limit else _DEFAULT_LIMIT

    def fetch_embedding(self, doc_id: int) -> np.ndarray | None:
        """Return the stored embedding for a document, or None when missing
        (doc doesn't exist, is soft-deleted, or enrichment hasn't completed)."""
        with self._pool.connection() as conn:
            row = conn.execute(
                "SELECT embedding FROM documents"
                " WHERE id = %s AND deleted_at IS NULL"
                " AND enrich_status = %s AND embedding IS NOT NULL",
                (doc_id, ENRICH_STATUS_DONE),
            ).fetchone()
        if row is None or row[0] is None:
            return None
        return row[0]

    def fetch_embeddings_bulk(self, doc_ids: list[int]) -> dict[int, np.ndarray]:
        """Return a {doc_id: embedding} map for the subset of doc_ids that have
        one (non-deleted, enrich_status=done). Missing IDs are silently absent;
        callers should iterate their original list to preserve order."""
        if not doc_ids:
            return {}
        with self._pool.connection() as conn:
            rows = conn.execute(
                "SELECT id, embedding FROM documents"
                " WHERE id = ANY(%s) AND deleted_at IS NULL"
                " AND enrich_status = %s AND embedding IS NOT NULL",
                (doc_ids, ENRICH_STATUS_DONE),
            ).fetchall()
        return {int(row[0]): row[1] for row in rows if row[1] is not None}

    def similar(
        self,
        source_doc_id: int,
        query_embedding: np.ndarray,
        user_id: int,
        lab_id: int,
        limit: int,
        min_similarity: float = _MIN_SIMILARITY,
    ) -> list[ScoredDocument]:
        """Find documents most similar to the source doc's embedding, excluding
        the source itself. Results ranked by cosine distance ascending."""
        limit = self._clamp_limit(limit)

        query = sql.SQL(
            "SELECT"
            "  d.id,"
            "  COALESCE(d.title, '') AS title,"
            "  d.original_file_name,"
            "  COALESCE(d.summary, '') AS summary,"
            "  COALESCE(d.authors, ARRAY[]::text[]) AS authors,"
            "  COALESCE(d.tags, ARRAY[]::text[]) AS tags,"
            "  1 - (d.embedding <=> %(query_vec)s) AS similarity"
            " FROM documents d"
            " WHERE {base}"
            "  AND d.id != %(source_id)s"
            "  AND d.embedding IS NOT NULL"
            "  AND 1 - (d.embedding <=> %(query_vec)s) >= %(min_sim)s"
            "  {access}"
            " ORDER BY d.embedding <=> %(query_vec)s ASC"
            " LIMIT %(limit)s"
        ).format(base=_BASE_WHERE, access=self._access(lab_id))

        params = {
            "query_vec": query_embedding,
            "source_id": source_doc_id,
            "user_id": user_id,
            "lab_id": lab_id,
            "limit": limit,
            "min_sim": min_similarity,
        }

        with self._pool.connection() as conn:
            rows = conn.execute(query, params).fetchall()

        return [
            ScoredDocument(
                doc_id=row[0],
                title=row[1],
                original_file_name=row[2],
                summary=row[3],
                authors=list(row[4]) if row[4] else [],
                tags=list(row[5]) if row[5] else [],
                similarity=float(row[6]),
            )
            for row in rows
        ]

    @staticmethod
    def _clamp_personalized_limit(limit: int) -> int:
        if not limit:
            return _DEFAULT_PERSONALIZED_LIMIT
        return min(max(limit, 1), _MAX_PERSONALIZED_LIMIT)

    @staticmethod
    def _clamp_collaborator_limit(limit: int) -> int:
        if not limit:
            return _DEFAULT_COLLABORATOR_LIMIT
        return min(max(limit, 1), _MAX_COLLABORATOR_LIMIT)

    def collaborators_search(
        self,
        query_embedding: np.ndarray,
        lab_id: int,
        exclude_user_id: int,
        limit: int,
        min_similarity: float = _MIN_COLLABORATOR_SIMILARITY,
    ) -> list[ScoredUser]:
        """Rank lab members by cosine similarity between the caller's profile
        centroid and each candidate's centroid (averaged embeddings of docs
        they liked or viewed). The caller is excluded; users without any
        like/view signal are excluded by the inner join. UNION ALL means a
        doc both liked and viewed contributes twice — that's intentional, more
        engagement → more weight in the candidate centroid."""
        limit = self._clamp_collaborator_limit(limit)

        query = sql.SQL(
            "WITH candidate_signals AS ("
            "  SELECT lm.user_id, d.embedding"
            "    FROM lab_members lm"
            "    JOIN document_likes dl"
            "      ON dl.user_id = lm.user_id AND dl.deleted_at IS NULL"
            "    JOIN documents d"
            "      ON d.id = dl.document_id"
            "     AND d.deleted_at IS NULL"
            "     AND d.enrich_status = {done}"
            "     AND d.embedding IS NOT NULL"
            "   WHERE lm.lab_id = %(lab_id)s"
            "     AND lm.deleted_at IS NULL"
            "     AND lm.user_id <> %(exclude_user_id)s"
            "  UNION ALL"
            "  SELECT lm.user_id, d.embedding"
            "    FROM lab_members lm"
            "    JOIN document_views dv"
            "      ON dv.user_id = lm.user_id AND dv.deleted_at IS NULL"
            "    JOIN documents d"
            "      ON d.id = dv.document_id"
            "     AND d.deleted_at IS NULL"
            "     AND d.enrich_status = {done}"
            "     AND d.embedding IS NOT NULL"
            "   WHERE lm.lab_id = %(lab_id)s"
            "     AND lm.deleted_at IS NULL"
            "     AND lm.user_id <> %(exclude_user_id)s"
            "), centroids AS ("
            "  SELECT user_id,"
            "         AVG(embedding)::vector(768) AS centroid,"
            "         COUNT(*)::int AS signal_count"
            "    FROM candidate_signals"
            "   GROUP BY user_id"
            ")"
            " SELECT u.id,"
            "        u.username,"
            "        COALESCE(up.nickname, '') AS nickname,"
            "        COALESCE(up.avatar_key, '') AS avatar_key,"
            "        1 - (c.centroid <=> %(query_vec)s) AS similarity,"
            "        c.signal_count"
            "   FROM centroids c"
            "   JOIN users u ON u.id = c.user_id AND u.deleted_at IS NULL"
            "   LEFT JOIN user_profiles up"
            "     ON up.user_id = u.id AND up.deleted_at IS NULL"
            "  WHERE 1 - (c.centroid <=> %(query_vec)s) >= %(min_sim)s"
            "  ORDER BY c.centroid <=> %(query_vec)s ASC"
            "  LIMIT %(limit)s"
        ).format(done=sql.Literal(ENRICH_STATUS_DONE))

        params = {
            "query_vec": query_embedding,
            "lab_id": lab_id,
            "exclude_user_id": exclude_user_id,
            "limit": limit,
            "min_sim": min_similarity,
        }

        with self._pool.connection() as conn:
            rows = conn.execute(query, params).fetchall()

        return [
            ScoredUser(
                user_id=int(row[0]),
                username=row[1],
                nickname=row[2],
                avatar_key=row[3],
                similarity=float(row[4]),
                signal_count=int(row[5]),
            )
            for row in rows
        ]

    def personalized_search(
        self,
        query_embedding: np.ndarray,
        user_id: int,
        lab_id: int,
        exclude_ids: list[int],
        limit: int,
        min_similarity: float = _MIN_PERSONALIZED_SIMILARITY,
    ) -> list[ScoredDocument]:
        """Find documents most similar to a profile vector, excluding any IDs
        the caller already engaged with (typically liked docs). Results ranked
        by cosine distance ascending, same access-control shape as similar()."""
        limit = self._clamp_personalized_limit(limit)
        # Always pass a list so `!= ALL(%s)` is well-defined; an empty array
        # means "exclude nothing".
        excludes = exclude_ids or []

        query = sql.SQL(
            "SELECT"
            "  d.id,"
            "  COALESCE(d.title, '') AS title,"
            "  d.original_file_name,"
            "  COALESCE(d.summary, '') AS summary,"
            "  COALESCE(d.authors, ARRAY[]::text[]) AS authors,"
            "  COALESCE(d.tags, ARRAY[]::text[]) AS tags,"
            "  1 - (d.embedding <=> %(query_vec)s) AS similarity"
            " FROM documents d"
            " WHERE {base}"
            "  AND d.id != ALL(%(exclude_ids)s)"
            "  AND d.embedding IS NOT NULL"
            "  AND 1 - (d.embedding <=> %(query_vec)s) >= %(min_sim)s"
            "  {access}"
            " ORDER BY d.embedding <=> %(query_vec)s ASC"
            " LIMIT %(limit)s"
        ).format(base=_BASE_WHERE, access=self._access(lab_id))

        params = {
            "query_vec": query_embedding,
            "exclude_ids": excludes,
            "user_id": user_id,
            "lab_id": lab_id,
            "limit": limit,
            "min_sim": min_similarity,
        }

        with self._pool.connection() as conn:
            rows = conn.execute(query, params).fetchall()

        return [
            ScoredDocument(
                doc_id=row[0],
                title=row[1],
                original_file_name=row[2],
                summary=row[3],
                authors=list(row[4]) if row[4] else [],
                tags=list(row[5]) if row[5] else [],
                similarity=float(row[6]),
            )
            for row in rows
        ]
