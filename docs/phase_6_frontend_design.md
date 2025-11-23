# Phase 6: Frontend (Next.js) Applications

## 1. Architecture Overview

All frontend applications will use **Next.js 14+ (App Router)** with **TypeScript**.
We will use a shared UI library (`frontend/packages/ui`) based on **Tailwind CSS** and **Shadcn/UI** to ensure consistency.

### 1.1 Shared Components
*   `Button`, `Input`, `Card`, `Modal`, `Table`.
*   `Layout`: Sidebar, Header, Footer.
*   `AuthGuard`: HOC for protecting routes.

## 2. Application Designs

### 2.1 Central Bank Console (`cbn-console`)
*   **Target Audience**: Central Bank Operators, Auditors.
*   **Key Features**:
    *   **Dashboard**: Total Supply, Velocity of Money, Intermediary Health.
    *   **Issuance**: Mint/Burn workflows (Multi-sig approval UI).
    *   **Governance**: Update scheme rules, freeze/unfreeze intermediaries.
*   **Sitemap**:
    *   `/`: Login.
    *   `/dashboard`: Overview.
    *   `/issuance`: Mint/Burn.
    *   `/intermediaries`: Manage banks.

### 2.2 Intermediary Portal (`intermediary-portal`)
*   **Target Audience**: Bank Operations Staff.
*   **Key Features**:
    *   **Customer Management**: KYC status, Wallet creation.
    *   **Liquidity Management**: Request CBDC from CBNO, redeem for reserves.
    *   **Reporting**: Transaction logs, compliance reports.
*   **Sitemap**:
    *   `/`: Login.
    *   `/customers`: List of retail users.
    *   `/liquidity`: Manage bank's own CBDC position.

### 2.3 Citizen Wallet App (`citizen-wallet-app`)
*   **Target Audience**: General Public (Mobile Web / PWA).
*   **Key Features**:
    *   **Wallet**: View Balance, QR Code.
    *   **Send**: P2P transfer (Phone/QR).
    *   **Offline**: Toggle offline mode (loads local purse UI).
*   **Sitemap**:
    *   `/`: Splash/Login.
    *   `/home`: Balance & Actions.
    *   `/scan`: QR Scanner.
    *   `/send`: Transfer form.

### 2.4 Merchant Portal (`merchant-portal`)
*   **Target Audience**: Business Owners.
*   **Key Features**:
    *   **POS**: Generate dynamic QRs for collection.
    *   **Settlement**: View daily totals, auto-sweep to bank account.

## 3. Integration Strategy
*   **API Client**: Generated from OpenAPI specs (or manual typed fetch wrapper).
*   **State Management**: React Query (TanStack Query) for server state, Zustand for local UI state.
*   **Authentication**: Store JWT in HttpOnly cookies (via Next.js Middleware proxy) or local storage (if simplified).
