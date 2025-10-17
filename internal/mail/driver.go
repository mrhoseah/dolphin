package mail

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Message represents an email message
type Message struct {
	To          []string          `json:"to"`
	Cc          []string          `json:"cc,omitempty"`
	Bcc         []string          `json:"bcc,omitempty"`
	From        string            `json:"from"`
	Subject     string            `json:"subject"`
	Text        string            `json:"text,omitempty"`
	HTML        string            `json:"html,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	Name        string
	ContentType string
	Data        []byte
}

// Driver defines the interface for mail drivers
type Driver interface {
	Send(ctx context.Context, message *Message) error
	SendBatch(ctx context.Context, messages []*Message) error
}

// SMTPDriver implements mail sending using SMTP
type SMTPDriver struct {
	host     string
	port     int
	username string
	password string
	auth     smtp.Auth
	logger   *zap.Logger
}

// NewSMTPDriver creates a new SMTP mail driver
func NewSMTPDriver(host string, port int, username, password string, logger *zap.Logger) *SMTPDriver {
	var auth smtp.Auth
	if username != "" && password != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	return &SMTPDriver{
		host:     host,
		port:     port,
		username: username,
		password: password,
		auth:     auth,
		logger:   logger,
	}
}

func (d *SMTPDriver) Send(ctx context.Context, message *Message) error {
	addr := fmt.Sprintf("%s:%d", d.host, d.port)

	// Build email headers
	headers := make(map[string]string)
	headers["From"] = message.From
	headers["To"] = strings.Join(message.To, ", ")
	headers["Subject"] = message.Subject
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	headers["MIME-Version"] = "1.0"

	// Add custom headers
	for k, v := range message.Headers {
		headers[k] = v
	}

	// Build email body
	var body bytes.Buffer

	// Write headers
	for k, v := range headers {
		body.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	body.WriteString("\r\n")

	// Write content
	if message.HTML != "" {
		body.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		body.WriteString("\r\n")
		body.WriteString(message.HTML)
	} else {
		body.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		body.WriteString("\r\n")
		body.WriteString(message.Text)
	}

	// Send email
	recipients := append(message.To, message.Cc...)
	recipients = append(recipients, message.Bcc...)

	err := smtp.SendMail(addr, d.auth, message.From, recipients, body.Bytes())
	if err != nil {
		d.logger.Error("Failed to send email via SMTP", zap.Error(err))
		return err
	}

	d.logger.Info("Email sent successfully via SMTP",
		zap.String("to", strings.Join(message.To, ",")),
		zap.String("subject", message.Subject))

	return nil
}

func (d *SMTPDriver) SendBatch(ctx context.Context, messages []*Message) error {
	for _, message := range messages {
		if err := d.Send(ctx, message); err != nil {
			return err
		}
	}
	return nil
}

// MailgunDriver implements mail sending using Mailgun
type MailgunDriver struct {
	domain     string
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewMailgunDriver creates a new Mailgun mail driver
func NewMailgunDriver(domain, apiKey string, logger *zap.Logger) *MailgunDriver {
	return &MailgunDriver{
		domain:     domain,
		apiKey:     apiKey,
		baseURL:    "https://api.mailgun.net/v3",
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
	}
}

func (d *MailgunDriver) Send(ctx context.Context, message *Message) error {
	// Build form data
	data := make(map[string]string)
	data["from"] = message.From
	data["subject"] = message.Subject
	data["to"] = strings.Join(message.To, ",")
	data["cc"] = strings.Join(message.Cc, ",")
	data["bcc"] = strings.Join(message.Bcc, ",")
	for k, v := range message.Headers {
		data[fmt.Sprintf("h:%s", k)] = v
	}

	// Convert map[string]string to url.Values
	form := make(map[string][]string)
	for k, v := range data {
		form[k] = []string{v}
	}

	// Send request
	resp, err := d.httpClient.PostForm(fmt.Sprintf("%s/%s/messages", d.baseURL, d.domain), form)
	if err != nil {
		d.logger.Error("Failed to send email via Mailgun", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		d.logger.Error("mailgun API error", zap.Int("status", resp.StatusCode), zap.Error(err))
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			d.logger.Error("failed to read mailgun API error body", zap.Error(err))
			return fmt.Errorf("mailgun API error: status %d, body %s, error %w", resp.StatusCode, string(body), err)
		}
		return fmt.Errorf("mailgun API error: status %d, body %s, error %w", resp.StatusCode, string(body), err)
	}

	d.logger.Info("Email sent successfully via Mailgun",
		zap.String("to", strings.Join(message.To, ",")),
		zap.String("subject", message.Subject),
		zap.String("domain", d.domain),
		zap.String("apiKey", d.apiKey),
		zap.String("response", resp.Status))
	return nil
}
