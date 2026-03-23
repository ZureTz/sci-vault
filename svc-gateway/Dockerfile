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
COPY --from=builder /app/config.docker.yaml /app/config.docker.yaml

EXPOSE 8080

CMD ["./gateway", "--config", "config.docker.yaml"]