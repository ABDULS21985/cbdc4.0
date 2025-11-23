# Phase 10: Implementation Roadmap & Prioritised Backlog

## 1. Implementation Phases

### Phase A: MVP Prototype (Weeks 1-4)
*   **Goal**: End-to-end flow on Ganache with basic UI.
*   **Deliverables**:
    *   `auth-service`, `wallet-service`, `payments-service` (Basic).
    *   Ganache Contracts (ERC-20).
    *   `citizen-wallet-app` (Web).
    *   Flow: User A sends to User B.

### Phase B: Fabric Testnet (Weeks 5-8)
*   **Goal**: Replace Ganache with Hyperledger Fabric.
*   **Deliverables**:
    *   Fabric Network (3 Orgs, Raft) on Docker Compose.
    *   Go Chaincode (`cbdc-core`).
    *   Integration of Backend with Fabric SDK.

### Phase C: Intermediary Integration (Weeks 9-12)
*   **Goal**: Onboard Banks and enable G2P.
*   **Deliverables**:
    *   `intermediary-gateway-service`.
    *   `intermediary-portal`.
    *   `cbn-console` (Issuance workflows).

### Phase D: Offline Capability (Weeks 13-16)
*   **Goal**: Enable offline P2P payments.
*   **Deliverables**:
    *   `offline-service`.
    *   Offline Purse Logic (Go & Device Stub).
    *   Reconciliation Engine.

### Phase E: Hardening & Compliance (Weeks 17-20)
*   **Goal**: Production readiness.
*   **Deliverables**:
    *   Security Audit.
    *   Performance Testing (TPS Benchmarks).
    *   DR/Failover Drills.

### Phase F: Pilot (Weeks 21+)
*   **Goal**: Real-world usage.
*   **Deliverables**:
    *   Limited rollout to 1000 users.
    *   Live Monitoring.

## 2. Prioritised Backlog

### Epic 1: Core Ledger (Fabric)
*   [ ] Setup Fabric Network (3 Orgs).
*   [ ] Implement `Issue` and `Transfer` chaincode.
*   [ ] Implement Private Data Collections for KYC.

### Epic 2: Wallet & Identity
*   [ ] Implement `auth-service` with JWT.
*   [ ] Implement `wallet-service` key management.
*   [ ] Build `citizen-wallet-app` UI.

### Epic 3: Payments & Gateway
*   [ ] Implement `payments-service` orchestration.
*   [ ] Build `intermediary-gateway` API.
*   [ ] Implement Limits & Checks.

### Epic 4: Offline
*   [ ] Design Offline Purse Protocol.
*   [ ] Implement `offline-service` reconciliation.
