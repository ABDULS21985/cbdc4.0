"use client";

import { useState, useEffect } from "react";

// Types
interface Transaction {
  id: string;
  amount: number;
  customerWallet: string;
  timestamp: string;
  status: "completed" | "pending" | "failed";
  type: "payment" | "refund";
}

interface MerchantStats {
  todayVolume: number;
  todayTransactions: number;
  weekVolume: number;
  pendingSettlement: number;
}

interface QRPaymentRequest {
  amount: number;
  reference: string;
  description: string;
}

// API Configuration
const PAYMENTS_SERVICE_URL = process.env.NEXT_PUBLIC_PAYMENTS_SERVICE_URL || "http://localhost:8082";
const WALLET_SERVICE_URL = process.env.NEXT_PUBLIC_WALLET_SERVICE_URL || "http://localhost:8081";

export default function MerchantPortal() {
  const [activeTab, setActiveTab] = useState<"dashboard" | "receive" | "transactions" | "settlements" | "settings">("dashboard");
  const [merchantId] = useState("MERCHANT_001");
  const [merchantName] = useState("Sample Merchant Store");
  const [balance, setBalance] = useState(0);
  const [stats, setStats] = useState<MerchantStats>({
    todayVolume: 0,
    todayTransactions: 0,
    weekVolume: 0,
    pendingSettlement: 0,
  });
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [qrRequest, setQrRequest] = useState<QRPaymentRequest>({
    amount: 0,
    reference: "",
    description: "",
  });
  const [generatedQR, setGeneratedQR] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [notification, setNotification] = useState<{ type: "success" | "error"; message: string } | null>(null);

  // Fetch merchant data on mount
  useEffect(() => {
    fetchMerchantData();
    fetchTransactions();
    // Poll for new payments every 10 seconds
    const interval = setInterval(fetchTransactions, 10000);
    return () => clearInterval(interval);
  }, []);

  const fetchMerchantData = async () => {
    try {
      const response = await fetch(`${WALLET_SERVICE_URL}/wallet/${merchantId}`);
      if (response.ok) {
        const data = await response.json();
        setBalance(data.balance || 0);
      }
    } catch (error) {
      console.error("Failed to fetch merchant data:", error);
    }
  };

  const fetchTransactions = async () => {
    try {
      const response = await fetch(`${PAYMENTS_SERVICE_URL}/payments/history/${merchantId}`);
      if (response.ok) {
        const data = await response.json();
        setTransactions(data.transactions || []);

        // Calculate stats
        const now = new Date();
        const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
        const weekStart = new Date(todayStart.getTime() - 7 * 24 * 60 * 60 * 1000);

        const todayTxs = (data.transactions || []).filter(
          (tx: Transaction) => new Date(tx.timestamp) >= todayStart
        );
        const weekTxs = (data.transactions || []).filter(
          (tx: Transaction) => new Date(tx.timestamp) >= weekStart
        );

        setStats({
          todayVolume: todayTxs.reduce((sum: number, tx: Transaction) => sum + tx.amount, 0),
          todayTransactions: todayTxs.length,
          weekVolume: weekTxs.reduce((sum: number, tx: Transaction) => sum + tx.amount, 0),
          pendingSettlement: (data.transactions || [])
            .filter((tx: Transaction) => tx.status === "pending")
            .reduce((sum: number, tx: Transaction) => sum + tx.amount, 0),
        });
      }
    } catch (error) {
      console.error("Failed to fetch transactions:", error);
    }
  };

  const generatePaymentQR = async () => {
    if (qrRequest.amount <= 0) {
      setNotification({ type: "error", message: "Please enter a valid amount" });
      return;
    }

    setLoading(true);
    try {
      const response = await fetch(`${PAYMENTS_SERVICE_URL}/payments/qr/generate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          merchantId,
          amount: qrRequest.amount,
          reference: qrRequest.reference || `REF-${Date.now()}`,
          description: qrRequest.description,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        setGeneratedQR(data.qrCode);
        setNotification({ type: "success", message: "Payment QR generated successfully" });
      } else {
        throw new Error("Failed to generate QR");
      }
    } catch (error) {
      setNotification({ type: "error", message: "Failed to generate payment QR" });
    } finally {
      setLoading(false);
    }
  };

  const initiateRefund = async (transactionId: string, amount: number) => {
    setLoading(true);
    try {
      const response = await fetch(`${PAYMENTS_SERVICE_URL}/payments/refund`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          originalTransactionId: transactionId,
          merchantId,
          amount,
        }),
      });

      if (response.ok) {
        setNotification({ type: "success", message: "Refund initiated successfully" });
        fetchTransactions();
      } else {
        throw new Error("Refund failed");
      }
    } catch (error) {
      setNotification({ type: "error", message: "Failed to process refund" });
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
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-white rounded-xl p-6 shadow-sm border">
          <p className="text-gray-500 text-sm">Available Balance</p>
          <p className="text-2xl font-bold text-green-600">{formatCurrency(balance)}</p>
        </div>
        <div className="bg-white rounded-xl p-6 shadow-sm border">
          <p className="text-gray-500 text-sm">Today&apos;s Volume</p>
          <p className="text-2xl font-bold">{formatCurrency(stats.todayVolume)}</p>
          <p className="text-xs text-gray-400">{stats.todayTransactions} transactions</p>
        </div>
        <div className="bg-white rounded-xl p-6 shadow-sm border">
          <p className="text-gray-500 text-sm">This Week</p>
          <p className="text-2xl font-bold">{formatCurrency(stats.weekVolume)}</p>
        </div>
        <div className="bg-white rounded-xl p-6 shadow-sm border">
          <p className="text-gray-500 text-sm">Pending Settlement</p>
          <p className="text-2xl font-bold text-orange-500">{formatCurrency(stats.pendingSettlement)}</p>
        </div>
      </div>

      {/* Recent Transactions */}
      <div className="bg-white rounded-xl shadow-sm border">
        <div className="p-4 border-b">
          <h3 className="font-semibold">Recent Payments</h3>
        </div>
        <div className="divide-y max-h-96 overflow-y-auto">
          {transactions.slice(0, 10).map((tx) => (
            <div key={tx.id} className="p-4 flex items-center justify-between hover:bg-gray-50">
              <div className="flex items-center gap-3">
                <div className={`w-10 h-10 rounded-full flex items-center justify-center ${
                  tx.type === "payment" ? "bg-green-100" : "bg-red-100"
                }`}>
                  {tx.type === "payment" ? "+" : "-"}
                </div>
                <div>
                  <p className="font-medium">{tx.customerWallet.slice(0, 8)}...{tx.customerWallet.slice(-4)}</p>
                  <p className="text-xs text-gray-500">{new Date(tx.timestamp).toLocaleString()}</p>
                </div>
              </div>
              <div className="text-right">
                <p className={`font-semibold ${tx.type === "payment" ? "text-green-600" : "text-red-600"}`}>
                  {tx.type === "payment" ? "+" : "-"}{formatCurrency(tx.amount)}
                </p>
                <span className={`text-xs px-2 py-1 rounded-full ${
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
  );

  const renderReceivePayment = () => (
    <div className="max-w-xl mx-auto space-y-6">
      <div className="bg-white rounded-xl shadow-sm border p-6">
        <h3 className="font-semibold mb-4">Generate Payment Request</h3>

        <div className="space-y-4">
          <div>
            <label className="block text-sm text-gray-600 mb-1">Amount (kobo)</label>
            <input
              type="number"
              value={qrRequest.amount || ""}
              onChange={(e) => setQrRequest({ ...qrRequest, amount: parseInt(e.target.value) || 0 })}
              className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-green-500"
              placeholder="Enter amount"
            />
            <p className="text-xs text-gray-500 mt-1">{formatCurrency(qrRequest.amount)}</p>
          </div>

          <div>
            <label className="block text-sm text-gray-600 mb-1">Reference (optional)</label>
            <input
              type="text"
              value={qrRequest.reference}
              onChange={(e) => setQrRequest({ ...qrRequest, reference: e.target.value })}
              className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-green-500"
              placeholder="Order ID or reference"
            />
          </div>

          <div>
            <label className="block text-sm text-gray-600 mb-1">Description (optional)</label>
            <input
              type="text"
              value={qrRequest.description}
              onChange={(e) => setQrRequest({ ...qrRequest, description: e.target.value })}
              className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-green-500"
              placeholder="Payment description"
            />
          </div>

          <button
            onClick={generatePaymentQR}
            disabled={loading || qrRequest.amount <= 0}
            className="w-full py-3 bg-green-600 text-white rounded-lg font-medium hover:bg-green-700 disabled:opacity-50"
          >
            {loading ? "Generating..." : "Generate QR Code"}
          </button>
        </div>
      </div>

      {generatedQR && (
        <div className="bg-white rounded-xl shadow-sm border p-6 text-center">
          <h3 className="font-semibold mb-4">Scan to Pay</h3>
          <div className="bg-gray-100 p-4 rounded-lg inline-block">
            {/* QR Code would be rendered here - using placeholder */}
            <div className="w-48 h-48 bg-white border-2 border-dashed border-gray-300 flex items-center justify-center">
              <span className="text-gray-400 text-sm">QR: {generatedQR.slice(0, 20)}...</span>
            </div>
          </div>
          <p className="mt-4 text-lg font-semibold">{formatCurrency(qrRequest.amount)}</p>
          <p className="text-sm text-gray-500">{qrRequest.description || "Payment Request"}</p>
        </div>
      )}
    </div>
  );

  const renderTransactions = () => (
    <div className="bg-white rounded-xl shadow-sm border">
      <div className="p-4 border-b flex items-center justify-between">
        <h3 className="font-semibold">All Transactions</h3>
        <button
          onClick={fetchTransactions}
          className="text-sm text-green-600 hover:underline"
        >
          Refresh
        </button>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Customer</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Time</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {transactions.map((tx) => (
              <tr key={tx.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-sm font-mono">{tx.id.slice(0, 8)}...</td>
                <td className="px-4 py-3 text-sm">{tx.customerWallet.slice(0, 12)}...</td>
                <td className="px-4 py-3">
                  <span className={`text-xs px-2 py-1 rounded ${
                    tx.type === "payment" ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"
                  }`}>
                    {tx.type}
                  </span>
                </td>
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
                <td className="px-4 py-3">
                  {tx.type === "payment" && tx.status === "completed" && (
                    <button
                      onClick={() => initiateRefund(tx.id, tx.amount)}
                      className="text-xs text-red-600 hover:underline"
                    >
                      Refund
                    </button>
                  )}
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

  const renderSettlements = () => (
    <div className="space-y-6">
      <div className="bg-white rounded-xl shadow-sm border p-6">
        <h3 className="font-semibold mb-4">Settlement Schedule</h3>
        <p className="text-gray-600 mb-4">
          Settlements are processed automatically every business day at 2:00 PM WAT.
        </p>

        <div className="grid grid-cols-2 gap-4">
          <div className="bg-gray-50 rounded-lg p-4">
            <p className="text-sm text-gray-500">Next Settlement</p>
            <p className="text-lg font-semibold">Tomorrow, 2:00 PM</p>
          </div>
          <div className="bg-gray-50 rounded-lg p-4">
            <p className="text-sm text-gray-500">Pending Amount</p>
            <p className="text-lg font-semibold text-orange-600">{formatCurrency(stats.pendingSettlement)}</p>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm border p-6">
        <h3 className="font-semibold mb-4">Bank Account Details</h3>
        <div className="space-y-3">
          <div className="flex justify-between">
            <span className="text-gray-500">Bank</span>
            <span className="font-medium">First Bank Nigeria</span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-500">Account Number</span>
            <span className="font-medium">****4521</span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-500">Account Name</span>
            <span className="font-medium">{merchantName}</span>
          </div>
        </div>
      </div>
    </div>
  );

  const renderSettings = () => (
    <div className="max-w-xl mx-auto space-y-6">
      <div className="bg-white rounded-xl shadow-sm border p-6">
        <h3 className="font-semibold mb-4">Business Information</h3>
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-gray-600 mb-1">Business Name</label>
            <input
              type="text"
              value={merchantName}
              disabled
              className="w-full px-4 py-3 border rounded-lg bg-gray-50"
            />
          </div>
          <div>
            <label className="block text-sm text-gray-600 mb-1">Merchant ID</label>
            <input
              type="text"
              value={merchantId}
              disabled
              className="w-full px-4 py-3 border rounded-lg bg-gray-50 font-mono"
            />
          </div>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm border p-6">
        <h3 className="font-semibold mb-4">Notification Preferences</h3>
        <div className="space-y-3">
          <label className="flex items-center gap-3">
            <input type="checkbox" defaultChecked className="w-4 h-4" />
            <span>Email notifications for payments</span>
          </label>
          <label className="flex items-center gap-3">
            <input type="checkbox" defaultChecked className="w-4 h-4" />
            <span>SMS notifications for large payments (&gt; N100,000)</span>
          </label>
          <label className="flex items-center gap-3">
            <input type="checkbox" defaultChecked className="w-4 h-4" />
            <span>Daily settlement reports</span>
          </label>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm border p-6">
        <h3 className="font-semibold mb-4">API Integration</h3>
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-gray-600 mb-1">API Key</label>
            <div className="flex gap-2">
              <input
                type="password"
                value="api_key_hidden_xxxxx"
                disabled
                className="flex-1 px-4 py-3 border rounded-lg bg-gray-50 font-mono"
              />
              <button className="px-4 py-2 border rounded-lg hover:bg-gray-50">
                Show
              </button>
            </div>
          </div>
          <div>
            <label className="block text-sm text-gray-600 mb-1">Webhook URL</label>
            <input
              type="text"
              placeholder="https://your-server.com/webhook"
              className="w-full px-4 py-3 border rounded-lg"
            />
          </div>
          <button className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700">
            Save Webhook
          </button>
        </div>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-green-600 rounded-lg flex items-center justify-center text-white font-bold">
              eN
            </div>
            <div>
              <h1 className="font-bold text-lg">eNaira Merchant</h1>
              <p className="text-xs text-gray-500">{merchantName}</p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">Balance: {formatCurrency(balance)}</span>
            <button className="p-2 hover:bg-gray-100 rounded-full">
              <span className="sr-only">Notifications</span>
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
              </svg>
            </button>
          </div>
        </div>
      </header>

      {/* Navigation Tabs */}
      <nav className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4">
          <div className="flex gap-6 overflow-x-auto">
            {[
              { id: "dashboard", label: "Dashboard" },
              { id: "receive", label: "Receive Payment" },
              { id: "transactions", label: "Transactions" },
              { id: "settlements", label: "Settlements" },
              { id: "settings", label: "Settings" },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as typeof activeTab)}
                className={`py-4 px-2 border-b-2 font-medium text-sm whitespace-nowrap ${
                  activeTab === tab.id
                    ? "border-green-600 text-green-600"
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
        {activeTab === "receive" && renderReceivePayment()}
        {activeTab === "transactions" && renderTransactions()}
        {activeTab === "settlements" && renderSettlements()}
        {activeTab === "settings" && renderSettings()}
      </main>

      {/* Notification Toast */}
      {notification && (
        <div className={`fixed bottom-4 right-4 px-4 py-3 rounded-lg shadow-lg ${
          notification.type === "success" ? "bg-green-600" : "bg-red-600"
        } text-white`}>
          {notification.message}
          <button
            onClick={() => setNotification(null)}
            className="ml-3 font-bold"
          >
            x
          </button>
        </div>
      )}
    </div>
  );
}
