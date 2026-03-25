FROM bufbuild/buf:1.47.2 AS pb-builder
WORKDIR /workspace
COPY --from=workspace_root proto/ ./proto/
COPY --from=workspace_root buf.yaml buf.gen.yaml ./
RUN buf generate

FROM python:3.14-slim

COPY --from=ghcr.io/astral-sh/uv:0.8.15 /uv /uvx /bin/

WORKDIR /app

ENV PYTHONUNBUFFERED=1
ENV UV_COMPILE_BYTECODE=1

COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev

COPY . .
# Bring in the generated python stubs
COPY --from=pb-builder /workspace/svc-recommender/src/pb/ ./src/pb/

EXPOSE 50051

CMD ["uv", "run", "main.py"]