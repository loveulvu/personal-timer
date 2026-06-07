package main

import (
	"log"
	"personal/internal/dailytasks"
	"personal/internal/db"
	"personal/internal/handler"
	"personal/internal/projects"

	"github.com/gin-gonic/gin"
)

func main() {
	mysqlDB, err := db.NewMySQL()
	if err != nil {
		log.Fatal(err)
	}
	defer mysqlDB.Close()
	r := gin.Default()
	api := r.Group("/api")
	api.GET("/health", handler.Health)
	api.GET("/health/db", handler.HealthDB(mysqlDB))
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
	if err := r.Run(":8085"); err != nil {
		log.Fatal(err)
	}
}
