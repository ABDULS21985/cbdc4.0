package main

import (
	"log"

	"github.com/centralbank/cbdc/backend/chaincode/cbdc-core/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	cbdcChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating CBDC chaincode: %v", err)
	}

	if err := cbdcChaincode.Start(); err != nil {
		log.Panicf("Error starting CBDC chaincode: %v", err)
	}
}
