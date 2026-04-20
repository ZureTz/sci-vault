FROM astral/uv:python3.14-alpine

WORKDIR /app

ENV PYTHONUNBUFFERED=1 \
    UV_COMPILE_BYTECODE=1 \
    UV_SYNC_TIMEOUT=180

COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev

COPY . .

EXPOSE 50051

CMD ["uv", "run", "main.py"]
