package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/probe-system/core/internal/models"
)

type TaskRepository struct {
	db *DB
}

func NewTaskRepository(db *DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	paramsJSON, _ := json.Marshal(task.Params)
	agentIDsJSON, _ := json.Marshal(task.AgentIDs)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (id, type, name, target, script_id, params, interval_sec, 
			timeout_sec, status, agent_ids, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, task.ID, task.Type, task.Name, task.Target, task.ScriptID, string(paramsJSON),
		task.Interval, task.Timeout, task.Status, string(agentIDsJSON),
		task.CreatedAt, task.UpdatedAt)

	return err
}

func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	paramsJSON, _ := json.Marshal(task.Params)
	agentIDsJSON, _ := json.Marshal(task.AgentIDs)

	_, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET type=?, name=?, target=?, script_id=?, params=?, 
			interval_sec=?, timeout_sec=?, status=?, agent_ids=?, updated_at=?
		WHERE id=?
	`, task.Type, task.Name, task.Target, task.ScriptID, string(paramsJSON),
		task.Interval, task.Timeout, task.Status, string(agentIDsJSON),
		time.Now(), task.ID)

	return err
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	task := &models.Task{}
	var paramsJSON, agentIDsJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, type, name, target, script_id, params, interval_sec, 
			timeout_sec, status, agent_ids, created_at, updated_at
		FROM tasks WHERE id = ?
	`, id).Scan(&task.ID, &task.Type, &task.Name, &task.Target, &task.ScriptID,
		&paramsJSON, &task.Interval, &task.Timeout, &task.Status, &agentIDsJSON,
		&task.CreatedAt, &task.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(paramsJSON), &task.Params)
	json.Unmarshal([]byte(agentIDsJSON), &task.AgentIDs)

	return task, nil
}

func (r *TaskRepository) List(ctx context.Context) ([]*models.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, name, target, script_id, params, interval_sec, 
			timeout_sec, status, agent_ids, created_at, updated_at
		FROM tasks ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*models.Task{}
	for rows.Next() {
		task := &models.Task{}
		var paramsJSON, agentIDsJSON string

		if err := rows.Scan(&task.ID, &task.Type, &task.Name, &task.Target, &task.ScriptID,
			&paramsJSON, &task.Interval, &task.Timeout, &task.Status, &agentIDsJSON,
			&task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(paramsJSON), &task.Params)
		json.Unmarshal([]byte(agentIDsJSON), &task.AgentIDs)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *TaskRepository) ListByAgent(ctx context.Context, agentID string) ([]*models.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, name, target, script_id, params, interval_sec, 
			timeout_sec, status, agent_ids, created_at, updated_at
		FROM tasks 
		WHERE agent_ids LIKE ? AND status IN ('pending', 'running')
		ORDER BY created_at DESC
	`, "%"+agentID+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*models.Task{}
	for rows.Next() {
		task := &models.Task{}
		var paramsJSON, agentIDsJSON string

		if err := rows.Scan(&task.ID, &task.Type, &task.Name, &task.Target, &task.ScriptID,
			&paramsJSON, &task.Interval, &task.Timeout, &task.Status, &agentIDsJSON,
			&task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(paramsJSON), &task.Params)
		json.Unmarshal([]byte(agentIDsJSON), &task.AgentIDs)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id string, status models.TaskStatus) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?
	`, status, time.Now(), id)
	return err
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	return err
}

// Task Results
func (r *TaskRepository) RecordResult(ctx context.Context, result *models.TaskResult) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO task_results (id, task_id, agent_id, success, output, error, duration_ms, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, result.ID, result.TaskID, result.AgentID, result.Success, result.Output,
		result.Error, result.Duration, result.Timestamp)
	return err
}

func (r *TaskRepository) GetResults(ctx context.Context, taskID string, limit int) ([]*models.TaskResult, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, task_id, agent_id, success, output, error, duration_ms, timestamp
		FROM task_results WHERE task_id = ?
		ORDER BY timestamp DESC LIMIT ?
	`, taskID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []*models.TaskResult{}
	for rows.Next() {
		result := &models.TaskResult{}
		if err := rows.Scan(&result.ID, &result.TaskID, &result.AgentID, &result.Success,
			&result.Output, &result.Error, &result.Duration, &result.Timestamp); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// Script Repository
type ScriptRepository struct {
	db *DB
}

func NewScriptRepository(db *DB) *ScriptRepository {
	return &ScriptRepository{db: db}
}

func (r *ScriptRepository) Create(ctx context.Context, script *models.Script) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO scripts (id, name, description, content, checksum, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, script.ID, script.Name, script.Description, script.Content, script.Checksum,
		script.CreatedAt, script.UpdatedAt)
	return err
}

func (r *ScriptRepository) Update(ctx context.Context, script *models.Script) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE scripts SET name=?, description=?, content=?, checksum=?, updated_at=?
		WHERE id=?
	`, script.Name, script.Description, script.Content, script.Checksum, time.Now(), script.ID)
	return err
}

func (r *ScriptRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM scripts WHERE id = ?`, id)
	return err
}

func (r *ScriptRepository) GetByID(ctx context.Context, id string) (*models.Script, error) {
	script := &models.Script{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, content, checksum, created_at, updated_at
		FROM scripts WHERE id = ?
	`, id).Scan(&script.ID, &script.Name, &script.Description, &script.Content,
		&script.Checksum, &script.CreatedAt, &script.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return script, nil
}

func (r *ScriptRepository) List(ctx context.Context) ([]*models.Script, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, checksum, created_at, updated_at
		FROM scripts ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scripts := []*models.Script{}
	for rows.Next() {
		script := &models.Script{}
		if err := rows.Scan(&script.ID, &script.Name, &script.Description,
			&script.Checksum, &script.CreatedAt, &script.UpdatedAt); err != nil {
			return nil, err
		}
		scripts = append(scripts, script)
	}

	return scripts, nil
}
