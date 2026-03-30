"""Database DSN factory."""

import logging

from config import Config

log = logging.getLogger(__name__)


def build_db_dsn(cfg: Config) -> str:
    """Return the PostgreSQL connection string derived from *cfg*.

    Centralises DSN construction so callers do not depend directly on
    :class:`~config.Config` internals.
    """
    log.info(
        "building DB DSN (%s@%s:%d/%s sslmode=%s)",
        cfg.db_user,
        cfg.db_host,
        cfg.db_port,
        cfg.db_name,
        cfg.db_ssl_mode,
    )
    return cfg.db_dsn
