package tasks

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) EstimatePreview(c *gin.Context) {
	var req EstimatePreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid json",
		})
		return
	}

	result, err := h.service.EstimatePreview(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidProjectID) || errors.Is(err, ErrInvalidEstimatedMinutes) {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		if errors.Is(err, ErrProjectNotFound) {
			c.JSON(404, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "estimate preview failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   result,
	})
}
