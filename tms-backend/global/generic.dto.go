package global

// GenericNotificationResponse represents a generic response with a notification class and message
type GenericNotificationResponse struct {
	NotificationClass   string `json:"notificationClass"`   // CSS class for notification: "is-success", "is-danger", "is-warning", "is-info"
	NotificationMessage string `json:"notificationMessage"` // Message to display
}
