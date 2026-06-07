package stats

import (
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetDailyStats(c *gin.Context) {
	date := c.Query("date")

	if date == "" {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "date is required",
		})
		return
	}

	if _, err := time.Parse("2006-01-02", date); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "date must be YYYY-MM-DD",
		})
		return
	}

	result, err := h.service.GetDailyStats(date)
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "get daily stats failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   result,
	})
}

func (h *Handler) GetWeeklyStats(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "start_date must be YYYY-MM-DD",
		})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
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

	result, err := h.service.GetWeeklyStats(startDate, endDate)
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "get weekly stats failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   result,
	})
}
