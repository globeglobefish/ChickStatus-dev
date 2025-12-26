package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/probe-system/core/internal/models"
)

type AgentRepository struct {
	db *DB
}

func NewAgentRepository(db *DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) Create(ctx context.Context, agent *models.Agent) error {
	tagsJSON, _ := json.Marshal(agent.Tags)
	locationJSON, _ := json.Marshal(agent.Location)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO agents (id, hostname, ip, os, arch, version, status, group_id, 
			custom_name, description, tags, location, public_visible, last_seen_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, agent.ID, agent.Hostname, agent.IP, agent.OS, agent.Arch, agent.Version,
		agent.Status, agent.GroupID, agent.CustomName, agent.Description,
		string(tagsJSON), string(locationJSON), agent.PublicVisible,
		agent.LastSeenAt, agent.CreatedAt, agent.UpdatedAt)

	return err
}

func (r *AgentRepository) Update(ctx context.Context, agent *models.Agent) error {
	tagsJSON, _ := json.Marshal(agent.Tags)
	locationJSON, _ := json.Marshal(agent.Location)

	_, err := r.db.ExecContext(ctx, `
		UPDATE agents SET hostname=?, ip=?, os=?, arch=?, version=?, status=?, 
			group_id=?, custom_name=?, description=?, tags=?, location=?, 
			public_visible=?, last_seen_at=?, updated_at=?
		WHERE id=?
	`, agent.Hostname, agent.IP, agent.OS, agent.Arch, agent.Version, agent.Status,
		agent.GroupID, agent.CustomName, agent.Description, string(tagsJSON),
		string(locationJSON), agent.PublicVisible, agent.LastSeenAt, time.Now(), agent.ID)

	return err
}

func (r *AgentRepository) GetByID(ctx context.Context, id string) (*models.Agent, error) {
	agent := &models.Agent{}
	var tagsJSON, locationJSON string
	var groupID sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, hostname, ip, os, arch, version, status, group_id, custom_name, 
			description, tags, location, public_visible, last_seen_at, created_at, updated_at
		FROM agents WHERE id = ?
	`, id).Scan(&agent.ID, &agent.Hostname, &agent.IP, &agent.OS, &agent.Arch,
		&agent.Version, &agent.Status, &groupID, &agent.CustomName, &agent.Description,
		&tagsJSON, &locationJSON, &agent.PublicVisible, &agent.LastSeenAt,
		&agent.CreatedAt, &agent.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if groupID.Valid {
		agent.GroupID = &groupID.String
	}
	json.Unmarshal([]byte(tagsJSON), &agent.Tags)
	json.Unmarshal([]byte(locationJSON), &agent.Location)

	return agent, nil
}

func (r *AgentRepository) List(ctx context.Context, filter *models.AgentFilter) ([]*models.Agent, error) {
	query := `SELECT id, hostname, ip, os, arch, version, status, group_id, custom_name, 
		description, tags, location, public_visible, last_seen_at, created_at, updated_at FROM agents WHERE 1=1`
	args := []interface{}{}

	if filter != nil {
		if filter.GroupID != nil {
			query += " AND group_id = ?"
			args = append(args, *filter.GroupID)
		}
		if filter.Status != nil {
			query += " AND status = ?"
			args = append(args, *filter.Status)
		}
		if filter.Search != "" {
			query += " AND (hostname LIKE ? OR custom_name LIKE ? OR ip LIKE ?)"
			search := "%" + filter.Search + "%"
			args = append(args, search, search, search)
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agents := []*models.Agent{}
	for rows.Next() {
		agent := &models.Agent{}
		var tagsJSON, locationJSON string
		var groupID sql.NullString

		err := rows.Scan(&agent.ID, &agent.Hostname, &agent.IP, &agent.OS, &agent.Arch,
			&agent.Version, &agent.Status, &groupID, &agent.CustomName, &agent.Description,
			&tagsJSON, &locationJSON, &agent.PublicVisible, &agent.LastSeenAt,
			&agent.CreatedAt, &agent.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if groupID.Valid {
			agent.GroupID = &groupID.String
		}
		json.Unmarshal([]byte(tagsJSON), &agent.Tags)
		json.Unmarshal([]byte(locationJSON), &agent.Location)

		// Filter by tags if specified
		if filter != nil && len(filter.Tags) > 0 {
			if !r.hasAllTags(agent.Tags, filter.Tags) {
				continue
			}
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

func (r *AgentRepository) hasAllTags(agentTags, filterTags []string) bool {
	tagSet := make(map[string]bool)
	for _, t := range agentTags {
		tagSet[strings.ToLower(t)] = true
	}
	for _, t := range filterTags {
		if !tagSet[strings.ToLower(t)] {
			return false
		}
	}
	return true
}

func (r *AgentRepository) ListPublic(ctx context.Context) ([]*models.Agent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, hostname, os, arch, version, status, group_id, custom_name, 
			description, tags, location, last_seen_at, created_at
		FROM agents WHERE public_visible = 1
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agents := []*models.Agent{}
	for rows.Next() {
		agent := &models.Agent{}
		var tagsJSON, locationJSON string
		var groupID sql.NullString

		err := rows.Scan(&agent.ID, &agent.Hostname, &agent.OS, &agent.Arch,
			&agent.Version, &agent.Status, &groupID, &agent.CustomName, &agent.Description,
			&tagsJSON, &locationJSON, &agent.LastSeenAt, &agent.CreatedAt)
		if err != nil {
			return nil, err
		}

		if groupID.Valid {
			agent.GroupID = &groupID.String
		}
		json.Unmarshal([]byte(tagsJSON), &agent.Tags)
		json.Unmarshal([]byte(locationJSON), &agent.Location)
		agent.IP = "" // Never expose IP in public API

		agents = append(agents, agent)
	}

	return agents, nil
}

func (r *AgentRepository) UpdateStatus(ctx context.Context, id string, status models.AgentStatus) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE agents SET status = ?, updated_at = ? WHERE id = ?
	`, status, time.Now(), id)
	return err
}

func (r *AgentRepository) UpdateLastSeen(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE agents SET last_seen_at = ?, status = 'online', updated_at = ? WHERE id = ?
	`, time.Now(), time.Now(), id)
	return err
}

func (r *AgentRepository) UpdateRemark(ctx context.Context, id string, remark *models.AgentRemark) error {
	tagsJSON, _ := json.Marshal(remark.Tags)
	_, err := r.db.ExecContext(ctx, `
		UPDATE agents SET custom_name = ?, description = ?, tags = ?, updated_at = ? WHERE id = ?
	`, remark.CustomName, remark.Description, string(tagsJSON), time.Now(), id)
	return err
}

func (r *AgentRepository) AssignGroup(ctx context.Context, agentID, groupID string) error {
	var gid interface{} = groupID
	if groupID == "" {
		gid = nil
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE agents SET group_id = ?, updated_at = ? WHERE id = ?
	`, gid, time.Now(), agentID)
	return err
}

func (r *AgentRepository) SetPublicVisible(ctx context.Context, id string, visible bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE agents SET public_visible = ?, updated_at = ? WHERE id = ?
	`, visible, time.Now(), id)
	return err
}

func (r *AgentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM agents WHERE id = ?`, id)
	return err
}

func (r *AgentRepository) CountByStatus(ctx context.Context) (online, offline int, err error) {
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM agents WHERE status = 'online'`).Scan(&online)
	if err != nil {
		return
	}
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM agents WHERE status = 'offline'`).Scan(&offline)
	return
}

// Group Repository
type GroupRepository struct {
	db *DB
}

func NewGroupRepository(db *DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(ctx context.Context, group *models.Group) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO groups (id, name, description, display_order, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, group.ID, group.Name, group.Description, group.DisplayOrder, group.CreatedAt)
	return err
}

func (r *GroupRepository) Update(ctx context.Context, group *models.Group) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE groups SET name = ?, description = ?, display_order = ? WHERE id = ?
	`, group.Name, group.Description, group.DisplayOrder, group.ID)
	return err
}

func (r *GroupRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM groups WHERE id = ?`, id)
	return err
}

func (r *GroupRepository) GetByID(ctx context.Context, id string) (*models.Group, error) {
	group := &models.Group{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, display_order, created_at FROM groups WHERE id = ?
	`, id).Scan(&group.ID, &group.Name, &group.Description, &group.DisplayOrder, &group.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (r *GroupRepository) List(ctx context.Context) ([]*models.Group, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, display_order, created_at 
		FROM groups ORDER BY display_order, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []*models.Group{}
	for rows.Next() {
		group := &models.Group{}
		if err := rows.Scan(&group.ID, &group.Name, &group.Description, 
			&group.DisplayOrder, &group.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (r *GroupRepository) GetAgentCount(ctx context.Context, groupID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM agents WHERE group_id = ?
	`, groupID).Scan(&count)
	return count, err
}
