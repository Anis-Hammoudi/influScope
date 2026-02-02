# InfluScope

A distributed influencer discovery and search platform designed with an event-driven microservices architecture.

![CI Status](https://github.com/hammo/influScope/actions/workflows/ci.yml/badge.svg)

## Overview

InfluScope mimics a production-grade ingestion pipeline. It decouples data discovery (scraping) from data indexing using a message broker, ensuring system resilience and scalability. The system allows users to search for influencers based on bio keywords, categories, and usernames via a REST API, while providing real-time metrics on data throughput.

## Architecture

The system follows a writer/reader pattern decoupled by RabbitMQ. A sidecar Prometheus instance scrapes metrics from all microservices.
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Scraper   │─────▶│  RabbitMQ   │─────▶│   Indexer   │
│  (Service)  │      │  (Fanout)   │      │  (Service)  │
└─────────────┘      └─────────────┘      └──────┬──────┘
                                                  │
                                                  ▼
┌─────────────┐                          ┌─────────────┐
│     API     │─────────────────────────▶│Elasticsearch│
│  (Gateway)  │                          │   (NoSQL)   │
└─────────────┘                          └─────────────┘
```


## Services

* **Scraper:** Generates influencer profiles and publishes events to the `influencer-events` exchange. Exposes ingestion metrics on `:8081/metrics`.
* **Indexer:** Consumes messages from the queue and performs bulk indexing operations into Elasticsearch. Exposes indexing counters and error rates on `:8082/metrics`.
* **API:** A lightweight HTTP gateway that translates user search queries into Elasticsearch DSL.
* **Prometheus:** Aggregates metrics from the Scraper and Indexer to visualize system throughput and bottlenecks.

## Tech Stack

* **Language:** Go (Golang) 1.25
* **Messaging:** RabbitMQ (utilizing `upfluence/amqp` for connection pooling)
* **Search Engine:** Elasticsearch 7.17
* **Observability:** Prometheus
* **CI/CD:** GitHub Actions (Automated build & test pipelines)
* **Infrastructure:** Docker & Docker Compose

## Design Decisions

* **Decoupling:** RabbitMQ is used to separate the scraping logic from the indexing logic. This prevents data loss if the search engine is under heavy load; messages simply accumulate in the queue.
* **Observability:** Custom Prometheus instrumentation tracks `influencers_discovered_total` vs `influencers_indexed_total`, allowing for immediate detection of pipeline latency or dropped messages.
* **Resilience:** Services implement application-level retry logic with exponential backoff to handle infrastructure startup latency and temporary network partitions.
* **Connection Pooling:** Utilizes the `upfluence/amqp` library to manage efficient channel reuse and graceful recovery.

## Quick Start

### Prerequisites

* Docker Desktop
* Go 1.25+ (for local development)

### Running the Stack

1.  **Clone the repository**
    ```bash
    git clone [https://github.com/hammo/influScope.git](https://github.com/hammo/influScope.git)
    cd influScope
    ```

2.  **Start services**
    ```bash
    docker-compose up --build
    ```

3.  **Verify Status**
    * **Search API:** `http://localhost:8080/search?q=tech`
    * **RabbitMQ Dashboard:** `http://localhost:15672` (guest/guest)
    * **Prometheus Dashboard:** `http://localhost:9090`

### Observability

To verify the pipeline health, open Prometheus (`http://localhost:9090`) and query the custom metrics:

* `influencers_discovered_total` (Scraper output)
* `influencers_indexed_total` (Indexer throughput)

*If "Discovered" > "Indexed", the queue is building backpressure.*

## CI/CD Pipeline

This project uses **GitHub Actions** to enforce code quality. On every push to `main`:
1.  Dependencies are verified (`go mod download`).
2.  All microservices (Scraper, Indexer, API) are built in parallel to ensure no breaking changes were introduced.
## License

MIT