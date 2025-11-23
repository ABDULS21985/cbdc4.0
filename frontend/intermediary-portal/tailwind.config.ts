import type { Config } from 'tailwindcss';

const config: Config = {
    content: [
        './pages/**/*.{js,ts,jsx,tsx,mdx}',
        './components/**/*.{js,ts,jsx,tsx,mdx}',
        './app/**/*.{js,ts,jsx,tsx,mdx}',
        '../packages/ui/**/*.{js,ts,jsx,tsx}',
    ],
    theme: {
        extend: {
            colors: {
                intermediary: {
                    blue: '#2563eb',
                    light: '#dbeafe',
                },
            },
        },
    },
    plugins: [],
};

export default config;
