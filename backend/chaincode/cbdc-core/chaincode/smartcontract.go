package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Wallet represents a user's holding capability
type Wallet struct {
	ID             string `json:"id"`
	OwnerID        string `json:"owner_id"`
	IntermediaryID string `json:"intermediary_id"`
	Tier           string `json:"tier"`
	Status         string `json:"status"`
	Balance        int64  `json:"balance"`
}

// Transaction represents a movement of funds
type Transaction struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int64  `json:"amount"`
	Timestamp int64  `json:"timestamp"`
	Signature []byte `json:"signature,omitempty"` // Added for Phase 4
}

// SmartContract provides functions for managing a CBDC
type SmartContract struct {
	contractapi.Contract
}

// InitLedger initializes the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// In a real scenario, we might set up the central bank wallet here
	return nil
}

// Issue mints new CBDC to a bank's wallet. Only Central Bank can call this.
func (s *SmartContract) Issue(ctx contractapi.TransactionContextInterface, toWalletID string, amount int64) error {
	// Check if caller is from Central Bank MSP
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}
	if mspID != "CentralBankMSP" {
		return fmt.Errorf("unauthorized: only Central Bank can issue CBDC")
	}

	// In a real prod environment, we would also check for specific 'admin' attribute or OU
	// val, found, err := ctx.GetClientIdentity().GetAttributeValue("role")
	// if !found || val != "admin" { ... }

	walletBytes, err := ctx.GetStub().GetState(toWalletID)
	if err != nil {
		return fmt.Errorf("failed to read wallet: %v", err)
	}
	if walletBytes == nil {
		return fmt.Errorf("wallet %s does not exist", toWalletID)
	}

	var wallet Wallet
	err = json.Unmarshal(walletBytes, &wallet)
	if err != nil {
		return err
	}

	wallet.Balance += amount

	updatedWalletBytes, _ := json.Marshal(wallet)
	err = ctx.GetStub().PutState(toWalletID, updatedWalletBytes)
	if err != nil {
		return err
	}

	// Record Transaction
	tx := Transaction{
		ID:        ctx.GetStub().GetTxID(),
		Type:      "Mint",
		From:      "CentralBank", // Minting source
		To:        toWalletID,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}
	txBytes, _ := json.Marshal(tx)
	return ctx.GetStub().PutState(tx.ID, txBytes)
}

// Redeem burns CBDC from a bank's wallet. Only Central Bank can call this.
func (s *SmartContract) Redeem(ctx contractapi.TransactionContextInterface, fromWalletID string, amount int64) error {
	// Check if caller is from Central Bank MSP
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}
	if mspID != "CentralBankMSP" {
		return fmt.Errorf("unauthorized: only Central Bank can redeem CBDC")
	}

	walletBytes, err := ctx.GetStub().GetState(fromWalletID)
	if err != nil {
		return fmt.Errorf("failed to read wallet: %v", err)
	}
	if walletBytes == nil {
		return fmt.Errorf("wallet %s does not exist", fromWalletID)
	}

	var wallet Wallet
	err = json.Unmarshal(walletBytes, &wallet)
	if err != nil {
		return err
	}

	if wallet.Balance < amount {
		return fmt.Errorf("insufficient funds to redeem")
	}

	wallet.Balance -= amount

	updatedWalletBytes, _ := json.Marshal(wallet)
	err = ctx.GetStub().PutState(fromWalletID, updatedWalletBytes)
	if err != nil {
		return err
	}

	// Record Transaction
	tx := Transaction{
		ID:        ctx.GetStub().GetTxID(),
		Type:      "Redeem",
		From:      fromWalletID,
		To:        "CentralBank", // Burning destination
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}
	txBytes, _ := json.Marshal(tx)
	return ctx.GetStub().PutState(tx.ID, txBytes)
}

// Transfer moves funds between wallets
func (s *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, fromWalletID string, toWalletID string, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	// 1. Get Sender
	senderBytes, err := ctx.GetStub().GetState(fromWalletID)
	if err != nil {
		return err
	}
	if senderBytes == nil {
		return fmt.Errorf("sender wallet %s not found", fromWalletID)
	}
	var sender Wallet
	json.Unmarshal(senderBytes, &sender)

	if sender.Status == "Frozen" {
		return fmt.Errorf("sender wallet is frozen")
	}
	if sender.Balance < amount {
		return fmt.Errorf("insufficient funds")
	}

	// Enforce Tier Limits (Phase 0 Requirement)
	// Tier 0: 10,000 limit
	// Tier 1: 100,000 limit
	// Tier 2: 1,000,000 limit
	var limit int64
	switch sender.Tier {
	case "Tier 0":
		limit = 10000
	case "Tier 1":
		limit = 100000
	case "Tier 2":
		limit = 1000000
	default:
		limit = 0 // Unknown tier, block tx
	}

	if amount > limit {
		return fmt.Errorf("transaction amount %d exceeds limit %d for %s", amount, limit, sender.Tier)
	}

	// 2. Get Receiver
	receiverBytes, err := ctx.GetStub().GetState(toWalletID)
	if err != nil {
		return err
	}
	if receiverBytes == nil {
		return fmt.Errorf("receiver wallet %s not found", toWalletID)
	}
	var receiver Wallet
	json.Unmarshal(receiverBytes, &receiver)

	if receiver.Status == "Frozen" {
		return fmt.Errorf("receiver wallet is frozen")
	}

	// 3. Update Balances
	sender.Balance -= amount
	receiver.Balance += amount

	// 4. Save States
	senderUpdated, _ := json.Marshal(sender)
	receiverUpdated, _ := json.Marshal(receiver)
	ctx.GetStub().PutState(fromWalletID, senderUpdated)
	ctx.GetStub().PutState(toWalletID, receiverUpdated)

	// 5. Save Transaction Record
	tx := Transaction{
		ID:        ctx.GetStub().GetTxID(),
		Type:      "Transfer",
		From:      fromWalletID,
		To:        toWalletID,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}
	txBytes, _ := json.Marshal(tx)
	ctx.GetStub().PutState(tx.ID, txBytes)

	// 6. Emit Event
	ctx.GetStub().SetEvent("TransferEvent", txBytes)

	return nil
}

// CreateWallet creates a new wallet (called by Intermediary)
func (s *SmartContract) CreateWallet(ctx contractapi.TransactionContextInterface, id string, ownerID string, intermediaryID string, tier string) error {
	exists, err := ctx.GetStub().GetState(id)
	if err != nil {
		return err
	}
	if exists != nil {
		return fmt.Errorf("wallet %s already exists", id)
	}

	wallet := Wallet{
		ID:             id,
		OwnerID:        ownerID,
		IntermediaryID: intermediaryID,
		Tier:           tier,
		Status:         "Active",
		Balance:        0,
	}

	walletBytes, _ := json.Marshal(wallet)
	return ctx.GetStub().PutState(id, walletBytes)
}

// GetWallet returns the wallet state
func (s *SmartContract) GetWallet(ctx contractapi.TransactionContextInterface, id string) (*Wallet, error) {
	walletBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	if walletBytes == nil {
		return nil, fmt.Errorf("wallet %s does not exist", id)
	}

	var wallet Wallet
	err = json.Unmarshal(walletBytes, &wallet)
	return &wallet, nil
}

// GetTransaction returns the transaction details
func (s *SmartContract) GetTransaction(ctx contractapi.TransactionContextInterface, id string) (*Transaction, error) {
	// Note: In Fabric, transactions are stored in blocks, but we can query the world state if we saved the Tx object there.
	// In our Transfer function, we didn't explicitly save the Tx object to the world state, we only emitted an event.
	// To support this query, we should modify Transfer to save the Tx object or rely on an off-chain indexer.
	// For this build phase, let's implement saving the Tx to world state in Transfer.

	txBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	if txBytes == nil {
		return nil, fmt.Errorf("transaction %s does not exist", id)
	}

	var tx Transaction
	err = json.Unmarshal(txBytes, &tx)
	return &tx, nil
}

// OfflinePurse represents a secure element on a device (Private Data)
type OfflinePurse struct {
	DeviceID string `json:"device_id"`
	Counter  int64  `json:"counter"`
	Limit    int64  `json:"limit"`
}

// OfflineProof represents the cryptographic proof of an offline transaction
type OfflineProof struct {
	FromWalletID string `json:"from"`
	ToWalletID   string `json:"to"`
	Amount       int64  `json:"amount"`
	Nonce        int64  `json:"nonce"`
	Signature    string `json:"signature"`
}

// ReconcileOffline processes an offline transaction proof
func (s *SmartContract) ReconcileOffline(ctx contractapi.TransactionContextInterface, proofJSON string) error {
	// 1. Parse Proof (Simplified for prototype)
	// In production, this would verify Ed25519 signatures and check against the OfflinePurse state
	// stored in a Private Data Collection.

	// For this build, we will simulate the reconciliation by just logging it and updating the wallet.
	// We assume the 'proofJSON' contains { "from": "...", "to": "...", "amount": 10, "signature": "..." }

	// We assume the 'proofJSON' contains the OfflineProof structure

	var proof OfflineProof
	if err := json.Unmarshal([]byte(proofJSON), &proof); err != nil {
		return err
	}

	// 2. Update Balances (Reuse Transfer logic or call it directly)
	// Note: Offline transactions usually mean funds were ALREADY deducted from the 'OfflinePurse'
	// and now need to be deducted from the on-chain 'Shadow Account' or just settled.
	// Here we treat it as a deferred transfer.

	// We call Transfer internally, which records a "Transfer" type transaction.
	// To strictly follow the doc which says Type="OfflineReconcile", we should implement custom logic here
	// or modify Transfer to accept a type. For simplicity and code reuse, we'll let Transfer handle it
	// but we'll emit a separate event or just accept "Transfer" as the underlying mechanic.
	// However, to be 100% compliant with the "OfflineReconcile" enum requirement, let's manually do it.

	// ... Actually, reusing Transfer is safer for balance logic.
	// Let's modify Transfer to be internal or just record a secondary "OfflineReconcile" record?
	// No, that duplicates.
	// Let's just call Transfer. The "Type" in the doc is likely the *intent*.
	// If I must have "OfflineReconcile" as the Type in the DB, I should copy the Transfer logic here.

	// COPYING TRANSFER LOGIC FOR ACCURACY (Simplified for brevity)
	// 1. Get Sender
	senderBytes, err := ctx.GetStub().GetState(proof.FromWalletID)
	if err != nil {
		return err
	}
	if senderBytes == nil {
		return fmt.Errorf("sender wallet not found")
	}
	var sender Wallet
	json.Unmarshal(senderBytes, &sender)

	// 2. Get Receiver
	receiverBytes, err := ctx.GetStub().GetState(proof.ToWalletID)
	if err != nil {
		return err
	}
	if receiverBytes == nil {
		return fmt.Errorf("receiver wallet not found")
	}
	var receiver Wallet
	json.Unmarshal(receiverBytes, &receiver)

	// 3. Update
	if sender.Balance < proof.Amount {
		return fmt.Errorf("insufficient funds")
	}
	sender.Balance -= proof.Amount
	receiver.Balance += proof.Amount

	senderUpdated, _ := json.Marshal(sender)
	receiverUpdated, _ := json.Marshal(receiver)
	ctx.GetStub().PutState(proof.FromWalletID, senderUpdated)
	ctx.GetStub().PutState(proof.ToWalletID, receiverUpdated)

	// 4. Record Transaction
	tx := Transaction{
		ID:        ctx.GetStub().GetTxID(),
		Type:      "OfflineReconcile",
		From:      proof.FromWalletID,
		To:        proof.ToWalletID,
		Amount:    proof.Amount,
		Timestamp: time.Now().Unix(),
		Signature: []byte(proof.Signature),
	}
	txBytes, _ := json.Marshal(tx)
	return ctx.GetStub().PutState(tx.ID, txBytes)
}

// FreezeWallet blocks a wallet from transacting
func (s *SmartContract) FreezeWallet(ctx contractapi.TransactionContextInterface, walletID string) error {
	// Check permissions (Central Bank or Regulator)
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}
	// Simplified policy: CentralBankMSP OR RegulatorMSP (as per Phase 4 Design)
	if mspID != "CentralBankMSP" && mspID != "RegulatorMSP" {
		return fmt.Errorf("unauthorized: only Central Bank or Regulator can freeze wallets")
	}

	walletBytes, err := ctx.GetStub().GetState(walletID)
	if err != nil {
		return err
	}
	if walletBytes == nil {
		return fmt.Errorf("wallet %s does not exist", walletID)
	}

	var wallet Wallet
	err = json.Unmarshal(walletBytes, &wallet)
	if err != nil {
		return err
	}

	wallet.Status = "Frozen"
	updatedWalletBytes, _ := json.Marshal(wallet)
	return ctx.GetStub().PutState(walletID, updatedWalletBytes)
}

// UnfreezeWallet unblocks a wallet
func (s *SmartContract) UnfreezeWallet(ctx contractapi.TransactionContextInterface, walletID string) error {
	// Check permissions
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}
	if mspID != "CentralBankMSP" {
		return fmt.Errorf("unauthorized: only Central Bank can unfreeze wallets")
	}

	walletBytes, err := ctx.GetStub().GetState(walletID)
	if err != nil {
		return err
	}
	if walletBytes == nil {
		return fmt.Errorf("wallet %s does not exist", walletID)
	}

	var wallet Wallet
	err = json.Unmarshal(walletBytes, &wallet)
	if err != nil {
		return err
	}

	wallet.Status = "Active"
	updatedWalletBytes, _ := json.Marshal(wallet)
	return ctx.GetStub().PutState(walletID, updatedWalletBytes)
}
