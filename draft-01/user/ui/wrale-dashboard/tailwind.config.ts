import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        'wrale-primary': '#1a365d',
        'wrale-secondary': '#2d3748',
        'wrale-accent': '#4fd1c5',
        'wrale-danger': '#f56565',
        'wrale-warning': '#ed8936',
        'wrale-success': '#48bb78'
      }
    },
  },
  plugins: [],
}
export default config