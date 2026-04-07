package main

import (
	"config"
	"handlers"
	"middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {
	config.InitDB()
	if err := autoMigrate(); err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(corsMiddleware())
	r.POST("/map/convert-osm", handlers.ConvertUploadedOSM)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.GET("/overview", handlers.GetSimulationOverview)
	protected.POST("/uploadmap", handlers.UpLoadMap)
	protected.POST("/map/import-bounds", handlers.ImportMapByBounds)

	project := protected.Group("/project")
	project.POST("/generate", handlers.GenerateProject)
	project.GET("/projects", handlers.ListProjects)
	project.POST("/:id/run", handlers.EnqueueProjectTask)
	project.GET("/:id/tasks", handlers.ListProjectTasks)

	r.Run(":8082")
}
