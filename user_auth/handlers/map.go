package handlers

import (
	"config"
	"fmt"
	"models"
	"net/http"
	"os"
	"time"
	"utils"

	"github.com/gin-gonic/gin"
)

// uploadmap 上传地图项目
// @Summary 生成项目配置文件
// @Description 生成项目配置文件
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce plain
// @Param netFile formData file false "路网文件"
// @Success 200 {string} string "路网文件保存成功"
// @Failure 400 {string} string "请求错误"
// @Failure 500 {string} string "服务器内部错误"
// @Router /uploadmap [post]
// @Security BearerAuth
func UpLoadMap(c *gin.Context) {
	// 目前没有实现逻辑，保留原样
	// 获取 POST 表单字段
	// 获取文件字段

	username := c.GetString("username")

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	netFile, err := c.FormFile("netFile")
	if err != nil {
		c.String(http.StatusBadRequest, "netFile 获取失败: %v", err)
		return
	}
	mapName, _ := utils.GenerateRandomString(5)
	workDir := "SimulationConfig"
	if !utils.DirExists(workDir) {
		err := os.Mkdir(workDir, 0755)
		if err != nil {
			c.String(http.StatusBadRequest, "creator dir fixed ")
			return
		}
	}
	netFilePath := fmt.Sprintf("%s/%s", workDir, mapName)

	c.SaveUploadedFile(netFile, netFilePath)

	// 创建管理员用户，关联租户ID
	mapdata := models.MapData{
		TenantID:   user.ID,
		Name:       mapName,
		FileURL:    "",
		UploadedAt: time.Now(),
	}
	if err := config.DB.Create(&mapdata).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "路网文件保存失败: " + err.Error()})
		return
	}
	c.String(http.StatusOK, mapName)

}
