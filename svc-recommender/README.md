# SVC-Recommender

`svc-recommender` is a recommender system that provides personalized recommendations based on user preferences and historical data. The system utilizes embedding models and vectorization to represent the semantic content of articles, and leverages a vector database (`pgvector`) to efficiently store and retrieve these embeddings.

Acting as a core component of a larger microservices architecture, this service is implemented as a highly scalable gRPC server designed to deliver real-time recommendations to users.

## Tech Stack

- **Language**: Python 3.14+
- **RPC**: gRPC (`grpcio`)
- **AI / LLM**: Google GenAI (Gemini for metadata enrichment and translation, `gemini-embedding-001` for vector embedding)
- **Configuration**: YAML + environment variables
- **Storage / Data**: PostgreSQL (`psycopg`) with `pgvector`, Redis, RustFS (S3 via `boto3`)
- **Document conversion**: PDF/text pass through unchanged; DOCX/PPTX/XLSX are converted to PDF via headless LibreOffice (input to Gemini only — never persisted)
- **Package & Runtime Tooling**: `uv`

## Getting Started

Follow the instructions below to set up and run the `svc-recommender` service locally.

### Prerequisites

- [uv](https://docs.astral.sh/uv/getting-started/installation/): An extremely fast Python package installer and resolver.
- **LibreOffice** + **CJK fonts** — required to convert DOCX/PPTX/XLSX documents to PDF for enrichment. The Docker image installs these for you, so this is only needed when running the service **outside Docker**.

If you haven't installed `uv` yet, you can do so by running the following command:

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

> **Running without Docker?** Install `libreoffice` and the Noto CJK fonts via your system package manager so Office-format conversion (and Chinese-language document rendering) works. Examples:
>
> ```bash
> # Debian / Ubuntu
> sudo apt-get install -y libreoffice fonts-noto-cjk
>
> # macOS (Homebrew)
> brew install --cask libreoffice
> brew install font-noto-sans-cjk font-noto-serif-cjk
>
> # Alpine
> apk add libreoffice font-noto-cjk
> ```
>
> Without LibreOffice, only PDF/TXT/Markdown uploads can be enriched; Office formats will be marked `failed`.

### Install Dependencies

Navigate to the project directory and install the required dependencies using `uv`:

```bash
uv sync
```

### Configuration

The service loads configuration in the following priority order (highest to lowest):

1. **Environment variables** — override individual keys
2. **`config.yaml`** in the project root (optional)
3. **Built-in defaults** — host `0.0.0.0`, port `50051`, max workers `10`

To use a config file, copy the example and update it as needed:

```bash
cp config.example.yaml config.yaml
```

The full configuration reference is:

```yaml
host: "0.0.0.0"
port: 50051
max_workers: 10

redis_host: "localhost"
redis_port: 6379
redis_password: ""
redis_db: 0

db_host: "localhost"
db_port: 5432
db_user: "sci_vault"
db_password: "sci_vault_secret"
db_name: "sci_vault"
db_ssl_mode: "disable"
db_timezone: "UTC"

s3_endpoint: "http://0.0.0.0:9000"
s3_access_key: "rustfsadmin"
s3_secret_key: "rustfsadmin"
s3_private_bucket: "sci-vault"

# Decides whether to use the Vertex AI endpoint or the Gemini API endpoint
google_genai_use_vertexai: true
# Obtain a key at https://aistudio.google.com/app/api-keys or the Vertex AI console
google_genai_api_key: "YOUR_GEMINI_API_KEY"
# Optional proxy settings for the Google GenAI API
google_genai_use_proxy: false
google_genai_proxy_url: "http://your-proxy:port"
```

These keys may also be supplied via environment variables (e.g. in a `.env` file consumed
by `uv run --env-file .env`), which take precedence over `config.yaml`.

### Run the Service

Start the gRPC server (the service will automatically pick up `config.yaml` and any environment variables):

```bash
uv run --env-file .env main.py
```

The server is now up and ready to receive gRPC requests on port `50051` by default.

## gRPC API

The service exposes the `RecommenderService` defined in [`proto/recommender/v1/recommender.proto`](../proto/recommender/v1/recommender.proto).

| RPC                      | Description                                                                                                                                |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `Health`                 | Returns liveness status (`ok`) and service name                                                                                            |
| `EnrichDocument`         | Fire-and-forget async enrichment (authors, summary, tags, embedding); ACKs immediately and processes in the background, status via Redis   |
| `TranslateText`          | Server-streaming LLM translation; streams translated chunks as they are generated                                                          |
| `SemanticSearch`         | Embeds the query (`RETRIEVAL_QUERY`) and runs vector search with a keyword fallback; access-scoped to the caller's private + lab docs      |
| `RecommendSimilar`       | Nearest neighbours to a source document's embedding (source excluded); same access scoping as `SemanticSearch`                             |
| `RecommendForUser`       | Personalised feed built from the caller's like/view/search-query signals averaged into a profile centroid; liked docs excluded             |
| `RecommendCollaborators` | Ranks the caller's lab-mates by interest-profile similarity; candidate centroids built server-side in SQL; scoped to one lab               |

## Directory Structure

```text
svc-recommender/
├── src/
│   ├── cache/                # Redis wrappers (query-embedding cache, enrich status)
│   ├── conversion/           # Document format conversion
│   │   ├── converter.py      # Dispatch on content type → Gemini-ready payload
│   │   └── office.py         # OOXML → PDF via headless LibreOffice
│   ├── genai/                # Google Gemini integration
│   │   ├── document.py       # Metadata extraction + document embedding
│   │   ├── embedding_resolver.py # 3-tier query-embedding cache (Redis → Postgres → Gemini)
│   │   ├── query_embedder.py # Query embedding with task-type asymmetry
│   │   └── translate.py      # Streaming LLM translation
│   ├── infrastructure/       # Database and storage connection setup
│   ├── interceptor/          # gRPC interceptors (logging)
│   ├── pb/                   # Protobuf-generated stubs (gitignored)
│   ├── repository/           # SQL access layer (psycopg)
│   │   ├── document.py       # Document queries
│   │   ├── query_embedding.py # query_embeddings table backstop
│   │   ├── recommend.py      # Similar / personalised / collaborator queries
│   │   ├── search.py         # Semantic + keyword search
│   │   └── types.py          # ScoredDocument / ScoredUser row types
│   ├── servicer/             # gRPC service implementations
│   │   ├── document.py       # EnrichDocument
│   │   ├── health.py         # Health
│   │   ├── recommend.py      # RecommendSimilar / ForUser / Collaborators
│   │   ├── search.py         # SemanticSearch
│   │   └── translate.py      # TranslateText
│   ├── storage/              # S3-compatible object download (RustFS)
│   ├── utils/                # Helpers (query-embedding key derivation, etc.)
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

If you build this service directly with `docker build` (outside Compose), make sure you either:

1. Keep the generated stubs committed under `src/pb`, or
2. Run `buf generate` in the repository root, if you changed the proto definitions.

Otherwise, protobuf stubs will be missing and the service will fail to start.

The Docker image bundles `libreoffice` and `font-noto-cjk` so that DOCX/PPTX/XLSX (including
CJK-language documents) can be converted to PDF for enrichment.

## License

This project is licensed under the [LICENSE](../LICENSE) file in the root directory.
