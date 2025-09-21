package handlers

import (
	"SumoConfig"
	"config"
	"encoding/xml"
	"fmt"
	"models"
	"net/http"
	"os"
	"port"
	"time"

	"github.com/gin-gonic/gin"
)

var portManager *port.PortManager = port.NewPortManager(8000, 9000)

// GenerateRequest 表单字段
type GenerateRequest struct {
	Name           string `form:"name" json:"name" binding:"required"`
	SimulationTime int32  `form:"simulationTime" json:"simulationTime" binding:"required"`
	IsNowRun       bool   `form:"isNowRun" json:"isNowRun" binding:"required"`
}

// @Summary      项目生成
// @Description  生成一个仿真项目，包含参数和上传文件
// @Tags         项目
// @Accept       multipart/form-data
// @Produce      json
// @Param        name formData string true "项目名称"
// @Param        simulationTime formData int true "仿真时长"
// @Param        isNowRun formData bool true "是否立即运行"
// @Param        netFile formData file false "路网文件"
// @Param        routeFiles formData file false "路线文件（可多个）" collectionFormat(multi)
// @Param        additionalFiles formData file false "额外的配置文件（可多个）" collectionFormat(multi)
// @Success      200 {object} map[string]string "生成成功提示"
// @Failure      500 {object} map[string]string "服务器内部错误"
// @Router       /project/generate [post]
// @Security     BearerAuth
func Generate(c *gin.Context) {
	fmt.Println("Generate")

	username := c.GetString("username")

	var user models.User

	var sumoConfig SumoConfig.SumoConfiguration
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	var req GenerateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取表单失败: " + err.Error()})
		return
	}

	netFileHeaders := form.File["netFile"]
	routeFileHeaders := form.File["routeFiles"]
	additionalFileHeaders := form.File["additionalFiles"]

	workDir := "SimulationConfig"
	userDir := fmt.Sprintf("%s/%s", workDir, user.Username)
	projectDir := fmt.Sprintf("%s/%s", userDir, req.Name)

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败"})
		return
	}

	sumoConfig.Input = &SumoConfig.Input{}

	// 保存 netFile（如果存在）
	if len(netFileHeaders) > 0 {
		netFile := netFileHeaders[0]
		netFilePath := fmt.Sprintf("%s/%s", projectDir, netFile.Filename)
		if err := c.SaveUploadedFile(netFile, netFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存 netFile 失败"})
			return
		}

		sumoConfig.Input.NetFile = SumoConfig.StringAttr{
			Value: netFile.Filename,
		}
	}

	if len(routeFileHeaders) > 0 {
		var routeFiles string
		for i, file := range routeFileHeaders {
			savePath := fmt.Sprintf("%s/%s", projectDir, file.Filename)
			if err := c.SaveUploadedFile(file, savePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存 routeFile %s 失败", file.Filename)})
				return
			}
			if i > 0 {
				routeFiles += ","
			}
			routeFiles += file.Filename

		}
		sumoConfig.Input.RouteFiles = SumoConfig.StringAttr{
			Value: routeFiles,
		}
	}

	if len(additionalFileHeaders) > 0 {
		var additionalFiles string
		for i, file := range additionalFileHeaders {
			savePath := fmt.Sprintf("%s/%s", projectDir, file.Filename)
			if err := c.SaveUploadedFile(file, savePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存 routeFile %s 失败", file.Filename)})
				return
			}
			if i > 0 {
				additionalFiles += ","
			}
			additionalFiles += file.Filename
		}
		sumoConfig.Input.AdditionalFiles = &SumoConfig.StringAttr{
			Value: additionalFiles,
		}
	}
	output, err := xml.MarshalIndent(sumoConfig, "", "    ")
	if err != nil {
		panic(err)
	}
	sumosavePath := fmt.Sprintf("%s/%s", projectDir, "sumo.sumocfg")
	if err := os.WriteFile(sumosavePath, output, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存 sumo 配置文件失败"})
		return
	}

	Status := "pending"
	if req.IsNowRun {
		Status = "running"
	}

	simulationProject := models.SimulationProject{
		TenantID:    user.TenantID,
		Name:        req.Name,
		Status:      Status,
		Description: fmt.Sprintf("仿真项目 %s 的描述", req.Name),
		MapDataID:   0,                // 假设没有 MapDataID，实际应用中可能需要根据实际情况设置
		MapData:     models.MapData{}, // 假设没有 MapData，实际应用中可能需要根据实际情况设置
		Tenant:      user.Tenant,
		CreatedAt:   time.Now(),
	}
	if err := config.DB.Create(&simulationProject).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建项目失败: " + err.Error()})
		return
	}

	if req.IsNowRun {
		portManager.StartProcess("py", 2, "../../sim_server/sim_server.py", fmt.Sprintf("--config=%s", sumosavePath), fmt.Sprintf("--projectId=%d", simulationProject.ID))

	}

	c.JSON(http.StatusOK, gin.H{"message": "generate project successfully"})
}
