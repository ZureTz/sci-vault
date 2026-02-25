# ── Proto generation stage ──────────────────────────────────────────────────
FROM python:3.14.3-slim AS proto-gen

WORKDIR /proto-gen

# Install grpcio-tools (build-time only)
RUN pip install --no-cache-dir grpcio-tools

COPY proto/ ./proto/

RUN mkdir gen

RUN python -m grpc_tools.protoc \
    -I./proto \
    --python_out=gen \
    --pyi_out=gen \
    ./proto/health/health.proto

# ── Build stage ─────────────────────────────────────────────────────────────
FROM python:3.14.3-slim AS builder

WORKDIR /app

# Install uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /usr/local/bin/

# Copy dependency manifests first (layer cache)
COPY svc-recommender/pyproject.toml svc-recommender/uv.lock ./

# Install production deps into /app/.venv (no dev extras)
RUN uv sync --frozen --no-dev

# Copy generated protobuf bindings from proto-gen stage
COPY --from=proto-gen /proto-gen/gen ./gen/

# Copy source
COPY svc-recommender/main.py ./
COPY svc-recommender/config.yaml ./

# ── Runtime stage ────────────────────────────────────────────────────────────
FROM python:3.14.3-slim

WORKDIR /app

COPY --from=builder /app /app

ENV PATH="/app/.venv/bin:$PATH" \
    PYTHONUNBUFFERED=1 \
    CONFIG_PATH=/app/config.yaml

CMD ["python", "main.py"]
