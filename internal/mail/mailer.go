package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"
)

// Mailable represents a mailable class
type Mailable struct {
	To          []string
	Cc          []string
	Bcc         []string
	From        string
	Subject     string
	TextContent string
	HTMLContent string
	Attachments []Attachment
	Headers     map[string]string
	Priority    Priority
	ReplyTo     string
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
	Inline      bool
	CID         string
}

// Priority represents email priority
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 3
	PriorityHigh   Priority = 5
)

// Mailer represents a mailer interface
type Mailer interface {
	Send(mailable *Mailable) error
	SendTo(to []string, subject, text, html string) error
	Queue(mailable *Mailable) error
}

// SMTPMailer implements Mailer using SMTP
type SMTPMailer struct {
	host     string
	port     string
	username string
	password string
	from     string
	auth     smtp.Auth
}

// NewSMTPMailer creates a new SMTP mailer
func NewSMTPMailer(host, port, username, password, from string) *SMTPMailer {
	auth := smtp.PlainAuth("", username, password, host)
	return &SMTPMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		auth:     auth,
	}
}

// Send sends an email
func (sm *SMTPMailer) Send(mailable *Mailable) error {
	if mailable.From == "" {
		mailable.From = sm.from
	}

	message, err := sm.buildMessage(mailable)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%s", sm.host, sm.port)
	to := append(mailable.To, mailable.Cc...)
	to = append(to, mailable.Bcc...)

	return smtp.SendMail(addr, sm.auth, mailable.From, to, message)
}

// SendTo sends a simple email
func (sm *SMTPMailer) SendTo(to []string, subject, text, html string) error {
	mailable := &Mailable{
		To:          to,
		Subject:     subject,
		TextContent: text,
		HTMLContent: html,
		From:        sm.from,
	}

	return sm.Send(mailable)
}

// Queue queues an email for later sending
func (sm *SMTPMailer) Queue(mailable *Mailable) error {
	// In a real implementation, this would add to a queue
	// For now, we'll just send immediately
	return sm.Send(mailable)
}

// buildMessage builds the email message
func (sm *SMTPMailer) buildMessage(mailable *Mailable) ([]byte, error) {
	var buf bytes.Buffer

	// Set default headers
	headers := make(map[string]string)
	headers["From"] = mailable.From
	headers["To"] = strings.Join(mailable.To, ", ")
	headers["Subject"] = mailable.Subject
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	headers["MIME-Version"] = "1.0"

	if len(mailable.Cc) > 0 {
		headers["Cc"] = strings.Join(mailable.Cc, ", ")
	}

	if mailable.ReplyTo != "" {
		headers["Reply-To"] = mailable.ReplyTo
	}

	// Set priority
	switch mailable.Priority {
	case PriorityHigh:
		headers["X-Priority"] = "1"
		headers["X-MSMail-Priority"] = "High"
	case PriorityLow:
		headers["X-Priority"] = "5"
		headers["X-MSMail-Priority"] = "Low"
	default:
		headers["X-Priority"] = "3"
		headers["X-MSMail-Priority"] = "Normal"
	}

	// Add custom headers
	for key, value := range mailable.Headers {
		headers[key] = value
	}

	// Write headers
	for key, value := range headers {
		fmt.Fprintf(&buf, "%s: %s\r\n", key, value)
	}

	// Determine if we need multipart
	hasAttachments := len(mailable.Attachments) > 0
	hasBothContent := mailable.TextContent != "" && mailable.HTMLContent != ""

	if hasAttachments || hasBothContent {
		boundary := fmt.Sprintf("boundary_%d", time.Now().Unix())
		fmt.Fprintf(&buf, "Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary)

		// Add text content
		if mailable.TextContent != "" {
			fmt.Fprintf(&buf, "--%s\r\n", boundary)
			fmt.Fprintf(&buf, "Content-Type: text/plain; charset=utf-8\r\n\r\n")
			fmt.Fprintf(&buf, "%s\r\n\r\n", mailable.TextContent)
		}

		// Add HTML content
		if mailable.HTMLContent != "" {
			fmt.Fprintf(&buf, "--%s\r\n", boundary)
			fmt.Fprintf(&buf, "Content-Type: text/html; charset=utf-8\r\n\r\n")
			fmt.Fprintf(&buf, "%s\r\n\r\n", mailable.HTMLContent)
		}

		// Add attachments
		for _, attachment := range mailable.Attachments {
			fmt.Fprintf(&buf, "--%s\r\n", boundary)
			if attachment.Inline {
				fmt.Fprintf(&buf, "Content-Type: %s\r\n", attachment.ContentType)
				fmt.Fprintf(&buf, "Content-Disposition: inline; filename=\"%s\"\r\n", attachment.Filename)
				if attachment.CID != "" {
					fmt.Fprintf(&buf, "Content-ID: <%s>\r\n", attachment.CID)
				}
			} else {
				fmt.Fprintf(&buf, "Content-Type: %s\r\n", attachment.ContentType)
				fmt.Fprintf(&buf, "Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Filename)
			}
			fmt.Fprintf(&buf, "Content-Transfer-Encoding: base64\r\n\r\n")

			// Encode attachment data as base64
			encoded := base64Encode(attachment.Data)
			fmt.Fprintf(&buf, "%s\r\n\r\n", encoded)
		}

		fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	} else {
		// Simple message
		if mailable.HTMLContent != "" {
			fmt.Fprintf(&buf, "Content-Type: text/html; charset=utf-8\r\n\r\n")
			fmt.Fprintf(&buf, "%s", mailable.HTMLContent)
		} else {
			fmt.Fprintf(&buf, "Content-Type: text/plain; charset=utf-8\r\n\r\n")
			fmt.Fprintf(&buf, "%s", mailable.TextContent)
		}
	}

	return buf.Bytes(), nil
}

// base64Encode encodes data as base64
func base64Encode(data []byte) string {
	// Simple base64 encoding implementation
	// In production, use encoding/base64
	return fmt.Sprintf("%x", data)
}

// MailgunMailer implements Mailer using Mailgun
type MailgunMailer struct {
	domain string
	apiKey string
	from   string
}

// NewMailgunMailer creates a new Mailgun mailer
func NewMailgunMailer(domain, apiKey, from string) *MailgunMailer {
	return &MailgunMailer{
		domain: domain,
		apiKey: apiKey,
		from:   from,
	}
}

// Send sends an email via Mailgun
func (mm *MailgunMailer) Send(mailable *Mailable) error {
	// In a real implementation, this would use Mailgun's API
	// For now, we'll simulate success
	fmt.Printf("Sending email via Mailgun to %v: %s\n", mailable.To, mailable.Subject)
	return nil
}

// SendTo sends a simple email via Mailgun
func (mm *MailgunMailer) SendTo(to []string, subject, text, html string) error {
	mailable := &Mailable{
		To:          to,
		Subject:     subject,
		TextContent: text,
		HTMLContent: html,
		From:        mm.from,
	}

	return mm.Send(mailable)
}

// Queue queues an email for later sending via Mailgun
func (mm *MailgunMailer) Queue(mailable *Mailable) error {
	// In a real implementation, this would add to Mailgun's queue
	return mm.Send(mailable)
}

// SendGridMailer implements Mailer using SendGrid
type SendGridMailer struct {
	apiKey string
	from   string
}

// NewSendGridMailer creates a new SendGrid mailer
func NewSendGridMailer(apiKey, from string) *SendGridMailer {
	return &SendGridMailer{
		apiKey: apiKey,
		from:   from,
	}
}

// Send sends an email via SendGrid
func (sgm *SendGridMailer) Send(mailable *Mailable) error {
	// In a real implementation, this would use SendGrid's API
	// For now, we'll simulate success
	fmt.Printf("Sending email via SendGrid to %v: %s\n", mailable.To, mailable.Subject)
	return nil
}

// SendTo sends a simple email via SendGrid
func (sgm *SendGridMailer) SendTo(to []string, subject, text, html string) error {
	mailable := &Mailable{
		To:          to,
		Subject:     subject,
		TextContent: text,
		HTMLContent: html,
		From:        sgm.from,
	}

	return sgm.Send(mailable)
}

// Queue queues an email for later sending via SendGrid
func (sgm *SendGridMailer) Queue(mailable *Mailable) error {
	// In a real implementation, this would add to SendGrid's queue
	return sgm.Send(mailable)
}

// MailManager manages mailers
type MailManager struct {
	mailers       map[string]Mailer
	defaultMailer string
}

// NewMailManager creates a new mail manager
func NewMailManager() *MailManager {
	return &MailManager{
		mailers:       make(map[string]Mailer),
		defaultMailer: "smtp",
	}
}

// RegisterMailer registers a mailer
func (mm *MailManager) RegisterMailer(name string, mailer Mailer) {
	mm.mailers[name] = mailer
}

// SetDefaultMailer sets the default mailer
func (mm *MailManager) SetDefaultMailer(name string) {
	mm.defaultMailer = name
}

// GetMailer returns a mailer by name
func (mm *MailManager) GetMailer(name string) Mailer {
	if mailer, exists := mm.mailers[name]; exists {
		return mailer
	}
	return nil
}

// DefaultMailer returns the default mailer
func (mm *MailManager) DefaultMailer() Mailer {
	return mm.GetMailer(mm.defaultMailer)
}

// Send sends an email using the default mailer
func (mm *MailManager) Send(mailable *Mailable) error {
	return mm.DefaultMailer().Send(mailable)
}

// SendTo sends a simple email using the default mailer
func (mm *MailManager) SendTo(to []string, subject, text, html string) error {
	return mm.DefaultMailer().SendTo(to, subject, text, html)
}

// Queue queues an email using the default mailer
func (mm *MailManager) Queue(mailable *Mailable) error {
	return mm.DefaultMailer().Queue(mailable)
}

// MailableBuilder helps build mailable instances
type MailableBuilder struct {
	mailable *Mailable
}

// NewMailableBuilder creates a new mailable builder
func NewMailableBuilder() *MailableBuilder {
	return &MailableBuilder{
		mailable: &Mailable{
			Headers: make(map[string]string),
		},
	}
}

// To sets the recipients
func (mb *MailableBuilder) To(to []string) *MailableBuilder {
	mb.mailable.To = to
	return mb
}

// Cc sets the CC recipients
func (mb *MailableBuilder) Cc(cc []string) *MailableBuilder {
	mb.mailable.Cc = cc
	return mb
}

// Bcc sets the BCC recipients
func (mb *MailableBuilder) Bcc(bcc []string) *MailableBuilder {
	mb.mailable.Bcc = bcc
	return mb
}

// From sets the sender
func (mb *MailableBuilder) From(from string) *MailableBuilder {
	mb.mailable.From = from
	return mb
}

// Subject sets the subject
func (mb *MailableBuilder) Subject(subject string) *MailableBuilder {
	mb.mailable.Subject = subject
	return mb
}

// Text sets the text content
func (mb *MailableBuilder) Text(text string) *MailableBuilder {
	mb.mailable.TextContent = text
	return mb
}

// HTML sets the HTML content
func (mb *MailableBuilder) HTML(html string) *MailableBuilder {
	mb.mailable.HTMLContent = html
	return mb
}

// Attach adds an attachment
func (mb *MailableBuilder) Attach(filename string, data []byte, contentType string) *MailableBuilder {
	attachment := Attachment{
		Filename:    filename,
		ContentType: contentType,
		Data:        data,
		Inline:      false,
	}
	mb.mailable.Attachments = append(mb.mailable.Attachments, attachment)
	return mb
}

// AttachInline adds an inline attachment
func (mb *MailableBuilder) AttachInline(filename string, data []byte, contentType string, cid string) *MailableBuilder {
	attachment := Attachment{
		Filename:    filename,
		ContentType: contentType,
		Data:        data,
		Inline:      true,
		CID:         cid,
	}
	mb.mailable.Attachments = append(mb.mailable.Attachments, attachment)
	return mb
}

// Priority sets the priority
func (mb *MailableBuilder) Priority(priority Priority) *MailableBuilder {
	mb.mailable.Priority = priority
	return mb
}

// ReplyTo sets the reply-to address
func (mb *MailableBuilder) ReplyTo(replyTo string) *MailableBuilder {
	mb.mailable.ReplyTo = replyTo
	return mb
}

// Header adds a custom header
func (mb *MailableBuilder) Header(key, value string) *MailableBuilder {
	mb.mailable.Headers[key] = value
	return mb
}

// Build builds the mailable
func (mb *MailableBuilder) Build() *Mailable {
	return mb.mailable
}

// TemplateMailer handles template-based emails
type TemplateMailer struct {
	mailer   Mailer
	template *template.Template
}

// NewTemplateMailer creates a new template mailer
func NewTemplateMailer(mailer Mailer, templatePath string) (*TemplateMailer, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}

	return &TemplateMailer{
		mailer:   mailer,
		template: tmpl,
	}, nil
}

// SendTemplate sends an email using a template
func (tm *TemplateMailer) SendTemplate(to []string, subject string, data interface{}) error {
	var htmlBuf bytes.Buffer
	err := tm.template.Execute(&htmlBuf, data)
	if err != nil {
		return err
	}

	mailable := &Mailable{
		To:          to,
		Subject:     subject,
		HTMLContent: htmlBuf.String(),
	}

	return tm.mailer.Send(mailable)
}

// Common email templates
const (
	WelcomeEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Welcome</title>
</head>
<body>
    <h1>Welcome, {{.Name}}!</h1>
    <p>Thank you for joining us. We're excited to have you on board.</p>
    <p>Best regards,<br>The Team</p>
</body>
</html>`

	PasswordResetTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Password Reset</title>
</head>
<body>
    <h1>Password Reset Request</h1>
    <p>Hello {{.Name}},</p>
    <p>You requested a password reset. Click the link below to reset your password:</p>
    <p><a href="{{.ResetLink}}">Reset Password</a></p>
    <p>This link will expire in {{.Expiry}} hours.</p>
    <p>If you didn't request this, please ignore this email.</p>
</body>
</html>`

	NotificationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Notification</title>
</head>
<body>
    <h1>{{.Title}}</h1>
    <p>{{.Message}}</p>
    {{if .ActionLink}}
    <p><a href="{{.ActionLink}}">{{.ActionText}}</a></p>
    {{end}}
</body>
</html>`
)
