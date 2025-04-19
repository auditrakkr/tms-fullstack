package global

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// SMTP configuration variables
var (
    SMTP_HOST     string
    SMTP_PORT     string
    SMTP_USERNAME string
    SMTP_PASSWORD string
)


func init() {
	 // Try to load environment variables from .env file
    // This is useful for local development
    err := godotenv.Load()
    if err != nil {
        // Just log the error and continue - not fatal if .env file is missing
        // as we might be in a production environment with environment variables
        // set by other means (e.g., Docker, K8s, etc.)
        log.Println("Warning: Error loading .env file:", err)
    }
	// Load SMTP configuration from environment variables
    loadSMTPConfig()

}

// loadSMTPConfig loads SMTP configuration from environment variables
func loadSMTPConfig() {
    SMTP_HOST = getEnvWithDefault("SMTP_HOST", "smtp.gmail.com")
    SMTP_PORT = getEnvWithDefault("SMTP_PORT", "587")
    SMTP_USERNAME = os.Getenv("SMTP_USERNAME")
    SMTP_PASSWORD = os.Getenv("SMTP_PASSWORD")

    // Validate required configuration
    if SMTP_USERNAME == "" || SMTP_PASSWORD == "" {
        log.Println("Warning: SMTP credentials are not set. Email functionality will not work.")
    }
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}



// MailOptions defines the structure for email sending parameters
type MailOptions struct {
    To       string
    From     string
    Subject  string
    Text     string
    Html     string // Optional HTML content
    ReplyTo  string // Optional reply-to address
    Cc       []string
    Bcc      []string
    Attachments []Attachment // Optional attachments
}

// Attachment represents an email attachment
type Attachment struct {
    Filename string
    Content  []byte
    ContentType string
}

// MailOptionSettings holds templates and default values for various email types
type MailOptionSettings struct {
    From         string
    Subject      string
    TextTemplate string
    HtmlTemplate string // Optional HTML template
}

// Define mail templates for different types of notifications
var (
    ConfirmEmailMailOptionSettings = MailOptionSettings{
        From:         "noreply@auditrakkr.com",
        Subject:      "Email Verification",
        TextTemplate: "Please verify your email by clicking this link: {url}",
        HtmlTemplate: "<p>Please verify your email by clicking this link: <a href=\"{url}\">Verify Email</a></p>",
    }

    ResetPasswordMailOptionSettings = MailOptionSettings{
        From:         "noreply@auditrakkr.com",
        Subject:      "Password Reset Request",
        TextTemplate: "Reset your password by clicking this link: {url}",
        HtmlTemplate: "<p>Reset your password by clicking this link: <a href=\"{url}\">Reset Password</a></p>",
    }
)

// SendMail sends an email using the provided options
func SendMail(options MailOptions) error {
	// Check if SMTP is properly configured
    if SMTP_USERNAME == "" || SMTP_PASSWORD == "" {
        return fmt.Errorf("SMTP is not configured properly. Check environment variables")
    }

    // Set up authentication information
    auth := smtp.PlainAuth("", SMTP_USERNAME, SMTP_PASSWORD, SMTP_HOST)

    // Build email headers
    headers := make(map[string]string)
    headers["From"] = options.From
    headers["To"] = options.To
    headers["Subject"] = options.Subject
    headers["MIME-Version"] = "1.0"

    // Handle CC and BCC if provided
    if len(options.Cc) > 0 {
        headers["Cc"] = strings.Join(options.Cc, ",")
    }

    if options.ReplyTo != "" {
        headers["Reply-To"] = options.ReplyTo
    }

    // Determine if we should send HTML or plain text
    var contentType string
    var body string

    if options.Html != "" {
        contentType = "text/html; charset=UTF-8"
        body = options.Html
    } else {
        contentType = "text/plain; charset=UTF-8"
        body = options.Text
    }

    headers["Content-Type"] = contentType

    // Build the message
    message := ""
    for key, value := range headers {
        message += fmt.Sprintf("%s: %s\r\n", key, value)
    }
    message += "\r\n" + body

    // Get all recipients (To + Cc + Bcc)
    var recipients []string
    recipients = append(recipients, strings.Split(options.To, ",")...)
    recipients = append(recipients, options.Cc...)
    recipients = append(recipients, options.Bcc...)

    // Connect to the SMTP server with TLS
    // Create a custom TLS config
    tlsConfig := &tls.Config{
        ServerName: SMTP_HOST,
    }

    // Connect to the server
    smtpServer := fmt.Sprintf("%s:%s", SMTP_HOST, SMTP_PORT)
    client, err := smtp.Dial(smtpServer)
    if err != nil {
        return fmt.Errorf("failed to connect to SMTP server: %v", err)
    }
    defer client.Close()

    // Start TLS
    if err = client.StartTLS(tlsConfig); err != nil {
        return fmt.Errorf("failed to start TLS: %v", err)
    }

    // Authenticate
    if err = client.Auth(auth); err != nil {
        return fmt.Errorf("authentication failed: %v", err)
    }

    // Set the sender
    if err = client.Mail(options.From); err != nil {
        return fmt.Errorf("failed to set sender: %v", err)
    }

    // Set the recipients
    for _, recipient := range recipients {
        recipient = strings.TrimSpace(recipient)
        if recipient != "" {
            if err = client.Rcpt(recipient); err != nil {
                return fmt.Errorf("failed to add recipient %s: %v", recipient, err)
            }
        }
    }

    // Send the email body
    wc, err := client.Data()
    if err != nil {
        return fmt.Errorf("failed to start data session: %v", err)
    }

    _, err = wc.Write([]byte(message))
    if err != nil {
        return fmt.Errorf("failed to write email data: %v", err)
    }

    err = wc.Close()
    if err != nil {
        return fmt.Errorf("failed to close data writer: %v", err)
    }

    // Close the connection
    err = client.Quit()
    if err != nil {
        return fmt.Errorf("failed to close connection: %v", err)
    }

    return nil
}

// SendMailAsync sends an email asynchronously using a goroutine
func SendMailAsync(options MailOptions) {
    go func() {
        err := SendMail(options)
        if err != nil {
            log.Printf("Error sending email: %v\n", err)
        }
    }()
}
