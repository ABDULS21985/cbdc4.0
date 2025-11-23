"use client";

import React, { useState, useEffect } from 'react';
import { Button } from '@cbdc/ui/Button'; // Assuming we configured aliases or workspace linking
import { Card } from '@cbdc/ui/Card';
import { Input } from '@cbdc/ui/Input';

export default function Home() {
    const [balance, setBalance] = useState<number | null>(null);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        // Fetch balance on load
        fetchBalance();
    }, []);

    const fetchBalance = async () => {
        setLoading(true);
        try {
            // In a real app, we'd use a proxy or env var for the API URL
            const res = await fetch('http://localhost:8082/wallets/wallet-alice-123'); // Mocked ID
            if (res.ok) {
                const data = await res.json();
                setBalance(data.balance);
            }
        } catch (err) {
            console.error("Failed to fetch balance", err);
        } finally {
            setLoading(false);
        }
    };

    return (
        <main className="flex min-h-screen flex-col items-center justify-center p-24 bg-gray-50">
            <Card className="w-full max-w-md">
                <h1 className="text-2xl font-bold mb-6 text-center">My CBDC Wallet</h1>

                <div className="mb-8 text-center">
                    <p className="text-gray-500">Current Balance</p>
                    <div className="text-4xl font-bold text-blue-600">
                        {loading ? 'Loading...' : `₦ ${balance ?? '0.00'}`}
                    </div>
                </div>

                <div className="space-y-4">
                    <Input placeholder="Recipient Address" label="Send To" />
                    <Input placeholder="Amount" type="number" label="Amount" />
                    <Button className="w-full" onClick={() => alert('Transfer initiated!')}>
                        Send Money
                    </Button>
                </div>
            </Card>
        </main>
    );
}
