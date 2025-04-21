module.exports = {
  content: [
    './views/**/*.html',     // Go template files
    './assets/sass/**/*.scss',  // Custom SCSS paths
    './**/*.go'                 // If you use class names inside Go files
  ],
  theme: {
    extend: {}
  },
  darkMode: 'class', // or 'media'
  plugins: []
}
