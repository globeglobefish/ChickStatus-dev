package service

import (
	"context"
	"time"

	"github.com/probe-system/core/internal/models"
)

type AgentService interface {
	Register(ctx context.Context, req *RegisterRequest) (*models.Agent, error)
	UpdateStatus(ctx context.Context, agentID string, status models.AgentStatus) error
	UpdateLastSeen(ctx context.Context, agentID string) error
	GetByID(ctx context.Context, agentID string) (*models.Agent, error)
	List(ctx context.Context, filter *models.AgentFilter) ([]*models.Agent, error)
	ListPublic(ctx context.Context) ([]*models.Agent, error)
	UpdateRemark(ctx context.Context, agentID string, remark *models.AgentRemark) error
	AssignGroup(ctx context.Context, agentID, groupID string) error
	SetPublicVisible(ctx context.Context, agentID string, visible bool) error
	Delete(ctx context.Context, agentID string) error
}

type RegisterRequest struct {
	Hostname string
	IP       string
	OS       string
	Arch     string
	Version  string
}

type GroupService interface {
	Create(ctx context.Context, group *models.Group) error
	Update(ctx context.Context, group *models.Group) error
	Delete(ctx context.Context, groupID string) error
	GetByID(ctx context.Context, groupID string) (*models.Group, error)
	List(ctx context.Context) ([]*models.Group, error)
}

type MetricService interface {
	Store(ctx context.Context, agentID string, metrics *models.Metrics) error
	GetLatest(ctx context.Context, agentID string) (*models.Metrics, error)
	GetHistory(ctx context.Context, agentID string, from, to time.Time) ([]*models.Metrics, error)
	Cleanup(ctx context.Context, retentionDays int) (int64, error)
}

type TrafficService interface {
	RecordTraffic(ctx context.Context, agentID string, bytesSent, bytesRecv uint64) error
	GetStats(ctx context.Context, agentID string) (*models.TrafficStats, error)
	ConfigureCycle(ctx context.Context, agentID string, startDate time.Time, durationDays int, limitBytes uint64) error
	CheckAndResetCycles(ctx context.Context) error
}

type TaskService interface {
	Create(ctx context.Context, task *models.Task) error
	Update(ctx context.Context, task *models.Task) error
	Cancel(ctx context.Context, taskID string) error
	GetByID(ctx context.Context, taskID string) (*models.Task, error)
	List(ctx context.Context) ([]*models.Task, error)
	ListByAgent(ctx context.Context, agentID string) ([]*models.Task, error)
	RecordResult(ctx context.Context, result *models.TaskResult) error
	GetResults(ctx context.Context, taskID string, limit int) ([]*models.TaskResult, error)
}

type ScriptService interface {
	Create(ctx context.Context, script *models.Script) error
	Update(ctx context.Context, script *models.Script) error
	Delete(ctx context.Context, scriptID string) error
	GetByID(ctx context.Context, scriptID string) (*models.Script, error)
	List(ctx context.Context) ([]*models.Script, error)
}

type AlertService interface {
	CreateRule(ctx context.Context, rule *models.AlertRule) error
	UpdateRule(ctx context.Context, rule *models.AlertRule) error
	DeleteRule(ctx context.Context, ruleID string) error
	GetRule(ctx context.Context, ruleID string) (*models.AlertRule, error)
	ListRules(ctx context.Context) ([]*models.AlertRule, error)
	CheckAndTrigger(ctx context.Context, agentID string, metrics *models.Metrics) error
	ResolveAlert(ctx context.Context, alertID string) error
	GetActiveAlerts(ctx context.Context) ([]*models.Alert, error)
	GetAlertHistory(ctx context.Context, limit int) ([]*models.Alert, error)
	AddNotifier(n Notifier)
}

type GeoService interface {
	Lookup(ip string) (*models.GeoLocation, error)
}

type Notifier interface {
	Send(ctx context.Context, alert *models.Alert) error
	SendRecovery(ctx context.Context, alert *models.Alert) error
}

type SettingsService interface {
	Get(ctx context.Context) (*models.Settings, error)
	Update(ctx context.Context, settings *models.Settings) error
}

type AuthService interface {
	Login(ctx context.Context, username, password string) (string, error)
	ValidateToken(token string) (*models.User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
}
