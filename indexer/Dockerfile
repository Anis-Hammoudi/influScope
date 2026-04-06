FROM golang:1.25-alpine

# Install git (required by Go to fetch dependencies)
RUN apk add --no-cache git

# Set the working directory to the root of the app
WORKDIR /app

# 1. Copy your entire project
COPY . .

# 2. Move specifically into the indexer directory
WORKDIR /app/indexer

# 3. Generate go.sum and download dependencies automatically
RUN go mod tidy

# 4. Build the executable and place it in the root /app folder
RUN go build -o /app/indexer-app ./cmd/indexer/main.go

# 5. Move back to the root app folder and run it
WORKDIR /app
CMD ["./indexer-app"]