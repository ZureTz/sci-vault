# SVC-Recommender

`svc-recommender` is a recommender system that provides personalized recommendations based on user preferences and historical data. The system utilizes embedding models and vectorization to represent the semantic content of articles, and leverages a vector database (`pgvector`) to efficiently store and retrieve these embeddings.

Acting as a core component of a larger microservices architecture, this service is implemented as a highly scalable gRPC server designed to deliver real-time recommendations to users.

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

| RPC | Description |
|-----|-------------|
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

## Roadmap: Docker Integration

A `Dockerfile` and `docker-compose.yaml` configuration will be added soon to containerize the service, ensuring a seamless and reliable deployment process across all environments.

```bash
# Future usage
docker compose up -d
```

## License

This project is licensed under the [LICENSE](../LICENSE) file in the root directory.
