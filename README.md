# sci-vault

`sci-vault` is an AI-powered collaborative platform for intelligent management and discovery of laboratory research data. The system leverages modern microservices architecture with gRPC communication, embedding-based recommendations, and vector search capabilities to provide researchers with powerful tools for data exploration and insight extraction.

## Overview

This monorepo contains a complete microservices-based application for research data management:

- **svc-gateway**: REST API gateway built with Go and Gin, serving as the primary entry point for client requests
- **svc-recommender**: gRPC-based recommendation engine built with Python, utilizing embedding models and pgvector for intelligent content discovery
- **frontend**: Modern SvelteKit-based web application for user-friendly data exploration and interaction

## Technology Stack

- **Backend**: Go (Gateway), Python (Recommender Service)
- **Frontend**: SvelteKit, Vite, Tailwind CSS, TypeScript
- **Communication**: gRPC, REST API
- **Database**: PostgreSQL with pgvector extension
- **Build & Code Gen**: Buf (Protocol Buffers), Vite, Go modules, Python UV

## Prerequisites

Before getting started, ensure you have the following tools installed:

- [Buf](https://buf.build/docs/cli/installation/) - For code generation from protobuf definitions
- [Go](https://go.dev/doc/install) 1.25+ - For the gateway service
- [Python](https://www.python.org/downloads/) 3.10+ with [uv](https://docs.astral.sh/uv/getting-started/installation/) - For the recommender service
- [Bun](https://bun.sh/) 1.0+ - For the frontend application
- [PostgreSQL](https://www.postgresql.org/download/) with [pgvector](https://github.com/pgvector/pgvector) extension - For vector storage

## Quick Start

### 1. Generate gRPC Code

The first step is to generate the necessary gRPC code for both services using Buf:

```bash
buf generate
```

This reads the protobuf definitions from the `proto/` directory and generates the required gRPC stubs and client code for `svc-gateway` and `svc-recommender`. **Re-run this command whenever you modify any `.proto` files.**

### 2. Set Up Individual Services

Each service has its own setup and runtime requirements. Navigate to the respective service directories and follow their specific instructions:

#### svc-gateway (API Gateway)

```bash
cd svc-gateway
go mod tidy
go run .
```

See [svc-gateway/README.md](./svc-gateway/README.md) for detailed instructions.

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
├── buf.yaml               # Buf configuration for code generation
├── buf.gen.yaml           # Buf generation settings
└── README.md              # This file
```

## Development Workflow

1. **Update Proto Definitions**: Modify files in `proto/` as needed
2. **Generate Code**: Run `buf generate` to update generated code
3. **Develop Services**: Make changes to individual service code
4. **Test**: Run services independently or deploy with Docker (coming soon)

## Service Communication

- **Frontend ↔ Gateway**: REST API over HTTP
- **Gateway ↔ Recommender**: gRPC over TCP
- **All Services ↔ Database**: PostgreSQL connections

## Roadmap

Upcoming improvements include:
- Additional recommendation algorithms and personalization features
- Docker and docker-compose configuration for containerized development and deployment
- Enhanced CI/CD pipeline with automated testing and building

## License

This project is licensed under the [LICENSE](LICENSE) file in the root directory.
