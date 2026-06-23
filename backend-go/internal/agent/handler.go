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
	proposals      *ProposalService
}

func NewHandler(registry *ToolRegistry, builder *ContextPackBuilder, runners ...*Runner) *Handler {
	h := &Handler{registry: registry, contextBuilder: builder}
	if len(runners) > 0 {
		h.runner = runners[0]
	}
	return h
}

func (h *Handler) SetProposalService(service *ProposalService) {
	h.proposals = service
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

func (h *Handler) ListRuns(c *gin.Context) {
	if h.runner == nil {
		c.JSON(500, gin.H{"status": "error", "message": "agent runner unavailable"})
		return
	}
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			c.JSON(400, gin.H{"status": "error", "message": "invalid limit"})
			return
		}
		limit = parsed
	}
	items, err := h.runner.ListRuns(c.Request.Context(), c.Query("status"), limit)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"runs": items})
}

func (h *Handler) GetRunTrajectory(c *gin.Context) {
	if h.runner == nil {
		c.JSON(500, gin.H{"status": "error", "message": "agent runner unavailable"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"status": "error", "message": "invalid agent run id"})
		return
	}
	result, err := h.runner.GetTrajectory(c.Request.Context(), id)
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

func (h *Handler) ListActionProposals(c *gin.Context) {
	if h.proposals == nil {
		c.JSON(500, gin.H{"status": "error", "message": "proposal service unavailable"})
		return
	}
	items, err := h.proposals.List(c.Request.Context(), c.Query("status"))
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"proposals": items})
}

func (h *Handler) GetActionProposal(c *gin.Context) {
	if h.proposals == nil {
		c.JSON(500, gin.H{"status": "error", "message": "proposal service unavailable"})
		return
	}
	id, ok := parsePositiveID(c)
	if !ok {
		return
	}
	proposal, err := h.proposals.Get(c.Request.Context(), id)
	if err != nil {
		writeProposalError(c, err)
		return
	}
	c.JSON(200, gin.H{"proposal": proposal})
}

func (h *Handler) AcceptActionProposal(c *gin.Context) {
	if h.proposals == nil {
		c.JSON(500, gin.H{"status": "error", "message": "proposal service unavailable"})
		return
	}
	id, ok := parsePositiveID(c)
	if !ok {
		return
	}
	proposal, err := h.proposals.Accept(c.Request.Context(), id)
	if err != nil {
		writeProposalError(c, err)
		return
	}
	c.JSON(200, gin.H{"proposal": proposal})
}

func (h *Handler) RejectActionProposal(c *gin.Context) {
	if h.proposals == nil {
		c.JSON(500, gin.H{"status": "error", "message": "proposal service unavailable"})
		return
	}
	id, ok := parsePositiveID(c)
	if !ok {
		return
	}
	proposal, err := h.proposals.Reject(c.Request.Context(), id)
	if err != nil {
		writeProposalError(c, err)
		return
	}
	c.JSON(200, gin.H{"proposal": proposal})
}

func parsePositiveID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"status": "error", "message": "invalid id"})
		return 0, false
	}
	return id, true
}

func writeProposalError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrProposalNotFound):
		c.JSON(404, gin.H{"status": "error", "message": err.Error()})
	case errors.Is(err, ErrProposalConflict):
		c.JSON(409, gin.H{"status": "error", "message": err.Error()})
	case errors.Is(err, ErrInvalidToolInput):
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
	default:
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
	}
}
