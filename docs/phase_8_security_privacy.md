# Phase 8: Security, Privacy & Compliance

## 1. Security Architecture

### 1.1 Network Security
*   **Zero Trust**: All service-to-service communication is mutually authenticated via mTLS.
*   **Segmentation**:
    *   **Zone A (Core)**: Fabric Peers, Orderers, HSMs. No direct internet access.
    *   **Zone B (Services)**: Backend Microservices. Access only to Zone A and Zone C.
    *   **Zone C (DMZ)**: API Gateways, Load Balancers. Public facing.

### 1.2 Cryptography & PKI
*   **Algorithms**: ECDSA (secp256r1) for Fabric, Ed25519 for Offline Purses.
*   **HSM Integration**:
    *   Root CAs and Orderer Signing Keys stored in FIPS 140-2 Level 3 HSMs (simulated with SoftHSM for dev).
    *   Wallet Service uses Cloud KMS or HSM for custodial keys.

## 2. Privacy Model

### 2.1 Data Minimization
*   **Ledger**: Stores only `WalletID` (Pseudonym) and `Balance`. No names, addresses, or tax IDs.
*   **Intermediaries**: Store the link between `WalletID` and Real Identity.

### 2.2 Private Data Collections (PDC)
*   **Usage**: To share transaction details between the transacting banks and the regulator without broadcasting to the entire network.
*   **Policy**: `OR('BankA', 'BankB', 'Regulator')`.

## 3. Compliance Strategy

### 3.1 AML/CFT
*   **Transaction Monitoring**: Real-time analysis of transaction graphs in the `Data Warehouse` (Layer 5).
*   **Sanctions Screening**: Performed by the `Intermediary Gateway` before submitting any transaction to the ledger.

### 3.2 Limits & Controls
*   **Tier 0**: Max balance $500, Max daily tx $100.
*   **Tier 1**: Max balance $10,000, Max daily tx $2,000.
*   **Enforcement**:
    *   **Ingress**: API Gateway rejects requests exceeding limits.
    *   **Chaincode**: Final check on-chain to prevent bypass.
