# SVC Gateway

`svc-gateway` is the API gateway service for the `sci-vault` application. It is built using the [Gin](https://github.com/gin-gonic/gin) framework and uses [Viper](https://github.com/spf13/viper) for configuration management.

## Directory Structure

```text
svc-gateway/
├── internal/
│   ├── config/       # Configuration loading logic (Viper)
│   ├── handler/      # Request handlers
│   ├── middleware/   # Gin middleware (Logger, Recovery)
│   ├── pb/           # Protobuf generated code
│   │   └── health/   # Generated from proto/health/health.proto
│   ├── producer/     # Kafka producer (protobuf serialization)
│   └── router/       # Route definitions
├── config.example.yaml # Configuration template
├── config.yaml         # Local configuration file (ignored by git)
├── main.go             # Entry point
└── go.mod              # Dependency management
```

## Quick Start

### Prerequisites

- Go 1.25+
- [protoc](https://grpc.io/docs/protoc-installation/) (Protocol Buffers compiler)
- `protoc-gen-go` plugin:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  ```
- A running Kafka broker (default: `localhost:9092`)

### Compile Protobuf

Run this from the **repository root** whenever `.proto` files change:

```bash
# from sci-vault/
protoc \
  --proto_path=proto \
  --go_out=svc-gateway/internal/pb \
  --go_opt=paths=source_relative \
  proto/health/health.proto
```

### Install Dependencies

```bash
go mod tidy
```

### Run the Service

```bash
# Run directly
go run .

# Or use custom configuration
PORT=9090 go run .
```

## Configuration Details

The service attempts to load configuration in the following order of priority:

1.  **Environment Variables**: Highest priority (e.g., `HOST`, `PORT`).
2.  **Configuration File**: Reads `config.yaml` in the current directory.
3.  **Defaults**: HOST defaults to `0.0.0.0`, PORT defaults to `8080`.

### Using Local Configuration

1.  Copy `config.example.yaml` to `config.yaml`.
2.  Modify the contents of `config.yaml` as needed.

```yaml
host: "0.0.0.0"
port: "8080"

kafka:
  brokers:
    - "localhost:9092"
  topic: "health-events"
```

## API Endpoints

- `GET /health`: Health check endpoint (plain JSON).
- `GET /health-protobuf`: Health check endpoint that publishes a `HealthEvent` protobuf message to Kafka and returns a JSON acknowledgement.
- `GET /api/v1/...`: Reserved for business API routes.

## License

This project is licensed under the [LICENSE](../../LICENSE) file in the root directory.
