package handlers

import (
	"config"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"models"
	"net/http"
	"strings"
	"time"
	"utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RegisterWorkerRequest struct {
	Name           string `json:"name" binding:"required"`
	Address        string `json:"address"`
	QueueName      string `json:"queueName"`
	LabelsJSON     string `json:"labelsJson"`
	MaxConcurrency int    `json:"maxConcurrency"`
}

type WorkerAuthRequest struct {
	WorkerID       *uint  `json:"workerId"`
	Name           string `json:"name"`
	Secret         string `json:"secret" binding:"required"`
	Address        string `json:"address"`
	QueueName      string `json:"queueName"`
	LabelsJSON     string `json:"labelsJson"`
	MaxConcurrency int    `json:"maxConcurrency"`
}

type WorkerHeartbeatRequest struct {
	Status       string `json:"status"`
	RunningTasks int    `json:"runningTasks"`
}

type CompleteTaskRequest struct {
	OutputArtifactID *uint  `json:"outputArtifactId"`
	LogURL           string `json:"logUrl"`
}

type FailTaskRequest struct {
	Error  string `json:"error" binding:"required"`
	LogURL string `json:"logUrl"`
}

type TaskProgressRequest struct {
	Progress int `json:"progress"`
}

type WorkerListItem struct {
	ID             uint       `json:"id"`
	Name           string     `json:"name"`
	Status         string     `json:"status"`
	Address        string     `json:"address"`
	QueueName      string     `json:"queueName"`
	MaxConcurrency int        `json:"maxConcurrency"`
	RunningTasks   int        `json:"runningTasks"`
	LastHeartbeat  *time.Time `json:"lastHeartbeat"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

const workerOfflineTimeout = 30 * time.Second

func RegisterWorker(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	var req RegisterWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if req.MaxConcurrency <= 0 {
		req.MaxConcurrency = 1
	}
	if req.QueueName == "" {
		req.QueueName = "default"
	}

	now := time.Now()
	var worker models.WorkerNode
	err := config.DB.Where("tenant_id = ? AND name = ?", user.TenantID, req.Name).First(&worker).Error
	workerSecret, secretHash, err := generateWorkerSecret()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generate worker secret failed: " + err.Error()})
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		worker = models.WorkerNode{
			TenantID:       user.TenantID,
			Name:           req.Name,
			SecretHash:     secretHash,
			Status:         "offline",
			Address:        req.Address,
			QueueName:      req.QueueName,
			LabelsJSON:     req.LabelsJSON,
			MaxConcurrency: req.MaxConcurrency,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := config.DB.Create(&worker).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create worker failed: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"workerId": worker.ID, "workerSecret": workerSecret})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query worker failed: " + err.Error()})
		return
	}

	if err := config.DB.Model(&worker).Updates(map[string]any{
		"secret_hash":     secretHash,
		"status":          "offline",
		"address":         req.Address,
		"queue_name":      req.QueueName,
		"labels_json":     req.LabelsJSON,
		"max_concurrency": req.MaxConcurrency,
		"last_heartbeat":  nil,
		"running_tasks":   0,
		"updated_at":      now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update worker failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workerId": worker.ID, "workerSecret": workerSecret})
}

func AuthenticateWorker(c *gin.Context) {
	var req WorkerAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if req.MaxConcurrency <= 0 {
		req.MaxConcurrency = 1
	}
	if strings.TrimSpace(req.QueueName) == "" {
		req.QueueName = "default"
	}
	if (req.WorkerID == nil || *req.WorkerID == 0) && strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workerId or name is required"})
		return
	}

	var worker models.WorkerNode
	query := config.DB
	if req.WorkerID != nil && *req.WorkerID != 0 {
		query = query.Where("id = ?", *req.WorkerID)
	} else {
		query = query.Where("name = ?", strings.TrimSpace(req.Name))
	}
	if err := query.First(&worker).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid worker credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query worker failed: " + err.Error()})
		return
	}

	if worker.SecretHash == "" || hashWorkerSecret(req.Secret) != worker.SecretHash {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid worker credentials"})
		return
	}

	now := time.Now()
	status := "online"
	if worker.RunningTasks > 0 {
		status = "busy"
	}
	if err := config.DB.Model(&worker).Updates(map[string]any{
		"status":          status,
		"address":         strings.TrimSpace(req.Address),
		"queue_name":      strings.TrimSpace(req.QueueName),
		"labels_json":     strings.TrimSpace(req.LabelsJSON),
		"max_concurrency": req.MaxConcurrency,
		"last_heartbeat":  now,
		"updated_at":      now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update worker failed: " + err.Error()})
		return
	}

	token, err := utils.GenerateWorkerToken(worker.Name, worker.ID, worker.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generate worker token failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workerId": worker.ID,
		"token":    "Bearer " + token,
	})
}

func ListWorkers(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	var workers []models.WorkerNode
	if err := config.DB.
		Where("tenant_id = ?", user.TenantID).
		Order("updated_at desc, id asc").
		Find(&workers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query workers failed: " + err.Error()})
		return
	}

	items := make([]WorkerListItem, 0, len(workers))
	now := time.Now()
	for _, worker := range workers {
		status := effectiveWorkerStatus(worker, now)
		items = append(items, WorkerListItem{
			ID:             worker.ID,
			Name:           worker.Name,
			Status:         status,
			Address:        worker.Address,
			QueueName:      worker.QueueName,
			MaxConcurrency: worker.MaxConcurrency,
			RunningTasks:   worker.RunningTasks,
			LastHeartbeat:  worker.LastHeartbeat,
			UpdatedAt:      worker.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"workers": items})
}

func WorkerHeartbeat(c *gin.Context) {
	worker, ok := findWorkerFromParam(c)
	if !ok {
		return
	}

	var req WorkerHeartbeatRequest
	_ = c.ShouldBindJSON(&req)
	if req.Status == "" {
		req.Status = "online"
	}

	now := time.Now()
	if err := config.DB.Model(&worker).Updates(map[string]any{
		"status":         req.Status,
		"running_tasks":  req.RunningTasks,
		"last_heartbeat": now,
		"updated_at":     now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "heartbeat failed: " + err.Error()})
		return
	}

	if req.RunningTasks == 0 {
		if err := reconcileOrphanedWorkerTasks(worker.ID, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "reconcile orphaned tasks failed: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func ClaimNextTask(c *gin.Context) {
	worker, ok := findWorkerFromParam(c)
	if !ok {
		return
	}

	var task models.SimulationTask
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		query := tx.Preload("Config").
			Where("status = ?", "queued").
			Order("priority desc, id asc")
		if worker.QueueName != "" {
			query = query.Where("queue_name = ? OR queue_name = ''", worker.QueueName)
		}

		if err := query.First(&task).Error; err != nil {
			return err
		}

		now := time.Now()
		result := tx.Model(&models.SimulationTask{}).
			Where("id = ? AND status = ?", task.ID, "queued").
			Updates(map[string]any{
				"status":       "running",
				"progress":     1,
				"worker_id":    worker.ID,
				"scheduled_at": now,
				"started_at":   now,
				"updated_at":   now,
				"last_error":   "",
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "claim task failed: " + err.Error()})
		return
	}

	if err := syncProjectStatus(task.ProjectID, time.Now()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update project status failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

func CompleteTask(c *gin.Context) {
	task, ok := findWorkerTask(c)
	if !ok {
		return
	}

	var req CompleteTaskRequest
	_ = c.ShouldBindJSON(&req)

	now := time.Now()
	if err := config.DB.Model(&models.SimulationTask{}).
		Where("id = ? AND worker_id = ?", task.ID, task.WorkerID).
		Updates(map[string]any{
			"status":             "succeeded",
			"progress":           100,
			"output_artifact_id": req.OutputArtifactID,
			"log_url":            req.LogURL,
			"ended_at":           now,
			"updated_at":         now,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "complete task failed: " + err.Error()})
		return
	}

	if err := syncProjectStatus(task.ProjectID, now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update project status failed: " + err.Error()})
		return
	}

	if err := generateSignalOptimizationSuggestion(task.ID); err != nil {
		fmt.Printf("generate signal optimization suggestion for task %d failed: %v\n", task.ID, err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func UpdateTaskProgress(c *gin.Context) {
	task, ok := findWorkerTask(c)
	if !ok {
		return
	}

	var req TaskProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	progress := req.Progress
	if progress < 0 {
		progress = 0
	}
	if progress > 99 {
		progress = 99
	}

	now := time.Now()
	if err := config.DB.Model(&models.SimulationTask{}).
		Where("id = ? AND worker_id = ?", task.ID, task.WorkerID).
		Updates(map[string]any{
			"progress":   progress,
			"updated_at": now,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update progress failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func FailTask(c *gin.Context) {
	task, ok := findWorkerTask(c)
	if !ok {
		return
	}

	var req FailTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	now := time.Now()
	if err := config.DB.Model(&models.SimulationTask{}).
		Where("id = ? AND worker_id = ?", task.ID, task.WorkerID).
		Updates(map[string]any{
			"status":     "failed",
			"last_error": req.Error,
			"log_url":    req.LogURL,
			"ended_at":   now,
			"updated_at": now,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fail task failed: " + err.Error()})
		return
	}

	if err := syncProjectStatus(task.ProjectID, now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update project status failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func findWorkerFromParam(c *gin.Context) (models.WorkerNode, bool) {
	var worker models.WorkerNode
	tenantID, _ := c.Get("tenantId")
	query := config.DB
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&worker, c.Param("workerId")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "worker not found"})
		return models.WorkerNode{}, false
	}
	return worker, true
}

func findWorkerTask(c *gin.Context) (models.SimulationTask, bool) {
	var task models.SimulationTask
	if err := config.DB.Preload("Project").
		Where("id = ? AND worker_id = ?", c.Param("taskId"), c.Param("workerId")).
		First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found for worker"})
		return models.SimulationTask{}, false
	}

	if tenantID, exists := c.Get("tenantId"); exists {
		if tenantIDValue, ok := tenantID.(uint); ok && task.Project.TenantID != tenantIDValue {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found for worker"})
			return models.SimulationTask{}, false
		}
	}

	return task, true
}

func effectiveWorkerStatus(worker models.WorkerNode, now time.Time) string {
	if worker.LastHeartbeat == nil {
		return "offline"
	}
	if now.Sub(*worker.LastHeartbeat) > workerOfflineTimeout {
		return "offline"
	}
	if worker.Status == "" {
		return "offline"
	}
	return worker.Status
}

func syncProjectStatus(projectID uint, now time.Time) error {
	projectStatus, err := deriveProjectStatus(projectID)
	if err != nil {
		return err
	}

	return config.DB.Model(&models.SimulationProject{}).
		Where("id = ?", projectID).
		Updates(map[string]any{
			"status":    projectStatus,
			"update_at": now,
		}).Error
}

func deriveProjectStatus(projectID uint) (string, error) {
	var tasks []models.SimulationTask
	if err := config.DB.
		Where("project_id = ?", projectID).
		Order("created_at desc").
		Find(&tasks).Error; err != nil {
		return "", err
	}

	projectStatus := "ready"
	hasSucceeded := false
	hasFailed := false
	hasQueued := false
	hasRunning := false

	for _, task := range tasks {
		switch task.Status {
		case "running":
			hasRunning = true
		case "queued", "dispatching":
			hasQueued = true
		case "failed":
			hasFailed = true
		case "succeeded":
			hasSucceeded = true
		}
	}

	switch {
	case hasRunning:
		projectStatus = "running"
	case hasQueued:
		projectStatus = "queued"
	case hasFailed:
		projectStatus = "failed"
	case hasSucceeded:
		projectStatus = "succeeded"
	}

	return projectStatus, nil
}

func reconcileOrphanedWorkerTasks(workerID uint, now time.Time) error {
	staleBefore := now.Add(-45 * time.Second)

	var tasks []models.SimulationTask
	if err := config.DB.
		Where("worker_id = ? AND status = ? AND updated_at < ?", workerID, "running", staleBefore).
		Find(&tasks).Error; err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}

	projectIDs := make(map[uint]struct{}, len(tasks))
	taskIDs := make([]uint, 0, len(tasks))
	for _, task := range tasks {
		projectIDs[task.ProjectID] = struct{}{}
		taskIDs = append(taskIDs, task.ID)
	}

	if err := config.DB.Model(&models.SimulationTask{}).
		Where("id IN ?", taskIDs).
		Updates(map[string]any{
			"status":       "queued",
			"progress":     0,
			"worker_id":    nil,
			"scheduled_at": nil,
			"started_at":   nil,
			"updated_at":   now,
			"last_error":   "Task was re-queued automatically after worker heartbeat reported no running tasks.",
		}).Error; err != nil {
		return err
	}

	for projectID := range projectIDs {
		if err := syncProjectStatus(projectID, now); err != nil {
			return err
		}
	}
	return nil
}

func generateWorkerSecret() (string, string, error) {
	secretBytes := make([]byte, 24)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", err
	}
	secret := hex.EncodeToString(secretBytes)
	return secret, hashWorkerSecret(secret), nil
}

func hashWorkerSecret(secret string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(secret)))
	return hex.EncodeToString(sum[:])
}
