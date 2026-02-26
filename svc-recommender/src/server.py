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


def serve(cfg: Config | None = None) -> None:
    """Start the gRPC server and block until a shutdown signal is received."""
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    )

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
