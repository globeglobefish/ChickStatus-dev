package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/probe-system/core/internal/service"
)

type ScriptHandler struct {
	scriptSvc service.ScriptService
}

func NewScriptHandler(scriptSvc service.ScriptService) *ScriptHandler {
	return &ScriptHandler{scriptSvc: scriptSvc}
}

// GetScriptContent returns the script content for agent download
// This endpoint should be protected by agent token validation
func (h *ScriptHandler) GetScriptContent(c *gin.Context) {
	script, err := h.scriptSvc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if script == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("X-Script-Checksum", script.Checksum)
	c.String(http.StatusOK, script.Content)
}
