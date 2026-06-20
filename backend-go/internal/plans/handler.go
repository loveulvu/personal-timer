package plans

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

func (h *Handler) GetPlanRisk(c *gin.Context) {
	result, err := h.service.GetPlanRisk(c.Request.Context(), c.Query("date"))
	if err != nil {
		if errors.Is(err, ErrInvalidPlanRiskDate) {
			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(500, gin.H{"status": "error", "message": "get plan risk failed"})
		return
	}
	c.JSON(200, gin.H{"status": "ok", "data": result})
}
