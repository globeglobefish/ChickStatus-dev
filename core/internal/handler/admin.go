package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/probe-system/core/internal/models"
	"github.com/probe-system/core/internal/service"
	"github.com/probe-system/core/internal/ws"
)

type AdminHandler struct {
	agentSvc    service.AgentService
	groupSvc    service.GroupService
	metricSvc   service.MetricService
	trafficSvc  service.TrafficService
	taskSvc     service.TaskService
	scriptSvc   service.ScriptService
	alertSvc    service.AlertService
	settingsSvc service.SettingsService
	authSvc     service.AuthService
	wsHandler   *ws.Handler
}

func NewAdminHandler(
	agentSvc service.AgentService,
	groupSvc service.GroupService,
	metricSvc service.MetricService,
	trafficSvc service.TrafficService,
	taskSvc service.TaskService,
	scriptSvc service.ScriptService,
	alertSvc service.AlertService,
	settingsSvc service.SettingsService,
	authSvc service.AuthService,
	wsHandler *ws.Handler,
) *AdminHandler {
	return &AdminHandler{
		agentSvc:    agentSvc,
		groupSvc:    groupSvc,
		metricSvc:   metricSvc,
		trafficSvc:  trafficSvc,
		taskSvc:     taskSvc,
		scriptSvc:   scriptSvc,
		alertSvc:    alertSvc,
		settingsSvc: settingsSvc,
		authSvc:     authSvc,
		wsHandler:   wsHandler,
	}
}

// Auth
func (h *AdminHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authSvc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Agents
func (h *AdminHandler) ListAgents(c *gin.Context) {
	filter := &models.AgentFilter{}

	if groupID := c.Query("group_id"); groupID != "" {
		filter.GroupID = &groupID
	}
	if status := c.Query("status"); status != "" {
		s := models.AgentStatus(status)
		filter.Status = &s
	}
	if search := c.Query("search"); search != "" {
		filter.Search = search
	}

	agents, err := h.agentSvc.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, agents)
}

func (h *AdminHandler) GetAgent(c *gin.Context) {
	agent, err := h.agentSvc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if agent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func (h *AdminHandler) UpdateAgentRemark(c *gin.Context) {
	var req models.AgentRemark
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.agentSvc.UpdateRemark(c.Request.Context(), c.Param("id"), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AdminHandler) AssignAgentGroup(c *gin.Context) {
	var req struct {
		GroupID string `json:"group_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.agentSvc.AssignGroup(c.Request.Context(), c.Param("id"), req.GroupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AdminHandler) SetAgentVisibility(c *gin.Context) {
	var req struct {
		Visible bool `json:"visible"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.agentSvc.SetPublicVisible(c.Request.Context(), c.Param("id"), req.Visible); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AdminHandler) DeleteAgent(c *gin.Context) {
	if err := h.agentSvc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Groups
func (h *AdminHandler) ListGroups(c *gin.Context) {
	groups, err := h.groupSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}

func (h *AdminHandler) CreateGroup(c *gin.Context) {
	var group models.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.groupSvc.Create(c.Request.Context(), &group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, group)
}

func (h *AdminHandler) UpdateGroup(c *gin.Context) {
	var group models.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	group.ID = c.Param("id")

	if err := h.groupSvc.Update(c.Request.Context(), &group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *AdminHandler) DeleteGroup(c *gin.Context) {
	if err := h.groupSvc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Metrics
func (h *AdminHandler) GetAgentMetrics(c *gin.Context) {
	agentID := c.Param("id")

	latest, err := h.metricSvc.GetLatest(c.Request.Context(), agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, latest)
}

func (h *AdminHandler) GetAgentMetricsHistory(c *gin.Context) {
	agentID := c.Param("id")

	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	to := time.Now()
	from := to.Add(-time.Duration(hours) * time.Hour)

	history, err := h.metricSvc.GetHistory(c.Request.Context(), agentID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *AdminHandler) GetAgentTraffic(c *gin.Context) {
	stats, err := h.trafficSvc.GetStats(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AdminHandler) ConfigureTrafficCycle(c *gin.Context) {
	var req struct {
		StartDate string `json:"start_date"`
		Duration  int    `json:"duration"`
		Limit     uint64 `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
		return
	}

	if err := h.trafficSvc.ConfigureCycle(c.Request.Context(), c.Param("id"), startDate, req.Duration, req.Limit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Tasks
func (h *AdminHandler) ListTasks(c *gin.Context) {
	tasks, err := h.taskSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *AdminHandler) CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskSvc.Create(c.Request.Context(), &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Assign to agents
	for _, agentID := range task.AgentIDs {
		h.wsHandler.AssignTask(agentID, &task)
	}

	c.JSON(http.StatusCreated, task)
}

func (h *AdminHandler) GetTaskResults(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	results, err := h.taskSvc.GetResults(c.Request.Context(), c.Param("id"), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *AdminHandler) CancelTask(c *gin.Context) {
	if err := h.taskSvc.Cancel(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Scripts
func (h *AdminHandler) ListScripts(c *gin.Context) {
	scripts, err := h.scriptSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, scripts)
}

func (h *AdminHandler) CreateScript(c *gin.Context) {
	var script models.Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.scriptSvc.Create(c.Request.Context(), &script); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, script)
}

func (h *AdminHandler) GetScript(c *gin.Context) {
	script, err := h.scriptSvc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if script == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
		return
	}

	c.JSON(http.StatusOK, script)
}

func (h *AdminHandler) DeleteScript(c *gin.Context) {
	if err := h.scriptSvc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Alerts
func (h *AdminHandler) ListAlertRules(c *gin.Context) {
	rules, err := h.alertSvc.ListRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}

func (h *AdminHandler) CreateAlertRule(c *gin.Context) {
	var rule models.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.alertSvc.CreateRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

func (h *AdminHandler) UpdateAlertRule(c *gin.Context) {
	var rule models.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule.ID = c.Param("id")

	if err := h.alertSvc.UpdateRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (h *AdminHandler) DeleteAlertRule(c *gin.Context) {
	if err := h.alertSvc.DeleteRule(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AdminHandler) GetActiveAlerts(c *gin.Context) {
	alerts, err := h.alertSvc.GetActiveAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

func (h *AdminHandler) GetAlertHistory(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	alerts, err := h.alertSvc.GetAlertHistory(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// Settings
func (h *AdminHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingsSvc.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *AdminHandler) UpdateSettings(c *gin.Context) {
	var settings models.Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.settingsSvc.Update(c.Request.Context(), &settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	if err := h.authSvc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
