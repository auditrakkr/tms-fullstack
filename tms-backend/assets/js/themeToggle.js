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

  const iconElement = themeToggleButton.querySelector('svg');
  if (!iconElement) return;

  if (isDark) {
    // Show sun icon when in dark mode
    iconElement.innerHTML = `
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
    `;
  } else {
    // Show moon icon when in light mode
    iconElement.innerHTML = `
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
    `;
  }
}

// Initialize the theme on page load
document.addEventListener('DOMContentLoaded', function() {
  // Check for saved theme preference or system preference
  const savedTheme = localStorage.getItem('theme');
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

  // Apply the theme
  if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }

  // Setup the theme toggle
  setupThemeToggle();

  // Handle authentication UI
  handleAuthentication();
});

// Handle authentication UI based on the token
function handleAuthentication() {
  const accessToken = localStorage.getItem('accessToken');
  const signInBtn = document.getElementById('sign-in');
  const signOutBtn = document.getElementById('sign-out');
  const welcomeMessage = document.getElementById('welcome-message');

  if (!signInBtn || !signOutBtn || !welcomeMessage) return;

  if (accessToken) {
    try {
      const decodedToken = jwt_decode(accessToken);
      const now = Math.floor(Date.now() / 1000); // Current time in seconds

      if (decodedToken.exp > now) {
        // Token is valid
        signInBtn.classList.add('hidden');
        signOutBtn.classList.remove('hidden');
        welcomeMessage.textContent = `Welcome ${decodedToken.sub.firstName || 'User'}!`;
        return;
      }
    } catch (error) {
      console.error('Error decoding token:', error);
    }
  }

  // No valid token
  signInBtn.classList.remove('hidden');
  signOutBtn.classList.add('hidden');
  welcomeMessage.textContent = 'Welcome Guest!';
}