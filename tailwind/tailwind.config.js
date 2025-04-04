/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "../internal/ui/views/**/*.templ",
  ],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
    require('@tailwindcss/aspect-ratio'),
  ],
} 