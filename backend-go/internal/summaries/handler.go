package summaries

import (
	"errors"
	"strconv"
	"time"

	"personal/internal/llm"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GenerateDailySummary(c *gin.Context) {
	var req GenerateDailySummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, 400, "invalid json")
		return
	}

	if req.Date == "" {
		writeError(c, 400, "date is required")
		return
	}

	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		writeError(c, 400, "date must be YYYY-MM-DD")
		return
	}

	result, err := h.service.GenerateDailySummary(c.Request.Context(), req.Date)
	if err != nil {
		writeGenerateError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   result,
	})
}

func (h *Handler) GenerateWeeklySummary(c *gin.Context) {
	var req GenerateWeeklySummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, 400, "invalid json")
		return
	}

	if req.StartDate == "" || req.EndDate == "" {
		writeError(c, 400, "start_date and end_date are required")
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		writeError(c, 400, "start_date must be YYYY-MM-DD")
		return
	}

	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		writeError(c, 400, "end_date must be YYYY-MM-DD")
		return
	}

	if end.Before(start) {
		writeError(c, 400, "end_date cannot be earlier than start_date")
		return
	}

	result, err := h.service.GenerateWeeklySummary(c.Request.Context(), req.StartDate, req.EndDate)
	if err != nil {
		writeGenerateError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   result,
	})
}

func (h *Handler) ListSummaries(c *gin.Context) {
	summaryType := c.Query("type")
	if summaryType != "" && summaryType != "daily" && summaryType != "weekly" {
		writeError(c, 400, "type must be daily or weekly")
		return
	}

	result, err := h.service.ListSummaries(c.Request.Context(), summaryType)
	if err != nil {
		writeError(c, 500, "list summaries failed")
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   result,
	})
}

func (h *Handler) GetSummaryByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, 400, "invalid summary id")
		return
	}

	result, err := h.service.GetSummaryByID(c.Request.Context(), id)
	if errors.Is(err, ErrSummaryNotFound) {
		writeError(c, 404, err.Error())
		return
	}
	if err != nil {
		writeError(c, 500, "get summary failed")
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   result,
	})
}

func (h *Handler) AcceptActionItem(c *gin.Context) {
	summaryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || summaryID <= 0 {
		writeError(c, 400, "invalid summary id")
		return
	}
	itemIndex, err := strconv.Atoi(c.Param("item_index"))
	if err != nil || itemIndex < 0 {
		writeError(c, 400, "invalid action item index")
		return
	}
	var req AcceptActionItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, 400, "invalid json")
		return
	}
	if req.TargetDate == "" {
		writeError(c, 400, "target_date is required")
		return
	}

	result, err := h.service.AcceptActionItem(c.Request.Context(), summaryID, itemIndex, req.TargetDate)
	if errors.Is(err, ErrSummaryNotFound) {
		writeError(c, 404, err.Error())
		return
	}
	if errors.Is(err, ErrActionItemIndexInvalid) ||
		errors.Is(err, ErrActionItemNotAcceptable) ||
		errors.Is(err, ErrActionItemProjectInvalid) ||
		errors.Is(err, ErrActionItemTargetDateInvalid) {
		writeError(c, 400, err.Error())
		return
	}
	if err != nil {
		writeError(c, 500, "accept action item failed")
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   result,
	})
}

func (h *Handler) ListActionItemAcceptances(c *gin.Context) {
	summaryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || summaryID <= 0 {
		writeError(c, 400, "invalid summary id")
		return
	}
	result, err := h.service.ListActionItemAcceptances(c.Request.Context(), summaryID)
	if err != nil {
		writeError(c, 500, "list action item acceptances failed")
		return
	}
	c.JSON(200, gin.H{"status": "ok", "data": result})
}

func (h *Handler) DeleteSummary(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, 400, "invalid summary id")
		return
	}

	if err := h.service.DeleteSummary(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrSummaryNotFound) {
			writeError(c, 404, err.Error())
			return
		}
		writeError(c, 500, "delete summary failed")
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func writeGenerateError(c *gin.Context, err error) {
	if errors.Is(err, ErrSummaryAlreadyExists) {
		writeError(c, 409, "summary already exists")
		return
	}
	if errors.Is(err, ErrStatsQueryFailed) {
		writeError(c, 500, err.Error())
		return
	}
	if errors.Is(err, ErrSummaryPersistFailed) {
		writeError(c, 500, err.Error())
		return
	}
	if errors.Is(err, llm.ErrMissingAPIKey) {
		writeError(c, 500, "LLM_API_KEY is required")
		return
	}
	if errors.Is(err, llm.ErrMissingBaseURL) {
		writeError(c, 500, "LLM_BASE_URL is required")
		return
	}
	if errors.Is(err, llm.ErrMissingModel) {
		writeError(c, 500, "LLM_MODEL is required")
		return
	}
	if errors.Is(err, llm.ErrNotConfigured) {
		writeError(c, 500, "LLM is not configured")
		return
	}
	if errors.Is(err, llm.ErrRequestTimeout) {
		writeError(c, 502, "LLM request timed out")
		return
	}
	if errors.Is(err, llm.ErrRequestFailed) {
		writeError(c, 502, "LLM request failed")
		return
	}
	if errors.Is(err, llm.ErrEmptyResponse) {
		writeError(c, 502, "LLM returned empty content")
		return
	}
	if errors.Is(err, ErrLLMGenerationFailed) {
		writeError(c, 502, "LLM generation failed")
		return
	}

	writeError(c, 500, "generate summary failed")
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": message,
	})
}
