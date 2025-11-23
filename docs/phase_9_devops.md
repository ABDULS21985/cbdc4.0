# Phase 9: DevOps, CI/CD, Environments

## 1. Environment Strategy

### 1.1 Environments
*   **Local Dev**:
    *   **DLT**: Ganache (Instant) or Local Fabric (Docker Compose).
    *   **Services**: Go binaries running locally.
    *   **Frontend**: `npm run dev`.
*   **Testnet (Shared Dev)**:
    *   **DLT**: Fabric Network (3 Orgs, 5 Orderers) on K8s.
    *   **Services**: Deployed via Helm Charts.
    *   **Data**: Reset nightly.
*   **UAT (Staging)**:
    *   **DLT**: Mirror of Prod topology.
    *   **Data**: Persistent test data.
*   **Production**:
    *   **DLT**: Geo-distributed nodes.
    *   **Security**: HSMs enforced.

## 2. CI/CD Pipelines

### 2.1 Workflows
*   **`ci-backend`**:
    *   Trigger: PR to `main`.
    *   Steps: `go test ./...`, `go vet`, `staticcheck`.
    *   Output: Docker Images pushed to Registry (tagged `sha-xyz`).
*   **`ci-frontend`**:
    *   Trigger: PR to `main`.
    *   Steps: `npm run lint`, `npm run test`, `npm run build`.
*   **`cd-deploy`**:
    *   Trigger: Tag release `v*`.
    *   Steps: Update Helm Chart versions, sync ArgoCD.

## 3. Observability & SRE

### 3.1 Stack
*   **Metrics**: Prometheus (scraping `/metrics` endpoints).
*   **Visualization**: Grafana (Dashboards for TPS, Latency, Error Rates).
*   **Tracing**: OpenTelemetry (Jaeger/Tempo).
*   **Logs**: ELK Stack or Loki.

### 3.2 Key Metrics (SLIs)
*   **Ledger Latency**: Time from `SubmitTransaction` to `BlockCommitted`.
*   **Payment Success Rate**: % of payments succeeding without error.
*   **Offline Reconciliation Lag**: Time to process offline batches.
