package dailytasks

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateDailyTask(c *gin.Context) {
	var req CreateDailyTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid json",
		})
		return
	}

	if req.ProjectID == nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "project_id is required",
		})
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "title is required",
		})
		return
	}

	if strings.TrimSpace(req.TaskDate) == "" {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "task_date is required",
		})
		return
	}

	if _, err := time.Parse("2006-01-02", req.TaskDate); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "task_date must be YYYY-MM-DD",
		})
		return
	}

	if req.EstimatedMinutes <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "estimated_minutes must be greater than 0",
		})
		return
	}

	id, err := h.service.CreateDailyTask(req)
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "create daily task failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(201, gin.H{
		"status": "ok",
		"id":     id,
	})
}

func (h *Handler) ListDailyTasksByDate(c *gin.Context) {
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

	tasks, err := h.service.ListDailyTasksByDate(date)
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "list daily tasks failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   tasks,
	})
}

func (h *Handler) GetDailyTaskByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid daily task id",
		})
		return
	}

	task, err := h.service.GetDailyTaskByID(id)
	if err != nil {
		if errors.Is(err, ErrDailyTaskNotFound) {
			c.JSON(404, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "get daily task failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   task,
	})
}

func (h *Handler) UpdateDailyTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid daily task id",
		})
		return
	}

	var req UpdateDailyTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid json",
		})
		return
	}

	if req.ProjectID == nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "project_id is required",
		})
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "title is required",
		})
		return
	}

	if strings.TrimSpace(req.TaskDate) == "" {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "task_date is required",
		})
		return
	}

	if _, err := time.Parse("2006-01-02", req.TaskDate); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "task_date must be YYYY-MM-DD",
		})
		return
	}

	if req.EstimatedMinutes <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "estimated_minutes must be greater than 0",
		})
		return
	}

	if err := h.service.UpdateDailyTask(id, req); err != nil {
		if errors.Is(err, ErrDailyTaskNotFound) {
			c.JSON(404, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		if errors.Is(err, ErrInvalidDailyTaskStatus) ||
			errors.Is(err, ErrInvalidDailyTaskStatusTransition) {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "update daily task failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func (h *Handler) DeleteDailyTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid daily task id",
		})
		return
	}

	if err := h.service.DeleteDailyTask(id); err != nil {
		if errors.Is(err, ErrDailyTaskNotFound) {
			c.JSON(404, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		if errors.Is(err, ErrDailyTaskMustBeCompleted) {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "delete daily task failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
	})
}
