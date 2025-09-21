package handlers

import (
	"SumoConfig"
	"config"
	"fmt"
	"models"
	"net/http"
	"os"
	"path/filepath"
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

// getGeojson 获取路网文件的Geojson数据
// @Summary 获取路网文件的Geojson数据
// @Description 获取路网文件的Geojson数据
// @Tags Geojson
// @Produce application/octet-stream
// @Param netfile query string true "仿真器索引" default(http://localhost:8081/SimulationConfig/j_lI-)
// @Success 200 {string} string "geojson转换成功"
// @Failure 400 {string} string "请求错误"
// @Failure 500 {string} string "服务器内部错误"
// @Router /getGeojson [get]
// @Security BearerAuth
func GetGeojson(c *gin.Context) {
	// 目前没有实现逻辑，保留原样
	// 获取文件字段``
	netfile := c.DefaultQuery("netfile", "")
	if netfile == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "netfile unavailable"})
		return
	}
	GeoJson := SumoConfig.ToGeojson(netfile)
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(GeoJson)
}

// DownloadSimulationConfig godoc
// @Summary      下载仿真配置文件
// @Description  根据文件名从 files 目录中下载指定仿真配置文件
// @Tags         SimulationConfig
// @Accept       json
// @Produce      octet-stream
// @Param        filename  path  string  true  "文件名"
// @Success      200  {file}  file
// @Failure      500  {string}  string  "内部错误"
// @Router       /SimulationConfig/{filename} [get]
func Download(c *gin.Context) {
	filename := c.Param("filename")
	cleanFilename := filepath.Base(filename) // 防止路径遍历攻击

	// 获取当前程序所在目录
	exePath, err := os.Executable()
	if err != nil {
		c.String(500, "内部错误")
		return
	}
	exeDir := filepath.Dir(exePath)

	// 拼接 full path
	fullPath := filepath.Join(exeDir, "SimulationConfig", cleanFilename)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.String(404, "文件不存在")
		return
	}

	// 发送文件作为下载附件
	c.FileAttachment(fullPath, cleanFilename)
}
