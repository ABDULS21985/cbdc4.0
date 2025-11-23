import React from 'react';

export default function Home() {
    return (
        <main className="flex min-h-screen flex-col items-center p-4 bg-gray-50">
            <div className="w-full max-w-md bg-white rounded-xl shadow-md overflow-hidden md:max-w-2xl mt-10 p-6">
                <div className="uppercase tracking-wide text-sm text-indigo-500 font-semibold">My Wallet</div>
                <div className="mt-2">
                    <h1 className="text-4xl font-bold text-gray-900">₦ 1,500.00</h1>
                    <p className="text-gray-500">Available Balance</p>
                </div>

                <div className="mt-8 grid grid-cols-2 gap-4">
                    <button className="bg-indigo-600 text-white py-3 rounded-lg font-medium hover:bg-indigo-700">
                        Send
                    </button>
                    <button className="bg-gray-200 text-gray-800 py-3 rounded-lg font-medium hover:bg-gray-300">
                        Receive
                    </button>
                </div>
            </div>
        </main>
    );
}
