package mail

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// Mailable defines the interface for mailable classes
type Mailable interface {
	Build() *Message
	GetSubject() string
	GetTo() []string
	GetFrom() string
}

// BaseMailable provides a base implementation for mailable classes
type BaseMailable struct {
	To      []string
	From    string
	Subject string
	Data    map[string]interface{}
}

// GetSubject returns the email subject
func (m *BaseMailable) GetSubject() string {
	return m.Subject
}

// GetTo returns the recipient email addresses
func (m *BaseMailable) GetTo() []string {
	return m.To
}

// GetFrom returns the sender email address
func (m *BaseMailable) GetFrom() string {
	return m.From
}

// MailManager manages mail sending
type MailManager struct {
	driver      Driver
	templates   map[string]*template.Template
	templateDir string
	logger      *zap.Logger
}

// NewMailManager creates a new mail manager
func NewMailManager(driver Driver, templateDir string, logger *zap.Logger) *MailManager {
	return &MailManager{
		driver:      driver,
		templates:   make(map[string]*template.Template),
		templateDir: templateDir,
		logger:      logger,
	}
}

// Send sends an email message
func (m *MailManager) Send(ctx context.Context, message *Message) error {
	return m.driver.Send(ctx, message)
}

// SendBatch sends multiple email messages
func (m *MailManager) SendBatch(ctx context.Context, messages []*Message) error {
	return m.driver.SendBatch(ctx, messages)
}

// SendMailable sends a mailable class
func (m *MailManager) SendMailable(ctx context.Context, mailable Mailable) error {
	message := mailable.Build()
	return m.driver.Send(ctx, message)
}

// SendMailableBatch sends multiple mailable classes
func (m *MailManager) SendMailableBatch(ctx context.Context, mailables []Mailable) error {
	messages := make([]*Message, len(mailables))
	for i, mailable := range mailables {
		messages[i] = mailable.Build()
	}
	return m.driver.SendBatch(ctx, messages)
}

// SendTemplate sends an email using a template
func (m *MailManager) SendTemplate(ctx context.Context, templateName string, data map[string]interface{}, to []string, subject string) error {
	// Load template if not already loaded
	tmpl, err := m.loadTemplate(templateName)
	if err != nil {
		return err
	}

	// Render template
	var html bytes.Buffer
	if err := tmpl.Execute(&html, data); err != nil {
		return err
	}

	// Create message
	message := &Message{
		To:      to,
		Subject: subject,
		HTML:    html.String(),
		From:    m.getDefaultFrom(),
	}

	return m.driver.Send(ctx, message)
}

// SendTemplateWithText sends an email using both HTML and text templates
func (m *MailManager) SendTemplateWithText(ctx context.Context, templateName string, data map[string]interface{}, to []string, subject string) error {
	// Load HTML template
	htmlTmpl, err := m.loadTemplate(templateName + ".html")
	if err != nil {
		return err
	}

	// Load text template
	textTmpl, err := m.loadTemplate(templateName + ".txt")
	if err != nil {
		return err
	}

	// Render HTML template
	var html bytes.Buffer
	if err := htmlTmpl.Execute(&html, data); err != nil {
		return err
	}

	// Render text template
	var text bytes.Buffer
	if err := textTmpl.Execute(&text, data); err != nil {
		return err
	}

	// Create message
	message := &Message{
		To:      to,
		Subject: subject,
		HTML:    html.String(),
		Text:    text.String(),
		From:    m.getDefaultFrom(),
	}

	return m.driver.Send(ctx, message)
}

// loadTemplate loads a template from the template directory
func (m *MailManager) loadTemplate(templateName string) (*template.Template, error) {
	// Check if template is already loaded
	if tmpl, exists := m.templates[templateName]; exists {
		return tmpl, nil
	}

	// Load template from file
	templatePath := filepath.Join(m.templateDir, templateName)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %s: %w", templateName, err)
	}

	// Cache template
	m.templates[templateName] = tmpl

	return tmpl, nil
}

// getDefaultFrom returns the default sender email address
func (m *MailManager) getDefaultFrom() string {
	// This could be configured via environment variables or config
	return "noreply@example.com"
}

// QueueMail queues an email for later sending
func (m *MailManager) QueueMail(ctx context.Context, message *Message, delay time.Duration) error {
	// This is a simplified implementation
	// In a real implementation, you'd use a proper queue system like Redis, RabbitMQ, etc.
	go func() {
		time.Sleep(delay)
		if err := m.driver.Send(context.Background(), message); err != nil {
			m.logger.Error("Failed to send queued email", zap.Error(err))
		}
	}()

	return nil
}

// QueueMailable queues a mailable for later sending
func (m *MailManager) QueueMailable(ctx context.Context, mailable Mailable, delay time.Duration) error {
	message := mailable.Build()
	return m.QueueMail(ctx, message, delay)
}

// TestConnection tests the mail driver connection
func (m *MailManager) TestConnection(ctx context.Context) error {
	// Send a test email
	testMessage := &Message{
		To:      []string{"test@example.com"},
		Subject: "Test Email",
		Text:    "This is a test email to verify the mail configuration.",
		From:    m.getDefaultFrom(),
	}

	return m.driver.Send(ctx, testMessage)
}

// GetDriver returns the current mail driver
func (m *MailManager) GetDriver() Driver {
	return m.driver
}

// SetDriver sets a new mail driver
func (m *MailManager) SetDriver(driver Driver) {
	m.driver = driver
}

// Common email templates and helpers

// WelcomeEmail represents a welcome email
type WelcomeEmail struct {
	BaseMailable
	UserName string
	LoginURL string
}

// Build builds the welcome email message
func (e *WelcomeEmail) Build() *Message {
	html := fmt.Sprintf(`
		<html>
		<body>
			<h1>Welcome to our platform!</h1>
			<p>Hello %s,</p>
			<p>Welcome to our platform! We're excited to have you on board.</p>
			<p>You can log in using this link: <a href="%s">Login</a></p>
			<p>Best regards,<br>The Team</p>
		</body>
		</html>
	`, e.UserName, e.LoginURL)

	return &Message{
		To:      e.To,
		From:    e.From,
		Subject: e.Subject,
		HTML:    html,
	}
}

// PasswordResetEmail represents a password reset email
type PasswordResetEmail struct {
	BaseMailable
	UserName  string
	ResetURL  string
	ExpiresAt time.Time
}

// Build builds the password reset email message
func (e *PasswordResetEmail) Build() *Message {
	html := fmt.Sprintf(`
		<html>
		<body>
			<h1>Password Reset Request</h1>
			<p>Hello %s,</p>
			<p>You requested a password reset. Click the link below to reset your password:</p>
			<p><a href="%s">Reset Password</a></p>
			<p>This link will expire at %s.</p>
			<p>If you didn't request this, please ignore this email.</p>
			<p>Best regards,<br>The Team</p>
		</body>
		</html>
	`, e.UserName, e.ResetURL, e.ExpiresAt.Format("2006-01-02 15:04:05"))

	return &Message{
		To:      e.To,
		From:    e.From,
		Subject: e.Subject,
		HTML:    html,
	}
}

// NotificationEmail represents a notification email
type NotificationEmail struct {
	BaseMailable
	Title      string
	Message    string
	ActionURL  string
	ActionText string
}

// Build builds the notification email message
func (e *NotificationEmail) Build() *Message {
	html := fmt.Sprintf(`
		<html>
		<body>
			<h1>%s</h1>
			<p>%s</p>
			%s
			<p>Best regards,<br>The Team</p>
		</body>
		</html>
	`, e.Title, e.Message, func() string {
		if e.ActionURL != "" && e.ActionText != "" {
			return fmt.Sprintf(`<p><a href="%s">%s</a></p>`, e.ActionURL, e.ActionText)
		}
		return ""
	}())

	return &Message{
		To:      e.To,
		From:    e.From,
		Subject: e.Subject,
		HTML:    html,
	}
}
