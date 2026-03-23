FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gateway main.go

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/gateway /app/gateway

EXPOSE 8080

# config.yaml is expected to be mounted at runtime via a volume bind mount.
# Example: -v ./svc-gateway/config.docker.yaml:/app/config.yaml:ro
CMD ["./gateway", "--config", "/app/config.yaml"]