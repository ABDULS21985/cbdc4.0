# Phase 4: CBDC Domain Model & Chaincode Design

## 1. Domain Model

### 1.1 Core Entities

#### `Wallet`
Represents a user's holding capability.
*   **ID**: `string` (UUID or Hashed Public Key).
*   **OwnerID**: `string` (Pseudonymous ID).
*   **IntermediaryID**: `string` (MSP ID of the bank managing this wallet).
*   **Tier**: `enum` (Tier0, Tier1, Tier2).
*   **Status**: `enum` (Active, Frozen, Suspended).
*   **Balance**: `int64` (in smallest unit, e.g. cents).

#### `Transaction`
Represents a movement of funds.
*   **ID**: `string` (TxID).
*   **Type**: `enum` (Mint, Transfer, Redeem, OfflineReconcile).
*   **FromWallet**: `string`.
*   **ToWallet**: `string`.
*   **Amount**: `int64`.
*   **Timestamp**: `int64`.
*   **Signature**: `bytes`.

#### `OfflinePurse` (Off-Chain / Private Data)
Represents a secure element on a device.
*   **DeviceID**: `string` (Public Key of SE).
*   **Counter**: `int64` (Monotonic counter to prevent replay).
*   **Limit**: `int64` (Max offline balance).

### 1.2 Data Placement
*   **Public Ledger (Fabric)**: `Wallet` (Balance only), `Transaction` (Pseudonymous).
*   **Private Data Collection**: `Wallet` (KYC Metadata), `OfflinePurse` (Device details).
*   **Off-Chain DB**: Full KYC, Transaction History for UI, Analytics.

## 2. Fabric Chaincode Design (Go)

### 2.1 Interface
The chaincode will implement the `CBDCChaincode` interface.

```go
type CBDCChaincode interface {
    // Lifecycle
    InitLedger(ctx ContractContext) error

    // Core Operations
    Issue(ctx ContractContext, amount int64, toBank string) error
    Redeem(ctx ContractContext, amount int64, fromBank string) error
    Transfer(ctx ContractContext, fromWallet, toWallet string, amount int64) error

    // Admin/Compliance
    FreezeWallet(ctx ContractContext, walletID string) error
    UnfreezeWallet(ctx ContractContext, walletID string) error

    // Offline
    ReconcileOffline(ctx ContractContext, proof OfflineProof) error
}
```

### 2.2 Endorsement Policies
*   `Issue`: Requires `OrgCentralBank`.
*   `Transfer`: Requires `OrgCentralBank` AND `OrgBankConsortium` (The intermediary managing the sender).
*   `Freeze`: Requires `OrgCentralBank` OR (`OrgBankConsortium` + `OrgRegulator`).

## 3. Ganache Prototype Contracts (Solidity)

### 3.1 `CBDC.sol`
An ERC-20 compliant token with extensions.

```solidity
interface ICBDC {
    function mint(address to, uint256 amount) external; // Only CentralBank
    function burn(uint256 amount) external;
    function transfer(address to, uint256 amount) external returns (bool);
    function freeze(address account) external; // Only Admin
    function unfreeze(address account) external; // Only Admin
    
    // Offline
    function depositFor(address to, uint256 amount, bytes memory signature) external;
}
```
