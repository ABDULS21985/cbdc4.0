# Phase 5: Backend Service Design (Golang)

## 1. Service Responsibilities & APIs

### 1.1 `auth-service`
*   **Role**: Identity Provider & Token Issuer.
*   **Responsibilities**:
    *   Authenticate users via National ID / Bio (simulated).
    *   Issue JWTs with claims (`role`, `tier`, `intermediary_id`).
    *   Manage API keys for service-to-service auth.
*   **API**:
    *   `POST /auth/login`: Exchange credentials for JWT.
    *   `POST /auth/refresh`: Refresh token.
    *   `GET /auth/verify`: Validate token (internal).

### 1.2 `wallet-service`
*   **Role**: Wallet Manager.
*   **Responsibilities**:
    *   Create/Onboard Wallets.
    *   Manage Keys (Custodial for Tier 0/1).
    *   Map `UserID` -> `WalletAddress`.
*   **API**:
    *   `POST /wallets`: Create new wallet.
    *   `GET /wallets/{id}`: Get wallet details.
    *   `GET /wallets/{id}/balance`: Get balance (cached or live from Fabric).

### 1.3 `payments-service`
*   **Role**: Transaction Orchestrator.
*   **Responsibilities**:
    *   Initiate Payments (P2P, P2M).
    *   Apply Limits (Daily/Monthly caps).
    *   Submit to Fabric SDK.
*   **API**:
    *   `POST /payments`: Initiate transfer.
    *   `GET /payments/{id}`: Get status.
    *   `GET /payments/history`: List transactions.

### 1.4 `offline-service`
*   **Role**: Offline Manager.
*   **Responsibilities**:
    *   Issue Offline Purses (Signed by CBNO).
    *   Process Offline Reconciliation (Batch upload of signed offline txs).
    *   Detect Double Spends.
*   **API**:
    *   `POST /offline/purse`: Request offline purse.
    *   `POST /offline/reconcile`: Submit offline txs.

### 1.5 `intermediary-gateway-service`
*   **Role**: Bank Integration Point.
*   **Responsibilities**:
    *   Standardized API for all banks.
    *   Webhook delivery for incoming payments.

## 2. API Specification (General)

*   **Protocol**: REST (JSON) for external clients, gRPC for internal service-to-service.
*   **Auth**: Bearer Token (JWT).
*   **Error Format**:
    ```json
    {
      "code": "insufficient_funds",
      "message": "Balance is too low",
      "trace_id": "abc-123"
    }
    ```

## 3. Data Storage Design

### 3.1 Off-Chain Database (Postgres)
We will use a shared Postgres instance with separate schemas per service (or separate DBs).

*   **`wallet_db`**:
    *   `users`: id, kyc_data, tier.
    *   `wallets`: id, user_id, address, encrypted_keys.
*   **`payments_db`**:
    *   `transactions`: id, from, to, amount, status, tx_hash (Fabric).
*   **`offline_db`**:
    *   `devices`: id, public_key, counter.
    *   `vouchers`: id, device_id, amount, status.

### 3.2 Consistency
*   **Optimistic**: Write to DB as "Pending", submit to Fabric.
*   **Async**: Listen to Fabric Block Events to update DB status to "Confirmed" or "Failed".
