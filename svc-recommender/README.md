# SVC-Recommender

`svc-recommender` is a recommender system that provides personalized recommendations based on user preferences and historical data. The system utilizes embedding models and vectorization to represent the semantic content of articles, and leverages a vector database (`pgvector`) to efficiently store and retrieve these embeddings.

Acting as a core component of a larger microservices architecture, this service is implemented as a highly scalable gRPC server designed to deliver real-time recommendations to users.

## Tech Stack

- **Language**: Python 3.14+
- **RPC**: gRPC (`grpcio`)
- **AI / LLM**: Google GenAI (Gemini 3 Flash Preview for metadata enrichment, `gemini-embedding-001` for vector embedding)
- **Configuration**: YAML + environment variables
- **Storage / Data**: PostgreSQL (`psycopg`) with `pgvector`, Redis, RustFS (S3 via `boto3`)
- **Package & Runtime Tooling**: `uv`

## Getting Started

Follow the instructions below to set up and run the `svc-recommender` service locally.

### Prerequisites

- [uv](https://docs.astral.sh/uv/getting-started/installation/): An extremely fast Python package installer and resolver.

If you haven't installed `uv` yet, you can do so by running the following command:

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

### Install Dependencies

Navigate to the project directory and install the required dependencies using `uv`:

```bash
uv sync
```

### Configuration

The service loads configuration in the following priority order (highest to lowest):

1. **Environment variables** — `GRPC_HOST`, `GRPC_PORT`, `GRPC_MAX_WORKERS`
2. **`config.yaml`** in the project root (optional)
3. **Built-in defaults** — host `0.0.0.0`, port `50051`, max workers `10`

To use a config file, copy the example and update it as needed:

```bash
cp config.example.yaml config.yaml
```

To override via environment variables (e.g. in a `.env` file):

```bash
GRPC_HOST=0.0.0.0
GRPC_PORT=50051
GRPC_MAX_WORKERS=10
```

### Run the Service

Start the gRPC server (the service will automatically pick up `config.yaml` and any environment variables):

```bash
uv run --env-file .env main.py
```

The server is now up and ready to receive gRPC requests on port `50051` by default.

## gRPC API

The service exposes the `RecommenderService` defined in the protobuf schema.

| RPC      | Description                                     |
| -------- | ----------------------------------------------- |
| `Health` | Returns liveness status (`ok`) and service name |

## Directory Structure

```text
svc-recommender/
├── src/
│   ├── interceptor/
│   │   └── logging.py        # gRPC logging interceptor
│   ├── pb/
│   │   └── recommender/      # Protobuf-generated stubs
│   ├── servicer/
│   │   └── health.py         # Health RPC implementation
│   ├── config.py             # Configuration loading (YAML + env vars)
│   └── server.py             # gRPC server factory and lifecycle
├── config.example.yaml       # Configuration template
├── config.yaml               # Local configuration file (ignored by git)
├── main.py                   # Entry point
└── pyproject.toml            # Dependency management
```

## Docker Deployment

This service is fully containerized. You can run it along with the rest of the application using Docker Compose from the root directory:

```bash
cd ..
docker compose up -d --build recommender
```

Before building or running the `recommender` service via Docker Compose, ensure that the protobuf stubs under `src/pb` have been generated on the host (for example, by running `buf generate` in the appropriate directory). The current Docker image does not generate these stubs during build, so missing stubs will prevent the service from starting correctly.
When built via the repository `docker-compose.yaml`, protobuf stubs are generated automatically inside the image using Buf and the `workspace_root` additional build context.

If you build this service directly with `docker build` (outside Compose), make sure you either:

1. Provide the same workspace context expected by the Dockerfile, or
2. Run `buf generate` in the repository root beforehand.

Otherwise, protobuf stubs will be missing and the service will fail to start.

Note: while `docker-compose.yaml` may define environment variables for services such as PostgreSQL and Redis, the `svc-recommender` service does not presently use these settings directly; it only runs the gRPC server and is wired into the broader stack via Docker networking and ports.
## License

This project is licensed under the [LICENSE](../LICENSE) file in the root directory.
