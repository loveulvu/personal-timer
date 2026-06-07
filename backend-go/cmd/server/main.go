package main

import (
	"log"
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
	if err := r.Run(":8085"); err != nil {
		log.Fatal(err)
	}
}
