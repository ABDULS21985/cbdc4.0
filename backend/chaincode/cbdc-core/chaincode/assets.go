package chaincode

// Wallet represents a user's holding capability
type Wallet struct {
	ID             string `json:"id"`
	OwnerID        string `json:"owner_id"` // Pseudonymous ID
	IntermediaryID string `json:"intermediary_id"`
	Tier           string `json:"tier"`   // Tier0, Tier1, Tier2
	Status         string `json:"status"` // Active, Frozen
	Balance        int64  `json:"balance"`
}

// Transaction represents a movement of funds
type Transaction struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // Mint, Transfer, Redeem
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int64  `json:"amount"`
	Timestamp int64  `json:"timestamp"`
}

const (
	DocTypeWallet = "WALLET"
	DocTypeTx     = "TX"
)
