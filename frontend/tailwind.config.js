/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{html,ts,js}",
  ],
  theme: {
    extend: {
      backgroundColor: {
        'logo-bg': "url('/assets/images/logo_us.png')",
      },
    },
  },
  plugins: [],
};