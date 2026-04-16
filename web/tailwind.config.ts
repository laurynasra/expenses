import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        violet: { 600: '#7D56F4' },
        emerald: { 500: '#04B575' },
        amber: { 400: '#FFA500' },
      },
    },
  },
  plugins: [],
} satisfies Config
