# Phase 7: Offline CBDC Design

## 1. Offline Wallet Model

### 1.1 The Offline Purse
The **Offline Purse** is a secure container (Secure Element / TEE) on the user's device that holds a balance of "Offline CBDC".

*   **Properties**:
    *   **DeviceID**: Unique Public Key of the hardware element.
    *   **Balance**: Current offline balance.
    *   **Counter**: Monotonic counter incremented on every spend.
    *   **LastSyncHash**: Hash of the last reconciliation event.

### 1.2 Funding (Online -> Offline)
1.  User requests `LoadOffline(50 CBDC)`.
2.  Online Core locks 50 CBDC in the user's main account.
3.  Online Core issues a `PurseUpdate` certificate (signed by CBNO) crediting the device.
4.  Device verifies signature and increments local balance.

## 2. Offline Transaction Protocol (P2P)

### 2.1 Protocol Steps
1.  **Handshake**: Payer and Payee devices exchange public keys and capabilities (NFC/BLE).
2.  **Proposal**: Payer creates a `PaymentIntent`:
    *   `Amount`: 10
    *   `PayeeID`: Bob_Device_PK
    *   `Counter`: Payer_Counter + 1
3.  **Signing**: Payer's Secure Element signs the intent: `Sign(PaymentIntent, DeviceKey)`.
4.  **Transfer**: Payer sends the `SignedPayment` to Payee.
5.  **Verification**: Payee verifies:
    *   Signature is valid (using Payer's Cert).
    *   Payer's Cert is not expired/revoked (using cached CRL).
6.  **Completion**: Payee stores the `SignedPayment` and updates local "Pending Balance".

## 3. Double-Spend & Reconciliation

### 3.1 Risk Controls
*   **Limits**: Max offline balance (e.g., $500), Max transaction size ($50).
*   **TTL**: Offline purses must sync every 7 days or they lock.

### 3.2 Reconciliation (Offline -> Online)
1.  Payee comes online.
2.  Payee App uploads `SignedPayment` blobs to `offline-service`.
3.  `offline-service`:
    *   Verifies signatures.
    *   Checks against `UsedCounters` DB to detect double-spending.
    *   If valid: Credits Payee's online wallet, Debits Payer's "Shadow Offline Balance".
    *   If double-spend detected: Flags Payer's account, triggers fraud alert.

## 4. Integration
*   **Fabric**: Records the net settlement of offline batches.
*   **Ganache**: Can verify the cryptographic proofs (ecrecover) during prototyping.
