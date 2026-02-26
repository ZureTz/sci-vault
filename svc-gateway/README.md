# SVC Gateway

`svc-gateway` is the API gateway service for the `sci-vault` application. It is built using the [Gin](https://github.com/gin-gonic/gin) framework and uses [Viper](https://github.com/spf13/viper) for configuration management.

## Directory Structure

```text
svc-gateway/
├── internal/
│   ├── config/       # Configuration loading logic (Viper)
│   ├── handler/      # Request handlers
│   ├── middleware/   # Gin middleware (Logger, Recovery)
│   └── router/       # Route definitions
├── config.example.yaml # Configuration template
├── config.yaml         # Local configuration file (ignored by git)
├── main.go             # Entry point
└── go.mod              # Dependency management
```

## Quick Start

### Prerequisites

- Go 1.25+

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
```

## API Endpoints

- `GET /health`: Health check endpoint.
- `GET /api/v1/...`: Reserved for business API routes.

## License

This project is licensed under the [LICENSE](../../LICENSE) file in the root directory.
