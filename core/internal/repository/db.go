package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) Migrate() error {
	migrations := []string{
		migrationGroups,
		migrationAgents,
		migrationMetrics,
		migrationBillingCycles,
		migrationTrafficRecords,
		migrationTasks,
		migrationTaskResults,
		migrationScripts,
		migrationAlertRules,
		migrationAlerts,
		migrationSettings,
		migrationUsers,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return db.seedDefaults()
}

func (db *DB) seedDefaults() error {
	// Seed default settings
	_, err := db.Exec(`
		INSERT OR IGNORE INTO settings (id, data_retention_days, smtp_port)
		VALUES (1, 7, 587)
	`)
	if err != nil {
		return err
	}

	// Seed default admin user (password: admin)
	_, err = db.Exec(`
		INSERT OR IGNORE INTO users (id, username, password_hash)
		VALUES ('admin', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQlPJmx2UtKVqcPe7u7oM8Xq5Dqe')
	`)
	return err
}

const migrationGroups = `
CREATE TABLE IF NOT EXISTS groups (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT DEFAULT '',
	display_order INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_groups_order ON groups(display_order);
`

const migrationAgents = `
CREATE TABLE IF NOT EXISTS agents (
	id TEXT PRIMARY KEY,
	hostname TEXT NOT NULL,
	ip TEXT NOT NULL,
	os TEXT DEFAULT '',
	arch TEXT DEFAULT '',
	version TEXT DEFAULT '',
	status TEXT DEFAULT 'offline',
	group_id TEXT,
	custom_name TEXT DEFAULT '',
	description TEXT DEFAULT '',
	tags TEXT DEFAULT '[]',
	location TEXT DEFAULT '{}',
	public_visible INTEGER DEFAULT 0,
	last_seen_at DATETIME,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_group ON agents(group_id);
`

const migrationMetrics = `
CREATE TABLE IF NOT EXISTS metrics (
	id TEXT PRIMARY KEY,
	agent_id TEXT NOT NULL,
	cpu REAL DEFAULT 0,
	memory TEXT DEFAULT '{}',
	disks TEXT DEFAULT '[]',
	network TEXT DEFAULT '{}',
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_metrics_agent_time ON metrics(agent_id, timestamp DESC);
`

const migrationBillingCycles = `
CREATE TABLE IF NOT EXISTS billing_cycles (
	id TEXT PRIMARY KEY,
	agent_id TEXT NOT NULL UNIQUE,
	start_date DATETIME NOT NULL,
	duration INTEGER DEFAULT 30,
	limit_bytes INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);
`

const migrationTrafficRecords = `
CREATE TABLE IF NOT EXISTS traffic_records (
	id TEXT PRIMARY KEY,
	cycle_id TEXT NOT NULL,
	agent_id TEXT NOT NULL,
	bytes_sent INTEGER DEFAULT 0,
	bytes_recv INTEGER DEFAULT 0,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (cycle_id) REFERENCES billing_cycles(id) ON DELETE CASCADE,
	FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_traffic_cycle ON traffic_records(cycle_id);
`

const migrationTasks = `
CREATE TABLE IF NOT EXISTS tasks (
	id TEXT PRIMARY KEY,
	type TEXT NOT NULL,
	name TEXT DEFAULT '',
	target TEXT DEFAULT '',
	script_id TEXT DEFAULT '',
	params TEXT DEFAULT '{}',
	interval_sec INTEGER DEFAULT 0,
	timeout_sec INTEGER DEFAULT 60,
	status TEXT DEFAULT 'pending',
	agent_ids TEXT DEFAULT '[]',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
`

const migrationTaskResults = `
CREATE TABLE IF NOT EXISTS task_results (
	id TEXT PRIMARY KEY,
	task_id TEXT NOT NULL,
	agent_id TEXT NOT NULL,
	success INTEGER DEFAULT 0,
	output TEXT DEFAULT '',
	error TEXT DEFAULT '',
	duration_ms INTEGER DEFAULT 0,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
	FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_results_task ON task_results(task_id, timestamp DESC);
`

const migrationScripts = `
CREATE TABLE IF NOT EXISTS scripts (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT DEFAULT '',
	content TEXT NOT NULL,
	checksum TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const migrationAlertRules = `
CREATE TABLE IF NOT EXISTS alert_rules (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	metric_type TEXT NOT NULL,
	operator TEXT NOT NULL,
	threshold REAL NOT NULL,
	duration_sec INTEGER DEFAULT 0,
	cooldown_sec INTEGER DEFAULT 300,
	agent_ids TEXT DEFAULT '[]',
	enabled INTEGER DEFAULT 1,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const migrationAlerts = `
CREATE TABLE IF NOT EXISTS alerts (
	id TEXT PRIMARY KEY,
	rule_id TEXT NOT NULL,
	agent_id TEXT NOT NULL,
	status TEXT DEFAULT 'firing',
	metric_type TEXT NOT NULL,
	value REAL DEFAULT 0,
	threshold REAL DEFAULT 0,
	message TEXT DEFAULT '',
	triggered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	resolved_at DATETIME,
	FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE,
	FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_agent ON alerts(agent_id);
`

const migrationSettings = `
CREATE TABLE IF NOT EXISTS settings (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	data_retention_days INTEGER DEFAULT 7,
	telegram_bot_token TEXT DEFAULT '',
	telegram_chat_id TEXT DEFAULT '',
	smtp_host TEXT DEFAULT '',
	smtp_port INTEGER DEFAULT 587,
	smtp_username TEXT DEFAULT '',
	smtp_password TEXT DEFAULT '',
	smtp_from TEXT DEFAULT '',
	alert_email_to TEXT DEFAULT ''
);
`

const migrationUsers = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	username TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL
);
`
