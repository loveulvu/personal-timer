package projects

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid json",
		})
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "need name",
		})
		return
	}

	id, err := h.service.CreateProject(req)
	if err != nil {
		if errors.Is(err, ErrInvalidProjectCategory) {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "create project failed",
		})
		return
	}

	c.JSON(201, gin.H{
		"status": "ok",
		"id":     id,
	})
}
func (h *Handler) ListProjects(c *gin.Context) {
	projects, err := h.service.ListProjects()
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "failed to list projects",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   projects,
	})
}
func (h *Handler) GetProjectByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid project id",
		})
		return
	}
	project, err := h.service.GetProjectByID(id)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			c.JSON(404, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "failed to get project",
		})
		return
	}
	c.JSON(200, gin.H{
		"status": "ok",
		"data":   project,
	})

}

func (h *Handler) UpdateProject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid project id",
		})
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid json",
		})
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "need name",
		})
		return
	}

	if err := h.service.UpdateProject(id, req); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			c.JSON(404, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		if errors.Is(err, ErrInvalidProjectCategory) {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "update project failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func (h *Handler) DeleteProject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid project id",
		})
		return
	}

	if err := h.service.DeleteProject(id); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			c.JSON(404, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "delete project failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
	})
}
