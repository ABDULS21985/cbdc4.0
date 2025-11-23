#!/bin/bash

# Network Setup Scripts matching Phase 2 requirements

MODE=$1

if [ "$MODE" == "up" ]; then
    echo "Starting Fabric Network (Phase 2 Topology)..."
    docker-compose -f ../docker/docker-compose.yaml up -d
    echo "Network started. Use 'deploy_chaincode.sh' to deploy contracts."
elif [ "$MODE" == "down" ]; then
    echo "Stopping Fabric Network..."
    docker-compose -f ../docker/docker-compose.yaml down --volumes --remove-orphans
    echo "Network stopped and volumes cleaned."
elif [ "$MODE" == "generate" ]; then
    echo "Generating Crypto Material (Mock)..."
    # In a real env, we would run cryptogen here using ../crypto-config/crypto-config.yaml
    # cryptogen generate --config=../crypto-config/crypto-config.yaml --output=../crypto-config
    echo "Crypto material generation skipped (using pre-canned or assuming existing)."
else
    echo "Usage: ./network.sh [up|down|generate]"
    exit 1
fi
