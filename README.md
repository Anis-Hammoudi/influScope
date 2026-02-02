
An influencer discovery and search platform built with Go microservices.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Scraper   │────▶│  RabbitMQ   │────▶│   Indexer   │
│  (Writer)   │     │   (Queue)   │     │  (Reader)   │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                                               ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│     API     │────▶│ Elasticsearch│     │  Postgres   │
│  (Gateway)  │     │  (Search)   │     │    (DB)     │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Services

- **Scraper**: Discovers influencer profiles and publishes to RabbitMQ
- **Indexer**: Consumes from RabbitMQ and indexes to Elasticsearch
- **API**: HTTP gateway for querying influencer data

## Tech Stack

- **Go** - Backend services
- **PostgreSQL** - Primary database
- **Elasticsearch** - Search engine
- **RabbitMQ** - Message queue
- **Prometheus** - Metrics collection
- **Docker** - Containerization

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+

### Running with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

### Local Development

```bash
# Install dependencies
go mod download

# Run a specific service
cd scraper && go run main.go
cd indexer && go run main.go
cd api && go run main.go
```

## Service Ports

| Service       | Port  | Description           |
|---------------|-------|-----------------------|
| API           | 8080  | HTTP API              |
| Scraper       | 8081  | Scraper metrics       |
| Indexer       | 8082  | Indexer metrics       |
| PostgreSQL    | 5432  | Database              |
| Elasticsearch | 9200  | Search engine         |
| RabbitMQ      | 5672  | Message queue         |
| RabbitMQ UI   | 15672 | Management UI         |
| Prometheus    | 9090  | Metrics dashboard     |

## Configuration

Copy `.env.example` to `.env` and update the values:

```bash
cp .env .env
```

## License

MIT
# InfluScope

