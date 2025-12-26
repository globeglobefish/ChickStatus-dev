package models

type Settings struct {
	DataRetentionDays int    `json:"data_retention_days" db:"data_retention_days"`
	TelegramBotToken  string `json:"telegram_bot_token" db:"telegram_bot_token"`
	TelegramChatID    string `json:"telegram_chat_id" db:"telegram_chat_id"`
	SMTPHost          string `json:"smtp_host" db:"smtp_host"`
	SMTPPort          int    `json:"smtp_port" db:"smtp_port"`
	SMTPUsername      string `json:"smtp_username" db:"smtp_username"`
	SMTPPassword      string `json:"smtp_password" db:"smtp_password"`
	SMTPFrom          string `json:"smtp_from" db:"smtp_from"`
	AlertEmailTo      string `json:"alert_email_to" db:"alert_email_to"`
}

func DefaultSettings() *Settings {
	return &Settings{
		DataRetentionDays: 7,
		SMTPPort:          587,
	}
}

type User struct {
	ID           string `json:"id" db:"id"`
	Username     string `json:"username" db:"username"`
	PasswordHash string `json:"-" db:"password_hash"`
}
