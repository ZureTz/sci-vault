# SVC-Gateway

`svc-gateway` is the API gateway service for the `sci-vault` application. It acts as the primary entry point for all client requests and exposes the platform's REST API: user authentication and profiles, document upload/CRUD with async AI enrichment, lab (research group) management, semantic search, personalised recommendations and collaborator suggestions, dashboard statistics, and LLM-powered translation. Recommendation, search, enrichment, and translation requests are delegated over gRPC to `svc-recommender`. The gateway is built with [Gin](https://github.com/gin-gonic/gin) for HTTP routing, [GORM](https://gorm.io/) for database access, and [Viper](https://github.com/spf13/viper) for configuration management.

## Getting Started

Follow the instructions below to set up and run the `svc-gateway` service locally.

### Prerequisites

- [Go](https://go.dev/doc/install) 1.26+
- A running PostgreSQL instance
- A running Redis instance
- A running [RustFS](https://rustfs.com/) (or any S3-compatible) object storage instance
- A running `svc-recommender` instance (gRPC)

### Install Dependencies

Navigate to the project directory and download the required Go modules:

```bash
go mod tidy
```

### Configuration

The service requires a `config.yaml` file in the project root. Copy the example and fill in your values:

```bash
cp config.example.yaml config.yaml
```

The full configuration reference is:

```yaml
host: "0.0.0.0"
port: "8080"
recommender_addr: "localhost:50051"

log:
  level: "info"   # debug | info | warn | error
  format: "json"  # json | text

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

database:
  host: "localhost"
  port: 5432
  user: "sci_vault"
  password: "sci_vault_secret"
  db_name: "sci_vault"
  ssl_mode: "disable"
  timezone: "UTC"

storage:
  endpoint: "http://0.0.0.0:9000"
  presign_endpoint: "http://127.0.0.1:9000"  # must match the Host header the Vite proxy sends to RustFS
  access_key: "your-access-key"   # from RustFS console
  secret_key: "your-secret-key"   # from RustFS console
  private_bucket: "sci-vault"
  public_bucket: "public-assets"
  use_ssl: false
  public_proxy_path: "/assets"
  private_proxy_path: "/private"

mailer:
  host: "smtp.example.com"
  port: 587
  username: "your-email@example.com"
  password: "your-password"

jwt:
  secret: "your-256-bit-secret"
  timeout: 24  # hours
```

#### RustFS (Object Storage)

The service uses RustFS as its S3-compatible object storage backend. After starting the RustFS container via `docker compose up -d`, generate access credentials via the RustFS Console at `http://localhost:9001`, then configure them in `config.yaml` under the `storage` section.

### Run the Service

```bash
go run .
```

Database tables are automatically migrated on startup.

## API Endpoints

All routes are prefixed with `/api/v1`. Every group below except the public user routes and
the health check requires a valid JWT (`Authorization: Bearer <token>`).

### Health Check

| Method | Path             | Description                            |
| ------ | ---------------- | -------------------------------------- |
| `GET`  | `/api/v1/health` | Gateway and recommender liveness check |

### Public User Routes (`/api/v1/user`)

| Method | Path                    | Description                               |
| ------ | ----------------------- | ----------------------------------------- |
| `POST` | `/user/send_email_code` | Send email verification code (rate-limited) |
| `POST` | `/user/login`           | Log in (username or email + password)     |
| `POST` | `/user/register`        | Register a new user (requires email code) |
| `POST` | `/user/reset_password`  | Reset password (requires email code)      |

### Protected User Routes (`/api/v1/user`)

| Method | Path                     | Description                         |
| ------ | ------------------------ | ----------------------------------- |
| `POST` | `/user/upload_avatar`    | Upload user avatar (multipart form) |
| `POST` | `/user/change_password`  | Change the current user's password  |
| `PUT`  | `/user/profile`          | Update user profile                 |
| `GET`  | `/user/avatar/:user_id`  | Get user avatar URL                 |
| `GET`  | `/user/profile/:user_id` | Get user profile                    |

### Auth Routes (`/api/v1/auth`)

| Method | Path         | Description               |
| ------ | ------------ | ------------------------- |
| `GET`  | `/auth/test` | Verify JWT authentication |

### Document Routes (`/api/v1/docs`)

| Method   | Path                          | Description                                       |
| -------- | ----------------------------- | ------------------------------------------------- |
| `POST`   | `/docs/upload`                | Upload a single document (triggers enrichment)    |
| `POST`   | `/docs/upload/batch`          | Batch document upload                             |
| `GET`    | `/docs/mine`                  | List the caller's documents                       |
| `GET`    | `/docs/pending`               | List documents pending enrichment                 |
| `POST`   | `/docs/visibility/batch`      | Batch update document visibility                  |
| `GET`    | `/docs/:doc_id`               | Get a document's metadata + presigned download    |
| `GET`    | `/docs/:doc_id/enrich_status` | Poll enrichment status                            |
| `POST`   | `/docs/:doc_id/restart_enrichment` | Re-trigger enrichment                        |
| `PATCH`  | `/docs/:doc_id/visibility`    | Change visibility (owner only)                    |
| `PATCH`  | `/docs/:doc_id`               | Update document metadata                          |
| `DELETE` | `/docs/:doc_id`               | Soft-delete a document                            |
| `GET`    | `/docs/:doc_id/similar`       | Recommend similar documents                       |
| `POST`   | `/docs/:doc_id/like`          | Like a document                                   |
| `DELETE` | `/docs/:doc_id/like`          | Remove a like                                     |

### Interaction History (`/api/v1/mine/history`)

| Method | Path                  | Description                |
| ------ | --------------------- | -------------------------- |
| `GET`  | `/mine/history/views` | List the caller's views    |
| `GET`  | `/mine/history/likes` | List the caller's likes    |

### Recommendations (`/api/v1/mine`)

| Method | Path                   | Description                                  |
| ------ | ---------------------- | -------------------------------------------- |
| `GET`  | `/mine/recommendations` | Personalised document feed                  |
| `GET`  | `/mine/collaborators`   | Suggested collaborators (lab-mate ranking)  |

### Search (`/api/v1/search`)

| Method   | Path              | Description                          |
| -------- | ----------------- | ------------------------------------ |
| `GET`    | `/search`         | Semantic search (with keyword fallback) |
| `GET`    | `/search/history` | List the caller's search history     |
| `DELETE` | `/search/history` | Clear the caller's search history    |

### Dashboard Stats (`/api/v1/stats`)

| Method | Path                          | Description                        |
| ------ | ----------------------------- | ---------------------------------- |
| `GET`  | `/stats/mine/dashboard`       | The caller's dashboard statistics  |
| `GET`  | `/stats/labs/:lab_id/dashboard` | A lab's dashboard statistics     |

### Translation (`/api/v1/translate`)

| Method | Path                 | Description                          |
| ------ | -------------------- | ------------------------------------ |
| `POST` | `/translate/summary` | Translate a document summary (LLM)   |

### Labs (`/api/v1/labs`)

| Method   | Path                              | Description                                        |
| -------- | --------------------------------- | -------------------------------------------------- |
| `GET`    | `/labs`                           | List the caller's labs                             |
| `POST`   | `/labs`                           | Create a lab                                       |
| `POST`   | `/labs/join`                      | Join a lab by invitation code                      |
| `GET`    | `/labs/:lab_id`                   | Get lab details                                    |
| `GET`    | `/labs/:lab_id/members`           | List lab members                                   |
| `POST`   | `/labs/:lab_id/leave-request`     | Request to leave (sends email confirmation code)   |
| `DELETE` | `/labs/:lab_id/members/me`        | Confirm leave (with email code)                    |
| `GET`    | `/labs/:lab_id/documents`         | List lab-visible documents                         |
| `PATCH`  | `/labs/:lab_id`                   | Update lab info (owner only)                        |
| `DELETE` | `/labs/:lab_id/members/:user_id`  | Kick a member (owner only)                          |
| `POST`   | `/labs/:lab_id/transfer`          | Transfer ownership (owner only)                     |
| `POST`   | `/labs/:lab_id/delete-request`    | Request lab deletion (sends email code, owner only) |
| `DELETE` | `/labs/:lab_id`                   | Confirm deletion (lab name + email code, owner only) |
| `POST`   | `/labs/:lab_id/invite-code/reset` | Reset the invitation code (owner only)              |

## Directory Structure

```text
svc-gateway/
├── app/
│   └── app.go              # Application bootstrap and dependency injection
├── internal/
│   ├── config/             # Configuration structs and Viper loading
│   ├── dto/                # Request/response data transfer objects
│   ├── handler/            # HTTP request handlers
│   ├── middleware/         # Gin middleware (JWT authentication)
│   ├── model/              # GORM data models
│   ├── pb/                 # Protobuf-generated gRPC stubs
│   ├── repo/               # Repository layer (database access)
│   ├── router/             # Route definitions and registration
│   └── service/            # Business logic layer
├── pkg/
│   ├── app_error/          # Sentinel error definitions
│   ├── cache/              # Redis client wrapper
│   ├── codegen/            # Code-generation utilities
│   ├── database/           # PostgreSQL connection (GORM)
│   ├── grpc_client/        # gRPC client for svc-recommender
│   ├── jwt/                # JWT generation and claim extraction
│   ├── logger/             # Structured logger (slog) setup
│   ├── mailer/             # Async email sender (SMTP)
│   ├── password/           # Password hashing utilities
│   ├── reenrich/           # Document re-enrichment triggers
│   ├── storage/            # S3-compatible storage client (RustFS)
│   ├── utils/              # HTTP response helpers
│   └── validator/          # Custom Gin validators (username, password)
├── config.example.yaml     # Configuration template
├── config.yaml             # Local configuration file (ignored by git)
├── main.go                 # Entry point
└── go.mod                  # Dependency management
```

## Docker Deployment

This service can be run locally using Docker and Docker Compose. 

Before building the Docker image, ensure you have created a `config.docker.yaml` file so the build process can copy it into the container:

```bash
cp config.example.yaml config.docker.yaml
```

Update your `config.docker.yaml` to point to the correct Docker service hostnames (e.g., `postgres` for database host, `redis` for Redis host, and `recommender:50051` for the recommender address).

To build and start the entire stack including the gateway:

```bash
cd ..
docker compose up -d --build gateway
```

## License

This project is licensed under the [LICENSE](../LICENSE) file in the root directory.
