package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/probe-system/core/internal/models"
	"github.com/probe-system/core/internal/repository"
)

type TaskServiceImpl struct {
	repo *repository.TaskRepository
}

func NewTaskService(repo *repository.TaskRepository) *TaskServiceImpl {
	return &TaskServiceImpl{repo: repo}
}

func (s *TaskServiceImpl) Create(ctx context.Context, task *models.Task) error {
	task.ID = uuid.New().String()
	task.Status = models.TaskStatusPending
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	if task.Timeout == 0 {
		task.Timeout = 60
	}
	return s.repo.Create(ctx, task)
}

func (s *TaskServiceImpl) Update(ctx context.Context, task *models.Task) error {
	task.UpdatedAt = time.Now()
	return s.repo.Update(ctx, task)
}

func (s *TaskServiceImpl) Cancel(ctx context.Context, taskID string) error {
	return s.repo.UpdateStatus(ctx, taskID, models.TaskStatusCanceled)
}

func (s *TaskServiceImpl) GetByID(ctx context.Context, taskID string) (*models.Task, error) {
	return s.repo.GetByID(ctx, taskID)
}

func (s *TaskServiceImpl) List(ctx context.Context) ([]*models.Task, error) {
	return s.repo.List(ctx)
}

func (s *TaskServiceImpl) ListByAgent(ctx context.Context, agentID string) ([]*models.Task, error) {
	return s.repo.ListByAgent(ctx, agentID)
}

func (s *TaskServiceImpl) RecordResult(ctx context.Context, result *models.TaskResult) error {
	result.ID = uuid.New().String()
	result.Timestamp = time.Now()
	return s.repo.RecordResult(ctx, result)
}

func (s *TaskServiceImpl) GetResults(ctx context.Context, taskID string, limit int) ([]*models.TaskResult, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.GetResults(ctx, taskID, limit)
}

// Script Service
type ScriptServiceImpl struct {
	repo *repository.ScriptRepository
}

func NewScriptService(repo *repository.ScriptRepository) *ScriptServiceImpl {
	return &ScriptServiceImpl{repo: repo}
}

func (s *ScriptServiceImpl) Create(ctx context.Context, script *models.Script) error {
	script.ID = uuid.New().String()
	script.Checksum = s.computeChecksum(script.Content)
	script.CreatedAt = time.Now()
	script.UpdatedAt = time.Now()
	return s.repo.Create(ctx, script)
}

func (s *ScriptServiceImpl) Update(ctx context.Context, script *models.Script) error {
	script.Checksum = s.computeChecksum(script.Content)
	script.UpdatedAt = time.Now()
	return s.repo.Update(ctx, script)
}

func (s *ScriptServiceImpl) Delete(ctx context.Context, scriptID string) error {
	return s.repo.Delete(ctx, scriptID)
}

func (s *ScriptServiceImpl) GetByID(ctx context.Context, scriptID string) (*models.Script, error) {
	return s.repo.GetByID(ctx, scriptID)
}

func (s *ScriptServiceImpl) List(ctx context.Context) ([]*models.Script, error) {
	return s.repo.List(ctx)
}

func (s *ScriptServiceImpl) computeChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
