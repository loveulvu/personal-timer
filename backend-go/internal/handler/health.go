package handler

import (
	"database/sql"
	"errors"
	"os"
	"strings"

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

func writeLLMTestError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, llm.ErrMissingAPIKey):
		c.JSON(500, gin.H{"error": "LLM_API_KEY is required"})
	case errors.Is(err, llm.ErrMissingBaseURL):
		c.JSON(500, gin.H{"error": "LLM_BASE_URL is required"})
	case errors.Is(err, llm.ErrMissingModel):
		c.JSON(500, gin.H{"error": "LLM_MODEL is required"})
	case errors.Is(err, llm.ErrNotConfigured):
		c.JSON(500, gin.H{"error": "LLM is not configured"})
	case errors.Is(err, llm.ErrRequestTimeout):
		c.JSON(502, gin.H{"error": "LLM request timed out"})
	case errors.Is(err, llm.ErrRequestFailed):
		c.JSON(502, gin.H{"error": "LLM request failed"})
	case errors.Is(err, llm.ErrEmptyResponse):
		c.JSON(502, gin.H{"error": "LLM returned empty content"})
	default:
		c.JSON(502, gin.H{"error": "LLM test failed"})
	}
}

func isEnvConfigured(key string) bool {
	return strings.TrimSpace(os.Getenv(key)) != ""
}
