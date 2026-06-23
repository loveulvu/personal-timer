package agent

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	registry       *ToolRegistry
	contextBuilder *ContextPackBuilder
	runner         *Runner
}

func NewHandler(registry *ToolRegistry, builder *ContextPackBuilder, runners ...*Runner) *Handler {
	h := &Handler{registry: registry, contextBuilder: builder}
	if len(runners) > 0 {
		h.runner = runners[0]
	}
	return h
}

func (h *Handler) ListTools(c *gin.Context) {
	c.JSON(200, gin.H{"tools": h.registry.ListTools()})
}

func (h *Handler) CallTool(c *gin.Context) {
	var call ToolCall
	if err := c.ShouldBindJSON(&call); err != nil {
		c.JSON(400, gin.H{"success": false, "error_message": "invalid tool call json"})
		return
	}

	result, err := h.registry.Call(c.Request.Context(), call)
	if err != nil {
		switch {
		case errors.Is(err, ErrToolNotFound):
			c.JSON(400, gin.H{"success": false, "error_message": "unknown tool"})
		case errors.Is(err, ErrInvalidToolInput):
			c.JSON(400, gin.H{"success": false, "error_message": "invalid tool input"})
		default:
			c.JSON(500, gin.H{"success": false, "error_message": err.Error()})
		}
		return
	}
	c.JSON(200, result)
}

func (h *Handler) ContextPreview(c *gin.Context) {
	var req ContextPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "invalid context preview json"})
		return
	}
	if h.contextBuilder == nil {
		c.JSON(500, gin.H{"status": "error", "message": "context builder unavailable"})
		return
	}
	pack, err := h.contextBuilder.Build(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidContextPreviewInput) {
			c.JSON(400, gin.H{"status": "error", "message": "invalid context preview input"})
			return
		}
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(200, ContextPreviewResponse{ContextPack: pack})
}

func (h *Handler) CreateRun(c *gin.Context) {
	var req AgentRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "invalid agent run json"})
		return
	}
	if h.runner == nil {
		c.JSON(500, gin.H{"status": "error", "message": "agent runner unavailable"})
		return
	}
	result, err := h.runner.Start(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidContextPreviewInput) {
			c.JSON(400, gin.H{"status": "error", "message": "invalid agent run input"})
			return
		}
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(200, result)
}

func (h *Handler) GetRun(c *gin.Context) {
	if h.runner == nil {
		c.JSON(500, gin.H{"status": "error", "message": "agent runner unavailable"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"status": "error", "message": "invalid agent run id"})
		return
	}
	result, err := h.runner.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrAgentRunNotFound) {
			c.JSON(404, gin.H{"status": "error", "message": "agent run not found"})
			return
		}
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(200, result)
}
