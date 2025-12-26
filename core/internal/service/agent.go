package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/probe-system/core/internal/models"
	"github.com/probe-system/core/internal/repository"
)

type AgentServiceImpl struct {
	repo    *repository.AgentRepository
	geoSvc  GeoService
}

func NewAgentService(repo *repository.AgentRepository, geoSvc GeoService) *AgentServiceImpl {
	return &AgentServiceImpl{
		repo:   repo,
		geoSvc: geoSvc,
	}
}

func (s *AgentServiceImpl) Register(ctx context.Context, req *RegisterRequest) (*models.Agent, error) {
	// Check if agent already exists by hostname+IP
	existing, err := s.findExisting(ctx, req.Hostname, req.IP)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Update existing agent
		existing.OS = req.OS
		existing.Arch = req.Arch
		existing.Version = req.Version
		existing.Status = models.AgentStatusOnline
		existing.LastSeenAt = time.Now()
		existing.UpdatedAt = time.Now()

		// Update location if IP changed
		if existing.IP != req.IP {
			existing.IP = req.IP
			if loc, err := s.geoSvc.Lookup(req.IP); err == nil {
				existing.Location = loc
			}
		}

		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Create new agent
	agent := &models.Agent{
		ID:         uuid.New().String(),
		Hostname:   req.Hostname,
		IP:         req.IP,
		OS:         req.OS,
		Arch:       req.Arch,
		Version:    req.Version,
		Status:     models.AgentStatusOnline,
		CustomName: req.Hostname,
		Tags:       []string{},
		LastSeenAt: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Lookup geolocation
	if loc, err := s.geoSvc.Lookup(req.IP); err == nil {
		agent.Location = loc
	}

	if err := s.repo.Create(ctx, agent); err != nil {
		return nil, err
	}

	return agent, nil
}

func (s *AgentServiceImpl) findExisting(ctx context.Context, hostname, ip string) (*models.Agent, error) {
	filter := &models.AgentFilter{Search: hostname}
	agents, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	for _, a := range agents {
		if a.Hostname == hostname || a.IP == ip {
			return a, nil
		}
	}
	return nil, nil
}

func (s *AgentServiceImpl) UpdateStatus(ctx context.Context, agentID string, status models.AgentStatus) error {
	return s.repo.UpdateStatus(ctx, agentID, status)
}

func (s *AgentServiceImpl) UpdateLastSeen(ctx context.Context, agentID string) error {
	return s.repo.UpdateLastSeen(ctx, agentID)
}

func (s *AgentServiceImpl) GetByID(ctx context.Context, agentID string) (*models.Agent, error) {
	return s.repo.GetByID(ctx, agentID)
}

func (s *AgentServiceImpl) List(ctx context.Context, filter *models.AgentFilter) ([]*models.Agent, error) {
	return s.repo.List(ctx, filter)
}

func (s *AgentServiceImpl) ListPublic(ctx context.Context) ([]*models.Agent, error) {
	return s.repo.ListPublic(ctx)
}

func (s *AgentServiceImpl) UpdateRemark(ctx context.Context, agentID string, remark *models.AgentRemark) error {
	return s.repo.UpdateRemark(ctx, agentID, remark)
}

func (s *AgentServiceImpl) AssignGroup(ctx context.Context, agentID, groupID string) error {
	return s.repo.AssignGroup(ctx, agentID, groupID)
}

func (s *AgentServiceImpl) SetPublicVisible(ctx context.Context, agentID string, visible bool) error {
	return s.repo.SetPublicVisible(ctx, agentID, visible)
}

func (s *AgentServiceImpl) Delete(ctx context.Context, agentID string) error {
	return s.repo.Delete(ctx, agentID)
}

// Group Service
type GroupServiceImpl struct {
	repo *repository.GroupRepository
}

func NewGroupService(repo *repository.GroupRepository) *GroupServiceImpl {
	return &GroupServiceImpl{repo: repo}
}

func (s *GroupServiceImpl) Create(ctx context.Context, group *models.Group) error {
	group.ID = uuid.New().String()
	group.CreatedAt = time.Now()
	return s.repo.Create(ctx, group)
}

func (s *GroupServiceImpl) Update(ctx context.Context, group *models.Group) error {
	return s.repo.Update(ctx, group)
}

func (s *GroupServiceImpl) Delete(ctx context.Context, groupID string) error {
	return s.repo.Delete(ctx, groupID)
}

func (s *GroupServiceImpl) GetByID(ctx context.Context, groupID string) (*models.Group, error) {
	return s.repo.GetByID(ctx, groupID)
}

func (s *GroupServiceImpl) List(ctx context.Context) ([]*models.Group, error) {
	return s.repo.List(ctx)
}
