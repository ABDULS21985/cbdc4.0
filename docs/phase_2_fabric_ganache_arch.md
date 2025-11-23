# Phase 2: Fabric & Ganache Architecture Design

## 1. Hyperledger Fabric Network Design (Production)

### 1.1 Organizations
The network is permissioned and composed of the following organizations:

1.  **`OrgCentralBank` (CBNO)**
    *   **Role**: Network Operator, Issuer, Regulator.
    *   **Peers**: 2 Endorsing Peers, 2 Committing Peers (High Availability).
    *   **CA**: `ca.centralbank.cbdc` (Root CA for CBNO identities).
    *   **MSP ID**: `CentralBankMSP`.

2.  **`OrgBankConsortium`** (Initially one org representing early adopters, later split)
    *   **Role**: Intermediaries (Banks/PSPs) validating transactions.
    *   **Peers**: 2+ Peers (hosted by lead banks or a consortium operator).
    *   **CA**: `ca.consortium.cbdc`.
    *   **MSP ID**: `BankConsortiumMSP`.

3.  **`OrgRegulatorAuditor`**
    *   **Role**: Passive observer / Auditor.
    *   **Peers**: 1 Committing Peer (ReadOnly access to specific channels).
    *   **MSP ID**: `RegulatorMSP`.

### 1.2 Ordering Service
*   **Consensus**: **Raft** (EtcdRaft).
*   **Nodes**: 5 Orderer nodes distributed across `OrgCentralBank` (3 nodes) and `OrgBankConsortium` (2 nodes) to ensure crash fault tolerance and prevent single-party censorship (though CBNO holds majority).

### 1.3 Channels
1.  **`cbdc-main-channel`**
    *   **Purpose**: Retail CBDC balances, transfers, issuance, redemption.
    *   **Members**: All Orgs.
    *   **Chaincode**: `cbdc-core`.

2.  **`ops-governance-channel`**
    *   **Purpose**: Scheme rules, participant onboarding, fee structures, global limits.
    *   **Members**: `OrgCentralBank` (Admin), `OrgBankConsortium` (Read/Propose), `OrgRegulatorAuditor` (Read).
    *   **Chaincode**: `governance-cc`.

### 1.4 Private Data Collections (PDC)
To preserve privacy while maintaining a shared ledger:

*   **`pdc-retail-wallets`**:
    *   **Owner**: `OrgCentralBank` + Specific Intermediary.
    *   **Content**: Mapping of `WalletID` <-> `EncryptedKYCData`. The ledger only sees `WalletID` and `Balance`.
*   **`pdc-intermediary-positions`**:
    *   **Owner**: `OrgCentralBank` + `OrgBankConsortium`.
    *   **Content**: Detailed liquidity positions if deemed sensitive.

### 1.5 Endorsement Policies
*   **Issuance/Minting**: `AND('CentralBankMSP.admin')` - Only CBNO can mint.
*   **Transfer**: `AND('CentralBankMSP.peer', 'BankConsortiumMSP.peer')` - Requires validation from both the central authority and the intermediary layer.

## 2. Ganache/Ethereum Dev Architecture

### 2.1 Purpose
*   **Rapid Prototyping**: Test token logic (mint, burn, transfer, freeze) without waiting for Fabric block times or complex chaincode deployments.
*   **Offline Voucher Logic**: Prototype the cryptographic proofs for offline payments using Solidity's `ecrecover` before porting to Go.

### 2.2 Setup
*   **Chain ID**: `1337` (Local Dev).
*   **Token Standard**: Modified `ERC-20` with:
    *   `blacklist(address)` / `pause()` (Regulatory controls).
    *   `depositFor(address, amount, signature)` (Offline reconciliation).

## 3. Network Diagram

```mermaid
graph TD
    subgraph "Ordering Service (Raft)"
        O1[Orderer 1]
        O2[Orderer 2]
        O3[Orderer 3]
        O4[Orderer 4]
        O5[Orderer 5]
    end

    subgraph "OrgCentralBank"
        CB_P1[Peer 1 (Endorser)]
        CB_P2[Peer 2 (Endorser)]
        CB_CA[CA Server]
        CB_App[Ops Console]
    end

    subgraph "OrgBankConsortium"
        BK_P1[Peer 1]
        BK_P2[Peer 2]
        BK_CA[CA Server]
        BK_App[Intermediary Gateway]
    end

    subgraph "OrgRegulator"
        REG_P1[Peer 1 (Committing)]
    end

    CB_App --> CB_P1
    BK_App --> BK_P1
    CB_P1 --- O1
    BK_P1 --- O1
    
    classDef ord fill:#f9f,stroke:#333,stroke-width:2px;
    class O1,O2,O3,O4,O5 ord;
```

## 4. Infrastructure Directory Structure

We will create the following structure in the monorepo:

```text
/infra
  /fabric
    /configtx         # Channel & Genesis block config
    /crypto-config    # Cryptogen specs (for dev/test)
    /docker           # Docker Compose files for local Fabric
    /scripts          # Setup/Teardown scripts
  /ganache
    /contracts        # Solidity prototypes
    /migrations       # Truffle/Hardhat deploy scripts
    /tests            # JS/TS tests for contracts
```
