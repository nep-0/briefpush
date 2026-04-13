package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FeishuNotifier sends messages to a Feishu bot webhook.
type FeishuNotifier struct {
	webhookURL string
	httpClient *http.Client
}

// NewFeishuNotifier validates config and creates a Feishu notifier.
func NewFeishuNotifier(cfg FeishuConfig) (*FeishuNotifier, error) {
	webhookURL := strings.TrimSpace(cfg.WebhookURL)
	if webhookURL == "" {
		return nil, fmt.Errorf("notification feishu.webhook_url is required when enabled")
	}

	return &FeishuNotifier{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// Kind returns provider type.
func (n *FeishuNotifier) Kind() string {
	return ProviderFeishu
}

// Notify sends a text message to Feishu bot.
func (n *FeishuNotifier) Notify(ctx context.Context, msg Message) error {
	text := strings.TrimSpace(msg.Subject)
	body := strings.TrimSpace(msg.Body)
	if body != "" {
		if text == "" {
			text = body
		} else {
			text = text + "\n\n" + body
		}
	}

	payload := map[string]any{
		"msg_type": "text",
		"content": map[string]string{
			"text": text,
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal feishu payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu webhook status=%d body=%s", resp.StatusCode, string(respBytes))
	}

	var common struct {
		Code          int    `json:"code"`
		Msg           string `json:"msg"`
		StatusCode    int    `json:"StatusCode"`
		StatusMessage string `json:"StatusMessage"`
	}
	if err := json.Unmarshal(respBytes, &common); err == nil {
		if common.Code != 0 {
			return fmt.Errorf("feishu webhook code=%d msg=%s", common.Code, common.Msg)
		}
		if common.StatusCode != 0 {
			return fmt.Errorf("feishu webhook StatusCode=%d StatusMessage=%s", common.StatusCode, common.StatusMessage)
		}
	}

	return nil
}
