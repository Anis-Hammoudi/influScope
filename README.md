# InfluScope

A distributed influencer discovery and search platform designed with an event-driven microservices architecture.

## Overview

InfluScope mimics a production-grade ingestion pipeline. It decouples data discovery (scraping) from data indexing using a message broker, ensuring system resilience and scalability. The system allows users to search for influencers based on bio keywords, categories, and usernames via a REST API.

## Architecture

The system follows a writer/reader pattern decoupled by RabbitMQ. The Indexer consumes events and stores them directly in the search engine.

```
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

* **Scraper:** Generates influencer profiles and publishes events to the `influencer-events` exchange. Implements exponential backoff for connection reliability.
* **Indexer:** Consumes messages from the queue and performs bulk indexing operations into Elasticsearch.
* **API:** A lightweight HTTP gateway that translates user search queries into Elasticsearch DSL.

## Tech Stack

* **Language:** Go (Golang) 1.25
* **Messaging:** RabbitMQ (utilizing `upfluence/amqp` for connection pooling)
* **Search Engine:** Elasticsearch 7.17
* **Infrastructure:** Docker & Docker Compose

## Design Decisions

* **Decoupling:** RabbitMQ is used to separate the scraping logic from the indexing logic. This prevents data loss if the search engine is under heavy load; messages simply accumulate in the queue.
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
    * RabbitMQ Dashboard: `http://localhost:15672` (guest/guest)
    * Elasticsearch: `http://localhost:9200`

### API Usage

Once the stack is running and the scraper has generated data, you can search via the API:

```bash
# Search for tech influencers
curl "http://localhost:8080/search?q=tech"