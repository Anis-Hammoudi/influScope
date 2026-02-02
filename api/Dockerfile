FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY pkg ./pkg
COPY api ./api

WORKDIR /app/api
RUN go mod download
RUN go build -o api-bin main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api/api-bin .
CMD ["./api-bin"]