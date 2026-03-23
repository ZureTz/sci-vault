# SVC-Gateway

`svc-gateway` is the API gateway service for the `sci-vault` application. It acts as the primary entry point for all client requests, handling user authentication, profile management, avatar storage, and routing to the gRPC-based `svc-recommender`. The gateway is built with [Gin](https://github.com/gin-gonic/gin) for HTTP routing, [GORM](https://gorm.io/) for database access, and [Viper](https://github.com/spf13/viper) for configuration management.

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
  endpoint: "http://localhost:9000"
  access_key: "your-access-key"   # from RustFS console
  secret_key: "your-secret-key"   # from RustFS console
  private_bucket: "sci-vault"
  public_bucket: "public-assets"
  use_ssl: false

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

All routes are prefixed with `/api/v1`.

### Public User Routes (`/api/v1/user`)

| Method | Path                    | Description                               |
| ------ | ----------------------- | ----------------------------------------- |
| `POST` | `/user/send_email_code` | Send email verification code              |
| `POST` | `/user/login`           | Log in (username or email + password)     |
| `POST` | `/user/register`        | Register a new user (requires email code) |
| `POST` | `/user/reset_password`  | Reset password (requires email code)      |

### Protected User Routes (`/api/v1/user`, requires JWT)

| Method | Path                     | Description                         |
| ------ | ------------------------ | ----------------------------------- |
| `POST` | `/user/upload_avatar`    | Upload user avatar (multipart form) |
| `PUT`  | `/user/profile`          | Update user profile                 |
| `GET`  | `/user/avatar/:user_id`  | Get user avatar URL                 |
| `GET`  | `/user/profile/:user_id` | Get user profile                    |

### Protected Auth Routes (`/api/v1/auth`, requires JWT)

| Method | Path         | Description               |
| ------ | ------------ | ------------------------- |
| `GET`  | `/auth/test` | Verify JWT authentication |

### Health Check

| Method | Path             | Description                            |
| ------ | ---------------- | -------------------------------------- |
| `GET`  | `/api/v1/health` | Gateway and recommender liveness check |

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
│   ├── cache/              # Redis client wrapper
│   ├── database/           # PostgreSQL connection (GORM)
│   ├── grpcclient/         # gRPC client for svc-recommender
│   ├── jwt/                # JWT generation and claim extraction
│   ├── logger/             # Structured logger (slog) setup
│   ├── mailer/             # Async email sender (SMTP)
│   ├── password/           # Password hashing utilities
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
