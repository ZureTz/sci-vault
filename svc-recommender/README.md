# svc-recommender

Python microservice (managed by [uv](https://github.com/astral-sh/uv)) that consumes `HealthEvent` protobuf messages from the `health-events` Kafka topic produced by `svc-gateway`.

## Project layout

```
svc-recommender/
├── main.py              # Entry point — Kafka consumer loop
├── health/
│   └── health_pb2.py   # Generated from proto/health/health.proto
├── config.yaml          # Local config (gitignored)
├── config.example.yaml  # Template
├── Makefile             # make proto | make run
├── Dockerfile
├── pyproject.toml
└── uv.lock
```

## Quick start (local)

```bash
# 1. Install deps
uv sync

# 2. Start Kafka (from repo root)
docker compose up -d kafka

# 3. Run the consumer
make run
# or
uv run python main.py
```

## Regenerate protobuf bindings

```bash
make proto
```

## Configuration

Priority order (highest first):

| Source | Variable |
|--------|----------|
| Env var | `KAFKA_BROKERS` (comma-separated), `KAFKA_TOPIC`, `KAFKA_GROUP_ID` |
| `config.yaml` | `kafka.brokers`, `kafka.topic`, `kafka.group_id` |
| Defaults | `localhost:9092`, `health-events`, `svc-recommender` |

## Docker

```bash
docker compose up --build svc-recommender
```
