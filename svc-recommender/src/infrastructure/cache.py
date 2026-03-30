"""Redis client factory."""

import logging

import redis

from config import Config

log = logging.getLogger(__name__)


def build_redis_client(cfg: Config) -> redis.Redis:
    """Create and return a configured :class:`redis.Redis` client."""
    log.info(
        "building Redis client (%s:%d db=%d)",
        cfg.redis_host,
        cfg.redis_port,
        cfg.redis_db,
    )
    return redis.Redis(
        host=cfg.redis_host,
        port=cfg.redis_port,
        password=cfg.redis_password or None,
        db=cfg.redis_db,
        decode_responses=True,
    )
