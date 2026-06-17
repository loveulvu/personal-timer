package handler

import (
	"database/sql"
	"errors"
	"os"
	"strings"
	"time"

	"personal/internal/llm"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
func HealthDB(mysqlDB *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := mysqlDB.Ping(); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"db":     "err",
			})
			return
		}
		c.JSON(200, gin.H{
			"status": "ok",
			"db":     "mysql",
		})

	}
}

func Version(c *gin.Context) {
	c.JSON(200, gin.H{
		"name":    "personal-study-timer",
		"version": "2.2.0",
		"mode":    "local-api-server",
	})
}

func ConfigStatus(mysqlDB *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := gin.H{
			"database":                "ok",
			"llm_configured":          isEnvConfigured("LLM_API_KEY"),
			"llm_base_url_configured": isEnvConfigured("LLM_BASE_URL"),
			"llm_model_configured":    isEnvConfigured("LLM_MODEL"),
		}

		if err := mysqlDB.Ping(); err != nil {
			status["database"] = "error"
			status["error"] = "database ping failed"
			c.JSON(500, status)
			return
		}

		c.JSON(200, status)
	}
}

func TestLLM(c *gin.Context) {
	client := llm.NewClientFromEnv()
	_, err := client.GenerateSummary(c.Request.Context(), "Reply with OK if the LLM connection works.")
	if err != nil {
		writeLLMTestError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "LLM connection works",
	})
}

func TestLLMSummary(c *gin.Context) {
	client := llm.NewClientFromEnv()
	prompt := summaryLLMTestPrompt()
	start := time.Now()
	response, err := client.GenerateSummary(c.Request.Context(), prompt)
	elapsed := time.Since(start)
	if err != nil {
		c.JSON(llmTestErrorStatus(err), gin.H{
			"status":       "error",
			"elapsed_ms":   elapsed.Milliseconds(),
			"prompt_chars": len([]rune(prompt)),
			"error":        llmTestErrorMessage(err),
		})
		return
	}

	c.JSON(200, gin.H{
		"status":           "ok",
		"elapsed_ms":       elapsed.Milliseconds(),
		"prompt_chars":     len([]rune(prompt)),
		"response_preview": truncateForResponse(response, 500),
	})
}

func writeLLMTestError(c *gin.Context, err error) {
	c.JSON(llmTestErrorStatus(err), gin.H{"error": llmTestErrorMessage(err)})
}

func llmTestErrorStatus(err error) int {
	switch {
	case errors.Is(err, llm.ErrMissingAPIKey),
		errors.Is(err, llm.ErrMissingBaseURL),
		errors.Is(err, llm.ErrMissingModel),
		errors.Is(err, llm.ErrNotConfigured):
		return 500
	default:
		return 502
	}
}

func llmTestErrorMessage(err error) string {
	switch {
	case errors.Is(err, llm.ErrMissingAPIKey):
		return "LLM_API_KEY is required"
	case errors.Is(err, llm.ErrMissingBaseURL):
		return "LLM_BASE_URL is required"
	case errors.Is(err, llm.ErrMissingModel):
		return "LLM_MODEL is required"
	case errors.Is(err, llm.ErrNotConfigured):
		return "LLM is not configured"
	case errors.Is(err, llm.ErrRequestTimeout):
		return "LLM request timed out"
	case errors.Is(err, llm.ErrRequestFailed):
		return "LLM request failed"
	case errors.Is(err, llm.ErrEmptyResponse):
		return "LLM returned empty content"
	default:
		return "LLM test failed"
	}
}

func summaryLLMTestPrompt() string {
	block := `Daily/Weekly summary test data:
- Date range: 2026-06-10 to 2026-06-16
- Study focus: Go backend reliability, SQL review, English reading, frontend cleanup
- Completed sessions: 11 blocks, 1470 minutes total
- Interruptions: unstable network, one long debugging session, two unfinished review tasks
- Notes: compare planned vs actual time, point out risks, mention overdue items, avoid praise, output concise action items.
`
	return strings.Repeat(block, 8)
}

func truncateForResponse(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "...(truncated)"
}

func isEnvConfigured(key string) bool {
	return strings.TrimSpace(os.Getenv(key)) != ""
}
