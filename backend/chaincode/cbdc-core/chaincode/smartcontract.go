package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

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
	// TODO: Check if caller is Central Bank Admin

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
	return ctx.GetStub().PutState(toWalletID, updatedWalletBytes)
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

	// 5. Emit Event (Optional)
	tx := Transaction{
		ID:        ctx.GetStub().GetTxID(),
		Type:      "Transfer",
		From:      fromWalletID,
		To:        toWalletID,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}
	txBytes, _ := json.Marshal(tx)
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
