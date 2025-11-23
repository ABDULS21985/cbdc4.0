#!/bin/bash

# Environment variables for Fabric binaries
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config

# Channel Name
CHANNEL_NAME="cbdc-main-channel"
CC_NAME="cbdc-core"
CC_VERSION="1.0"
CC_SEQUENCE="1"
CC_SRC_PATH="../../backend/chaincode/cbdc-core"
COLLECTIONS_CONFIG="../chaincode/collections_config.json"

# Endorsement Policy: Requires endorsement from Central Bank AND Consortium
ENDORSEMENT_POLICY="AND('CentralBankMSP.peer','BankConsortiumMSP.peer')"

echo "Deploying Chaincode: $CC_NAME to Channel: $CHANNEL_NAME"

# 1. Package Chaincode
peer lifecycle chaincode package ${CC_NAME}.tar.gz --path ${CC_SRC_PATH} --lang golang --label ${CC_NAME}_${CC_VERSION}

# 2. Install Chaincode (Simulated for OrgCentralBank)
export CORE_PEER_LOCALMSPID="CentralBankMSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/../crypto-config/peerOrganizations/centralbank.cbdc/users/Admin@centralbank.cbdc/msp
export CORE_PEER_ADDRESS=peer0.centralbank.cbdc:7051
peer lifecycle chaincode install ${CC_NAME}.tar.gz

# 3. Install Chaincode (Simulated for OrgBankConsortium)
export CORE_PEER_LOCALMSPID="BankConsortiumMSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/../crypto-config/peerOrganizations/consortium.cbdc/users/Admin@consortium.cbdc/msp
export CORE_PEER_ADDRESS=peer0.consortium.cbdc:9051
peer lifecycle chaincode install ${CC_NAME}.tar.gz

# 4. Approve Chaincode Definition (OrgCentralBank)
# Note: Including --collections-config and --signature-policy
export CORE_PEER_LOCALMSPID="CentralBankMSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/../crypto-config/peerOrganizations/centralbank.cbdc/users/Admin@centralbank.cbdc/msp
export CORE_PEER_ADDRESS=peer0.centralbank.cbdc:7051

PACKAGE_ID=$(peer lifecycle chaincode queryinstalled | grep ${CC_NAME}_${CC_VERSION} | awk '{print $3}' | sed 's/,//')

peer lifecycle chaincode approveformyorg -o orderer.centralbank.cbdc:7050 --ordererTLSHostnameOverride orderer.centralbank.cbdc --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version ${CC_VERSION} --package-id ${PACKAGE_ID} --sequence ${CC_SEQUENCE} --collections-config ${COLLECTIONS_CONFIG} --signature-policy ${ENDORSEMENT_POLICY}

# 5. Approve Chaincode Definition (OrgBankConsortium)
export CORE_PEER_LOCALMSPID="BankConsortiumMSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/../crypto-config/peerOrganizations/consortium.cbdc/users/Admin@consortium.cbdc/msp
export CORE_PEER_ADDRESS=peer0.consortium.cbdc:9051

peer lifecycle chaincode approveformyorg -o orderer.centralbank.cbdc:7050 --ordererTLSHostnameOverride orderer.centralbank.cbdc --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version ${CC_VERSION} --package-id ${PACKAGE_ID} --sequence ${CC_SEQUENCE} --collections-config ${COLLECTIONS_CONFIG} --signature-policy ${ENDORSEMENT_POLICY}

# 6. Commit Chaincode Definition
export CORE_PEER_LOCALMSPID="CentralBankMSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/../crypto-config/peerOrganizations/centralbank.cbdc/users/Admin@centralbank.cbdc/msp
export CORE_PEER_ADDRESS=peer0.centralbank.cbdc:7051

peer lifecycle chaincode commit -o orderer.centralbank.cbdc:7050 --ordererTLSHostnameOverride orderer.centralbank.cbdc --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version ${CC_VERSION} --sequence ${CC_SEQUENCE} --collections-config ${COLLECTIONS_CONFIG} --signature-policy ${ENDORSEMENT_POLICY} --peerAddresses peer0.centralbank.cbdc:7051 --peerAddresses peer0.consortium.cbdc:9051

echo "Chaincode Deployed Successfully with Endorsement Policy: $ENDORSEMENT_POLICY"
