package handler

import (
	"database/sql"

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
