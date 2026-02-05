# Stage 1: Builder
FROM golang:1.24-alpine AS builder
WORKDIR /app
# 1. Copy Dependency Files
COPY pkg ./pkg
COPY scraper/go.mod scraper/go.sum ./scraper/

# 2. Download Dependencies
WORKDIR /app/scraper
RUN go mod download

# 3. Copy Source Code & Build
COPY scraper .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o scraper-bin .

# Stage 2: Runner
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/scraper/scraper-bin .

CMD ["./scraper-bin"]