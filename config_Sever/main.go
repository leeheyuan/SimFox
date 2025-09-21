package main

import (
	"config"
	"config_Sever/docs"
	_ "config_Sever/docs"
	"handlers"
	"middleware"
	"net/http"
	"strings"
	"utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置CORS相关头部
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		// 如果是OPTIONS请求，直接返回
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		noAuthPaths := []string{
			"/login",
			"/register",
			"/SimulationConfig",
			"/create-tenant-user",
			"/swagger/", // 匹配 Swagger 静态资源路径
			"/swagger/index.html",
			"/swagger/doc.json",
		}

		requestPath := c.Request.URL.Path

		// 判断当前路径是否属于白名单
		for _, path := range noAuthPaths {
			if strings.HasPrefix(requestPath, path) {
				c.Next()
				return
			}
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		username, err := utils.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("username", username)

		// 继续处理
		c.Next()
	}
}

// @title Simulation platform  API
// @version 1.0
// @description Saas仿真平台RESTful.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	config.InitDB()

	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/SimulationConfig/:filename", handlers.Download)

	r.POST("/uploadmap", handlers.UpLoadMap)
	r.GET("/getGeojson", handlers.GetGeojson)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	auth.GET("/profile", func(c *gin.Context) {
		username := c.MustGet("username").(string)
		c.JSON(200, gin.H{"message": "Welcome " + username})
	})
	docs.SwaggerInfo.Host = "127.0.0.1:8081"
	r.Run(":8081")
}
