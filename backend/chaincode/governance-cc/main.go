package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type GovernanceContract struct {
	contractapi.Contract
}

type GlobalParams struct {
	MaxTransactionLimit int64 `json:"max_transaction_limit"`
	MinTransactionLimit int64 `json:"min_transaction_limit"`
	FeePercentage       int   `json:"fee_percentage"` // Basis points
}

func (c *GovernanceContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	params := GlobalParams{
		MaxTransactionLimit: 1000000,
		MinTransactionLimit: 1,
		FeePercentage:       0,
	}
	paramsBytes, _ := json.Marshal(params)
	return ctx.GetStub().PutState("GLOBAL_PARAMS", paramsBytes)
}

func (c *GovernanceContract) UpdateParams(ctx contractapi.TransactionContextInterface, maxLimit int64, minLimit int64, fee int) error {
	// TODO: Check if caller is Admin of Central Bank

	params := GlobalParams{
		MaxTransactionLimit: maxLimit,
		MinTransactionLimit: minLimit,
		FeePercentage:       fee,
	}
	paramsBytes, _ := json.Marshal(params)
	return ctx.GetStub().PutState("GLOBAL_PARAMS", paramsBytes)
}

func (c *GovernanceContract) GetParams(ctx contractapi.TransactionContextInterface) (*GlobalParams, error) {
	paramsBytes, err := ctx.GetStub().GetState("GLOBAL_PARAMS")
	if err != nil {
		return nil, err
	}
	if paramsBytes == nil {
		return nil, fmt.Errorf("params not set")
	}

	var params GlobalParams
	err = json.Unmarshal(paramsBytes, &params)
	return &params, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&GovernanceContract{})
	if err != nil {
		log.Panicf("Error creating governance chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting governance chaincode: %v", err)
	}
}
