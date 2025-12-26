package notify

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/probe-system/core/internal/models"
)

type EmailNotifier struct {
	host     string
	port     int
	username string
	password string
	from     string
	to       string
}

func NewEmailNotifier(host string, port int, username, password, from, to string) *EmailNotifier {
	return &EmailNotifier{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		to:       to,
	}
}

func (n *EmailNotifier) Send(ctx context.Context, alert *models.Alert) error {
	if n.host == "" || n.to == "" {
		return nil
	}

	subject := fmt.Sprintf("[ALERT] %s - %s", alert.RuleName, alert.AgentName)
	body := n.formatAlertHTML(alert)

	return n.sendEmail(subject, body)
}

func (n *EmailNotifier) SendRecovery(ctx context.Context, alert *models.Alert) error {
	if n.host == "" || n.to == "" {
		return nil
	}

	subject := fmt.Sprintf("[RESOLVED] %s - %s", alert.RuleName, alert.AgentName)
	body := n.formatRecoveryHTML(alert)

	return n.sendEmail(subject, body)
}

func (n *EmailNotifier) formatAlertHTML(alert *models.Alert) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { background: #dc3545; color: white; padding: 20px; }
        .header h1 { margin: 0; font-size: 24px; }
        .content { padding: 20px; }
        .field { margin-bottom: 15px; }
        .label { font-weight: bold; color: #666; font-size: 12px; text-transform: uppercase; }
        .value { font-size: 16px; color: #333; margin-top: 5px; }
        .footer { background: #f8f9fa; padding: 15px 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚨 Alert Triggered</h1>
        </div>
        <div class="content">
            <div class="field">
                <div class="label">Rule</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Agent</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Metric</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Current Value</div>
                <div class="value">%.2f</div>
            </div>
            <div class="field">
                <div class="label">Threshold</div>
                <div class="value">%.2f</div>
            </div>
            <div class="field">
                <div class="label">Time</div>
                <div class="value">%s</div>
            </div>
        </div>
        <div class="footer">
            This is an automated alert from Probe System.
        </div>
    </div>
</body>
</html>`,
		alert.RuleName,
		alert.AgentName,
		alert.MetricType,
		alert.Value,
		alert.Threshold,
		alert.TriggeredAt.Format("2006-01-02 15:04:05 MST"),
	)
}

func (n *EmailNotifier) formatRecoveryHTML(alert *models.Alert) string {
	duration := ""
	if alert.ResolvedAt != nil {
		d := alert.ResolvedAt.Sub(alert.TriggeredAt)
		duration = d.Round(time.Second).String()
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { background: #28a745; color: white; padding: 20px; }
        .header h1 { margin: 0; font-size: 24px; }
        .content { padding: 20px; }
        .field { margin-bottom: 15px; }
        .label { font-weight: bold; color: #666; font-size: 12px; text-transform: uppercase; }
        .value { font-size: 16px; color: #333; margin-top: 5px; }
        .footer { background: #f8f9fa; padding: 15px 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Alert Resolved</h1>
        </div>
        <div class="content">
            <div class="field">
                <div class="label">Rule</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Agent</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Metric</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Duration</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Resolved At</div>
                <div class="value">%s</div>
            </div>
        </div>
        <div class="footer">
            This is an automated alert from Probe System.
        </div>
    </div>
</body>
</html>`,
		alert.RuleName,
		alert.AgentName,
		alert.MetricType,
		duration,
		time.Now().Format("2006-01-02 15:04:05 MST"),
	)
}

func (n *EmailNotifier) sendEmail(subject, body string) error {
	addr := fmt.Sprintf("%s:%d", n.host, n.port)

	headers := make(map[string]string)
	headers["From"] = n.from
	headers["To"] = n.to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	var auth smtp.Auth
	if n.username != "" && n.password != "" {
		auth = smtp.PlainAuth("", n.username, n.password, n.host)
	}

	return smtp.SendMail(addr, auth, n.from, []string{n.to}, []byte(msg.String()))
}
