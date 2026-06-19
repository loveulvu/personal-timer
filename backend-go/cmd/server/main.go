package main

import (
	"log"
	"personal/internal/dailytasks"
	"personal/internal/db"
	"personal/internal/handler"
	"personal/internal/llm"
	"personal/internal/memories"
	"personal/internal/projects"
	"personal/internal/stats"
	"personal/internal/summaries"
	"personal/internal/timer"
	"personal/internal/timesessions"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	mysqlDB, err := db.NewMySQL()
	if err != nil {
		log.Fatal(err)
	}
	defer mysqlDB.Close()
	r := gin.Default()
	api := r.Group("/api")
	api.GET("/health", handler.Health)
	api.GET("/health/db", handler.HealthDB(mysqlDB))
	api.GET("/version", handler.Version)
	api.GET("/config/status", handler.ConfigStatus(mysqlDB))
	api.POST("/llm/test", handler.TestLLM)
	api.POST("/llm/test-summary", handler.TestLLMSummary)
	projectRepo := projects.NewRepository(mysqlDB)
	projectService := projects.NewService(projectRepo)
	projectHandler := projects.NewHandler(projectService)
	api.POST("/projects", projectHandler.CreateProject)
	api.GET("/projects", projectHandler.ListProjects)
	api.GET("/projects/:id", projectHandler.GetProjectByID)
	api.PUT("/projects/:id", projectHandler.UpdateProject)
	api.DELETE("/projects/:id", projectHandler.DeleteProject)
	dailyTaskRepo := dailytasks.NewRepository(mysqlDB)
	dailyTaskService := dailytasks.NewService(dailyTaskRepo)
	dailyTaskHandler := dailytasks.NewHandler(dailyTaskService)

	api.POST("/daily-tasks", dailyTaskHandler.CreateDailyTask)
	api.GET("/daily-tasks", dailyTaskHandler.ListDailyTasksByDate)
	api.GET("/daily-tasks/:id", dailyTaskHandler.GetDailyTaskByID)
	api.PUT("/daily-tasks/:id", dailyTaskHandler.UpdateDailyTask)
	api.DELETE("/daily-tasks/:id", dailyTaskHandler.DeleteDailyTask)
	timerRepo := timer.NewRepository(mysqlDB)
	timerService := timer.NewService(timerRepo)
	timerHandler := timer.NewHandler(timerService)

	api.POST("/daily-tasks/:id/start", timerHandler.StartTask)
	api.POST("/daily-tasks/:id/pause", timerHandler.PauseTask)
	api.POST("/daily-tasks/:id/resume", timerHandler.ResumeTask)
	api.POST("/daily-tasks/:id/finish", timerHandler.FinishTask)
	api.PUT("/daily-tasks/:id/completion", timerHandler.UpdateCompletedTask)
	api.DELETE("/daily-tasks/:id/completion", timerHandler.DeleteCompletedTask)
	timeSessionRepo := timesessions.NewRepository(mysqlDB)
	timeSessionService := timesessions.NewService(timeSessionRepo)
	timeSessionHandler := timesessions.NewHandler(timeSessionService)

	api.PUT("/time-sessions/:id", timeSessionHandler.UpdateTimeSession)
	statsRepo := stats.NewRepository(mysqlDB)
	statsService := stats.NewService(statsRepo)
	statsHandler := stats.NewHandler(statsService)

	api.GET("/stats/daily", statsHandler.GetDailyStats)
	api.GET("/stats/weekly", statsHandler.GetWeeklyStats)

	llmClient := llm.NewClientFromEnv()
	summaryRepo := summaries.NewRepository(mysqlDB)
	summaryService := summaries.NewService(summaryRepo, statsService, llmClient)
	memoryRepo := memories.NewRepository(mysqlDB)
	memoryExtractor := memories.NewExtractor(memoryRepo)
	memoryRecall := memories.NewRecallService(memoryRepo)
	summaryService.SetMemoryExtractor(memoryExtractor)
	summaryService.SetMemoryRecall(memoryRecall)
	summaryHandler := summaries.NewHandler(summaryService)

	api.POST("/summaries/daily/generate", summaryHandler.GenerateDailySummary)
	api.POST("/summaries/weekly/generate", summaryHandler.GenerateWeeklySummary)
	api.GET("/summaries", summaryHandler.ListSummaries)
	api.GET("/summaries/:id", summaryHandler.GetSummaryByID)
	api.POST("/summaries/:summary_id/action-items/:item_index/accept", summaryHandler.AcceptActionItem)
	api.DELETE("/summaries/:id", summaryHandler.DeleteSummary)
	memoryHandler := memories.NewHandler(memoryExtractor)
	api.POST("/memories/extract/summary/:summary_id", memoryHandler.ExtractSummary)
	if err := r.Run(":8085"); err != nil {
		log.Fatal(err)
	}
}
func loadEnv() {
	paths := []string{
		"../.env",
		".env",
		"/home/u1/projects/personal-study-timer/.env",
	}

	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("loaded env file: %s", path)
			return
		}
	}

	log.Println("warning: no .env file loaded")
}
