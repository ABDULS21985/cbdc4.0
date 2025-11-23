"use client";

import React, { useState, useEffect } from 'react';
import { Button } from '@cbdc/ui/Button';
import { Card } from '@cbdc/ui/Card';
import { Input } from '@cbdc/ui/Input';

// API Base URL - would come from environment in production
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082';

interface Transaction {
    id: string;
    from_wallet: string;
    to_wallet: string;
    amount: number;
    type: string;
    status: string;
    created_at: string;
}

interface WalletData {
    id: string;
    balance: number;
    status: string;
    tier_level: string;
}

export default function Home() {
    const [balance, setBalance] = useState<number | null>(null);
    const [loading, setLoading] = useState(false);
    const [walletId] = useState('wallet-alice-123'); // Would come from auth
    const [recipient, setRecipient] = useState('');
    const [amount, setAmount] = useState('');
    const [sending, setSending] = useState(false);
    const [transactions, setTransactions] = useState<Transaction[]>([]);
    const [activeTab, setActiveTab] = useState<'home' | 'send' | 'scan' | 'history' | 'offline'>('home');
    const [offlineMode, setOfflineMode] = useState(false);
    const [offlineBalance, setOfflineBalance] = useState<number>(0);
    const [notification, setNotification] = useState<{type: 'success' | 'error', message: string} | null>(null);

    useEffect(() => {
        fetchBalance();
        fetchTransactions();
    }, []);

    const fetchBalance = async () => {
        setLoading(true);
        try {
            const res = await fetch(`${API_BASE}/wallets/${walletId}`);
            if (res.ok) {
                const data: WalletData = await res.json();
                setBalance(data.balance);
            }
        } catch (err) {
            console.error("Failed to fetch balance", err);
        } finally {
            setLoading(false);
        }
    };

    const fetchTransactions = async () => {
        try {
            const res = await fetch(`http://localhost:8083/payments/history?wallet_id=${walletId}`);
            if (res.ok) {
                const data = await res.json();
                setTransactions(data.data?.transactions || []);
            }
        } catch (err) {
            console.error("Failed to fetch transactions", err);
        }
    };

    const handleSend = async () => {
        if (!recipient || !amount) {
            showNotification('error', 'Please enter recipient and amount');
            return;
        }

        setSending(true);
        try {
            const res = await fetch('http://localhost:8083/payments', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    from_wallet: walletId,
                    to_wallet: recipient,
                    amount: parseInt(amount),
                    type: 'P2P',
                    channel: 'MOBILE'
                })
            });

            if (res.ok) {
                showNotification('success', `Successfully sent ₦${amount} to ${recipient}`);
                setRecipient('');
                setAmount('');
                fetchBalance();
                fetchTransactions();
            } else {
                const error = await res.json();
                showNotification('error', error.message || 'Transfer failed');
            }
        } catch (err) {
            showNotification('error', 'Network error. Please try again.');
        } finally {
            setSending(false);
        }
    };

    const handleLoadOffline = async () => {
        const offlineAmount = prompt('Enter amount to load offline (max 500):');
        if (!offlineAmount) return;

        const amt = parseInt(offlineAmount);
        if (amt > 500) {
            showNotification('error', 'Maximum offline balance is ₦500');
            return;
        }

        try {
            const res = await fetch('http://localhost:8080/offline/fund', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    user_id: 'alice-123',
                    device_id: 'dev-mobile-001',
                    amount: amt
                })
            });

            if (res.ok) {
                setOfflineBalance(prev => prev + amt);
                showNotification('success', `Loaded ₦${amt} for offline use`);
                fetchBalance();
            } else {
                showNotification('error', 'Failed to load offline balance');
            }
        } catch (err) {
            showNotification('error', 'Network error');
        }
    };

    const showNotification = (type: 'success' | 'error', message: string) => {
        setNotification({ type, message });
        setTimeout(() => setNotification(null), 3000);
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString('en-NG', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    return (
        <main className="flex min-h-screen flex-col bg-gray-50">
            {/* Notification */}
            {notification && (
                <div className={`fixed top-4 left-1/2 transform -translate-x-1/2 px-6 py-3 rounded-lg shadow-lg z-50 ${
                    notification.type === 'success' ? 'bg-green-500 text-white' : 'bg-red-500 text-white'
                }`}>
                    {notification.message}
                </div>
            )}

            {/* Header */}
            <header className="bg-blue-600 text-white p-4 shadow-lg">
                <div className="max-w-md mx-auto flex justify-between items-center">
                    <h1 className="text-xl font-bold">eNaira Wallet</h1>
                    <div className="flex items-center space-x-2">
                        {offlineMode && (
                            <span className="text-xs bg-yellow-500 px-2 py-1 rounded">OFFLINE</span>
                        )}
                    </div>
                </div>
            </header>

            {/* Main Content */}
            <div className="flex-1 max-w-md mx-auto w-full p-4">
                {activeTab === 'home' && (
                    <div className="space-y-4">
                        {/* Balance Card */}
                        <Card className="bg-gradient-to-r from-blue-600 to-blue-800 text-white">
                            <p className="text-sm opacity-80">Available Balance</p>
                            <div className="text-4xl font-bold my-2">
                                {loading ? '...' : `₦${(balance || 0).toLocaleString()}`}
                            </div>
                            <p className="text-xs opacity-60">Wallet ID: {walletId}</p>
                        </Card>

                        {/* Offline Balance */}
                        <Card className="bg-yellow-50 border border-yellow-200">
                            <div className="flex justify-between items-center">
                                <div>
                                    <p className="text-sm text-yellow-800">Offline Balance</p>
                                    <p className="text-2xl font-bold text-yellow-900">₦{offlineBalance}</p>
                                </div>
                                <Button onClick={handleLoadOffline} className="bg-yellow-500 hover:bg-yellow-600">
                                    Load Offline
                                </Button>
                            </div>
                        </Card>

                        {/* Quick Actions */}
                        <div className="grid grid-cols-3 gap-4">
                            <button
                                onClick={() => setActiveTab('send')}
                                className="bg-white p-4 rounded-lg shadow text-center hover:bg-gray-50"
                            >
                                <div className="text-2xl mb-1">💸</div>
                                <div className="text-sm">Send</div>
                            </button>
                            <button
                                onClick={() => setActiveTab('scan')}
                                className="bg-white p-4 rounded-lg shadow text-center hover:bg-gray-50"
                            >
                                <div className="text-2xl mb-1">📱</div>
                                <div className="text-sm">Scan QR</div>
                            </button>
                            <button
                                onClick={() => setActiveTab('history')}
                                className="bg-white p-4 rounded-lg shadow text-center hover:bg-gray-50"
                            >
                                <div className="text-2xl mb-1">📋</div>
                                <div className="text-sm">History</div>
                            </button>
                        </div>

                        {/* Recent Transactions */}
                        <Card>
                            <h3 className="font-semibold mb-3">Recent Transactions</h3>
                            <div className="space-y-2">
                                {transactions.slice(0, 5).map(tx => (
                                    <div key={tx.id} className="flex justify-between items-center py-2 border-b last:border-0">
                                        <div>
                                            <p className="text-sm font-medium">
                                                {tx.from_wallet === walletId ? `To: ${tx.to_wallet}` : `From: ${tx.from_wallet}`}
                                            </p>
                                            <p className="text-xs text-gray-500">{formatDate(tx.created_at)}</p>
                                        </div>
                                        <div className={tx.from_wallet === walletId ? 'text-red-500' : 'text-green-500'}>
                                            {tx.from_wallet === walletId ? '-' : '+'}₦{tx.amount}
                                        </div>
                                    </div>
                                ))}
                                {transactions.length === 0 && (
                                    <p className="text-gray-500 text-center py-4">No transactions yet</p>
                                )}
                            </div>
                        </Card>
                    </div>
                )}

                {activeTab === 'send' && (
                    <Card>
                        <button onClick={() => setActiveTab('home')} className="text-blue-600 mb-4">&larr; Back</button>
                        <h2 className="text-xl font-bold mb-4">Send Money</h2>
                        <div className="space-y-4">
                            <Input
                                placeholder="wallet-recipient-id"
                                label="Recipient Wallet ID"
                                value={recipient}
                                onChange={(e) => setRecipient(e.target.value)}
                            />
                            <Input
                                placeholder="0"
                                type="number"
                                label="Amount (₦)"
                                value={amount}
                                onChange={(e) => setAmount(e.target.value)}
                            />
                            <Button
                                className="w-full"
                                onClick={handleSend}
                                disabled={sending}
                            >
                                {sending ? 'Sending...' : 'Send Money'}
                            </Button>
                        </div>
                    </Card>
                )}

                {activeTab === 'scan' && (
                    <Card>
                        <button onClick={() => setActiveTab('home')} className="text-blue-600 mb-4">&larr; Back</button>
                        <h2 className="text-xl font-bold mb-4">Scan QR Code</h2>
                        <div className="bg-gray-100 rounded-lg p-8 text-center">
                            <div className="text-6xl mb-4">📷</div>
                            <p className="text-gray-500">Camera access required</p>
                            <p className="text-sm text-gray-400 mt-2">Point your camera at a payment QR code</p>
                        </div>
                        <div className="mt-4 text-center">
                            <p className="text-sm text-gray-500">Or show your QR to receive:</p>
                            <div className="bg-white border-2 border-dashed border-gray-300 rounded-lg p-8 mt-2">
                                <div className="text-6xl">🏦</div>
                                <p className="text-xs text-gray-500 mt-2">{walletId}</p>
                            </div>
                        </div>
                    </Card>
                )}

                {activeTab === 'history' && (
                    <Card>
                        <button onClick={() => setActiveTab('home')} className="text-blue-600 mb-4">&larr; Back</button>
                        <h2 className="text-xl font-bold mb-4">Transaction History</h2>
                        <div className="space-y-2">
                            {transactions.map(tx => (
                                <div key={tx.id} className="flex justify-between items-center py-3 border-b">
                                    <div>
                                        <p className="font-medium">
                                            {tx.from_wallet === walletId ? `Sent to ${tx.to_wallet}` : `Received from ${tx.from_wallet}`}
                                        </p>
                                        <p className="text-xs text-gray-500">{formatDate(tx.created_at)}</p>
                                        <span className={`text-xs px-2 py-0.5 rounded ${
                                            tx.status === 'CONFIRMED' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
                                        }`}>
                                            {tx.status}
                                        </span>
                                    </div>
                                    <div className={`text-lg font-semibold ${tx.from_wallet === walletId ? 'text-red-500' : 'text-green-500'}`}>
                                        {tx.from_wallet === walletId ? '-' : '+'}₦{tx.amount.toLocaleString()}
                                    </div>
                                </div>
                            ))}
                            {transactions.length === 0 && (
                                <p className="text-gray-500 text-center py-8">No transactions found</p>
                            )}
                        </div>
                    </Card>
                )}
            </div>

            {/* Bottom Navigation */}
            <nav className="bg-white border-t shadow-lg">
                <div className="max-w-md mx-auto flex justify-around py-3">
                    <button
                        onClick={() => setActiveTab('home')}
                        className={`flex flex-col items-center ${activeTab === 'home' ? 'text-blue-600' : 'text-gray-500'}`}
                    >
                        <span className="text-xl">🏠</span>
                        <span className="text-xs">Home</span>
                    </button>
                    <button
                        onClick={() => setActiveTab('send')}
                        className={`flex flex-col items-center ${activeTab === 'send' ? 'text-blue-600' : 'text-gray-500'}`}
                    >
                        <span className="text-xl">💸</span>
                        <span className="text-xs">Send</span>
                    </button>
                    <button
                        onClick={() => setActiveTab('scan')}
                        className={`flex flex-col items-center ${activeTab === 'scan' ? 'text-blue-600' : 'text-gray-500'}`}
                    >
                        <span className="text-xl">📱</span>
                        <span className="text-xs">Scan</span>
                    </button>
                    <button
                        onClick={() => setActiveTab('history')}
                        className={`flex flex-col items-center ${activeTab === 'history' ? 'text-blue-600' : 'text-gray-500'}`}
                    >
                        <span className="text-xl">📋</span>
                        <span className="text-xs">History</span>
                    </button>
                    <button
                        onClick={() => setOfflineMode(!offlineMode)}
                        className={`flex flex-col items-center ${offlineMode ? 'text-yellow-600' : 'text-gray-500'}`}
                    >
                        <span className="text-xl">{offlineMode ? '🔴' : '🟢'}</span>
                        <span className="text-xs">Offline</span>
                    </button>
                </div>
            </nav>
        </main>
    );
}
