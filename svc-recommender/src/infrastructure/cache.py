"""Redis client."""

import logging

import redis

from config import Config

log = logging.getLogger(__name__)


class Cache:
    """Wraps a Redis connection."""

    def __init__(self, cfg: Config) -> None:
        log.info(
            "building Redis client (%s:%d db=%d)",
            cfg.redis_host,
            cfg.redis_port,
            cfg.redis_db,
        )
        self._client = redis.Redis(
            host=cfg.redis_host,
            port=cfg.redis_port,
            password=cfg.redis_password or None,
            db=cfg.redis_db,
            decode_responses=True,
        )

    @property
    def client(self) -> redis.Redis:
        return self._client
