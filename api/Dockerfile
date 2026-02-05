# Stage 1: Builder
FROM golang:1.24-alpine AS builder
WORKDIR /app

# 1. Cache Dependencies
COPY pkg ./pkg
COPY api/go.mod api/go.sum ./api/
WORKDIR /app/api
RUN go mod download

# 2. Build
COPY api .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api-bin .

# Stage 2: Runner
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api/api-bin .
EXPOSE 8080
CMD ["./api-bin"]