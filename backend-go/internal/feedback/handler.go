package feedback

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

func (h *Handler) SubmitFeedback(c *gin.Context) {
	var req SubmitFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "invalid json"})
		return
	}
	result, err := h.service.SubmitFeedback(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidFeedbackTargetType) ||
			errors.Is(err, ErrInvalidFeedbackTargetID) ||
			errors.Is(err, ErrInvalidFeedbackTargetIndex) ||
			errors.Is(err, ErrInvalidFeedbackValue) ||
			errors.Is(err, ErrFeedbackNoteTooLong) {
			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(500, gin.H{"status": "error", "message": "submit feedback failed"})
		return
	}
	c.JSON(200, gin.H{"status": "ok", "data": result})
}
