/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#fdf4f5',
          100: '#fbe8eb',
          200: '#f7d5db',
          300: '#f1b4bf',
          400: '#e78598',
          500: '#db5b75',
          600: '#c53d5a',
          700: '#a52e46',
          800: '#89293d',
          900: '#752637',
          950: '#41101b',
        },
        dark: {
          800: '#181920',
          850: '#14151b',
          900: '#0f1015',
          950: '#0a0a0d',
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
      }
    },
  },
  plugins: [],
}
