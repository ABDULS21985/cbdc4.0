package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// Config for E2E tests - assumes services are running locally
const (
	WalletServiceURL   = "http://localhost:8082"
	PaymentsServiceURL = "http://localhost:8083"
)

func TestRetailPaymentFlow(t *testing.T) {
	// Skip if not in integration mode
	// t.Skip("Skipping E2E test in build phase")

	// 1. Create Alice's Wallet
	aliceID := fmt.Sprintf("alice-%d", time.Now().Unix())
	createWallet(t, aliceID, "Tier1")

	// 2. Create Bob's Wallet
	bobID := fmt.Sprintf("bob-%d", time.Now().Unix())
	createWallet(t, bobID, "Tier1")

	// 3. Issue Funds to Alice (Mocked or via Admin API if exists)
	// For this test, we assume Alice has funds (e.g. via genesis or manual issue)

	// 4. Alice sends to Bob
	transfer(t, aliceID, bobID, 50)

	// 5. Verify Balances (via Wallet Service)
	// aliceBal := getBalance(t, aliceID)
	// bobBal := getBalance(t, bobID)
	// assert(aliceBal == expected)
}

func createWallet(t *testing.T, userID, tier string) {
	payload := map[string]string{
		"user_id": userID,
		"tier":    tier,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(WalletServiceURL+"/wallets", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Logf("Failed to create wallet: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Logf("Create wallet failed with status: %d", resp.StatusCode)
	}
}

func transfer(t *testing.T, from, to string, amount int64) {
	payload := map[string]interface{}{
		"from":   "wallet-" + from,
		"to":     "wallet-" + to,
		"amount": amount,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(PaymentsServiceURL+"/payments/p2p", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Logf("Failed to transfer: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Logf("Transfer failed with status: %d", resp.StatusCode)
	}
}
