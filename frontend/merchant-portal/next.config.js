/** @type {import('next').NextConfig} */
const nextConfig = {
    reactStrictMode: true,
    transpilePackages: ['@cbdc/ui'],
    env: {
        NEXT_PUBLIC_PAYMENTS_SERVICE_URL: process.env.NEXT_PUBLIC_PAYMENTS_SERVICE_URL || 'http://localhost:8082',
        NEXT_PUBLIC_WALLET_SERVICE_URL: process.env.NEXT_PUBLIC_WALLET_SERVICE_URL || 'http://localhost:8081',
    },
    async headers() {
        return [
            {
                source: '/:path*',
                headers: [
                    { key: 'X-Frame-Options', value: 'DENY' },
                    { key: 'X-Content-Type-Options', value: 'nosniff' },
                ],
            },
        ];
    },
};

module.exports = nextConfig;
