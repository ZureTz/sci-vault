# sci-vault

`sci-vault` is an AI-powered collaborative platform for intelligent management and discovery of laboratory research data. The system leverages modern microservices architecture with gRPC communication, embedding-based recommendations, and vector search capabilities to provide researchers with powerful tools for data exploration and insight extraction.

## Overview

This monorepo contains a complete microservices-based application for research data management:

- **svc-gateway**: REST API gateway built with Go and Gin. Handles user authentication (registration, login, email verification, password reset), profile management, avatar uploads, JWT middleware, and routes recommendation requests to `svc-recommender` via gRPC.
- **svc-recommender**: gRPC-based recommendation engine built with Python, utilizing embedding models and pgvector for intelligent content discovery.
- **frontend**: Modern SvelteKit (Svelte 5) web application with dark/light theme, i18n support (en / zh-CN), and full integration with the gateway API.

## Technology Stack

| Layer             | Technologies                                                   |
| ----------------- | -------------------------------------------------------------- |
| **Gateway**       | Go 1.26, Gin, GORM, JWT, Redis, gomail                         |
| **Recommender**   | Python 3.14, gRPC, pgvector                                    |
| **Frontend**      | SvelteKit (Svelte 5), Vite, Tailwind CSS v4, TypeScript, Axios |
| **Database**      | PostgreSQL (with pgvector extension)                           |
| **Cache**         | Redis                                                          |
| **Storage**       | RustFS (S3-compatible object storage)                          |
| **Communication** | gRPC (gateway ↔ recommender), REST/HTTP (frontend ↔ gateway)   |
| **Code Gen**      | Buf (Protocol Buffers)                                         |

## Prerequisites

Before getting started, ensure you have the following tools installed:

- [Buf](https://buf.build/docs/cli/installation/) — For code generation from protobuf definitions
- [Go](https://go.dev/doc/install) 1.26+ — For the gateway service
- [Python](https://www.python.org/downloads/) 3.10+ with [uv](https://docs.astral.sh/uv/getting-started/installation/) — For the recommender service
- [Bun](https://bun.sh/) 1.0+ — For the frontend application
- [Docker](https://docs.docker.com/get-docker/) with Docker Compose — For running infrastructure services

## Quick Start

### 1. Generate gRPC Code

Generate the necessary gRPC stubs for both services using Buf. **This step is required before running the application:**

```bash
buf generate
```

This reads the protobuf definitions from the `proto/` directory and generates the required stubs for `svc-gateway` and `svc-recommender`. **Re-run this command whenever you modify any `.proto` files.**

### 2. Start Everything with Docker Compose

> **⚠️ WARNING for Production environments**: The provided `docker-compose.yaml` and default configurations are intended for **local development and testing only**. When deploying to a production environment, you MUST use your own secure parameters, strong passwords, and proper secrets management. It is highly recommended to create and use a dedicated `docker-compose-production.yaml` with hardened configurations.

A `docker-compose.yaml` is provided at the root to spin up the entire application (Frontend, Gateway, Recommender) along with all required infrastructure services (PostgreSQL, Redis, RustFS) in one command.

First, prepare the gateway configuration file:

```bash
cp svc-gateway/config.docker.example.yaml svc-gateway/config.docker.yaml
```

Then open `svc-gateway/config.docker.yaml` and fill in your secrets:

| Field                                       | What to set                                                                                                            |
| ------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `database.password`                         | Must match `POSTGRES_PASSWORD` in `docker-compose.yaml`                                                                |
| `storage.access_key` / `storage.secret_key` | Must match `RUSTFS_ACCESS_KEY` / `RUSTFS_SECRET_KEY` in `docker-compose.yaml` (default: `rustfsadmin` / `rustfsadmin`) |
| `mailer.username` / `mailer.password`       | Your SMTP credentials                                                                                                  |
| `jwt.secret`                                | Any strong random string                                                                                               |

> **RustFS S3 credentials**: RustFS uses the same `RUSTFS_ACCESS_KEY`/`RUSTFS_SECRET_KEY` values from `docker-compose.yaml` directly as its S3 API credentials — no need to log into the web console to generate separate keys. For local development the defaults (`rustfsadmin` / `rustfsadmin`) work out of the box.

Then run the cluster:

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

> **Updating configuration**: `config.docker.yaml` is mounted into the gateway container at runtime (not baked into the image). After editing it, only a restart is needed — no rebuild:
> ```bash
> docker compose restart gateway
> # or for production:
> docker compose -f docker-compose-production.yaml restart gateway
> ```

To stop the infrastructure:

```bash
# For Local Development
docker compose down

# For Production
docker compose -f docker-compose-production.yaml down
```

### 3. Set Up Local Infrastructure only (Optional)

If you prefer to develop services locally while running only the back-end infrastructure (DB, Redis, S3) in Docker:

```bash
# For Local Development
docker compose up -d postgres redis rustfs

# For Production
docker compose -f docker-compose-production.yaml up -d postgres redis rustfs
```

### 4. Set Up Individual Services for Local Development (Optional)

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
