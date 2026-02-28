"""gRPC server factory and lifecycle management."""

import logging
import signal
import sys
from concurrent import futures

import grpc

from recommender import recommender_pb2_grpc
from config import Config
from interceptor.logging import LoggingInterceptor
from servicer.health import HealthServicer

log = logging.getLogger(__name__)


def create_server(cfg: Config) -> grpc.Server:
    """Build and configure a gRPC Server from *cfg* (but do not start it)."""
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=cfg.max_workers),
        interceptors=[LoggingInterceptor()],
    )
    # Register servicers – add more here as new RPCs are defined.
    recommender_pb2_grpc.add_RecommenderServiceServicer_to_server(
        HealthServicer(),
        server,
    )
    server.add_insecure_port(cfg.addr)
    return server


class ColoredFormatter(logging.Formatter):
    COLORS = {
        logging.DEBUG: "\033[36m",  # Cyan
        logging.INFO: "\033[32m",  # Green
        logging.WARNING: "\033[33m",  # Yellow
        logging.ERROR: "\033[31m",  # Red
        logging.CRITICAL: "\033[1;31m",  # Bold Red
    }
    RESET = "\033[0m"

    def format(self, record: logging.LogRecord) -> str:
        color = self.COLORS.get(record.levelno, self.RESET)
        # Create a copy so we don't mutate the original record for other handlers
        record_copy = logging.makeLogRecord(record.__dict__)
        record_copy.levelname = f"{color}{record_copy.levelname}{self.RESET}"
        return super().format(record_copy)


def serve(cfg: Config | None = None) -> None:
    """Start the gRPC server and block until a shutdown signal is received."""
    fmt = "%(asctime)s %(levelname)s %(name)s: %(message)s"
    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(ColoredFormatter(fmt))

    # Remove existing handlers explicitly if basicConfig was already called
    logging.root.handlers.clear()
    logging.basicConfig(level=logging.INFO, handlers=[handler])

    if cfg is None:
        cfg = Config.load()

    server = create_server(cfg)

    log.info("starting svc-recommender on %s", cfg.addr)
    server.start()

    def _on_shutdown(signal_num: int, _frame: object) -> None:
        log.info("received signal %s, shutting down…", signal.Signals(signal_num).name)
        stopped = server.stop(grace=5)
        stopped.wait()
        sys.exit(0)

    signal.signal(signal.SIGINT, _on_shutdown)
    signal.signal(signal.SIGTERM, _on_shutdown)

    server.wait_for_termination()
