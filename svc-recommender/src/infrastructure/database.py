"""PostgreSQL connection pool factory."""

import logging

from psycopg_pool import ConnectionPool
from pgvector.psycopg import register_vector

from config import Config

log = logging.getLogger(__name__)


def configure(conn):
    register_vector(conn)


def build_db_pool(cfg: Config) -> ConnectionPool:
    """Create and return an open :class:`psycopg_pool.ConnectionPool`.

    The pool is opened eagerly (``open=True``) so connection errors surface
    at startup rather than on the first request.
    """
    log.info(
        "building DB pool (%s@%s:%d/%s sslmode=%s)",
        cfg.db_user,
        cfg.db_host,
        cfg.db_port,
        cfg.db_name,
        cfg.db_ssl_mode,
    )
    return ConnectionPool(conninfo=cfg.db_dsn, configure=configure, open=True)
