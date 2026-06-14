package timer

import (
	"errors"
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
		writeTimerError(c, err)
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
		writeTimerError(c, err)
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
		writeTimerError(c, err)
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
	var input FinishTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "invalid json"})
		return
	}
	if err := h.service.FinishTask(taskID, input); err != nil {
		writeTimerError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "task finished",
	})

}

func (h *Handler) UpdateCompletedTask(c *gin.Context) {
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || taskID <= 0 {
		c.JSON(400, gin.H{"status": "error", "message": "invalid task id"})
		return
	}
	var input UpdateCompletedTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "invalid json"})
		return
	}
	if err := h.service.UpdateCompletedTask(taskID, input); err != nil {
		writeTimerError(c, err)
		return
	}
	c.JSON(200, gin.H{"status": "ok", "message": "completed task updated"})
}

func (h *Handler) DeleteCompletedTask(c *gin.Context) {
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || taskID <= 0 {
		c.JSON(400, gin.H{"status": "error", "message": "invalid task id"})
		return
	}
	if err := h.service.DeleteCompletedTask(taskID); err != nil {
		writeTimerError(c, err)
		return
	}
	c.JSON(200, gin.H{"status": "ok", "message": "completed task deleted"})
}

func writeTimerError(c *gin.Context, err error) {
	if errors.Is(err, ErrTaskNotFound) ||
		errors.Is(err, ErrTaskMustBePlanned) ||
		errors.Is(err, ErrTaskMustBeRunning) ||
		errors.Is(err, ErrTaskMustBePaused) ||
		errors.Is(err, ErrTaskMustBeRunningPaused) ||
		errors.Is(err, ErrTaskMustBeCompleted) ||
		errors.Is(err, ErrRunningSessionNotFound) ||
		errors.Is(err, ErrFinishNoteRequired) ||
		errors.Is(err, ErrFinishDescriptionRequired) ||
		errors.Is(err, ErrActualMinutesInvalid) ||
		errors.Is(err, ErrActualMinutesConflict) {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(500, gin.H{
		"status":  "error",
		"message": "timer operation failed",
	})
}
