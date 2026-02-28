# SVC-Gateway

`svc-gateway` is the API gateway service for the `sci-vault` application. It acts as the primary entry point for client requests, routing them to the appropriate backend microservices (such as the gRPC-based `svc-recommender`). The gateway is built with performance and reliability in mind, utilizing the [Gin](https://github.com/gin-gonic/gin) web framework and [Viper](https://github.com/spf13/viper) for robust configuration management.

## Getting Started

Follow the instructions below to set up and run the `svc-gateway` service locally.

### Prerequisites

- [Go](https://go.dev/doc/install) 1.25+

### Install Dependencies

Navigate to the project directory and download the required Go modules:

```bash
go mod tidy
```

### Configuration

The service loads configurations prioritized in the following order: environment variables, the local `config.yaml` file, and finally fallback defaults (e.g., Host `0.0.0.0`, Port `8080`).

Copy the example configuration file to create your local `config.yaml` file:

```bash
cp config.example.yaml config.yaml
```

Update the `config.yaml` file with your specific environment details, such as target gRPC ports, host bindings, and API configurations.

### Run the Service

Start the API gateway directly using Go:

```bash
go run .
```

## API Endpoints

- `GET /health`: Health check endpoint.
- `GET /api/v1/...`: Reserved for business API routes.

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

## Roadmap: Docker Integration

Similar to other services in this architecture, Docker support will be added in the future to simplify containerized deployments and orchestrate the entire microservices stack locally.

## License

This project is licensed under the [LICENSE](../LICENSE) file in the root directory.
