package handlers

import (
	"config"
	"fmt"
	"models"
	"net/http"
	"time"
	"utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string `json:"username" example:"liheyuan"`
	Password string `json:"password" example:"123456"`
}

type CreateTenantUserRequest struct {
	TenantName string `json:"tenant_name" binding:"required"`
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// Register godoc
// @Summary      用户注册
// @Description  注册一个新用户，提供用户名和密码
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "注册请求体"
// @Success      200 {object} map[string]string "注册成功提示"
// @Failure      400 {object} map[string]string "请求参数错误"
// @Failure      409 {object} map[string]string "用户名已存在"
// @Failure      500 {object} map[string]string "服务器内部错误"
// @Router       /register [post]
func Register(c *gin.Context) {
	fmt.Println("Register")
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user := models.User{Username: req.Username, Password: req.Password}
	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

// Login godoc
// @Summary 用户登录
// @Description 用户通过用户名和密码进行登录，成功后返回 JWT Token
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "登录请求体"
// @Success 200 {object} map[string]string "{"token": "jwt_token"}"
// @Failure 400 {object} map[string]string "{"error": "Invalid request"}"
// @Failure 401 {object} map[string]string "{"error": "Invalid credentials"}"
// @Failure 500 {object} map[string]string "{"error": "Could not generate token"}"
// @Router /login [post]
func Login(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := config.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": "Bearer " + token})
}

// CreateTenantUser godoc
// @Summary 创建租户和管理员用户
// @Description 创建一个新的租户，并为其创建一个管理员用户
// @Tags 用户
// @Accept  json
// @Produce  json
// @Param   data  body  CreateTenantUserRequest  true  "租户与管理员信息"
// @Success 200 {object} map[string]string "{"token": "jwt_token"}"
// @Failure 400   {object} map[string]string
// @Failure 500   {object} map[string]string
// @Router /create-tenant-user [post]
func CreateTenantUser(c *gin.Context) {
	var req CreateTenantUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 先创建租户
	tenant := models.Tenant{
		Name:      req.TenantName,
		CreatedAt: time.Now(),
	}
	if err := config.DB.Create(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建租户失败: " + err.Error()})
		return
	}

	// 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建管理员用户，关联租户ID
	user := models.User{
		TenantID:  tenant.ID,
		Username:  req.Username,
		Password:  string(hashedPassword),
		Role:      "admin",
		CreatedAt: time.Now(),
	}
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败: " + err.Error()})
		return
	}

	token, err := utils.GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": "Bearer " + token})
}

// @Summary 获取用户首页信息
// @Tags 用户
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /overview [get]
// @Security BearerAuth
func GetUserOverview(c *gin.Context) {

	username := c.GetString("username")
	var projectCount int64
	var simulationCount int64
	var mapCount int64
	var recentProjects []models.SimulationProject

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	config.DB.Model(&models.SimulationProject{}).Where("tenant_id = ?", user.ID).Count(&projectCount)
	config.DB.Model(&models.Simulation{}).Where("tenant_id = ?", user.ID).Count(&simulationCount)
	config.DB.Model(&models.MapData{}).Where("tenant_id = ?", user.ID).Count(&mapCount)
	// 查询最近打开的项目，按更新时间倒序取前 5 个
	config.DB.Where("tenant_id = ?", user.ID).
		Order("updated_at desc").
		Limit(5).
		Find(&recentProjects)

	// 返回 JSON 响应
	c.JSON(http.StatusOK, gin.H{
		"project_count":    projectCount,
		"simulation_count": simulationCount,
		"map_count":        mapCount,
		"recent_projects":  recentProjects,
	})

}
