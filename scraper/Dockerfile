FROM golang:1.25-alpine

# Set the working directory to the root of the app
WORKDIR /app

# 1. Copy your entire project (this includes /scraper, /pkg, /go.work, etc.)
COPY . .

# 2. Move specifically into the scraper directory
WORKDIR /app/scraper

# 3. Generate go.sum and download dependencies automatically
RUN go mod tidy

# 4. Build the executable and place it in the root /app folder
RUN go build -o /app/scraper-app ./cmd/scrape/main.go

# 5. Move back to the root app folder and run it
WORKDIR /app
CMD ["./scraper-app"]