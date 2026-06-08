package timesessions

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const timeLayout = "2006-01-02 15:04:05"

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UpdateTimeSession(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, 400, "invalid time session id")
		return
	}

	req, err := parseUpdateTimeSessionRequest(c)
	if err != nil {
		writeError(c, 400, "invalid json")
		return
	}
	if req == nil {
		return
	}

	if req.StartedAt == "" {
		writeError(c, 400, "started_at is required")
		return
	}
	if req.EndedAt == "" {
		writeError(c, 400, "ended_at is required")
		return
	}

	startedAt, err := time.ParseInLocation(timeLayout, req.StartedAt, time.Local)
	if err != nil {
		writeError(c, 400, "started_at must be YYYY-MM-DD HH:MM:SS")
		return
	}
	endedAt, err := time.ParseInLocation(timeLayout, req.EndedAt, time.Local)
	if err != nil {
		writeError(c, 400, "ended_at must be YYYY-MM-DD HH:MM:SS")
		return
	}

	input := UpdateTimeSessionInput{
		StartedAt: startedAt,
		EndedAt:   endedAt,
	}
	if err := h.service.UpdateFinishedSession(c.Request.Context(), id, input); err != nil {
		if errors.Is(err, ErrInvalidTimeRange) ||
			errors.Is(err, ErrTimeSessionRunning) {
			writeError(c, 400, err.Error())
			return
		}
		if errors.Is(err, ErrTimeSessionNotFound) {
			writeError(c, 404, err.Error())
			return
		}
		writeError(c, 500, "update time session failed")
		return
	}

	c.JSON(200, gin.H{
		"status":           "ok",
		"duration_seconds": int(endedAt.Sub(startedAt).Seconds()),
	})
}

func parseUpdateTimeSessionRequest(c *gin.Context) (*UpdateTimeSessionRequest, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if _, ok := raw["daily_task_id"]; ok {
		writeError(c, 400, "daily_task_id cannot be updated")
		return nil, nil
	}
	if _, ok := raw["duration_seconds"]; ok {
		writeError(c, 400, "duration_seconds cannot be provided")
		return nil, nil
	}

	var req UpdateTimeSessionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": message,
	})
}
