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
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid json",
		})
		return
	}

	if req.Date == "" {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "date is required",
		})
		return
	}

	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "date must be YYYY-MM-DD",
		})
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
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid json",
		})
		return
	}

	if req.StartDate == "" || req.EndDate == "" {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "start_date must be YYYY-MM-DD",
		})
		return
	}

	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "end_date must be YYYY-MM-DD",
		})
		return
	}

	if end.Before(start) {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "end_date cannot be earlier than start_date",
		})
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
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "type must be daily or weekly",
		})
		return
	}

	result, err := h.service.ListSummaries(c.Request.Context(), summaryType)
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "list summaries failed",
		})
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
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid summary id",
		})
		return
	}

	result, err := h.service.GetSummaryByID(c.Request.Context(), id)
	if errors.Is(err, ErrSummaryNotFound) {
		c.JSON(404, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "get summary failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   result,
	})
}

func writeGenerateError(c *gin.Context, err error) {
	if errors.Is(err, llm.ErrNotConfigured) {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "LLM is not configured",
		})
		return
	}
	if errors.Is(err, ErrLLMGenerationFailed) {
		c.JSON(502, gin.H{
			"status":  "error",
			"message": "LLM generation failed",
		})
		return
	}

	c.JSON(500, gin.H{
		"status":  "error",
		"message": "generate summary failed",
	})
}
