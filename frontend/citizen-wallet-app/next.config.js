/** @type {import('next').NextConfig} */
const nextConfig = {
    reactStrictMode: true,
    transpilePackages: ['@cbdc/ui'],
    env: {
        NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082',
        NEXT_PUBLIC_WALLET_SERVICE_URL: process.env.NEXT_PUBLIC_WALLET_SERVICE_URL || 'http://localhost:8081',
        NEXT_PUBLIC_OFFLINE_SERVICE_URL: process.env.NEXT_PUBLIC_OFFLINE_SERVICE_URL || 'http://localhost:8083',
    },
    async headers() {
        return [
            {
                source: '/:path*',
                headers: [
                    {
                        key: 'X-Frame-Options',
                        value: 'DENY',
                    },
                    {
                        key: 'X-Content-Type-Options',
                        value: 'nosniff',
                    },
                    {
                        key: 'X-XSS-Protection',
                        value: '1; mode=block',
                    },
                ],
            },
        ];
    },
};

module.exports = nextConfig;
