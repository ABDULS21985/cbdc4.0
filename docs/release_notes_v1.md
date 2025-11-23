# CBDC Platform - Release Candidate V1

## Features Implemented
- **Core Ledger**: Hyperledger Fabric network with 3 orgs (Central Bank, Consortium, Regulator).
- **Chaincode**: Go-based smart contract for Issue, Transfer, Freeze, Redeem.
- **Identity**: JWT-based authentication with RBAC (Admin, Citizen, Merchant).
- **Wallet Service**: Account-based wallet management mapped to Fabric identities.
- **Payments**: P2P and Merchant payment flows with immediate settlement.
- **Offline**: Basic offline transaction reconciliation engine using Ed25519 signatures.
- **Frontend**: Citizen Wallet App with real API integration and shared UI components.
- **Security**: API Gateway (Nginx) and RBAC middleware.
- **DevOps**: Dockerfiles for all services and Prometheus monitoring stack.

## Known Limitations (TODOs)
- **Fabric Network**: Currently runs in "dev" mode with pre-generated crypto. Production requires CA integration.
- **Database**: Services currently use in-memory or mocked DBs. Postgres integration is stubbed.
- **Offline**: Double-spend protection relies on optimistic reconciliation; secure element integration is conceptual.
- **KYC**: Mocked checks; integration with National ID is pending.

## Next Steps
- Deploy to Testnet.
- Conduct security audit.
- Onboard pilot banks.
