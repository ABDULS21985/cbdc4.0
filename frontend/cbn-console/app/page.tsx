"use client";

import React, { useState, useEffect } from 'react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8087';

interface DashboardStats {
    total_supply: number;
    circulating_supply: number;
    total_transactions_24h: number;
    total_volume_24h: number;
    active_wallets: number;
    intermediaries: Intermediary[];
    last_updated: string;
}

interface Intermediary {
    id: string;
    name: string;
    status: string;
    cbdc_balance: number;
    customer_count: number;
}

interface Transaction {
    id: string;
    from: string;
    to: string;
    amount: number;
    type: string;
    status: string;
    created_at: string;
}

export default function CBNConsole() {
    const [activeTab, setActiveTab] = useState<'dashboard' | 'issuance' | 'intermediaries' | 'governance' | 'audit'>('dashboard');
    const [stats, setStats] = useState<DashboardStats | null>(null);
    const [loading, setLoading] = useState(false);
    const [issueAmount, setIssueAmount] = useState('');
    const [issueTarget, setIssueTarget] = useState('');
    const [redeemAmount, setRedeemAmount] = useState('');
    const [redeemFrom, setRedeemFrom] = useState('');
    const [freezeWallet, setFreezeWallet] = useState('');
    const [transactions, setTransactions] = useState<Transaction[]>([]);
    const [notification, setNotification] = useState<{type: 'success' | 'error', message: string} | null>(null);

    useEffect(() => {
        fetchDashboard();
    }, []);

    const fetchDashboard = async () => {
        setLoading(true);
        try {
            const res = await fetch(`${API_BASE}/ops/dashboard`);
            if (res.ok) {
                const data = await res.json();
                setStats(data.data || data);
            }
        } catch (err) {
            console.error("Failed to fetch dashboard", err);
        } finally {
            setLoading(false);
        }
    };

    const fetchTransactions = async () => {
        try {
            const res = await fetch(`${API_BASE}/ops/audit/transactions`);
            if (res.ok) {
                const data = await res.json();
                setTransactions(data.data || []);
            }
        } catch (err) {
            console.error("Failed to fetch transactions", err);
        }
    };

    const handleIssue = async () => {
        if (!issueAmount || !issueTarget) {
            showNotification('error', 'Please enter amount and target intermediary');
            return;
        }

        try {
            const res = await fetch(`${API_BASE}/ops/issue`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    amount: parseInt(issueAmount),
                    to_intermediary_id: issueTarget,
                    reason: 'Manual issuance',
                    approved_by: 'admin'
                })
            });

            if (res.ok) {
                showNotification('success', `Successfully issued ${issueAmount} CBDC to ${issueTarget}`);
                setIssueAmount('');
                setIssueTarget('');
                fetchDashboard();
            } else {
                showNotification('error', 'Issuance failed');
            }
        } catch (err) {
            showNotification('error', 'Network error');
        }
    };

    const handleRedeem = async () => {
        if (!redeemAmount || !redeemFrom) {
            showNotification('error', 'Please enter amount and source intermediary');
            return;
        }

        try {
            const res = await fetch(`${API_BASE}/ops/redeem`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    amount: parseInt(redeemAmount),
                    from_intermediary_id: redeemFrom,
                    reason: 'Manual redemption',
                    approved_by: 'admin'
                })
            });

            if (res.ok) {
                showNotification('success', `Successfully redeemed ${redeemAmount} CBDC from ${redeemFrom}`);
                setRedeemAmount('');
                setRedeemFrom('');
                fetchDashboard();
            } else {
                showNotification('error', 'Redemption failed');
            }
        } catch (err) {
            showNotification('error', 'Network error');
        }
    };

    const handleFreeze = async () => {
        if (!freezeWallet) {
            showNotification('error', 'Please enter wallet ID');
            return;
        }

        try {
            const res = await fetch(`${API_BASE}/ops/freeze`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    wallet_id: freezeWallet,
                    reason: 'Manual freeze'
                })
            });

            if (res.ok) {
                showNotification('success', `Wallet ${freezeWallet} frozen`);
                setFreezeWallet('');
            } else {
                showNotification('error', 'Freeze failed');
            }
        } catch (err) {
            showNotification('error', 'Network error');
        }
    };

    const showNotification = (type: 'success' | 'error', message: string) => {
        setNotification({ type, message });
        setTimeout(() => setNotification(null), 3000);
    };

    const formatNumber = (num: number) => {
        return new Intl.NumberFormat('en-NG', {
            style: 'currency',
            currency: 'NGN',
            minimumFractionDigits: 0
        }).format(num);
    };

    return (
        <div className="min-h-screen bg-gray-100">
            {/* Notification */}
            {notification && (
                <div className={`fixed top-4 right-4 px-6 py-3 rounded-lg shadow-lg z-50 ${
                    notification.type === 'success' ? 'bg-green-500 text-white' : 'bg-red-500 text-white'
                }`}>
                    {notification.message}
                </div>
            )}

            {/* Header */}
            <header className="bg-green-800 text-white shadow-lg">
                <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
                    <div className="flex items-center space-x-4">
                        <div className="text-2xl font-bold">CBN</div>
                        <div>
                            <h1 className="text-xl font-semibold">Central Bank Operations Console</h1>
                            <p className="text-xs text-green-200">eNaira Digital Currency Platform</p>
                        </div>
                    </div>
                    <div className="flex items-center space-x-4">
                        <span className="text-sm">Admin User</span>
                        <button className="bg-green-700 px-4 py-2 rounded hover:bg-green-600">Logout</button>
                    </div>
                </div>
            </header>

            {/* Navigation */}
            <nav className="bg-white shadow">
                <div className="max-w-7xl mx-auto px-4">
                    <div className="flex space-x-8">
                        {(['dashboard', 'issuance', 'intermediaries', 'governance', 'audit'] as const).map(tab => (
                            <button
                                key={tab}
                                onClick={() => {
                                    setActiveTab(tab);
                                    if (tab === 'audit') fetchTransactions();
                                }}
                                className={`py-4 px-2 border-b-2 font-medium text-sm capitalize ${
                                    activeTab === tab
                                        ? 'border-green-500 text-green-600'
                                        : 'border-transparent text-gray-500 hover:text-gray-700'
                                }`}
                            >
                                {tab}
                            </button>
                        ))}
                    </div>
                </div>
            </nav>

            {/* Main Content */}
            <main className="max-w-7xl mx-auto px-4 py-8">
                {activeTab === 'dashboard' && (
                    <div className="space-y-6">
                        {/* Stats Grid */}
                        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                            <div className="bg-white rounded-lg shadow p-6">
                                <p className="text-sm text-gray-500">Total Supply</p>
                                <p className="text-3xl font-bold text-green-600">
                                    {stats ? formatNumber(stats.total_supply) : '...'}
                                </p>
                            </div>
                            <div className="bg-white rounded-lg shadow p-6">
                                <p className="text-sm text-gray-500">24h Volume</p>
                                <p className="text-3xl font-bold text-blue-600">
                                    {stats ? formatNumber(stats.total_volume_24h) : '...'}
                                </p>
                            </div>
                            <div className="bg-white rounded-lg shadow p-6">
                                <p className="text-sm text-gray-500">24h Transactions</p>
                                <p className="text-3xl font-bold text-purple-600">
                                    {stats?.total_transactions_24h?.toLocaleString() || '...'}
                                </p>
                            </div>
                            <div className="bg-white rounded-lg shadow p-6">
                                <p className="text-sm text-gray-500">Active Wallets</p>
                                <p className="text-3xl font-bold text-orange-600">
                                    {stats?.active_wallets?.toLocaleString() || '...'}
                                </p>
                            </div>
                        </div>

                        {/* Intermediaries Table */}
                        <div className="bg-white rounded-lg shadow">
                            <div className="px-6 py-4 border-b">
                                <h2 className="text-lg font-semibold">Intermediary Positions</h2>
                            </div>
                            <table className="w-full">
                                <thead className="bg-gray-50">
                                    <tr>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">CBDC Balance</th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Customers</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y">
                                    {stats?.intermediaries?.map(int => (
                                        <tr key={int.id}>
                                            <td className="px-6 py-4">{int.name}</td>
                                            <td className="px-6 py-4">
                                                <span className={`px-2 py-1 rounded text-xs ${
                                                    int.status === 'ACTIVE' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                                                }`}>
                                                    {int.status}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4">{formatNumber(int.cbdc_balance)}</td>
                                            <td className="px-6 py-4">{int.customer_count?.toLocaleString()}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}

                {activeTab === 'issuance' && (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {/* Issue Card */}
                        <div className="bg-white rounded-lg shadow p-6">
                            <h2 className="text-xl font-semibold mb-4 text-green-700">Issue CBDC (Mint)</h2>
                            <p className="text-sm text-gray-500 mb-4">Mint new CBDC to an intermediary</p>
                            <div className="space-y-4">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">Amount</label>
                                    <input
                                        type="number"
                                        value={issueAmount}
                                        onChange={(e) => setIssueAmount(e.target.value)}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-green-500 focus:ring-green-500 p-2 border"
                                        placeholder="Enter amount"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">Target Intermediary</label>
                                    <select
                                        value={issueTarget}
                                        onChange={(e) => setIssueTarget(e.target.value)}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-green-500 focus:ring-green-500 p-2 border"
                                    >
                                        <option value="">Select intermediary</option>
                                        <option value="bank-a">Bank A</option>
                                        <option value="bank-b">Bank B</option>
                                        <option value="fintech-x">FintechX</option>
                                    </select>
                                </div>
                                <button
                                    onClick={handleIssue}
                                    className="w-full bg-green-600 text-white py-2 rounded hover:bg-green-700"
                                >
                                    Issue CBDC
                                </button>
                            </div>
                        </div>

                        {/* Redeem Card */}
                        <div className="bg-white rounded-lg shadow p-6">
                            <h2 className="text-xl font-semibold mb-4 text-red-700">Redeem CBDC (Burn)</h2>
                            <p className="text-sm text-gray-500 mb-4">Burn CBDC from an intermediary</p>
                            <div className="space-y-4">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">Amount</label>
                                    <input
                                        type="number"
                                        value={redeemAmount}
                                        onChange={(e) => setRedeemAmount(e.target.value)}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-red-500 focus:ring-red-500 p-2 border"
                                        placeholder="Enter amount"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">From Intermediary</label>
                                    <select
                                        value={redeemFrom}
                                        onChange={(e) => setRedeemFrom(e.target.value)}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-red-500 focus:ring-red-500 p-2 border"
                                    >
                                        <option value="">Select intermediary</option>
                                        <option value="bank-a">Bank A</option>
                                        <option value="bank-b">Bank B</option>
                                        <option value="fintech-x">FintechX</option>
                                    </select>
                                </div>
                                <button
                                    onClick={handleRedeem}
                                    className="w-full bg-red-600 text-white py-2 rounded hover:bg-red-700"
                                >
                                    Redeem CBDC
                                </button>
                            </div>
                        </div>
                    </div>
                )}

                {activeTab === 'intermediaries' && (
                    <div className="bg-white rounded-lg shadow">
                        <div className="px-6 py-4 border-b flex justify-between items-center">
                            <h2 className="text-lg font-semibold">Registered Intermediaries</h2>
                            <button className="bg-green-600 text-white px-4 py-2 rounded text-sm">+ Add New</button>
                        </div>
                        <table className="w-full">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y">
                                <tr>
                                    <td className="px-6 py-4">bank-a</td>
                                    <td className="px-6 py-4">Bank A</td>
                                    <td className="px-6 py-4">Commercial Bank</td>
                                    <td className="px-6 py-4"><span className="bg-green-100 text-green-800 px-2 py-1 rounded text-xs">ACTIVE</span></td>
                                    <td className="px-6 py-4">
                                        <button className="text-blue-600 hover:underline text-sm mr-2">Edit</button>
                                        <button className="text-red-600 hover:underline text-sm">Suspend</button>
                                    </td>
                                </tr>
                                <tr>
                                    <td className="px-6 py-4">bank-b</td>
                                    <td className="px-6 py-4">Bank B</td>
                                    <td className="px-6 py-4">Commercial Bank</td>
                                    <td className="px-6 py-4"><span className="bg-green-100 text-green-800 px-2 py-1 rounded text-xs">ACTIVE</span></td>
                                    <td className="px-6 py-4">
                                        <button className="text-blue-600 hover:underline text-sm mr-2">Edit</button>
                                        <button className="text-red-600 hover:underline text-sm">Suspend</button>
                                    </td>
                                </tr>
                                <tr>
                                    <td className="px-6 py-4">fintech-x</td>
                                    <td className="px-6 py-4">FintechX</td>
                                    <td className="px-6 py-4">Payment Service Provider</td>
                                    <td className="px-6 py-4"><span className="bg-green-100 text-green-800 px-2 py-1 rounded text-xs">ACTIVE</span></td>
                                    <td className="px-6 py-4">
                                        <button className="text-blue-600 hover:underline text-sm mr-2">Edit</button>
                                        <button className="text-red-600 hover:underline text-sm">Suspend</button>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                )}

                {activeTab === 'governance' && (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {/* Wallet Freeze */}
                        <div className="bg-white rounded-lg shadow p-6">
                            <h2 className="text-xl font-semibold mb-4">Freeze Wallet</h2>
                            <div className="space-y-4">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">Wallet ID</label>
                                    <input
                                        type="text"
                                        value={freezeWallet}
                                        onChange={(e) => setFreezeWallet(e.target.value)}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm p-2 border"
                                        placeholder="wallet-xxx-yyy"
                                    />
                                </div>
                                <button
                                    onClick={handleFreeze}
                                    className="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700"
                                >
                                    Freeze Wallet
                                </button>
                            </div>
                        </div>

                        {/* Scheme Parameters */}
                        <div className="bg-white rounded-lg shadow p-6">
                            <h2 className="text-xl font-semibold mb-4">Scheme Parameters</h2>
                            <div className="space-y-3 text-sm">
                                <div className="flex justify-between py-2 border-b">
                                    <span className="text-gray-600">Tier 0 Daily Limit</span>
                                    <span className="font-medium">10,000</span>
                                </div>
                                <div className="flex justify-between py-2 border-b">
                                    <span className="text-gray-600">Tier 1 Daily Limit</span>
                                    <span className="font-medium">100,000</span>
                                </div>
                                <div className="flex justify-between py-2 border-b">
                                    <span className="text-gray-600">Tier 2 Daily Limit</span>
                                    <span className="font-medium">1,000,000</span>
                                </div>
                                <div className="flex justify-between py-2 border-b">
                                    <span className="text-gray-600">Offline Max Balance</span>
                                    <span className="font-medium">500</span>
                                </div>
                                <div className="flex justify-between py-2 border-b">
                                    <span className="text-gray-600">Offline TX Limit</span>
                                    <span className="font-medium">50</span>
                                </div>
                                <div className="flex justify-between py-2">
                                    <span className="text-gray-600">Offline Sync TTL</span>
                                    <span className="font-medium">7 days</span>
                                </div>
                            </div>
                        </div>
                    </div>
                )}

                {activeTab === 'audit' && (
                    <div className="bg-white rounded-lg shadow">
                        <div className="px-6 py-4 border-b flex justify-between items-center">
                            <h2 className="text-lg font-semibold">Transaction Audit Log</h2>
                            <button className="bg-gray-600 text-white px-4 py-2 rounded text-sm">Export CSV</button>
                        </div>
                        <table className="w-full">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">TX ID</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">From</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">To</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y">
                                {transactions.map(tx => (
                                    <tr key={tx.id}>
                                        <td className="px-6 py-4 font-mono text-xs">{tx.id?.substring(0, 16)}...</td>
                                        <td className="px-6 py-4">{tx.from}</td>
                                        <td className="px-6 py-4">{tx.to}</td>
                                        <td className="px-6 py-4">{formatNumber(tx.amount)}</td>
                                        <td className="px-6 py-4">{tx.type}</td>
                                        <td className="px-6 py-4">
                                            <span className={`px-2 py-1 rounded text-xs ${
                                                tx.status === 'CONFIRMED' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
                                            }`}>
                                                {tx.status}
                                            </span>
                                        </td>
                                    </tr>
                                ))}
                                {transactions.length === 0 && (
                                    <tr>
                                        <td colSpan={6} className="px-6 py-8 text-center text-gray-500">No transactions found</td>
                                    </tr>
                                )}
                            </tbody>
                        </table>
                    </div>
                )}
            </main>
        </div>
    );
}
