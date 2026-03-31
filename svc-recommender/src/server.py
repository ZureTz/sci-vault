"""gRPC server lifecycle management."""

import logging
import signal
import sys
from concurrent import futures

import grpc

from pb.recommender.v1 import recommender_pb2_grpc
from config import Config
from interceptor.logging import LoggingInterceptor
from infrastructure.cache import Cache
from infrastructure.genai import GenAI
from infrastructure.database import Database
from infrastructure.storage import Storage
from cache.enrichment import EnrichmentStatusCache
from genai.document import DocumentGenAI
from repository.document import DocumentRepository
from storage.document import DocumentStorage
from servicer.document import DocumentServicer
from servicer.health import HealthServicer

log = logging.getLogger(__name__)


class RecommenderServer:
    """Owns all infrastructure and the gRPC server lifecycle."""

    def __init__(self, cfg: Config) -> None:
        self._db = Database(cfg)
        self._cache = Cache(cfg)
        self._storage = Storage(cfg)
        self._genai = GenAI(cfg)

        enrich_cache = EnrichmentStatusCache(self._cache.client)
        doc_repo = DocumentRepository(self._db.pool)
        doc_storage = DocumentStorage(self._storage.client, cfg.s3_private_bucket)
        doc_genai = DocumentGenAI(
            self._genai.metadata_client, self._genai.embedding_client
        )

        class _Servicer(DocumentServicer, HealthServicer):
            pass

        self._server = grpc.server(
            futures.ThreadPoolExecutor(max_workers=cfg.max_workers),
            interceptors=[LoggingInterceptor()],
        )
        recommender_pb2_grpc.add_RecommenderServiceServicer_to_server(
            _Servicer(enrich_cache, doc_repo, doc_storage, doc_genai),
            self._server,
        )
        self._server.add_insecure_port(cfg.addr)

    def start(self) -> None:
        self._server.start()

    def stop(self, grace: float = 5.0) -> None:
        stopped = self._server.stop(grace=grace)
        stopped.wait()
        self._db.close()

    def wait_for_termination(self) -> None:
        self._server.wait_for_termination()


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
        record_copy = logging.makeLogRecord(record.__dict__)
        record_copy.levelname = f"{color}{record_copy.levelname}{self.RESET}"
        return super().format(record_copy)


def serve(cfg: Config | None = None) -> None:
    """Start the gRPC server and block until a shutdown signal is received."""
    fmt = "%(asctime)s %(levelname)s %(name)s: %(message)s"
    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(ColoredFormatter(fmt))
    logging.root.handlers.clear()
    logging.basicConfig(level=logging.INFO, handlers=[handler])

    if cfg is None:
        cfg = Config.load()

    server = RecommenderServer(cfg)
    log.info("starting svc-recommender on %s", cfg.addr)
    server.start()

    def _on_shutdown(signal_num: int, _: object) -> None:
        log.info("received signal %s, shutting down…", signal.Signals(signal_num).name)
        server.stop()
        sys.exit(0)

    signal.signal(signal.SIGINT, _on_shutdown)
    signal.signal(signal.SIGTERM, _on_shutdown)

    server.wait_for_termination()
