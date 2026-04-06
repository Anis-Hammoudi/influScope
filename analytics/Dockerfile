# Stage 1: Builder
FROM golang:1.25-alpine AS builder

# Install git (required by Go to fetch dependencies)
RUN apk add --no-cache git

WORKDIR /app

# 1. Copy the entire monorepo to preserve relative local paths (../pkg, ../gen)
COPY . .

# 2. Move specifically into the analytics directory
WORKDIR /app/analytics

# 3. Dynamically generate go.sum and download dependencies
RUN go mod tidy

# 4. Build the binary from the NEW cmd/analytics entry point
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/analytics-bin ./cmd/analytics/main.go


# Stage 2: Runner
FROM --platform=linux/amd64 alpine:latest

# Optional: Add CA certificates if your analytics service ever needs to make external HTTPS calls
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 5. Copy the compiled binary from the builder stage
COPY --from=builder /app/analytics-bin .

RUN chmod +x analytics-bin

CMD ["./analytics-bin"]