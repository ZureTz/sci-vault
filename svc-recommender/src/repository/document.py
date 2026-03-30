"""Document repository — database access for document enrichment."""

import logging
from typing import Optional

import numpy as np
from psycopg_pool import ConnectionPool

log = logging.getLogger(__name__)

_DB_STATUS_DONE = "done"


class DocumentRepository:
    """Encapsulates all database operations for the document enrichment workflow."""

    def __init__(self, pool: ConnectionPool) -> None:
        self._pool = pool

    def write_enrichment_and_done(
        self,
        doc_id: int,
        authors: list[str],
        summary: str,
        tags: list[str],
        year: Optional[int],
        doi: Optional[str],
        embedding: np.ndarray,
    ) -> None:
        """Write AI-extracted fields and mark enrich_status=done in one UPDATE.

        Human-provided year/doi take priority via COALESCE:
          - year: kept if already set, otherwise filled with AI value
          - doi:  kept if non-empty, otherwise filled with AI value
        Raises on failure so the caller can propagate the error.
        """
        try:
            with self._pool.connection() as conn:
                conn.execute(
                    """
                    UPDATE documents SET
                        authors        = %s,
                        summary        = %s,
                        tags           = %s,
                        year           = COALESCE(year, %s),
                        doi            = COALESCE(NULLIF(doi, ''), %s),
                        embedding      = %s,
                        enrich_status  = %s
                    WHERE id = %s
                    """,
                    (
                        authors,
                        summary,
                        tags,
                        year,
                        doi,
                        embedding,
                        _DB_STATUS_DONE,
                        doc_id,
                    ),
                )
        except Exception:
            log.warning("DB enrichment write failed: doc_id=%d", doc_id)
            raise
