# Stage 1: Builder
FROM golang:1.25-alpine AS builder
WORKDIR /app
# 1. Copy Dependency Files
COPY pkg ./pkg
COPY gen ./gen
COPY indexer/go.mod indexer/go.sum ./indexer/

# 2. Download Dependencies
WORKDIR /app/indexer
RUN go mod download

# 3. Copy Source Code & Build
COPY indexer .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o indexer-bin .

# Stage 2: Runner (Tiny Alpine Image)
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/indexer/indexer-bin .

CMD ["./indexer-bin"]