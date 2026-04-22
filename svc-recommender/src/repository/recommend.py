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
from repository.search import SearchHit

log = logging.getLogger(__name__)

_DEFAULT_LIMIT = 5
_MAX_LIMIT = 20
_MIN_SIMILARITY = 0.6

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

    def similar(
        self,
        source_doc_id: int,
        query_embedding: np.ndarray,
        user_id: int,
        lab_id: int,
        limit: int,
        min_similarity: float = _MIN_SIMILARITY,
    ) -> list[SearchHit]:
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
            SearchHit(
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
