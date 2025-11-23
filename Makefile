# CBDC Platform Makefile

.PHONY: all build test lint clean up down

# Variables
GO_SERVICES := backend/services/auth-service backend/services/wallet-service backend/services/payments-service backend/services/offline-service backend/services/intermediary-gateway-service backend/services/rtgs-adapter backend/services/cbn-ops-service
CHAINCODE := backend/chaincode/cbdc-core
FRONTEND_APPS := frontend/cbn-console frontend/intermediary-portal frontend/citizen-wallet-app frontend/merchant-portal

all: build

# --- Backend (Go) ---

build-go:
	@echo "Building Go services..."
	@for dir in $(GO_SERVICES); do \
		echo "Building $$dir..."; \
		(cd $$dir && go build -v ./...); \
	done
	@echo "Building Chaincode..."
	(cd $(CHAINCODE) && go build -v ./...)

test-go:
	@echo "Testing Go services..."
	go test -v ./backend/...

lint-go:
	@echo "Linting Go code..."
	golangci-lint run ./backend/...

# --- Frontend (Next.js) ---

install-frontend:
	@echo "Installing frontend dependencies..."
	npm install

build-frontend:
	@echo "Building frontend apps..."
	npm run build

lint-frontend:
	@echo "Linting frontend apps..."
	npm run lint

# --- Infrastructure ---

up:
	@echo "Starting local environment..."
	docker-compose -f infra/fabric/docker/docker-compose.yaml up -d

down:
	@echo "Stopping local environment..."
	docker-compose -f infra/fabric/docker/docker-compose.yaml down

clean:
	@echo "Cleaning up..."
	rm -rf dist
	find . -name "node_modules" -type d -prune -exec rm -rf '{}' +

# --- Helpers ---

fmt:
	go fmt ./backend/...
	npm run format
