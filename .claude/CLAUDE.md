# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

sci-vault is an AI-powered microservices platform for laboratory research data management. It consists of three services:

- **svc-gateway** (Go/Gin) — REST API gateway, handles auth, document CRUD, user management
- **svc-recommender** (Python/gRPC) — AI-powered document enrichment (metadata extraction + vector embeddings via Google Gemini)
- **frontend** (SvelteKit 2 / Svelte 5) — Web UI with Tailwind CSS v4, shadcn-svelte + Bits UI components, i18n (en, zh-CN)

Infrastructure: PostgreSQL 18 + pgvector, Redis 8.6, RustFS (S3-compatible storage).

## Architecture

```
Browser → (REST) → svc-gateway → (gRPC) → svc-recommender
                        ↓                       ↓
                  PostgreSQL/Redis/RustFS   PostgreSQL/Redis/RustFS/GenAI
```

Both backend services follow layered architecture: handlers/servicers → services → repositories → infrastructure.

Document enrichment is async: gateway calls `EnrichDocument` RPC, recommender ACKs immediately, then processes in a background thread (extract PDF → Gemini metadata + embedding → store in pgvector).

## Common Commands

### Protobuf Code Generation
```bash
buf generate   # Regenerate Go + Python gRPC stubs from proto/
```

### Docker (Full Stack)
```bash
docker compose up -d --build
docker compose down
```

### svc-gateway (Go)
```bash
cd svc-gateway
go run .               # Run locally
go build -v ./...      # Build
go test ./...          # Run all tests
go vet ./...           # Lint
```

### svc-recommender (Python, uses uv)
```bash
cd svc-recommender
uv sync                              # Install dependencies
uv run --env-file .env main.py       # Run locally
uvx ruff check .                     # Lint
uvx ruff format --check .            # Format check
```

### frontend (SvelteKit, uses bun)
```bash
cd frontend
bun install            # Install dependencies
bun run dev            # Dev server (localhost:5173)
bun run build          # Production build
bun run check          # Svelte + TypeScript validation
bun run lint           # Prettier + ESLint
bun run format         # Auto-format
```

## Frontend UI Components (shadcn-svelte)

The frontend uses **shadcn-svelte** — a collection of copy-paste, accessible Svelte components built on Bits UI primitives and styled with Tailwind CSS v4. Components live in `frontend/src/lib/components/ui/`.

**Docs**: https://www.shadcn-svelte.com/llms.txt (LLM-friendly reference) | https://www.shadcn-svelte.com

**Available component categories:**
- **Form & Input**: Button, Calendar, Checkbox, Combobox, Date Picker, Input, Slider, Switch, Textarea
- **Layout & Navigation**: Accordion, Breadcrumb, Sidebar, Tabs, Resizable panels
- **Overlays & Dialogs**: Dialog, Drawer, Dropdown Menu, Tooltip, Popover, Alert Dialog, Sheet
- **Feedback & Status**: Alert, Badge, Progress, Skeleton, Spinner, Sonner (toast)
- **Display & Media**: Avatar, Card, Carousel, Chart, Data Table

**Usage conventions:**
- Import as namespace: `import * as Card from '$lib/components/ui/card'`, then use `<Card.Root>`, `<Card.Header>`, etc.
- Icons come from `lucide-svelte`
- Theme tokens use CSS custom properties (`--primary`, `--muted`, `--background`, etc.)
- Dark mode via `.dark` class on the `<html>` element (managed by `mode-watcher`)

## Key Conventions

- **Commit messages**: Conventional commits (`feat:`, `fix:`, `chore:`, `docs:`)
- **Config files**: Each service has `config.yaml` (local, gitignored), `config.docker.yaml` (Docker), and `config.example.yaml` (checked in)
- **Proto changes**: Always run `buf generate` after modifying `.proto` files; CI regenerates automatically
- **Gateway error handling**: Use `app_error` package for structured API error responses
- **Frontend auth**: JWT in localStorage, Axios interceptor adds Authorization header and redirects to login on 401
- **Frontend routing**: SvelteKit filesystem routes; `(dashboard)` group layout wraps authenticated pages
- **Svelte 5 runes**: Use the latest Svelte 5 runes API, not legacy reactive declarations
- **Validation**: go-playground/validator v10 with custom validators for username/password rules
- **Vector embeddings**: 1536-dimensional via `gemini-embedding-001`, stored with pgvector HNSW cosine index
