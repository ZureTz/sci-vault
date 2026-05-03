# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

sci-vault is an AI-powered microservices platform for laboratory research data management. It consists of three services:

- **svc-gateway** (Go/Gin) — REST API gateway, handles auth, document CRUD, user management, search, recommendations
- **svc-recommender** (Python/gRPC) — AI-powered document enrichment (metadata extraction + vector embeddings via Google Gemini), semantic search, similar-document recommendations, personalised feed
- **frontend** (SvelteKit 2 / Svelte 5) — Web UI with Tailwind CSS v4, shadcn-svelte + Bits UI components, i18n (en, zh-CN)

Infrastructure: PostgreSQL 18 + pgvector, Redis 8.6, RustFS (S3-compatible storage).

## Architecture

```
Browser → (REST) → svc-gateway → (gRPC) → svc-recommender
                        ↓                       ↓
                  PostgreSQL/Redis/RustFS   PostgreSQL/Redis/RustFS/GenAI
```

Both backend services follow layered architecture: `handler`/`servicer` → `service` → `repository` → infrastructure.

Document enrichment is async: gateway calls `EnrichDocument` RPC, recommender ACKs immediately, then processes in a background thread (extract PDF → Gemini metadata + embedding → store in pgvector). Status is tracked via Redis and polled by the frontend.

gRPC surface (`proto/recommender/v1/recommender.proto`):
- `Health` — liveness
- `EnrichDocument` — fire-and-forget async enrichment
- `TranslateText` — server-streaming LLM translation
- `SemanticSearch` — embed query + vector search (+ keyword fallback)
- `RecommendSimilar` — nearest neighbours to a source document's embedding
- `RecommendForUser` — personalised feed; the gateway forwards the caller's recent likes/views/search-queries and the recommender averages their embeddings into a profile centroid for nearest-neighbour search

## Common Commands

### Protobuf Code Generation
```bash
buf generate   # Regenerate Go + Python gRPC stubs from proto/
```
Go stubs land in `svc-gateway/internal/pb/recommender/v1/`, Python stubs in `svc-recommender/src/pb/recommender/v1/`. Always run after editing `.proto`.

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

Components live in `frontend/src/lib/components/ui/`. See https://www.shadcn-svelte.com/llms.txt for the LLM-friendly reference.

**Usage conventions:**
- Import as namespace: `import * as Card from '$lib/components/ui/card'`, then use `<Card.Root>`, `<Card.Header>`, etc.
- Icons come from `lucide-svelte`
- Theme tokens use CSS custom properties (`--primary`, `--muted`, `--background`, etc.)
- Dark mode via `.dark` class on `<html>` (managed by `mode-watcher`)
- Data tables use TanStack Table (`@tanstack/table-core`) via `$lib/components/ui/data-table` — server-side pagination/sort/filter is standard for list pages

## Backend Patterns (svc-gateway)

**Layered flow**: `handler` (Gin) → `service` (business logic) → `repo` (GORM) → infrastructure (DB / cache / S3 / gRPC client). Handlers do parameter binding + auth context extraction + error-code mapping; services own business rules and call other services/repos/gRPC; repos own SQL.

**URL parameter middleware**: `:doc_id` and `:lab_id` path params are parsed once by middleware (`middleware.ExtractDocID()` / `middleware.ExtractLabID()`) and stored in the gin context. Handlers read them via `c.GetUint("doc_id")` / `c.GetUint("lab_id")`. Routes group the middleware:
```go
docWithID := group.Group("/:doc_id").Use(middleware.ExtractDocID())
{
    docWithID.GET("",  deps.DocumentHandler.GetDocument)
    docWithID.PATCH("", deps.DocumentHandler.UpdateMetadata)
    ...
}
```
Don't reintroduce per-handler `ShouldBindUri` blocks for these params.

**Error handling**: sentinel errors live in `pkg/app_error`. Services return sentinels; handlers `errors.Is` against them and emit an i18n **code** string (not a prose message) — e.g. `service.update_document.forbidden`. The frontend looks up the code via `svelte-i18n` and the user sees the localised string. When adding a new error path:
1. Add a sentinel to `app_error`
2. Map it in the handler's `switch`
3. Add the i18n code to both `en.json` and `zh-CN.json`

**Typed repo patches**: when the service needs to update a subset of columns, the **repo** defines a typed struct (e.g. `repo.DocumentMetadataPatch`) and the repo decides which SQL columns to write. Do **not** pass `map[string]any` across the service/repo boundary — that invites column-name injection.

**SQL safety**: every `Where` / `Raw` / bound query uses `?` placeholders. Sort/filter whitelisting happens in two places: DTO `oneof` validators, and a `switch` in the repo that only assigns literal `ORDER BY` strings. Never splice user input into an identifier position.

**Gateway↔recommender access control**: the recommender enforces row-level access in its SQL (private docs owned by `user_id` OR lab-visible docs in `lab_id`). The **gateway** still guards any user-supplied `lab_id` by verifying membership — otherwise a caller could pass a lab they don't belong to.

**Async side-effects on writes**: document creates/updates/deletes also (a) update Redis enrichment status when appropriate, (b) invalidate the dashboard stats cache (`dashboardStatsKey(userID)`), (c) trigger `EnrichDocument` gRPC where relevant. These are best-effort — log-and-continue on failure; don't fail the request.

**Cross-service-owned schemas**: a few tables are declared as gateway models (so GORM `AutoMigrate` runs the DDL) but read/written only by svc-recommender — `QueryEmbedding` is the canonical example. The gateway never queries it; the recommender owns all reads/writes via psycopg. When changing such a model, remember GORM's `AutoMigrate` adds columns but **does not alter primary keys or constraints** on existing tables — for composite-PK changes in dev, drop the affected table and let AutoMigrate recreate it; production needs a real migration.

## Backend Patterns (svc-recommender)

**pgvector adapter**: `pgvector.psycopg.register_vector` is registered per-connection in `infrastructure/database.py`, so numpy arrays round-trip directly to/from `vector(768)` columns — no manual serialisation needed.

**SQL composition**: queries are built with `psycopg.sql.SQL(...).format(...)` using only `sql.Literal(<constant>)` fragments (e.g. `ENRICH_STATUS_DONE`, `DOC_VISIBILITY_PRIVATE`). All user-supplied values flow as `%(name)s` bind params.

**Access-control clauses** are shared between search and recommend repos as composable SQL snippets. When adding a new recommendation flow, reuse the same shape so private-vs-lab scoping stays consistent.

**Three-tier query embedding cache**: `genai/embedding_resolver.py:QueryEmbeddingResolver` resolves a `(text, task_type)` pair to a 768-dim vector via Redis (`cache/query_embedding.py`) → Postgres `query_embeddings` table (`repository/query_embedding.py`) → Gemini, persisting to both stores on miss. The mapping is deterministic so we never want to re-bill Gemini. **Always call `resolve_many(texts, task_type)` for batch flows** (e.g. `RecommendForUser`'s recent_queries) — each tier collapses to one round-trip (Redis MGET, Postgres `WHERE … = ANY(%s)`, batched Gemini call); calling `resolve()` in a loop is N+1 across all three tiers. The shared SHA-256 helpers live in `utils/query_embedding_key.py` so cache and repo agree on key form (raw bytes for the Postgres `bytea` PK, hex for the Redis string key).

**Embedding task-type asymmetry (critical)**: Gemini's `RETRIEVAL_QUERY` and `RETRIEVAL_DOCUMENT` produce vectors in deliberately *asymmetric* spaces — query embeddings are trained to be cosine-similar to documents about the same subject, **not** to other queries. Therefore:
- `SemanticSearch` embeds the typed query with `RETRIEVAL_QUERY` because it's matched against the corpus's `RETRIEVAL_DOCUMENT` vectors.
- `RecommendForUser` embeds historical search strings with `RETRIEVAL_DOCUMENT` because those vectors are averaged with liked/viewed *document* embeddings into a profile centroid — mixing the two spaces in a centroid is mathematically meaningless.
- The cache key includes `task_type` (Redis namespace + Postgres composite PK with `query_hash`) so the same string under two task types stores as two distinct entries and never collides. Constants `TASK_RETRIEVAL_QUERY` / `TASK_RETRIEVAL_DOCUMENT` are exported from `genai/query_embedder.py`.

**Shared row type**: `repository/types.py:ScoredDocument` is the dataclass returned by every embedding-based query (search results, similar-doc results, personalised feed results). Add new query helpers to this shape rather than introducing a parallel hierarchy.

## Frontend Patterns

**Svelte 5 runes only** — use `$state`, `$derived`, `$effect`, `$props`. No `$:` reactive declarations, no `export let`, no stores from `svelte/store` for new code (the runes-based `.svelte.ts` stores are the canonical pattern).

**Shared stores** (`$lib/stores/*.svelte.ts`):
- `lab.svelte.ts` — `getActiveLab / setActiveLab`, `getMyLabs / setMyLabs`, `getLabsVersion / invalidateLabs`. **The sidebar is the single source of truth for the lab list** — it calls `labApi.getMyLabs()` once on mount and whenever `labsVersion` bumps, then publishes via `setMyLabs`. Other pages read with `$derived(getMyLabs())` and must NOT fire their own `getMyLabs` request. To force a refresh (e.g. after create/join), call `invalidateLabs()`.
- `user.svelte.ts` — `getUser`, `setAvatarUrl`, etc.

**Effect IDs, not whole objects**: when an `$effect` depends on a store object that may be swapped for an equivalent same-ID copy (e.g. `activeLab` after `reloadLabs`), derive the ID first and depend on that — Svelte skips `$derived` notifications when the primitive output is strictly equal, avoiding spurious refetches:
```ts
let activeLabId = $derived(getActiveLab()?.id ?? null);
$effect(() => {
    const id = activeLabId;  // primitive — stable across object swaps
    if (id !== null) fetch(id);
});
```

**Route-param changes reuse the same component**: when SvelteKit navigates between `/foo/1` → `/foo/2`, `onMount` does NOT fire again. Data fetches that depend on `data.id` must live inside an `$effect` tracking `data.id`, not inside `onMount`.

**Back-navigation detection**: `document.referrer` is unreliable after client-side nav. Use `afterNavigate` to detect internal arrival:
```ts
let arrivedInternally = $state(false);
afterNavigate((nav) => {
    if (nav.from && nav.type !== 'enter') arrivedInternally = true;
});
```

**API layer** (`$lib/api/*.ts`):
- Axios instance at `$lib/api/request.ts` with a JWT request interceptor and a 401 → `/login` redirect response interceptor.
- Each resource exports a default object of methods: `documentApi.getSimilar`, `labApi.getMyLabs`, etc.
- Errors are surfaced to the user via `showApiErrors(err, fallbackI18nKey)` from `$lib/utils/api-error.ts`.

**i18n**: `svelte-i18n` with message bundles at `$lib/locales/{en,zh-CN}.json`. Use `$_('key.path')` in templates; keep key structure mirroring the backend's `service.<resource>.<outcome>` codes so API error codes resolve directly.

**Routing**: filesystem routes under `frontend/src/routes`. The `(dashboard)` group layout wraps all authenticated pages and injects `AppSidebar`. Nested route groups/params follow SvelteKit conventions.

## Data Model Notes

- **Document soft-delete**: `gorm.Model` supplies `deleted_at`; all dedup/access queries include `deleted_at IS NULL`.
- **Visibility**: `private` | `lab`. Lab-visible docs require `lab_id` set. Switching visibility is owner-only.
- **Enrichment status**: `not_started | pending | processing | done | failed`. Redis is the real-time source (key `doc:enrich:<id>`, TTL 24h); Postgres stores the final value. Only docs with `enrich_status = 'done'` carry a usable embedding.
- **Dedup indexes** (partial unique): `idx_documents_private_user_sha` on (`uploaded_by_user_id`, `content_sha256`) where `visibility='private'`; `idx_documents_lab_sha` on (`lab_id`, `content_sha256`) where `visibility='lab'`. Violations of each are distinguished in service-layer error mapping (`ErrDocumentDuplicate` vs `ErrDocumentDuplicateInLab`).
- **Embeddings**: 768-dim via `gemini-embedding-001`, stored with pgvector HNSW cosine index. Cosine *distance* is `<=>`, cosine *similarity* is `1 - (a <=> b)`.
- **User interactions**: `document_views` (throttled — one row per (user, doc, 15-min window) via `ViewThrottleWindow`; bumps `documents.view_count`) and `document_likes` (toggle-shaped soft-delete, partial unique index on non-deleted; bumps `documents.like_count`). Both repos expose `ListViewHistory` / `ListLikeHistory` for the activity-history UI and as inputs to the personalised recommendation flow.
- **Search history**: `search_histories` table — one row per unique `(user_id, query, lab_id)` upserted on each successful semantic search. Powers the search-page autocomplete and is the third signal source (alongside likes/views) for `RecommendForUser`.
- **`query_embeddings` table**: persistent backstop for the recommender's three-tier cache. Composite primary key `(query_hash, task_type)` because the same string under two Gemini task types is two distinct vectors. Schema is gateway-owned (`model.QueryEmbedding`); the gateway never reads or writes it — svc-recommender is the sole consumer.

## Key Conventions

- **Commit messages**: Conventional commits (`feat:`, `fix:`, `chore:`, `docs:`)
- **Config files**: Each service has `config.yaml` (local, gitignored), `config.docker.yaml` (Docker), and `config.example.yaml` (checked in)
- **Proto changes**: Always run `buf generate` after modifying `.proto` files; CI regenerates automatically
- **Frontend auth**: JWT in localStorage, Axios interceptor adds Authorization header and redirects to login on 401
- **Validation**: go-playground/validator v10 with custom validators for username/password rules; DTO tags use `binding:"oneof=..."` to whitelist enum-like inputs
