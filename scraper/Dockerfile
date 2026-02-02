# Start from Golang image
FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY pkg ./pkg

COPY scraper ./scraper

WORKDIR /app/scraper


RUN go mod download

RUN go build -o scraper-bin main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/scraper/scraper-bin .

CMD ["./scraper-bin"]