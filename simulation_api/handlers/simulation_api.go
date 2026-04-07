package handlers

import (
	"SumoConfig"
	"config"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"models"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GenerateProjectRequest struct {
	Name           string  `form:"name" json:"name" binding:"required"`
	SimulationTime int32   `form:"simulationTime" json:"simulationTime" binding:"required"`
	IsNowRun       bool    `form:"isNowRun" json:"isNowRun"`
	Speed          float64 `form:"speed" json:"speed"`
}

type CreateTaskRequest struct {
	SimulationTime int32   `json:"simulationTime"`
	Speed          float64 `json:"speed"`
}

type ProjectListItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TaskListItem struct {
	ID              uint       `json:"id"`
	Status          string     `json:"status"`
	DurationSeconds int32      `json:"durationSeconds"`
	Speed           float64    `json:"speed"`
	MonitorPort     int        `json:"monitorPort"`
	TraCIPort       int        `json:"traCIPort"`
	LastError       string     `json:"lastError"`
	CreatedAt       time.Time  `json:"createdAt"`
	StartedAt       *time.Time `json:"startedAt"`
	EndedAt         *time.Time `json:"endedAt"`
}

func GenerateProject(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	var req GenerateProjectRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	if req.Speed <= 0 {
		req.Speed = 1
	}

	projectDir, configPath, err := saveSimulationAssets(c, user, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	mapData := models.MapData{
		FileURL:    configPath,
		Name:       req.Name + "_map",
		TenantID:   user.TenantID,
		UploadedAt: now,
	}
	if err := config.DB.Create(&mapData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建地图数据失败: " + err.Error()})
		return
	}

	project := models.SimulationProject{
		TenantID:    user.TenantID,
		Name:        req.Name,
		Status:      "pending",
		Description: fmt.Sprintf("仿真项目 %s", req.Name),
		MapDataID:   mapData.ID,
		MapData:     mapData,
		UpdateAt:    now,
		CreatedAt:   now,
	}
	if err := config.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建项目失败: " + err.Error()})
		return
	}

	simConfig := models.SimulationConfig{
		ProjectID:  project.ID,
		Name:       req.Name + "_config",
		ConfigPath: configPath,
		CreatedAt:  now,
	}
	if err := config.DB.Create(&simConfig).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建仿真配置失败: " + err.Error()})
		return
	}

	var task *models.SimulationTask
	if req.IsNowRun {
		createdTask := models.SimulationTask{
			ProjectID:       project.ID,
			ConfigID:        simConfig.ID,
			Status:          "pending",
			DurationSeconds: req.SimulationTime,
			Speed:           req.Speed,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := config.DB.Create(&createdTask).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建仿真任务失败: " + err.Error()})
			return
		}
		task = &createdTask
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "generate project successfully",
		"projectId":  project.ID,
		"projectDir": projectDir,
		"configPath": configPath,
		"taskId":     taskID(task),
	})
}

func ListProjects(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	var projects []models.SimulationProject
	if err := config.DB.Where("tenant_id = ?", user.TenantID).Order("created_at desc").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询项目失败: " + err.Error()})
		return
	}

	items := make([]ProjectListItem, 0, len(projects))
	for _, project := range projects {
		items = append(items, ProjectListItem{
			ID:        project.ID,
			Name:      project.Name,
			Status:    project.Status,
			CreatedAt: project.CreatedAt,
			UpdatedAt: project.UpdateAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"projects": items})
}

func EnqueueProjectTask(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.SimulationTime <= 0 {
		req.SimulationTime = 60
	}
	if req.Speed <= 0 {
		req.Speed = 1
	}

	project, err := findProjectForUser(user.TenantID, uint(projectID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询项目失败: " + err.Error()})
		return
	}

	simConfig, err := ensureSimulationConfig(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建仿真配置失败: " + err.Error()})
		return
	}

	now := time.Now()
	task := models.SimulationTask{
		ProjectID:       project.ID,
		ConfigID:        simConfig.ID,
		Status:          "pending",
		DurationSeconds: req.SimulationTime,
		Speed:           req.Speed,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := config.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建仿真任务失败: " + err.Error()})
		return
	}

	if err := config.DB.Model(&models.SimulationProject{}).
		Where("id = ?", project.ID).
		Updates(map[string]any{"status": "pending", "update_at": now}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新项目状态失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "simulation task queued", "taskId": task.ID})
}

func ListProjectTasks(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	project, err := findProjectForUser(user.TenantID, uint(projectID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询项目失败: " + err.Error()})
		return
	}

	var tasks []models.SimulationTask
	if err := config.DB.Where("project_id = ?", project.ID).Order("created_at desc").Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务失败: " + err.Error()})
		return
	}

	items := make([]TaskListItem, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, TaskListItem{
			ID:              task.ID,
			Status:          task.Status,
			DurationSeconds: task.DurationSeconds,
			Speed:           task.Speed,
			MonitorPort:     task.MonitorPort,
			TraCIPort:       task.TraCIPort,
			LastError:       task.LastError,
			CreatedAt:       task.CreatedAt,
			StartedAt:       task.StartedAt,
			EndedAt:         task.EndedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tasks": items})
}

func currentUser(c *gin.Context) (models.User, bool) {
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return models.User{}, false
	}

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return models.User{}, false
	}

	return user, true
}

func saveSimulationAssets(c *gin.Context, user models.User, projectName string) (string, string, error) {
	var sumoConfig SumoConfig.SumoConfiguration
	form, err := c.MultipartForm()
	if err != nil {
		return "", "", fmt.Errorf("获取表单失败: %w", err)
	}

	netFileHeaders := form.File["netFile"]
	routeFileHeaders := form.File["routeFiles"]
	additionalFileHeaders := form.File["additionalFiles"]

	workDir := "SimulationConfig"
	userDir := fmt.Sprintf("%s/%s", workDir, user.Username)
	projectDir := fmt.Sprintf("%s/%s", userDir, projectName)

	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return "", "", fmt.Errorf("创建项目目录失败: %w", err)
	}

	sumoConfig.Input = &SumoConfig.Input{}

	if len(netFileHeaders) > 0 {
		netFile := netFileHeaders[0]
		netFilePath := fmt.Sprintf("%s/%s", projectDir, netFile.Filename)
		if err := c.SaveUploadedFile(netFile, netFilePath); err != nil {
			return "", "", fmt.Errorf("保存路网文件失败: %w", err)
		}
		sumoConfig.Input.NetFile = SumoConfig.StringAttr{Value: netFile.Filename}
	}

	if len(routeFileHeaders) > 0 {
		routeFiles := ""
		for i, file := range routeFileHeaders {
			savePath := fmt.Sprintf("%s/%s", projectDir, file.Filename)
			if err := c.SaveUploadedFile(file, savePath); err != nil {
				return "", "", fmt.Errorf("保存路径文件 %s 失败: %w", file.Filename, err)
			}
			if i > 0 {
				routeFiles += ","
			}
			routeFiles += file.Filename
		}
		sumoConfig.Input.RouteFiles = SumoConfig.StringAttr{Value: routeFiles}
	}

	if len(additionalFileHeaders) > 0 {
		additionalFiles := ""
		for i, file := range additionalFileHeaders {
			savePath := fmt.Sprintf("%s/%s", projectDir, file.Filename)
			if err := c.SaveUploadedFile(file, savePath); err != nil {
				return "", "", fmt.Errorf("保存附加文件 %s 失败: %w", file.Filename, err)
			}
			if i > 0 {
				additionalFiles += ","
			}
			additionalFiles += file.Filename
		}
		sumoConfig.Input.AdditionalFiles = &SumoConfig.StringAttr{Value: additionalFiles}
	}

	output, err := xml.MarshalIndent(sumoConfig, "", "    ")
	if err != nil {
		return "", "", fmt.Errorf("生成 SUMO 配置失败: %w", err)
	}

	configPath := fmt.Sprintf("%s/%s", projectDir, "sumo.sumocfg")
	if err := os.WriteFile(configPath, output, 0o644); err != nil {
		return "", "", fmt.Errorf("保存 SUMO 配置文件失败: %w", err)
	}

	return projectDir, configPath, nil
}

func findProjectForUser(tenantID uint, projectID uint) (models.SimulationProject, error) {
	var project models.SimulationProject
	err := config.DB.Preload("MapData").
		Where("id = ? AND tenant_id = ?", projectID, tenantID).
		First(&project).Error
	return project, err
}

func ensureSimulationConfig(project models.SimulationProject) (models.SimulationConfig, error) {
	var simConfig models.SimulationConfig
	err := config.DB.Where("project_id = ?", project.ID).Order("id desc").First(&simConfig).Error
	if err == nil {
		return simConfig, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.SimulationConfig{}, err
	}

	simConfig = models.SimulationConfig{
		ProjectID:  project.ID,
		Name:       project.Name + "_config",
		ConfigPath: project.MapData.FileURL,
		CreatedAt:  time.Now(),
	}
	return simConfig, config.DB.Create(&simConfig).Error
}

func taskID(task *models.SimulationTask) any {
	if task == nil {
		return nil
	}
	return task.ID
}
