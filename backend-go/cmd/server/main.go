package main

import (
	"log"
	"personal/internal/db"
	"personal/internal/handler"

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
	if err := r.Run(":8085"); err != nil {
		log.Fatal(err)
	}
}
