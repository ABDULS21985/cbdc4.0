"use client";

import { useState, useEffect } from "react";

// Types
interface Wallet {
  id: string;
  ownerName: string;
  tier: "Tier0" | "Tier1" | "Tier2";
  balance: number;
  status: "active" | "frozen" | "suspended";
  createdAt: string;
  kycLevel: number;
}

interface Transaction {
  id: string;
  from: string;
  to: string;
  amount: number;
  timestamp: string;
  status: "completed" | "pending" | "failed";
  type: "transfer" | "deposit" | "withdrawal" | "issuance";
}

interface IntermediaryStats {
  totalWallets: number;
  activeWallets: number;
  totalBalance: number;
  todayTransactions: number;
  todayVolume: number;
  pendingKyc: number;
}

interface KycRequest {
  id: string;
  walletId: string;
  customerName: string;
  requestedTier: "Tier1" | "Tier2";
  documents: string[];
  submittedAt: string;
  status: "pending" | "approved" | "rejected";
}

// API Configuration
const WALLET_SERVICE_URL = process.env.NEXT_PUBLIC_WALLET_SERVICE_URL || "http://localhost:8081";
const PAYMENTS_SERVICE_URL = process.env.NEXT_PUBLIC_PAYMENTS_SERVICE_URL || "http://localhost:8082";

export default function IntermediaryPortal() {
  const [activeTab, setActiveTab] = useState<"dashboard" | "wallets" | "kyc" | "transactions" | "liquidity" | "reports">("dashboard");
  const [intermediaryId] = useState("INTERMEDIARY_001");
  const [intermediaryName] = useState("Sample Bank");
  const [stats, setStats] = useState<IntermediaryStats>({
    totalWallets: 0,
    activeWallets: 0,
    totalBalance: 0,
    todayTransactions: 0,
    todayVolume: 0,
    pendingKyc: 0,
  });
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [kycRequests, setKycRequests] = useState<KycRequest[]>([]);
  const [loading, setLoading] = useState(false);
  const [notification, setNotification] = useState<{ type: "success" | "error"; message: string } | null>(null);
  const [liquidityBalance, setLiquidityBalance] = useState(0);

  // New wallet form
  const [newWalletForm, setNewWalletForm] = useState({
    ownerName: "",
    tier: "Tier0" as "Tier0" | "Tier1" | "Tier2",
    initialBalance: 0,
  });

  useEffect(() => {
    fetchDashboardData();
    fetchWallets();
    fetchKycRequests();
    fetchTransactions();
  }, []);

  const fetchDashboardData = async () => {
    try {
      const response = await fetch(`${WALLET_SERVICE_URL}/intermediary/${intermediaryId}/stats`);
      if (response.ok) {
        const data = await response.json();
        setStats(data);
      }

      const liqResponse = await fetch(`${WALLET_SERVICE_URL}/intermediary/${intermediaryId}/liquidity`);
      if (liqResponse.ok) {
        const liqData = await liqResponse.json();
        setLiquidityBalance(liqData.balance || 0);
      }
    } catch (error) {
      console.error("Failed to fetch dashboard data:", error);
    }
  };

  const fetchWallets = async () => {
    try {
      const response = await fetch(`${WALLET_SERVICE_URL}/intermediary/${intermediaryId}/wallets`);
      if (response.ok) {
        const data = await response.json();
        setWallets(data.wallets || []);
      }
    } catch (error) {
      console.error("Failed to fetch wallets:", error);
    }
  };

  const fetchTransactions = async () => {
    try {
      const response = await fetch(`${PAYMENTS_SERVICE_URL}/intermediary/${intermediaryId}/transactions`);
      if (response.ok) {
        const data = await response.json();
        setTransactions(data.transactions || []);
      }
    } catch (error) {
      console.error("Failed to fetch transactions:", error);
    }
  };

  const fetchKycRequests = async () => {
    try {
      const response = await fetch(`${WALLET_SERVICE_URL}/intermediary/${intermediaryId}/kyc/pending`);
      if (response.ok) {
        const data = await response.json();
        setKycRequests(data.requests || []);
      }
    } catch (error) {
      console.error("Failed to fetch KYC requests:", error);
    }
  };

  const createWallet = async () => {
    if (!newWalletForm.ownerName) {
      setNotification({ type: "error", message: "Please enter owner name" });
      return;
    }

    setLoading(true);
    try {
      const response = await fetch(`${WALLET_SERVICE_URL}/wallet/create`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          intermediaryId,
          ownerName: newWalletForm.ownerName,
          tier: newWalletForm.tier,
          initialBalance: newWalletForm.initialBalance,
        }),
      });

      if (response.ok) {
        setNotification({ type: "success", message: "Wallet created successfully" });
        setNewWalletForm({ ownerName: "", tier: "Tier0", initialBalance: 0 });
        fetchWallets();
        fetchDashboardData();
      } else {
        throw new Error("Failed to create wallet");
      }
    } catch (error) {
      setNotification({ type: "error", message: "Failed to create wallet" });
    } finally {
      setLoading(false);
    }
  };

  const processKyc = async (requestId: string, action: "approve" | "reject") => {
    setLoading(true);
    try {
      const response = await fetch(`${WALLET_SERVICE_URL}/kyc/${requestId}/${action}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ intermediaryId }),
      });

      if (response.ok) {
        setNotification({ type: "success", message: `KYC request ${action}d successfully` });
        fetchKycRequests();
        fetchDashboardData();
      } else {
        throw new Error(`Failed to ${action} KYC`);
      }
    } catch (error) {
      setNotification({ type: "error", message: `Failed to ${action} KYC request` });
    } finally {
      setLoading(false);
    }
  };

  const freezeWallet = async (walletId: string) => {
    setLoading(true);
    try {
      const response = await fetch(`${WALLET_SERVICE_URL}/wallet/${walletId}/freeze`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ intermediaryId, reason: "Intermediary action" }),
      });

      if (response.ok) {
        setNotification({ type: "success", message: "Wallet frozen successfully" });
        fetchWallets();
      } else {
        throw new Error("Failed to freeze wallet");
      }
    } catch (error) {
      setNotification({ type: "error", message: "Failed to freeze wallet" });
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("en-NG", {
      style: "currency",
      currency: "NGN",
      minimumFractionDigits: 2,
    }).format(amount / 100);
  };

  const renderDashboard = () => (
    <div className="space-y-6">
      {/* Stats Grid */}
      <div className="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
        <div className="bg-white rounded-xl p-5 shadow-sm border">
          <p className="text-gray-500 text-sm">Total Wallets</p>
          <p className="text-2xl font-bold">{stats.totalWallets.toLocaleString()}</p>
        </div>
        <div className="bg-white rounded-xl p-5 shadow-sm border">
          <p className="text-gray-500 text-sm">Active Wallets</p>
          <p className="text-2xl font-bold text-green-600">{stats.activeWallets.toLocaleString()}</p>
        </div>
        <div className="bg-white rounded-xl p-5 shadow-sm border">
          <p className="text-gray-500 text-sm">Total Balance</p>
          <p className="text-2xl font-bold">{formatCurrency(stats.totalBalance)}</p>
        </div>
        <div className="bg-white rounded-xl p-5 shadow-sm border">
          <p className="text-gray-500 text-sm">Today&apos;s Volume</p>
          <p className="text-2xl font-bold">{formatCurrency(stats.todayVolume)}</p>
          <p className="text-xs text-gray-400">{stats.todayTransactions} txns</p>
        </div>
        <div className="bg-white rounded-xl p-5 shadow-sm border">
          <p className="text-gray-500 text-sm">Pending KYC</p>
          <p className="text-2xl font-bold text-orange-500">{stats.pendingKyc}</p>
        </div>
        <div className="bg-white rounded-xl p-5 shadow-sm border">
          <p className="text-gray-500 text-sm">Liquidity Balance</p>
          <p className="text-2xl font-bold text-blue-600">{formatCurrency(liquidityBalance)}</p>
        </div>
      </div>

      {/* Quick Actions & Recent Activity */}
      <div className="grid lg:grid-cols-2 gap-6">
        {/* Quick Actions */}
        <div className="bg-white rounded-xl shadow-sm border p-6">
          <h3 className="font-semibold mb-4">Quick Actions</h3>
          <div className="grid grid-cols-2 gap-3">
            <button
              onClick={() => setActiveTab("wallets")}
              className="p-4 border rounded-lg hover:bg-gray-50 text-left"
            >
              <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center mb-2">
                <span className="text-green-600">+</span>
              </div>
              <p className="font-medium">Create Wallet</p>
              <p className="text-xs text-gray-500">Onboard new customer</p>
            </button>
            <button
              onClick={() => setActiveTab("kyc")}
              className="p-4 border rounded-lg hover:bg-gray-50 text-left"
            >
              <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center mb-2">
                <span className="text-blue-600">ID</span>
              </div>
              <p className="font-medium">Process KYC</p>
              <p className="text-xs text-gray-500">{stats.pendingKyc} pending</p>
            </button>
            <button
              onClick={() => setActiveTab("liquidity")}
              className="p-4 border rounded-lg hover:bg-gray-50 text-left"
            >
              <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center mb-2">
                <span className="text-purple-600">$</span>
              </div>
              <p className="font-medium">Manage Liquidity</p>
              <p className="text-xs text-gray-500">Request more eNaira</p>
            </button>
            <button
              onClick={() => setActiveTab("reports")}
              className="p-4 border rounded-lg hover:bg-gray-50 text-left"
            >
              <div className="w-10 h-10 bg-orange-100 rounded-lg flex items-center justify-center mb-2">
                <span className="text-orange-600">R</span>
              </div>
              <p className="font-medium">Generate Report</p>
              <p className="text-xs text-gray-500">Compliance reports</p>
            </button>
          </div>
        </div>

        {/* Recent Transactions */}
        <div className="bg-white rounded-xl shadow-sm border">
          <div className="p-4 border-b flex items-center justify-between">
            <h3 className="font-semibold">Recent Transactions</h3>
            <button onClick={() => setActiveTab("transactions")} className="text-sm text-green-600 hover:underline">
              View All
            </button>
          </div>
          <div className="divide-y max-h-80 overflow-y-auto">
            {transactions.slice(0, 5).map((tx) => (
              <div key={tx.id} className="p-4 flex items-center justify-between">
                <div>
                  <p className="font-medium text-sm">{tx.type.charAt(0).toUpperCase() + tx.type.slice(1)}</p>
                  <p className="text-xs text-gray-500">{new Date(tx.timestamp).toLocaleString()}</p>
                </div>
                <div className="text-right">
                  <p className="font-semibold">{formatCurrency(tx.amount)}</p>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${
                    tx.status === "completed" ? "bg-green-100 text-green-700" :
                    tx.status === "pending" ? "bg-yellow-100 text-yellow-700" :
                    "bg-red-100 text-red-700"
                  }`}>
                    {tx.status}
                  </span>
                </div>
              </div>
            ))}
            {transactions.length === 0 && (
              <div className="p-8 text-center text-gray-500">No transactions yet</div>
            )}
          </div>
        </div>
      </div>
    </div>
  );

  const renderWallets = () => (
    <div className="space-y-6">
      {/* Create Wallet Form */}
      <div className="bg-white rounded-xl shadow-sm border p-6">
        <h3 className="font-semibold mb-4">Create New Wallet</h3>
        <div className="grid md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm text-gray-600 mb-1">Customer Name</label>
            <input
              type="text"
              value={newWalletForm.ownerName}
              onChange={(e) => setNewWalletForm({ ...newWalletForm, ownerName: e.target.value })}
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-green-500"
              placeholder="Full name"
            />
          </div>
          <div>
            <label className="block text-sm text-gray-600 mb-1">Wallet Tier</label>
            <select
              value={newWalletForm.tier}
              onChange={(e) => setNewWalletForm({ ...newWalletForm, tier: e.target.value as "Tier0" | "Tier1" | "Tier2" })}
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-green-500"
            >
              <option value="Tier0">Tier 0 (Basic)</option>
              <option value="Tier1">Tier 1 (Standard)</option>
              <option value="Tier2">Tier 2 (Premium)</option>
            </select>
          </div>
          <div>
            <label className="block text-sm text-gray-600 mb-1">Initial Balance (kobo)</label>
            <input
              type="number"
              value={newWalletForm.initialBalance || ""}
              onChange={(e) => setNewWalletForm({ ...newWalletForm, initialBalance: parseInt(e.target.value) || 0 })}
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-green-500"
              placeholder="0"
            />
          </div>
          <div className="flex items-end">
            <button
              onClick={createWallet}
              disabled={loading}
              className="w-full py-2 bg-green-600 text-white rounded-lg font-medium hover:bg-green-700 disabled:opacity-50"
            >
              {loading ? "Creating..." : "Create Wallet"}
            </button>
          </div>
        </div>
      </div>

      {/* Wallets Table */}
      <div className="bg-white rounded-xl shadow-sm border overflow-hidden">
        <div className="p-4 border-b">
          <h3 className="font-semibold">Customer Wallets</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Wallet ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Owner</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Tier</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Balance</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {wallets.map((wallet) => (
                <tr key={wallet.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm font-mono">{wallet.id.slice(0, 12)}...</td>
                  <td className="px-4 py-3 text-sm">{wallet.ownerName}</td>
                  <td className="px-4 py-3">
                    <span className={`text-xs px-2 py-1 rounded ${
                      wallet.tier === "Tier2" ? "bg-purple-100 text-purple-700" :
                      wallet.tier === "Tier1" ? "bg-blue-100 text-blue-700" :
                      "bg-gray-100 text-gray-700"
                    }`}>
                      {wallet.tier}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm font-medium">{formatCurrency(wallet.balance)}</td>
                  <td className="px-4 py-3">
                    <span className={`text-xs px-2 py-1 rounded-full ${
                      wallet.status === "active" ? "bg-green-100 text-green-700" :
                      wallet.status === "frozen" ? "bg-blue-100 text-blue-700" :
                      "bg-red-100 text-red-700"
                    }`}>
                      {wallet.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {new Date(wallet.createdAt).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      {wallet.status === "active" && (
                        <button
                          onClick={() => freezeWallet(wallet.id)}
                          className="text-xs text-blue-600 hover:underline"
                        >
                          Freeze
                        </button>
                      )}
                      <button className="text-xs text-gray-600 hover:underline">
                        View
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {wallets.length === 0 && (
            <div className="p-8 text-center text-gray-500">No wallets found</div>
          )}
        </div>
      </div>
    </div>
  );

  const renderKyc = () => (
    <div className="space-y-6">
      <div className="bg-white rounded-xl shadow-sm border">
        <div className="p-4 border-b">
          <h3 className="font-semibold">Pending KYC Requests</h3>
        </div>
        <div className="divide-y">
          {kycRequests.filter(k => k.status === "pending").map((request) => (
            <div key={request.id} className="p-6">
              <div className="flex items-start justify-between">
                <div>
                  <h4 className="font-medium">{request.customerName}</h4>
                  <p className="text-sm text-gray-500">Wallet: {request.walletId}</p>
                  <p className="text-sm text-gray-500">
                    Requesting upgrade to: <span className="font-medium">{request.requestedTier}</span>
                  </p>
                  <p className="text-xs text-gray-400 mt-1">
                    Submitted: {new Date(request.submittedAt).toLocaleString()}
                  </p>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => processKyc(request.id, "approve")}
                    disabled={loading}
                    className="px-4 py-2 bg-green-600 text-white rounded-lg text-sm hover:bg-green-700 disabled:opacity-50"
                  >
                    Approve
                  </button>
                  <button
                    onClick={() => processKyc(request.id, "reject")}
                    disabled={loading}
                    className="px-4 py-2 border border-red-600 text-red-600 rounded-lg text-sm hover:bg-red-50 disabled:opacity-50"
                  >
                    Reject
                  </button>
                </div>
              </div>
              <div className="mt-4">
                <p className="text-sm font-medium text-gray-600 mb-2">Submitted Documents:</p>
                <div className="flex gap-2 flex-wrap">
                  {request.documents.map((doc, i) => (
                    <span key={i} className="px-3 py-1 bg-gray-100 rounded text-sm">
                      {doc}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          ))}
          {kycRequests.filter(k => k.status === "pending").length === 0 && (
            <div className="p-8 text-center text-gray-500">No pending KYC requests</div>
          )}
        </div>
      </div>

      {/* Processed KYC History */}
      <div className="bg-white rounded-xl shadow-sm border">
        <div className="p-4 border-b">
          <h3 className="font-semibold">Processed KYC History</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Customer</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Requested Tier</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Submitted</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {kycRequests.filter(k => k.status !== "pending").map((request) => (
                <tr key={request.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm">{request.customerName}</td>
                  <td className="px-4 py-3 text-sm">{request.requestedTier}</td>
                  <td className="px-4 py-3">
                    <span className={`text-xs px-2 py-1 rounded-full ${
                      request.status === "approved" ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"
                    }`}>
                      {request.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {new Date(request.submittedAt).toLocaleDateString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );

  const renderTransactions = () => (
    <div className="bg-white rounded-xl shadow-sm border">
      <div className="p-4 border-b flex items-center justify-between">
        <h3 className="font-semibold">All Transactions</h3>
        <button onClick={fetchTransactions} className="text-sm text-green-600 hover:underline">
          Refresh
        </button>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">From</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">To</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Time</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {transactions.map((tx) => (
              <tr key={tx.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-sm font-mono">{tx.id.slice(0, 8)}...</td>
                <td className="px-4 py-3">
                  <span className={`text-xs px-2 py-1 rounded ${
                    tx.type === "issuance" ? "bg-purple-100 text-purple-700" :
                    tx.type === "deposit" ? "bg-green-100 text-green-700" :
                    tx.type === "withdrawal" ? "bg-orange-100 text-orange-700" :
                    "bg-blue-100 text-blue-700"
                  }`}>
                    {tx.type}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm">{tx.from.slice(0, 12)}...</td>
                <td className="px-4 py-3 text-sm">{tx.to.slice(0, 12)}...</td>
                <td className="px-4 py-3 text-sm font-medium">{formatCurrency(tx.amount)}</td>
                <td className="px-4 py-3">
                  <span className={`text-xs px-2 py-1 rounded-full ${
                    tx.status === "completed" ? "bg-green-100 text-green-700" :
                    tx.status === "pending" ? "bg-yellow-100 text-yellow-700" :
                    "bg-red-100 text-red-700"
                  }`}>
                    {tx.status}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm text-gray-500">
                  {new Date(tx.timestamp).toLocaleString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {transactions.length === 0 && (
          <div className="p-8 text-center text-gray-500">No transactions found</div>
        )}
      </div>
    </div>
  );

  const renderLiquidity = () => (
    <div className="space-y-6">
      <div className="grid md:grid-cols-2 gap-6">
        {/* Current Liquidity */}
        <div className="bg-white rounded-xl shadow-sm border p-6">
          <h3 className="font-semibold mb-4">Current Liquidity Position</h3>
          <div className="text-center py-8">
            <p className="text-4xl font-bold text-blue-600">{formatCurrency(liquidityBalance)}</p>
            <p className="text-gray-500 mt-2">Available eNaira Balance</p>
          </div>
          <div className="mt-4 pt-4 border-t">
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">Daily Limit</span>
              <span className="font-medium">{formatCurrency(100000000)}</span>
            </div>
            <div className="flex justify-between text-sm mt-2">
              <span className="text-gray-500">Used Today</span>
              <span className="font-medium">{formatCurrency(stats.todayVolume)}</span>
            </div>
          </div>
        </div>

        {/* Request Liquidity */}
        <div className="bg-white rounded-xl shadow-sm border p-6">
          <h3 className="font-semibold mb-4">Request Additional Liquidity</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm text-gray-600 mb-1">Amount (kobo)</label>
              <input
                type="number"
                className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-green-500"
                placeholder="Enter amount to request"
              />
            </div>
            <div>
              <label className="block text-sm text-gray-600 mb-1">Justification</label>
              <textarea
                className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-green-500"
                rows={3}
                placeholder="Reason for liquidity request"
              />
            </div>
            <button className="w-full py-3 bg-green-600 text-white rounded-lg font-medium hover:bg-green-700">
              Submit Request to CBN
            </button>
          </div>
        </div>
      </div>

      {/* Liquidity History */}
      <div className="bg-white rounded-xl shadow-sm border">
        <div className="p-4 border-b">
          <h3 className="font-semibold">Liquidity History</h3>
        </div>
        <div className="p-6 text-center text-gray-500">
          No liquidity requests yet
        </div>
      </div>
    </div>
  );

  const renderReports = () => (
    <div className="space-y-6">
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
        {[
          { name: "Daily Transaction Report", desc: "All transactions for a specific date", icon: "D" },
          { name: "Monthly Summary", desc: "Aggregated monthly statistics", icon: "M" },
          { name: "KYC Compliance Report", desc: "Customer verification status", icon: "K" },
          { name: "AML/CTF Report", desc: "Suspicious activity monitoring", icon: "A" },
          { name: "Wallet Distribution", desc: "Breakdown by tier and status", icon: "W" },
          { name: "Settlement Report", desc: "Daily settlement reconciliation", icon: "S" },
        ].map((report) => (
          <div key={report.name} className="bg-white rounded-xl shadow-sm border p-6">
            <div className="w-12 h-12 bg-gray-100 rounded-lg flex items-center justify-center mb-4">
              <span className="text-xl font-bold text-gray-600">{report.icon}</span>
            </div>
            <h4 className="font-medium">{report.name}</h4>
            <p className="text-sm text-gray-500 mt-1">{report.desc}</p>
            <button className="mt-4 text-sm text-green-600 hover:underline">
              Generate Report
            </button>
          </div>
        ))}
      </div>

      {/* Scheduled Reports */}
      <div className="bg-white rounded-xl shadow-sm border p-6">
        <h3 className="font-semibold mb-4">Scheduled Reports</h3>
        <p className="text-gray-500 text-sm">
          Configure automated report generation and delivery to your email.
        </p>
        <button className="mt-4 px-4 py-2 border rounded-lg hover:bg-gray-50">
          Configure Scheduled Reports
        </button>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold">
              eN
            </div>
            <div>
              <h1 className="font-bold text-lg">eNaira Intermediary Portal</h1>
              <p className="text-xs text-gray-500">{intermediaryName}</p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">Liquidity: {formatCurrency(liquidityBalance)}</span>
            <button className="p-2 hover:bg-gray-100 rounded-full">
              <span className="sr-only">Settings</span>
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </button>
          </div>
        </div>
      </header>

      {/* Navigation */}
      <nav className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4">
          <div className="flex gap-6 overflow-x-auto">
            {[
              { id: "dashboard", label: "Dashboard" },
              { id: "wallets", label: "Wallets" },
              { id: "kyc", label: "KYC Verification" },
              { id: "transactions", label: "Transactions" },
              { id: "liquidity", label: "Liquidity" },
              { id: "reports", label: "Reports" },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as typeof activeTab)}
                className={`py-4 px-2 border-b-2 font-medium text-sm whitespace-nowrap ${
                  activeTab === tab.id
                    ? "border-blue-600 text-blue-600"
                    : "border-transparent text-gray-500 hover:text-gray-700"
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-6">
        {activeTab === "dashboard" && renderDashboard()}
        {activeTab === "wallets" && renderWallets()}
        {activeTab === "kyc" && renderKyc()}
        {activeTab === "transactions" && renderTransactions()}
        {activeTab === "liquidity" && renderLiquidity()}
        {activeTab === "reports" && renderReports()}
      </main>

      {/* Notification */}
      {notification && (
        <div className={`fixed bottom-4 right-4 px-4 py-3 rounded-lg shadow-lg ${
          notification.type === "success" ? "bg-green-600" : "bg-red-600"
        } text-white`}>
          {notification.message}
          <button onClick={() => setNotification(null)} className="ml-3 font-bold">x</button>
        </div>
      )}
    </div>
  );
}
