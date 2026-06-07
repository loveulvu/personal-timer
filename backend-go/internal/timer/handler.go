package timer

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) StartTask(c *gin.Context) {
	idStr := c.Param("id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || taskID <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid task id",
		})
		return
	}
	if err := h.service.StartTask(taskID); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "task started",
	})

}
func (h *Handler) PauseTask(c *gin.Context) {
	idStr := c.Param("id")

	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || taskID <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid task id",
		})
		return
	}

	if err := h.service.PauseTask(taskID); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "task paused",
	})
}
func (h *Handler) ResumeTask(c *gin.Context) {
	idStr := c.Param("id")

	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || taskID <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid task id",
		})
		return
	}
	if err := h.service.ResumeTask(taskID); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "task resume",
	})

}
func (h *Handler) FinishTask(c *gin.Context) {
	idStr := c.Param("id")

	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || taskID <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid task id",
		})
		return
	}
	if err := h.service.FinishTask(taskID); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "task finished",
	})

}
