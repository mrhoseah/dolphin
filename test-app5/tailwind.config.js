/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./views/**/*.{fin,fin.html,html,go}",
    "./public/**/*.html",
    "./internal/**/*.go",
  ],
  theme: {
    extend: {
      colors: {
        'dolphin-primary': '#009688',
        'dolphin-secondary': '#00BCD4',
        'dolphin-light': '#e0f7fa',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
}
