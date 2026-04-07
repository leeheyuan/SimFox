package main

import (
	"config"
	"handlers"
	"net/http"
	"strings"
	"utils"

	"github.com/gin-gonic/gin"
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
	if err := autoMigrate(); err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(corsMiddleware())
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.POST("/create-tenant-user", handlers.CreateTenantUser)
	r.Run(":8080")
}
