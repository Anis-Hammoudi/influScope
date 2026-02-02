FROM golang:1.25-alpine AS builder

WORKDIR /app

# 1. Copy Shared Pkg
COPY pkg ./pkg

# 2. Copy Indexer
COPY indexer ./indexer

# 3. Build
WORKDIR /app/indexer
RUN go mod download
RUN go build -o indexer-bin main.go

# --- Run Stage ---
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/indexer/indexer-bin .

CMD ["./indexer-bin"]