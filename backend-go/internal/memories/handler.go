package memories

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	extractor *Extractor
}

func NewHandler(extractor *Extractor) *Handler {
	return &Handler{extractor: extractor}
}

func (h *Handler) ExtractSummary(c *gin.Context) {
	summaryID, err := strconv.ParseInt(c.Param("summary_id"), 10, 64)
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
