package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/probe-system/core/internal/models"
)

type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

func NewTelegramNotifier(botToken, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *TelegramNotifier) Send(ctx context.Context, alert *models.Alert) error {
	if n.botToken == "" || n.chatID == "" {
		return nil
	}

	message := n.formatAlertMessage(alert)
	return n.sendMessage(ctx, message)
}

func (n *TelegramNotifier) SendRecovery(ctx context.Context, alert *models.Alert) error {
	if n.botToken == "" || n.chatID == "" {
		return nil
	}

	message := n.formatRecoveryMessage(alert)
	return n.sendMessage(ctx, message)
}

func (n *TelegramNotifier) formatAlertMessage(alert *models.Alert) string {
	return fmt.Sprintf(`🚨 *Alert Triggered*

*Rule:* %s
*Agent:* %s
*Metric:* %s
*Value:* %.2f
*Threshold:* %.2f
*Time:* %s

%s`,
		escapeMarkdown(alert.RuleName),
		escapeMarkdown(alert.AgentName),
		alert.MetricType,
		alert.Value,
		alert.Threshold,
		alert.TriggeredAt.Format("2006-01-02 15:04:05"),
		escapeMarkdown(alert.Message),
	)
}

func (n *TelegramNotifier) formatRecoveryMessage(alert *models.Alert) string {
	duration := ""
	if alert.ResolvedAt != nil {
		d := alert.ResolvedAt.Sub(alert.TriggeredAt)
		duration = fmt.Sprintf("Duration: %s", d.Round(time.Second))
	}

	return fmt.Sprintf(`✅ *Alert Resolved*

*Rule:* %s
*Agent:* %s
*Metric:* %s
*Time:* %s
%s`,
		escapeMarkdown(alert.RuleName),
		escapeMarkdown(alert.AgentName),
		alert.MetricType,
		time.Now().Format("2006-01-02 15:04:05"),
		duration,
	)
}

func (n *TelegramNotifier) sendMessage(ctx context.Context, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)

	payload := map[string]interface{}{
		"chat_id":    n.chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}

func escapeMarkdown(s string) string {
	replacer := []string{
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	}

	result := s
	for i := 0; i < len(replacer); i += 2 {
		result = replaceAll(result, replacer[i], replacer[i+1])
	}
	return result
}

func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if string(s[i]) == old {
			result += new
		} else {
			result += string(s[i])
		}
	}
	return result
}
