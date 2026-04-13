package notify

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	ProviderEmail  = "email"
	ProviderFeishu = "feishu"
)

// Message is the normalized notification payload shared across providers.
type Message struct {
	Subject string
	Body    string
}

// Notifier is implemented by each notification provider (email, feishu, etc.).
type Notifier interface {
	Kind() string
	Notify(ctx context.Context, msg Message) error
}

// FeishuConfig is reserved for future Feishu bot support.
type FeishuConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// ProviderConfig defines one notification provider.
type ProviderConfig struct {
	Type    string       `json:"type"`
	Enabled bool         `json:"enabled"`
	Email   EmailConfig  `json:"email"`
	Feishu  FeishuConfig `json:"feishu"`
}

// Config defines notification configuration.
type Config struct {
	Providers []ProviderConfig `json:"providers"`
}

// Dispatcher fans out one message to multiple providers.
type Dispatcher struct {
	notifiers []Notifier
}

// NewDispatcher builds notifier implementations from config.
func NewDispatcher(cfg Config) (*Dispatcher, error) {
	if len(cfg.Providers) == 0 {
		return &Dispatcher{}, nil
	}

	notifiers := make([]Notifier, 0, len(cfg.Providers))

	for i, provider := range cfg.Providers {
		if !provider.Enabled {
			continue
		}

		switch strings.ToLower(strings.TrimSpace(provider.Type)) {
		case ProviderEmail:
			n, err := NewEmailNotifier(provider.Email)
			if err != nil {
				return nil, fmt.Errorf("notification.providers[%d] email: %w", i, err)
			}
			notifiers = append(notifiers, n)
		case ProviderFeishu:
			n, err := NewFeishuNotifier(provider.Feishu)
			if err != nil {
				return nil, fmt.Errorf("notification.providers[%d] feishu: %w", i, err)
			}
			notifiers = append(notifiers, n)
		default:
			return nil, fmt.Errorf("notification.providers[%d]: unknown type %q", i, provider.Type)
		}
	}

	return &Dispatcher{notifiers: notifiers}, nil
}

// Enabled returns true when at least one notifier is configured.
func (d *Dispatcher) Enabled() bool {
	return d != nil && len(d.notifiers) > 0
}

// Notify sends the same message to all configured notifiers.
func (d *Dispatcher) Notify(ctx context.Context, msg Message) error {
	if !d.Enabled() {
		return nil
	}

	var errList []error
	for _, notifier := range d.notifiers {
		if err := notifier.Notify(ctx, msg); err != nil {
			errList = append(errList, fmt.Errorf("%s notifier: %w", notifier.Kind(), err))
		}
	}
	if len(errList) == 0 {
		return nil
	}
	return errors.Join(errList...)
}
