# Phase 3: Solution Architecture & Repository Structure

## 1. Overall Solution Architecture

The solution follows a **Microservices Architecture** for the backend, communicating with a **Permissioned Blockchain (Fabric)** for settlement and a **Next.js Frontend** layer for user interaction.

### 1.1 Components

#### DLT Layer
*   **Hyperledger Fabric**: The source of truth for CBDC balances and transactions.
*   **Ganache**: Used for rapid prototyping of smart contract logic and offline voucher verification.

#### Backend Services (Golang)
*   **`auth-service`**: Identity Provider (IdP) integration, JWT issuance, RBAC.
*   **`wallet-service`**: Manages user wallets, keys (if custodial), and maps users to on-chain addresses.
*   **`payments-service`**: Orchestrates payment flows, handles limits, and communicates with the DLT.
*   **`offline-service`**: Manages offline purse issuance and reconciles offline transactions.
*   **`intermediary-gateway-service`**: External API for Banks/PSPs to interact with the core.
*   **`rtgs-adapter`**: Simulates settlement with the Real-Time Gross Settlement system.
*   **`cbn-ops-service`**: Backend for the Central Bank Operations Console.

#### Frontend Applications (Next.js)
*   **`cbn-console`**: Administrative portal for the Central Bank.
*   **`intermediary-portal`**: Portal for Commercial Banks/PSPs.
*   **`citizen-wallet-app`**: PWA/Mobile-first wallet for end-users.
*   **`merchant-portal`**: Dashboard for merchants to view settlements and generate QRs.

#### Shared Libraries
*   **`pkg/cbdc-types`**: Shared Go structs and TypeScript interfaces.
*   **`pkg/crypto`**: Common cryptographic functions (signing, hashing).
*   **`ui-kit`**: Shared React components (Tailwind + Shadcn).

## 2. Monorepo Structure

We will use a monorepo to manage all components in a single repository.

```text
/
├── docs/                   # Architecture and design documentation
├── infra/                  # Infrastructure as Code
│   ├── fabric/             # Hyperledger Fabric network config
│   ├── ganache/            # Ganache/Truffle project
│   ├── k8s/                # Kubernetes manifests
│   └── monitoring/         # Prometheus/Grafana config
├── backend/                # Golang Microservices
│   ├── services/
│   │   ├── auth-service/
│   │   ├── wallet-service/
│   │   ├── payments-service/
│   │   ├── offline-service/
│   │   ├── intermediary-gateway-service/
│   │   ├── rtgs-adapter/
│   │   └── cbn-ops-service/
│   ├── chaincode/          # Fabric Chaincode
│   │   ├── cbdc-core/
│   │   └── governance/
│   └── pkg/                # Shared Go libraries
│       ├── api/            # Proto/OpenAPI definitions
│       └── common/         # Utilities
├── frontend/               # Next.js Applications
│   ├── cbn-console/
│   ├── intermediary-portal/
│   ├── citizen-wallet-app/
│   ├── merchant-portal/
│   └── packages/           # Shared Frontend Packages
│       ├── ui/
│       └── sdk/
├── scripts/                # Build and utility scripts
├── go.work                 # Go workspace file
└── package.json            # Root Node.js config (workspaces)
```

## 3. Component Interaction Diagram

```mermaid
graph TD
    subgraph "Frontend Layer"
        Wallet[Citizen Wallet]
        Console[CBN Console]
    end

    subgraph "API Gateway / Load Balancer"
        Gateway
    end

    subgraph "Backend Services"
        Auth[Auth Service]
        Pay[Payments Service]
        WalletSvc[Wallet Service]
        Offline[Offline Service]
    end

    subgraph "DLT Layer"
        Fabric[Hyperledger Fabric]
        Ganache[Ganache (Dev)]
    end

    Wallet --> Gateway
    Console --> Gateway
    Gateway --> Auth
    Gateway --> Pay
    Gateway --> WalletSvc
    
    Pay --> Fabric
    Pay --> WalletSvc
    Offline --> Fabric
    WalletSvc --> Auth
```
