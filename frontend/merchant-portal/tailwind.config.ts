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
                merchant: {
                    green: '#16a34a',
                    light: '#dcfce7',
                },
            },
        },
    },
    plugins: [],
};

export default config;
