# CBDC Platform Deployment Guide

## Prerequisites
- Docker & Docker Compose
- Go 1.21+
- Node.js 18+
- Make

## Local Development (Docker Compose)

1. **Start Infrastructure**
   ```bash
   make up
   ```
   This starts the Hyperledger Fabric network (Orderers, Peers, CA).

2. **Build & Run Backend Services**
   ```bash
   make build-go
   # Run services individually or via a separate compose file (TODO)
   ```

3. **Build & Run Frontends**
   ```bash
   make install-frontend
   make build-frontend
   cd frontend/citizen-wallet-app && npm run dev
   ```

## Production Deployment (Kubernetes)

1. **Build Docker Images**
   ```bash
   docker build -f backend/services/Dockerfile --build-arg SERVICE_PATH=auth-service -t cbdc/auth-service:v1 .
   docker build -f backend/services/Dockerfile --build-arg SERVICE_PATH=wallet-service -t cbdc/wallet-service:v1 .
   # ... repeat for all services
   ```

2. **Deploy to K8s**
   - Apply Fabric K8s manifests (using Operator or Helm).
   - Apply Service manifests (Deployment, Service, Ingress).
   - Configure Secrets (DB creds, Fabric MSP keys).

## Monitoring
- Access Grafana at `http://localhost:3000` (Default: admin/admin).
- Prometheus metrics available at `:9090`.
