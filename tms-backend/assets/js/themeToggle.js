// Theme toggle functionality
function setupThemeToggle() {
  const themeToggleButton = document.getElementById('theme-toggle');
  if (!themeToggleButton) return;

  // Toggle the theme when clicking the button
  themeToggleButton.addEventListener('click', function() {
    // Toggle the dark class on the html element
    document.documentElement.classList.toggle('dark');
    
    // Update the theme in localStorage
    const isDark = document.documentElement.classList.contains('dark');
    localStorage.setItem('theme', isDark ? 'dark' : 'light');
    
    // Load the appropriate stylesheet
    const themeLink = document.getElementById('theme-stylesheet');
    if (themeLink) {
      themeLink.href = `/assets/${isDark ? 'dark' : 'light'}-theme.css`;
    }
    
    // Update the button icon (sun vs moon)
    updateThemeButtonIcon(isDark);
  });
  
  // Update the button on initial load
  const isDark = document.documentElement.classList.contains('dark');
  updateThemeButtonIcon(isDark);
}

// Helper function to update the theme button icon
function updateThemeButtonIcon(isDark) {
  const themeToggleButton = document.getElementById('theme-toggle');
  if (!themeToggleButton) return;
  
  // Define SVG paths for sun and moon icons
  const moonPath = "M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z";
  const sunPath = "M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z";
  
  // Get the SVG path element
  const svgPath = themeToggleButton.querySelector('svg path');
  if (!svgPath) return;
  
  // Update the path data
  svgPath.setAttribute('d', isDark ? sunPath : moonPath);
  
  // Update the aria-label
  themeToggleButton.setAttribute('aria-label', isDark ? 'Switch to light mode' : 'Switch to dark mode');
}

// Initialize the theme on page load
document.addEventListener('DOMContentLoaded', function() {
  // Check for saved theme preference or system preference
  const savedTheme = localStorage.getItem('theme');
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  
  // Apply the theme
  if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
    document.documentElement.classList.add('dark');
    const themeLink = document.getElementById('theme-stylesheet');
    if (themeLink) {
      themeLink.href = '/assets/dark-theme.css';
    }
  }
  
  // Setup the theme toggle
  setupThemeToggle();
});
