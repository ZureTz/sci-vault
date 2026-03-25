"""Runtime configuration loaded from a YAML file with environment variable overrides."""

import logging
import os
from pathlib import Path

import yaml

log = logging.getLogger(__name__)

# Defaults (lowest priority)
_DEFAULTS: dict = {
    "host": "0.0.0.0",
    "port": 50051,
    "max_workers": 10,
    "redis_host": "localhost",
    "redis_port": 6379,
    "redis_password": "",
    "redis_db": 0,
    "db_host": "localhost",
    "db_port": 5432,
    "db_user": "sci_vault",
    "db_password": "",
    "db_name": "sci_vault",
    "db_ssl_mode": "disable",
    "db_timezone": "UTC",
}


class Config:
    """Centralized config for svc-recommender.

    Priority (highest → lowest):
      1. Environment variables (GRPC_HOST, GRPC_PORT, GRPC_MAX_WORKERS)
      2. config.yaml in the project root (optional)
      3. Built-in defaults
    """

    host: str
    port: int
    max_workers: int
    redis_host: str
    redis_port: int
    redis_password: str
    redis_db: int
    db_host: str
    db_port: int
    db_user: str
    db_password: str
    db_name: str
    db_ssl_mode: str
    db_timezone: str

    def __init__(
        self,
        host: str = _DEFAULTS["host"],
        port: int = _DEFAULTS["port"],
        max_workers: int = _DEFAULTS["max_workers"],
        redis_host: str = _DEFAULTS["redis_host"],
        redis_port: int = _DEFAULTS["redis_port"],
        redis_password: str = _DEFAULTS["redis_password"],
        redis_db: int = _DEFAULTS["redis_db"],
        db_host: str = _DEFAULTS["db_host"],
        db_port: int = _DEFAULTS["db_port"],
        db_user: str = _DEFAULTS["db_user"],
        db_password: str = _DEFAULTS["db_password"],
        db_name: str = _DEFAULTS["db_name"],
        db_ssl_mode: str = _DEFAULTS["db_ssl_mode"],
        db_timezone: str = _DEFAULTS["db_timezone"],
    ) -> None:
        self.host = host
        self.port = port
        self.max_workers = max_workers
        self.redis_host = redis_host
        self.redis_port = redis_port
        self.redis_password = redis_password
        self.redis_db = redis_db
        self.db_host = db_host
        self.db_port = db_port
        self.db_user = db_user
        self.db_password = db_password
        self.db_name = db_name
        self.db_ssl_mode = db_ssl_mode
        self.db_timezone = db_timezone

    @classmethod
    def load(cls, config_path: Path | None = None) -> "Config":
        """Load config from YAML file then apply environment variable overrides.

        *config_path* defaults to ``config.yaml`` next to the project root
        (i.e. the directory that contains ``main.py``).  The file is optional;
        if it is missing the service falls back to built-in defaults.
        """
        # Locate config.yaml: caller may supply an explicit path, otherwise
        # search the current working directory and the project root (src/../).
        if config_path is None:
            candidates = [
                Path.cwd() / "config.yaml",
                Path(__file__).parent.parent / "config.yaml",
            ]
            config_path = next((p for p in candidates if p.exists()), None)

        values: dict = dict(_DEFAULTS)

        if config_path and config_path.exists():
            with config_path.open() as fh:
                file_data = yaml.safe_load(fh) or {}
            values.update(file_data)
            log.info("loaded config from %s", config_path)
        else:
            log.info("no config.yaml found, using defaults and environment variables")

        # Environment variables override file values
        if (v := os.getenv("GRPC_HOST")) is not None:
            values["host"] = v
        if (v := os.getenv("GRPC_PORT")) is not None:
            values["port"] = int(v)
        if (v := os.getenv("GRPC_MAX_WORKERS")) is not None:
            values["max_workers"] = int(v)
        if (v := os.getenv("REDIS_HOST")) is not None:
            values["redis_host"] = v
        if (v := os.getenv("REDIS_PORT")) is not None:
            values["redis_port"] = int(v)
        if (v := os.getenv("REDIS_PASSWORD")) is not None:
            values["redis_password"] = v
        if (v := os.getenv("REDIS_DB")) is not None:
            values["redis_db"] = int(v)

        if (v := os.getenv("DB_HOST")) is not None:
            values["db_host"] = v
        if (v := os.getenv("DB_PORT")) is not None:
            values["db_port"] = int(v)
        if (v := os.getenv("DB_USER")) is not None:
            values["db_user"] = v
        if (v := os.getenv("DB_PASSWORD")) is not None:
            values["db_password"] = v
        if (v := os.getenv("DB_NAME")) is not None:
            values["db_name"] = v
        if (v := os.getenv("DB_SSL_MODE")) is not None:
            values["db_ssl_mode"] = v
        if (v := os.getenv("DB_TIMEZONE")) is not None:
            values["db_timezone"] = v

        return cls(
            host=str(values["host"]),
            port=int(values["port"]),
            max_workers=int(values["max_workers"]),
            redis_host=str(values["redis_host"]),
            redis_port=int(values["redis_port"]),
            redis_password=str(values["redis_password"]),
            redis_db=int(values["redis_db"]),
            db_host=str(values["db_host"]),
            db_port=int(values["db_port"]),
            db_user=str(values["db_user"]),
            db_password=str(values["db_password"]),
            db_name=str(values["db_name"]),
            db_ssl_mode=str(values["db_ssl_mode"]),
            db_timezone=str(values["db_timezone"]),
        )

    @classmethod
    def from_env(cls) -> "Config":
        """Backward-compatible alias for :meth:`load`."""
        return cls.load()

    @property
    def addr(self) -> str:
        """gRPC listen address, e.g. '0.0.0.0:50051'."""
        return f"{self.host}:{self.port}"

    @property
    def db_dsn(self) -> str:
        """PostgreSQL connection string for psycopg."""
        return (
            f"postgresql://{self.db_user}:{self.db_password}"
            f"@{self.db_host}:{self.db_port}/{self.db_name}"
            f"?sslmode={self.db_ssl_mode}"
        )
