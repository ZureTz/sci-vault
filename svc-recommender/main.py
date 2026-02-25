"""
svc-recommender — Kafka consumer for health events (protobuf).

Reads HealthEvent messages from the configured Kafka topic, deserialises
them from protobuf and logs each event.  Extend the `handle_event` function
to add real recommendation / processing logic.
"""

import logging
import os
import signal
import sys
from pathlib import Path
from types import FrameType
from typing import Optional

import yaml
from kafka import KafkaConsumer
from kafka.errors import NoBrokersAvailable

from gen.health.health_pb2 import HealthEvent

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s  %(levelname)-8s  %(name)s  %(message)s",
    datefmt="%Y-%m-%dT%H:%M:%S",
)
log = logging.getLogger("svc-recommender")


# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------


def load_config(path: str = "config.yaml") -> dict:
    cfg_path = Path(path)
    if not cfg_path.exists():
        log.warning("config file %s not found, using defaults", path)
        return {}
    with cfg_path.open() as f:
        return yaml.safe_load(f) or {}


# ---------------------------------------------------------------------------
# Business logic
# ---------------------------------------------------------------------------


def handle_event(event: HealthEvent) -> None:
    """Process a single HealthEvent.  Replace / extend this with real logic."""
    log.info(
        "received health event  service=%s  status=%s  timestamp=%d",
        event.service,
        event.status,
        event.timestamp,
    )


# ---------------------------------------------------------------------------
# Consumer
# ---------------------------------------------------------------------------


def build_consumer(brokers: list[str], topic: str, group_id: str) -> KafkaConsumer:
    return KafkaConsumer(
        topic,
        bootstrap_servers=brokers,
        group_id=group_id,
        auto_offset_reset="earliest",
        enable_auto_commit=True,
        # Raw bytes — we deserialise manually with protobuf
        value_deserializer=None,
        # Retry connection for up to ~30 s on startup
        reconnect_backoff_ms=500,
        reconnect_backoff_max_ms=5000,
    )


def run(consumer: KafkaConsumer) -> None:
    log.info("consumer started, waiting for messages …")
    for msg in consumer:
        if msg.value is None:
            continue
        event = HealthEvent()
        try:
            event.ParseFromString(msg.value)
        except Exception as exc:  # noqa: BLE001
            log.error("failed to parse protobuf message: %s", exc)
            continue
        handle_event(event)


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------


def main() -> None:
    cfg = load_config(os.getenv("CONFIG_PATH", "config.yaml"))

    kafka_cfg = cfg.get("kafka", {})
    # Environment variables take precedence over config file values (Docker-friendly).
    brokers_raw = os.getenv("KAFKA_BROKERS")
    brokers: list[str] = (
        brokers_raw.split(",")
        if brokers_raw
        else kafka_cfg.get("brokers", ["localhost:9092"])
    )
    topic: str = os.getenv("KAFKA_TOPIC") or kafka_cfg.get("topic", "health-events")
    group_id: str = os.getenv("KAFKA_GROUP_ID") or kafka_cfg.get(
        "group_id", "svc-recommender"
    )

    log.info("connecting to brokers=%s topic=%s group=%s", brokers, topic, group_id)

    consumer: Optional[KafkaConsumer] = None

    def _shutdown(sig: int, _frame: Optional[FrameType]) -> None:
        log.info("signal %d received, shutting down …", sig)
        if consumer:
            consumer.close()
        sys.exit(0)

    signal.signal(signal.SIGINT, _shutdown)
    signal.signal(signal.SIGTERM, _shutdown)

    try:
        consumer = build_consumer(brokers, topic, group_id)
    except NoBrokersAvailable as exc:
        log.error("cannot connect to Kafka: %s", exc)
        sys.exit(1)

    run(consumer)


if __name__ == "__main__":
    main()
