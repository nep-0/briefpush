package notify

import (
	"context"
	"fmt"
	"strings"

	gomail "gopkg.in/gomail.v2"
)

// SMTPConfig stores SMTP server details for email delivery.
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	UseTLS   bool   `json:"use_tls"`
}

// EmailConfig defines email notification settings.
type EmailConfig struct {
	From string     `json:"from"`
	To   []string   `json:"to"`
	SMTP SMTPConfig `json:"smtp"`
}

// EmailNotifier sends notifications via SMTP email.
type EmailNotifier struct {
	cfg EmailConfig
}

// NewEmailNotifier validates config and creates an email notifier.
func NewEmailNotifier(cfg EmailConfig) (*EmailNotifier, error) {
	if strings.TrimSpace(cfg.From) == "" {
		return nil, fmt.Errorf("notification email.from is required when enabled")
	}
	if len(cfg.To) == 0 {
		return nil, fmt.Errorf("notification email.to must contain at least one recipient")
	}
	if strings.TrimSpace(cfg.SMTP.Host) == "" || cfg.SMTP.Port <= 0 {
		return nil, fmt.Errorf("notification email.smtp.host and email.smtp.port are required")
	}
	return &EmailNotifier{cfg: cfg}, nil
}

// Kind returns provider type.
func (n *EmailNotifier) Kind() string {
	return "email"
}

// Notify sends one plaintext email message.
func (n *EmailNotifier) Notify(_ context.Context, msg Message) error {
	m := gomail.NewMessage()
	m.SetHeader("From", n.cfg.From)
	m.SetHeader("To", n.cfg.To...)
	m.SetHeader("Subject", strings.TrimSpace(msg.Subject))
	m.SetBody("text/plain", msg.Body)

	d := gomail.NewDialer(n.cfg.SMTP.Host, n.cfg.SMTP.Port, n.cfg.SMTP.Username, n.cfg.SMTP.Password)
	d.SSL = n.cfg.SMTP.UseTLS

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
