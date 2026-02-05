# InfluScope Cloud Architecture & Deployment

This document details the cloud infrastructure and deployment strategy for the InfluScope platform. The application is designed to run on a production-grade Kubernetes cluster hosted on AWS (Amazon EKS), utilizing cloud-native patterns for storage, networking, and orchestration.

## Architecture Overview

The system is deployed on AWS Elastic Kubernetes Service (EKS) to simulate a real-world high-availability environment.

### Infrastructure Components
* **Cloud Provider:** AWS (Region: us-east-1)
* **Orchestrator:** Amazon EKS (Managed Kubernetes Control Plane)
* **Compute:** Managed Node Group consisting of 2x `t3.medium` instances (Amazon Linux 2).
* **Storage:** AWS EBS (Elastic Block Store) GP3 volumes, provisioned dynamically via the EBS CSI Driver.
* **Networking:** AWS Classic Load Balancer (CLB) exposing the API Gateway to the public internet.
* **Container Registry:** GitHub Container Registry (GHCR) for secure, automated image storage.

### Data Flow
1. **Public Traffic:** Enters via the AWS Load Balancer -> Kubernetes Service (NodePort) -> API Pod.
2. **Internal Communication:**
    * **Async:** Scraper publishes metadata to RabbitMQ (StatefulSet).
    * **Sync:** Indexer calls Analytics Service via gRPC (ClusterIP).
    * **Storage:** Scraper uploads avatars to MinIO (S3-compatible object storage backed by EBS).
3. **Persistence:** Elasticsearch and RabbitMQ persist data to AWS EBS volumes to ensure data survival during pod restarts.

## Prerequisites

To manage this deployment, the following tools are required:
* **AWS CLI:** For credential management (`aws configure`).
* **eksctl:** For provisioning the cluster infrastructure.
* **kubectl:** For deploying application manifests.

## Configuration

### Cluster Configuration
The cluster is provisioned using `eksctl` with a dedicated configuration file (`cluster-config.yaml`).

**Key Settings:**
* **Node Group:** `workers` (2 nodes).
* **IAM Policies:** `ebs: true` (Allows nodes to mount hard drives).
* **Addons:** `aws-ebs-csi-driver` (Enables persistent storage for databases).

### CI/CD Pipeline
Deployment artifacts are built automatically via GitHub Actions.
1. **Trigger:** Push to `main` branch.
2. **Build:** Compiles Go binaries for `linux/amd64`.
3. **Test:** Runs unit tests for all microservices.
4. **Publish:** Pushes optimized Docker images to `ghcr.io/anis-hammoudi/influscope`.

## Deployment Instructions

### 1. Provision Infrastructure
Create the EKS cluster. This process takes approximately 15-20 minutes as AWS provisions the VPC, Control Plane, and EC2 instances.

```bash
eksctl create cluster -f cluster-config.yaml
```

### 2. Apply Configurations
Apply the ConfigMap containing environment variables and credentials.

```bash
kubectl apply -f k8s/00-config.yaml
```

### 3. Deploy Stateful Services
Deploy the database layer (Elasticsearch, RabbitMQ, MinIO). Wait 60 seconds after this step to allow AWS EBS volumes to attach to the nodes.

```bash
kubectl apply -f k8s/01-infrastructure.yaml
```

### 4. Deploy Microservices
Deploy the stateless Go applications (API, Scraper, Indexer, Analytics).

```bash
kubectl apply -f k8s/02-apps.yaml
```

### 5. Verification
Retrieve the external access URL for the API Gateway.

```bash
kubectl get svc api-service
```

Test the endpoint using curl or a browser:

```
http://a6a6579fe596e4a88ab127b0eb528b04-954164237.us-east-1.elb.amazonaws.com:8080/search?q=tech
```


## Design Decisions

**Why Kubernetes (EKS)?** To align with the target architecture of modern tech companies like Upfluence. Using EKS demonstrates the ability to manage managed control planes and worker nodes rather than simple Docker containers.

**Why StatefulSets?** RabbitMQ and Elasticsearch require stable network identities and persistent storage. Deploying them as StatefulSets ensures that if a pod crashes, it reconnects to the same EBS volume, preventing data loss.

**Why gRPC?** The Indexer and Analytics services communicate via gRPC to ensure low-latency, strictly typed communication for critical data enrichment paths.

## Teardown
To stop billing and remove all AWS resources (EC2, Load Balancers, Volumes):

```bash
eksctl delete cluster -f cluster-config.yaml
```
