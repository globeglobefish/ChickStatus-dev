package repository

import (
	"context"
	"database/sql"

	"github.com/probe-system/core/internal/models"
)

type SettingsRepository struct {
	db *DB
}

func NewSettingsRepository(db *DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Get(ctx context.Context) (*models.Settings, error) {
	settings := &models.Settings{}
	err := r.db.QueryRowContext(ctx, `
		SELECT data_retention_days, telegram_bot_token, telegram_chat_id,
			smtp_host, smtp_port, smtp_username, smtp_password, smtp_from, alert_email_to
		FROM settings WHERE id = 1
	`).Scan(&settings.DataRetentionDays, &settings.TelegramBotToken, &settings.TelegramChatID,
		&settings.SMTPHost, &settings.SMTPPort, &settings.SMTPUsername, &settings.SMTPPassword,
		&settings.SMTPFrom, &settings.AlertEmailTo)

	if err == sql.ErrNoRows {
		return models.DefaultSettings(), nil
	}
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *SettingsRepository) Update(ctx context.Context, settings *models.Settings) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE settings SET 
			data_retention_days = ?,
			telegram_bot_token = ?,
			telegram_chat_id = ?,
			smtp_host = ?,
			smtp_port = ?,
			smtp_username = ?,
			smtp_password = ?,
			smtp_from = ?,
			alert_email_to = ?
		WHERE id = 1
	`, settings.DataRetentionDays, settings.TelegramBotToken, settings.TelegramChatID,
		settings.SMTPHost, settings.SMTPPort, settings.SMTPUsername, settings.SMTPPassword,
		settings.SMTPFrom, settings.AlertEmailTo)

	return err
}

// User Repository
type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash FROM users WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET password_hash = ? WHERE id = ?
	`, passwordHash, id)
	return err
}
