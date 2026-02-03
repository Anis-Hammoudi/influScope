# InfluScope

A distributed influencer discovery and search platform designed with an event-driven microservices architecture.

## Overview

InfluScope mimics a production-grade ingestion pipeline. It decouples data discovery (scraping) from data indexing using a message broker, ensuring system resilience and scalability. The system allows users to search for influencers based on bio keywords, categories, and usernames via a REST API, while providing real-time metrics on data throughput.

## Architecture

The system follows a writer/reader pattern decoupled by RabbitMQ. A sidecar Prometheus instance scrapes metrics from all microservices.

```mermaid
graph LR
    subgraph "Infrastructure"
        ES[(Elasticsearch)]
        RMQ(RabbitMQ)
        S3[(S3 / MinIO)]
    end

    subgraph "Microservices"
        Scraper
        Indexer
        API
    end

    Scraper -->|1. Upload Avatar| S3
    Scraper -->|2. Publish Metadata| RMQ

    RMQ -->|3. Consume| Indexer
    Indexer -->|4. Index| ES

    API -->|5. Search| ES

    Scraper -.->|Metrics| Prom(Prometheus)
    Indexer -.->|Metrics| Prom
```

#Here is the final, formatted Markdown block for your README. Copy and paste this directly after your Mermaid diagram.

Markdown
## Services

* **Scraper:** Generates smart influencer profiles, **offloads avatars to S3 (MinIO)**, and publishes metadata events to RabbitMQ. Implements "self-healing" logic to initialize storage buckets automatically.
* **Indexer:** Consumes messages from the queue and performs bulk indexing operations into Elasticsearch.
* **API:** A lightweight HTTP gateway that translates user search queries into Elasticsearch DSL.
* **MinIO (S3):** Provides S3-compatible object storage for static assets (images), mimicking a production AWS environment.
* **Prometheus:** Aggregates metrics from the Scraper and Indexer to visualize system throughput and bottlenecks.

## Tech Stack

* **Language:** Go (Golang) 1.25
* **Messaging:** RabbitMQ (utilizing `upfluence/amqp` for connection pooling)
* **Search Engine:** Elasticsearch 7.17
* **Object Storage:** AWS S3 / MinIO (AWS SDK v2)
* **Observability:** Prometheus
* **Orchestration:** Kubernetes (StatefulSets & Deployments)
* **CI/CD:** GitHub Actions
* **Automation:** GNU Make

## Design Decisions

![AWS SAA](https://img.shields.io/badge/AWS-Solutions_Architect_Associate-FF9900?style=for-the-badge&logo=amazonaws&logoColor=white)

* **Hybrid Storage Pattern (AWS SAA):** Following AWS Well-Architected Framework best practices, I decoupled the storage of static assets from the metadata database.
    * **Structured Data (Bio, Follower Count):** Stored in **Elasticsearch** for low-latency indexing and search.
    * **Unstructured Data (Profile Images):** Offloaded to **S3** to reduce database bloat and leverage cost-effective tiered storage.
* **Self-Healing Infrastructure:** The Scraper service implements initialization logic to detect if the S3 bucket is missing and create it automatically, eliminating race conditions during startup.
* **Decoupling:** RabbitMQ separates scraping from indexing. This prevents data loss if the search engine is under heavy load; messages simply accumulate in the queue.
* **Observability:** Custom Prometheus instrumentation tracks `influencers_discovered_total` vs `influencers_indexed_total`, allowing for immediate detection of pipeline latency.

## Quick Start

### Prerequisites

* Docker Desktop
* Go 1.25+ (for local development)
* Make (Optional)

### Running the Stack

1.  **Clone the repository**
    ```bash
    git clone [https://github.com/hammo/influScope.git](https://github.com/hammo/influScope.git)
    cd influScope
    ```

2.  **Start Services (Automated)**
    ```bash
    make up
    ```
    *(Or manually: `docker-compose up --build`)*

3.  **Run Tests**
    ```bash
    make test
    ```

4.  **Verify Status**
    * **Search API:** `http://localhost:8080/search?q=tech`
    * **MinIO Console (S3):** `http://localhost:9001` (User: `admin` / Pass: `password`)
    * **RabbitMQ Dashboard:** `http://localhost:15672` (guest/guest)
    * **Prometheus:** `http://localhost:9090`

## Performance Benchmarks

To ensure scalability, the system was load-tested using **k6**.

* **Scenario:** 50 concurrent users (Virtual Users) performing continuous search queries.
* **Result:** 100% Success Rate (0 Failed Requests).
* **Latency:** Average response time of **6.04ms** (p95: 9.18ms).
* **Throughput:** ~40 req/s on local hardware.
### Observability

To verify the pipeline health, open Prometheus (`http://localhost:9090`) and query the custom metrics:

* `influencers_discovered_total` (Scraper output)
* `influencers_indexed_total` (Indexer throughput)

*If "Discovered" > "Indexed", the queue is building backpressure.*

## Kubernetes Deployment

This project includes production-ready Kubernetes manifests in the `k8s/` directory.

```bash
# Deploy to local cluster
kubectl apply -f k8s/00-config.yaml
kubectl apply -f k8s/01-infrastructure.yaml
kubectl apply -f k8s/02-apps.yaml
```
* StatefulSets: Used for RabbitMQ, Elasticsearch, and MinIO to ensure stable network IDs and persistent storage.

* Deployments: Used for stateless apps (API, Scraper, Indexer) to allow for easy scaling.

* ConfigMaps: Decouples configuration (URLs, Tuning params) from the application code.

## CI/CD Pipeline

This project uses **GitHub Actions** to enforce code quality. On every push to `main`:
1.  Dependencies are verified (`go mod download`).
2.  All microservices (Scraper, Indexer, API) are built in parallel to ensure no breaking changes were introduced.
## License

MIT
