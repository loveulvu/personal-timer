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
