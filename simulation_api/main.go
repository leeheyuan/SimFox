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
	r.POST("/worker/auth", handlers.AuthenticateWorker)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.GET("/overview", handlers.GetSimulationOverview)
	protected.POST("/uploadmap", handlers.UpLoadMap)
	protected.POST("/map/import-bounds", handlers.ImportMapByBounds)

	project := protected.Group("/project")
	project.POST("/generate", handlers.GenerateProject)
	project.POST("/import", handlers.ImportProject)
	project.GET("/projects", handlers.ListProjects)
	project.GET("/tasks", handlers.ListTasks)
	project.GET("/results", handlers.ListResults)
	project.GET("/results/:taskId/signal-optimization", handlers.GetSignalOptimization)
	project.POST("/results/:taskId/signal-optimization/accept", handlers.AcceptSignalOptimization)
	project.POST("/results/:taskId/signal-optimization/reject", handlers.RejectSignalOptimization)
	project.GET("/:id/files", handlers.ListProjectFiles)
	project.POST("/tasks/:taskId/cancel", handlers.CancelTask)
	project.POST("/tasks/:taskId/retry", handlers.RetryTask)
	project.DELETE("/:id", handlers.DeleteProject)
	project.POST("/:id/run", handlers.EnqueueProjectTask)
	project.GET("/:id/tasks", handlers.ListProjectTasks)

	workerAdmin := protected.Group("/worker")
	workerAdmin.GET("/list", handlers.ListWorkers)
	workerAdmin.POST("/register", handlers.RegisterWorker)

	workerRuntime := r.Group("/worker")
	workerRuntime.Use(middleware.WorkerAuthMiddleware())
	workerRuntime.POST("/:workerId/heartbeat", handlers.WorkerHeartbeat)
	workerRuntime.POST("/:workerId/tasks/next", handlers.ClaimNextTask)
	workerRuntime.POST("/:workerId/tasks/:taskId/progress", handlers.UpdateTaskProgress)
	workerRuntime.POST("/:workerId/tasks/:taskId/complete", handlers.CompleteTask)
	workerRuntime.POST("/:workerId/tasks/:taskId/fail", handlers.FailTask)

	r.Run(":8082")
}
