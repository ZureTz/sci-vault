"""PostgreSQL connection pool."""

import logging
from typing import Any

from psycopg_pool import ConnectionPool
from pgvector.psycopg import register_vector

from config import Config

log = logging.getLogger(__name__)


class Database:
    """Manages a psycopg connection pool. Call close() on shutdown."""

    def __init__(self, cfg: Config) -> None:
        log.info(
            "building DB pool (%s@%s:%d/%s sslmode=%s)",
            cfg.db_user,
            cfg.db_host,
            cfg.db_port,
            cfg.db_name,
            cfg.db_ssl_mode,
        )
        self._pool = ConnectionPool(
            conninfo=cfg.db_dsn,
            configure=lambda conn: register_vector(conn),
            open=True,
        )

    @property
    def pool(self) -> ConnectionPool[Any]:
        return self._pool

    def close(self) -> None:
        log.info("closing DB pool")
        self._pool.close()
