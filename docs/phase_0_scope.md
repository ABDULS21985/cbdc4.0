# Phase 0: Clarifying Assumptions & Scope

## 1. Assumptions & Scope

### 1.1 CBDC Policy & Scope
- **Jurisdiction**: `COUNTRY_X` (Placeholder).
- **Model**: Retail, Intermediated (Two-Tier).
  - **Central Bank (CBNO)**: Issues CBDC, maintains core ledger, sets policy.
  - **Intermediaries (Banks/PSPs)**: Distribute CBDC, manage KYC/AML, provide wallets.
- **Monetary Policy**: Non-interest bearing initially (0% remuneration).
- **Coexistence**: Complementary to cash and commercial bank money.
- **Wallet Tiers**:
  - **Tier 0**: Low limits, minimal KYC (e.g. phone number only).
  - **Tier 1**: Standard limits, full KYC.
  - **Tier 2**: High limits, business/corporate use.

### 1.2 Participant Model
- **Core Operator**: Central Bank (CBNO).
- **Intermediaries**:
  - `BankA`, `BankB` (Commercial Banks).
  - `FintechX` (PSP).
- **Roles**:
  - **PIP (Payment Interface Provider)**: End-user wallet providers.
  - **ESIP (External Service Interface Provider)**: 3rd party services (programmability).
- **External FMIs**:
  - RTGS (Real-Time Gross Settlement) for wholesale settlement.
  - Instant Payment Rail (IPS) for interoperability (stubbed initially).

### 1.3 Technical Scope & Stack
- **Production Ledger**: **Hyperledger Fabric**
  - **Consensus**: Raft (Crash Fault Tolerance).
  - **Privacy**: Private Data Collections (PDC) for user balances/transactions.
  - **Channels**:
    - `cbdc-main-channel`: Core transactional ledger.
    - `ops-governance-channel`: Network config and rules.
  - **Chaincode**: Golang.
- **Prototyping Ledger**: **Ganache (Ethereum)**
  - **Purpose**: Rapid logic validation, token standard prototyping (ERC-20+), offline voucher logic testing.
- **Backend**: **Golang**
  - Architecture: Modular microservices (Identity, Wallet, Payments, Offline, Gateway).
  - Communication: gRPC/REST.
- **Frontend**: **Next.js** (TypeScript, App Router)
  - Apps: Central Bank Console, Intermediary Portal, Citizen Wallet, Merchant Portal.

## 2. Key Design Decisions

1.  **Dual-Ledger Strategy**: We will use Ganache for rapid "smart contract" logic iteration and Fabric for the secure, permissioned production system. This allows us to move fast on logic (Solidity/EVM) while building the robust infrastructure (Fabric) in parallel.
2.  **Account-Based Core with Token Extensions**: The core ledger will be account-based (UTXO or Balance model in Fabric) to support high throughput and regulatory controls. "Token" behavior will be simulated for offline scenarios using cryptographic vouchers.
3.  **Privacy by Design**: User identities are **never** stored on the core ledger in plaintext. We will use pseudonymous identifiers (e.g., public keys or hashed IDs) on-chain, with the mapping held strictly by Intermediaries (and available to Regulators/CBNO via private channels/PDCs upon court order/audit).
4.  **Offline-First Capability**: Offline payments will be treated as a first-class citizen, utilizing a "secure purse" model on devices and a "voucher" reconciliation system on the backend.
5.  **Strict Tiered Access**: The system will enforce wallet tiers at the protocol level (chaincode) where possible, or at the ingress gateway, to ensure compliance with `COUNTRY_X` regulations.
