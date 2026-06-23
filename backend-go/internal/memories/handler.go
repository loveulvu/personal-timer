package memories

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	extractor *Extractor
	repo      *Repository
}

func NewHandler(extractor *Extractor, repo *Repository) *Handler {
	return &Handler{extractor: extractor, repo: repo}
}

func (h *Handler) ListMemories(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.repo.ListMemoriesForUI(c.Request.Context(), ListMemoryItemsFilter{
		Status:     c.DefaultQuery("status", "active"),
		MemoryType: c.Query("memory_type"),
		Limit:      limit,
	})
	if errors.Is(err, ErrInvalidMemoryInput) {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "list memories failed"})
		return
	}
	c.JSON(200, gin.H{"status": "ok", "data": items})
}

func (h *Handler) ListMemoryEvidence(c *gin.Context) {
	memoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || memoryID <= 0 {
		c.JSON(400, gin.H{"status": "error", "message": "invalid memory id"})
		return
	}
	items, err := h.repo.ListMemoryEvidence(c.Request.Context(), memoryID)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "list memory evidence failed"})
		return
	}
	c.JSON(200, gin.H{"status": "ok", "data": items})
}

func (h *Handler) ExtractSummary(c *gin.Context) {
	summaryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || summaryID <= 0 {
		c.JSON(400, gin.H{"error": "invalid summary id"})
		return
	}

	result, err := h.extractor.ExtractFromSummary(c.Request.Context(), summaryID)
	if errors.Is(err, ErrSummaryNotFound) {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, ErrInvalidSourceData) {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "extract memories failed"})
		return
	}

	c.JSON(200, gin.H{"status": "success", "data": result})
}
