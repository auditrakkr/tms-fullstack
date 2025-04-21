// This component can be included in any page that needs notifications
// Usage:
// 1. Include this script in your HTML
// 2. Add a notification container div with id="notification"
// 3. Call showNotification(message, type) to display a notification

/**
 * Shows a notification with the specified message and type
 * @param {string} message - The message to display
 * @param {string} type - The type of notification (success, error, warning, info)
 * @param {number} duration - Time in milliseconds before auto-dismissing (0 for no auto-dismiss)
 */
function showNotification(message, type = 'info', duration = 5000) {
  const notificationContainer = document.getElementById('notification');
  const notificationMessage = document.getElementById('notificationMessage');

  if (!notificationContainer || !notificationMessage) {
    console.error('Notification container not found');
    return;
  }

  // Reset classes
  notificationContainer.className = 'relative px-4 py-3 rounded-md border shadow-sm animate-slide-in-up';

  // Add appropriate styling based on type
  switch (type) {
    case 'success':
      notificationContainer.classList.add('bg-green-100', 'border-green-500', 'text-green-800',
                                          'dark:bg-green-900', 'dark:border-green-700', 'dark:text-green-200');
      break;
    case 'error':
      notificationContainer.classList.add('bg-red-100', 'border-red-500', 'text-red-800',
                                          'dark:bg-red-900', 'dark:border-red-700', 'dark:text-red-200');
      break;
    case 'warning':
      notificationContainer.classList.add('bg-yellow-100', 'border-yellow-500', 'text-yellow-800',
                                          'dark:bg-yellow-900', 'dark:border-yellow-700', 'dark:text-yellow-200');
      break;
    case 'info':
    default:
      notificationContainer.classList.add('bg-blue-100', 'border-blue-500', 'text-blue-800',
                                          'dark:bg-blue-900', 'dark:border-blue-700', 'dark:text-blue-200');
      break;
  }

  // Set the message
  notificationMessage.textContent = message;

  // Show the notification
  notificationContainer.classList.remove('hidden');

  // Auto-dismiss after duration (if duration > 0)
  if (duration > 0) {
    setTimeout(() => {
      dismissNotification();
    }, duration);
  }
}

/**
 * Dismisses the notification
 */
function dismissNotification() {
  const notification = document.getElementById('notification');
  if (notification) {
    // Add fade-out animation
    notification.classList.add('opacity-0', 'transition-opacity', 'duration-300');

    // After animation completes, hide the notification
    setTimeout(() => {
      notification.classList.add('hidden');
      notification.classList.remove('opacity-0', 'transition-opacity', 'duration-300');
    }, 300);
  }
}

// Set up event listeners for dismiss buttons
document.addEventListener('DOMContentLoaded', () => {
  const dismissButtons = document.querySelectorAll('#notification button');

  dismissButtons.forEach(button => {
    button.addEventListener('click', dismissNotification);
  });
});