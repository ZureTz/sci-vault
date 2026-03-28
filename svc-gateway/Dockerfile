FROM bufbuild/buf:1.66.1 AS pb-builder
WORKDIR /workspace
# Copy proto definitions and buf config from the additional context
COPY --from=workspace_root proto/ ./proto/
COPY --from=workspace_root buf.yaml buf.gen.yaml ./
RUN buf generate

FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Bring in the generated Go stubs
COPY --from=pb-builder /workspace/svc-gateway/internal/pb/ ./internal/pb/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gateway main.go

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/gateway /app/gateway

EXPOSE 8080

# config.yaml is expected to be mounted at runtime via a volume bind mount.
# Example: -v ./svc-gateway/config.docker.yaml:/app/config.yaml:ro
CMD ["./gateway", "--config", "/app/config.yaml"]