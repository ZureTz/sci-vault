"""Search repository — pgvector cosine similarity + full-text keyword queries."""

import logging
from dataclasses import dataclass

import numpy as np
from psycopg import sql
from psycopg_pool import ConnectionPool

from repository import (
    DOC_VISIBILITY_LAB,
    DOC_VISIBILITY_PRIVATE,
    ENRICH_STATUS_DONE,
)

log = logging.getLogger(__name__)

_DEFAULT_LIMIT = 10
_MAX_LIMIT = 50
_MIN_SIMILARITY = 0.6

# ── Shared fragments ──────────────────────────────────────────────────────────

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

_TEXT_VECTOR = sql.SQL(
    "to_tsvector('english',"
    "  COALESCE(d.title, '') || ' ' ||"
    "  COALESCE(d.summary, '') || ' ' ||"
    "  array_to_string(COALESCE(d.tags, ARRAY[]::text[]), ' ') || ' ' ||"
    "  array_to_string(COALESCE(d.authors, ARRAY[]::text[]), ' ')"
    ")"
)

_TEXT_QUERY = sql.SQL("websearch_to_tsquery('english', %(text_query)s)")


@dataclass
class SearchHit:
    doc_id: int
    title: str
    original_file_name: str
    summary: str
    authors: list[str]
    tags: list[str]
    similarity: float  # 1 - cosine_distance (0 for keyword-only matches)


class SearchRepository:
    """Vector similarity + keyword search against the documents table."""

    def __init__(self, pool: ConnectionPool) -> None:
        self._pool = pool

    @staticmethod
    def _access(lab_id: int) -> sql.Composable:
        return _ACCESS_PRIVATE_OR_LAB if lab_id > 0 else _ACCESS_PRIVATE_ONLY

    @staticmethod
    def _clamp_limit(limit: int) -> int:
        return min(max(limit, 1), _MAX_LIMIT) if limit else _DEFAULT_LIMIT

    def vector_search(
        self,
        query_embedding: np.ndarray,
        user_id: int,
        lab_id: int,
        limit: int,
        min_similarity: float = _MIN_SIMILARITY,
    ) -> list[SearchHit]:
        """Find documents by vector cosine similarity, filtered by a minimum threshold."""
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
            "  AND d.embedding IS NOT NULL"
            "  AND 1 - (d.embedding <=> %(query_vec)s) >= %(min_sim)s"
            "  {access}"
            " ORDER BY d.embedding <=> %(query_vec)s ASC"
            " LIMIT %(limit)s"
        ).format(base=_BASE_WHERE, access=self._access(lab_id))

        params = {
            "query_vec": query_embedding,
            "user_id": user_id,
            "lab_id": lab_id,
            "limit": limit,
            "min_sim": min_similarity,
        }

        with self._pool.connection() as conn:
            rows = conn.execute(query, params).fetchall()

        return [self._row_to_hit(row) for row in rows]

    def keyword_search(
        self,
        query_text: str,
        user_id: int,
        lab_id: int,
        limit: int,
        exclude_ids: list[int] | None = None,
    ) -> list[SearchHit]:
        """Full-text keyword search across title, summary, tags, and authors.

        Uses websearch_to_tsquery for natural-language query parsing.
        Results are ranked by ts_rank_cd and returned with similarity=0.
        """
        limit = self._clamp_limit(limit)

        exclude = (
            sql.SQL("AND d.id != ALL(%(exclude_ids)s)") if exclude_ids else sql.SQL("")
        )

        query = sql.SQL(
            "SELECT"
            "  d.id,"
            "  COALESCE(d.title, '') AS title,"
            "  d.original_file_name,"
            "  COALESCE(d.summary, '') AS summary,"
            "  COALESCE(d.authors, ARRAY[]::text[]) AS authors,"
            "  COALESCE(d.tags, ARRAY[]::text[]) AS tags,"
            "  0::float AS similarity"
            " FROM documents d"
            " WHERE {base}"
            "  AND {tsvec} @@ {tsq}"
            "  {access}"
            "  {exclude}"
            " ORDER BY ts_rank_cd({tsvec}, {tsq}) DESC"
            " LIMIT %(limit)s"
        ).format(
            base=_BASE_WHERE,
            access=self._access(lab_id),
            exclude=exclude,
            tsvec=_TEXT_VECTOR,
            tsq=_TEXT_QUERY,
        )

        params: dict = {
            "text_query": query_text,
            "user_id": user_id,
            "lab_id": lab_id,
            "limit": limit,
            "exclude_ids": exclude_ids or [],
        }

        with self._pool.connection() as conn:
            rows = conn.execute(query, params).fetchall()

        return [self._row_to_hit(row) for row in rows]

    @staticmethod
    def _row_to_hit(row: tuple) -> SearchHit:
        return SearchHit(
            doc_id=row[0],
            title=row[1],
            original_file_name=row[2],
            summary=row[3],
            authors=list(row[4]) if row[4] else [],
            tags=list(row[5]) if row[5] else [],
            similarity=float(row[6]),
        )
