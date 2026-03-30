# sci-vault

`sci-vault` is an AI-powered collaborative platform for intelligent management and discovery of laboratory research data. The system leverages modern microservices architecture with gRPC communication, embedding-based recommendations, and vector search capabilities to provide researchers with powerful tools for data exploration and insight extraction.

## Overview

This monorepo contains a complete microservices-based application for research data management:

- **svc-gateway**: REST API gateway built with Go and Gin. Handles user authentication (registration, login, email verification, password reset), profile management, avatar uploads, JWT middleware, and routes recommendation requests to `svc-recommender` via gRPC.
- **svc-recommender**: gRPC-based recommendation engine built with Python, utilizing embedding models and pgvector for intelligent content discovery.
- **frontend**: Modern SvelteKit (Svelte 5) web application with dark/light theme, i18n support (en / zh-CN), and full integration with the gateway API.

## Technology Stack

| Layer             | Technologies                                                                  |
| ----------------- | ----------------------------------------------------------------------------- |
| **Gateway**       | Go 1.26, Gin, GORM, Viper, JWT, Redis, AWS SDK v2 (S3-compatible RustFS), gomail |
| **Recommender**   | Python 3.14, gRPC (`grpcio`), psycopg, pgvector, Redis                        |
| **Frontend**      | SvelteKit 2 (Svelte 5), Vite 8, Tailwind CSS v4, TypeScript, Axios, Bits UI   |
| **Database**      | PostgreSQL 18 (with pgvector extension)                                        |
| **Cache**         | Redis 8.6                                                                       |
| **Storage**       | RustFS 1.0.0-alpha.89 (S3-compatible object storage)                           |
| **Communication** | gRPC (gateway ↔ recommender), REST/HTTP (frontend ↔ gateway)                  |
| **Code Gen**      | Buf (Protocol Buffers)                                                          |

## Prerequisites

**For Docker Deployment (Recommended):**
- [Docker](https://docs.docker.com/get-docker/) with Docker Compose

**For Manual Local Development:**
- [Buf](https://buf.build/docs/cli/installation/) — For gRPC protobuf code generation
- [Go](https://go.dev/doc/install) 1.26+ — For the gateway service
- [Python](https://www.python.org/downloads/) 3.14+ with [uv](https://docs.astral.sh/uv/getting-started/installation/) — For the recommender service
- [Bun](https://bun.sh/) 1.0+ — For the frontend application

## Quick Start Using Docker Compose

> **⚠️ WARNING for Production environments**: The provided `docker-compose.yaml` and default configurations are intended for **local development and testing only**. When deploying to a production environment, you MUST use your own secure parameters, strong passwords, and proper secrets management. It is highly recommended to create and use a dedicated `docker-compose-production.yaml` with hardened configurations.

A `docker-compose.yaml` is provided at the root to spin up the entire application (Frontend, Gateway, Recommender) along with all required infrastructure services (PostgreSQL, Redis, RustFS) in one command.

*Note: When you build the services via the provided `docker-compose.yaml`, the gRPC stubs are generated automatically inside the Docker images using the `workspace_root` additional build context. If you instead run `docker build` directly, you must either pass the same additional context used by Compose or run `buf generate` locally before building; otherwise, the `COPY --from=workspace_root ...` steps will fail and gRPC stubs will not be generated.*

### 1. Prepare the configuration files for each service:

```bash
cp svc-gateway/config.docker.example.yaml svc-gateway/config.docker.yaml
cp svc-recommender/config.docker.example.yaml svc-recommender/config.docker.yaml
cp frontend/nginx.example.conf frontend/nginx.conf
```

### 2. Open `svc-gateway/config.docker.yaml` and fill in your secrets:

| Field                                       | What to set                                                                                                            |
| ------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `database.password`                         | Must match `POSTGRES_PASSWORD` in `docker-compose.yaml`                                                                |
| `storage.access_key` / `storage.secret_key` | Must match `RUSTFS_ACCESS_KEY` / `RUSTFS_SECRET_KEY` in `docker-compose.yaml` (default: `rustfsadmin` / `rustfsadmin`) |
| `mailer.username` / `mailer.password`       | Your SMTP credentials                                                                                                  |
| `jwt.secret`                                | Any strong random string                                                                                               |

> **RustFS S3 credentials**: RustFS uses the same `RUSTFS_ACCESS_KEY`/`RUSTFS_SECRET_KEY` values from `docker-compose.yaml` directly as its S3 API credentials — no need to log into the web console to generate separate keys. For local development the defaults (`rustfsadmin` / `rustfsadmin`) work out of the box.

### 3. Run the cluster:

```bash
# For Local Development
docker compose up -d --build

# For Production (using production-specific config)
cp docker-compose.yaml docker-compose-production.yaml # If you haven't created one yet
# Edit docker-compose-production.yaml and svc-gateway/config.docker.yaml with production-grade secrets
docker compose -f docker-compose-production.yaml up -d --build
```

This starts the following containers:

| Container               | Service                        | Ports                          |
| ----------------------- | ------------------------------ | ------------------------------ |
| `sci-vault-postgres`    | PostgreSQL 18                  | `5432`                         |
| `sci-vault-redis`       | Redis 8                        | `6379`                         |
| `sci-vault-rustfs`      | RustFS (S3-compatible storage) | `9000` (API), `9001` (Console) |
| `sci-vault-recommender` | Recommender (gRPC)             | `50051`                        |
| `sci-vault-gateway`     | API Gateway (REST)             | `8080`                         |
| `sci-vault-frontend`    | Frontend Web Client            | `80`                           |

> **Updating configuration**: `config.docker.yaml` is mounted into each service container at runtime (not baked into the image). After editing, only a restart is needed — no rebuild:
> ```bash
> docker compose restart gateway recommender
> # or for production:
> docker compose -f docker-compose-production.yaml restart gateway recommender
> ```

> **Frontend nginx config**: `frontend/nginx.conf` (copied from `nginx.example.conf`) is mounted at runtime. You can edit it (e.g. tune proxy settings, add security headers) and restart without rebuilding:
> ```bash
> docker compose restart frontend
> ```

> **HTTPS / TLS**: To enable HTTPS for the frontend:
> 1. Place your certificate and private key at `frontend/ssl/cert.pem` and `frontend/ssl/key.pem`
> 2. Uncomment the `listen 443 ssl` block and certificate directives in `frontend/nginx.conf`
> 3. Update the `server_name` in `frontend/nginx.conf` to your actual domain
> 4. Restart the frontend container: `docker compose restart frontend`
>
> The `frontend/ssl/` directory is mounted read-only into the container at `/etc/nginx/ssl/`. The directory is git-ignored — never commit your private key.

To stop the infrastructure:

```bash
# For Local Development
docker compose down

# For Production
docker compose -f docker-compose-production.yaml down
```

## Manual Local Development

If you prefer to run the services directly on your host machine instead of using Docker, follow these steps.

### 1. Start Infrastructure Dependencies
If you prefer to develop services locally while running only the back-end infrastructure (DB, Redis, S3) in Docker:

```bash
docker compose up -d postgres redis rustfs # Add rustfs-volume-helper if first time
```

### 2. Generate gRPC Code
Make sure [Buf](https://buf.build/docs/cli/installation/) is installed.

```bash
buf generate
```
**Important:** Re-run this command whenever you modify any `.proto` files in the `proto/` directory.

### 3. Set Up Individual Services

Each service has its own setup and runtime requirements. Navigate to the respective directories and follow their specific README:

#### svc-gateway (API Gateway)

```bash
cd svc-gateway
cp config.example.yaml config.yaml  # then fill in your values
go mod tidy
go run .
```

See [svc-gateway/README.md](./svc-gateway/README.md) for the full configuration reference and API endpoint documentation.

#### svc-recommender (Recommender Engine)

```bash
cd svc-recommender
uv sync
uv run --env-file .env main.py
```

See [svc-recommender/README.md](./svc-recommender/README.md) for detailed instructions.

#### frontend (Web Client)

```bash
cd frontend
bun install
bun run dev
```

See [frontend/README.md](./frontend/README.md) for detailed instructions.

## Project Structure

```text
sci-vault/
├── proto/                  # Protocol buffer definitions for gRPC
│   └── recommender/        # Recommender service proto
├── svc-gateway/            # Go-based API gateway
├── svc-recommender/        # Python-based recommendation engine
├── frontend/               # SvelteKit web application
├── docker-compose.yaml     # Infrastructure services (PostgreSQL, Redis, RustFS)
├── buf.yaml                # Buf configuration for code generation
├── buf.gen.yaml            # Buf generation settings
└── README.md               # This file
```

## Development Workflow

1. **Start infrastructure**: `docker compose up -d`
2. **Update Proto Definitions**: Modify files in `proto/` as needed, then run `buf generate`
3. **Develop Services**: Make changes to individual service code
4. **Run Services**: Start `svc-recommender`, `svc-gateway`, and `frontend` independently

## Service Communication

```
Browser
  │  REST/HTTP
  ▼
svc-gateway ──gRPC──► svc-recommender
  │
  ├── PostgreSQL  (user data, profiles)
  ├── Redis       (email verification codes, cache)
  └── RustFS      (avatar & asset storage)
```

## Roadmap

- Additional recommendation algorithms and personalization features
- Enhanced CI/CD pipeline with automated testing and building

## License

This project is licensed under the [LICENSE](LICENSE) file in the root directory.
