# Stage 1: Builder
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy necessary source files
COPY pkg ./pkg
COPY gen ./gen
COPY analytics/go.mod analytics/go.sum ./analytics/

WORKDIR /app/analytics
RUN go mod download

# Build the binary
COPY analytics .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o analytics-bin .

# Stage 2: Runner
FROM --platform=linux/amd64 alpine:latest

WORKDIR /root/
COPY --from=builder /app/analytics/analytics-bin .
RUN chmod +x analytics-bin
CMD ["./analytics-bin"]