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

    def __init__(
        self,
        host: str = _DEFAULTS["host"],
        port: int = _DEFAULTS["port"],
        max_workers: int = _DEFAULTS["max_workers"],
    ) -> None:
        self.host = host
        self.port = port
        self.max_workers = max_workers

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

        return cls(
            host=str(values["host"]),
            port=int(values["port"]),
            max_workers=int(values["max_workers"]),
        )

    @classmethod
    def from_env(cls) -> "Config":
        """Backward-compatible alias for :meth:`load`."""
        return cls.load()

    @property
    def addr(self) -> str:
        """gRPC listen address, e.g. '0.0.0.0:50051'."""
        return f"{self.host}:{self.port}"
