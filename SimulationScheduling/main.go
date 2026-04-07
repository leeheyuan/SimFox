package main

import (
	"config"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"models"
	"net/http"
	"os"
	"os/exec"
	"port"
	"strconv"
	"sync"
	"time"

	"gorm.io/gorm"
)

type runtimeTask struct {
	TaskID      uint      `json:"taskId"`
	ProjectID   uint      `json:"projectId"`
	MonitorPort int       `json:"monitorPort"`
	TraCIPort   int       `json:"traciPort"`
	StartedAt   time.Time `json:"startedAt"`
}

type scheduler struct {
	portManager  *port.PortManager
	pollInterval time.Duration
	maxWorkers   int
	pythonCmd    string
	scriptPath   string
	httpAddr     string

	mu      sync.Mutex
	running map[uint]runtimeTask
}

func main() {
	config.InitDB()
	if err := autoMigrate(); err != nil {
		log.Fatal(err)
	}

	s := &scheduler{
		portManager:  port.NewPortManager(8000, 9000),
		pollInterval: durationFromEnv("SCHEDULER_POLL_INTERVAL", 3*time.Second),
		maxWorkers:   intFromEnv("SCHEDULER_MAX_WORKERS", 2),
		pythonCmd:    stringFromEnv("SIM_PYTHON_CMD", "py"),
		scriptPath:   stringFromEnv("SIM_SERVER_ENTRY", "../sim_server/sim_server.py"),
		httpAddr:     stringFromEnv("SCHEDULER_HTTP_ADDR", ":8090"),
		running:      make(map[uint]runtimeTask),
	}

	go s.startHTTPServer()

	log.Printf("simulation scheduler started, poll=%s maxWorkers=%d monitor=%s", s.pollInterval, s.maxWorkers, s.httpAddr)
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		s.dispatchAvailableTasks()
		<-ticker.C
	}
}

func autoMigrate() error {
	return config.DB.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.SimulationProject{},
		&models.MapData{},
		&models.SimulationConfig{},
		&models.SimulationTask{},
		&models.SimulationResult{},
		&models.Simulation{},
	)
}

func (s *scheduler) dispatchAvailableTasks() {
	for s.runningCount() < s.maxWorkers {
		dispatched, err := s.dispatchOneTask()
		if err != nil {
			log.Printf("dispatch task failed: %v", err)
			return
		}
		if !dispatched {
			return
		}
	}
}

func (s *scheduler) dispatchOneTask() (bool, error) {
	var task models.SimulationTask
	err := config.DB.Preload("Config").Order("id asc").Where("status = ?", "pending").First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	now := time.Now()
	result := config.DB.Model(&models.SimulationTask{}).
		Where("id = ? AND status = ?", task.ID, "pending").
		Updates(map[string]any{
			"status":     "running",
			"started_at": now,
			"updated_at": now,
			"last_error": "",
		})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return true, nil
	}

	if err := config.DB.Model(&models.SimulationProject{}).
		Where("id = ?", task.ProjectID).
		Updates(map[string]any{"status": "running", "update_at": now}).Error; err != nil {
		return false, err
	}

	go s.executeTask(task.ID)
	return true, nil
}

func (s *scheduler) executeTask(taskID uint) {
	var task models.SimulationTask
	if err := config.DB.Preload("Config").First(&task, taskID).Error; err != nil {
		log.Printf("load task %d failed: %v", taskID, err)
		return
	}

	ports, err := s.portManager.AllocatePorts(2)
	if err != nil {
		s.failTask(task, fmt.Errorf("allocate ports: %w", err))
		return
	}
	defer s.portManager.ReleasePorts(ports)

	now := time.Now()
	task.MonitorPort = ports[0]
	task.TraCIPort = ports[1]
	task.StartedAt = &now
	task.UpdatedAt = now
	if err := config.DB.Model(&models.SimulationTask{}).
		Where("id = ?", task.ID).
		Updates(map[string]any{
			"monitor_port": task.MonitorPort,
			"tra_ci_port":  task.TraCIPort,
			"started_at":   now,
			"updated_at":   now,
		}).Error; err != nil {
		s.failTask(task, fmt.Errorf("update task ports: %w", err))
		return
	}

	runtime := runtimeTask{
		TaskID:      task.ID,
		ProjectID:   task.ProjectID,
		MonitorPort: task.MonitorPort,
		TraCIPort:   task.TraCIPort,
		StartedAt:   now,
	}
	s.setRunning(runtime)
	defer s.clearRunning(task.ID)

	args := []string{
		s.scriptPath,
		fmt.Sprintf("--config=%s", task.Config.ConfigPath),
		fmt.Sprintf("--duration=%d", task.DurationSeconds),
		fmt.Sprintf("--speed=%g", task.Speed),
		fmt.Sprintf("--projectId=%d", task.ProjectID),
		fmt.Sprintf("--ports=%d,%d", task.MonitorPort, task.TraCIPort),
	}

	cmd := exec.Command(s.pythonCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("starting task=%d project=%d monitorPort=%d traciPort=%d", task.ID, task.ProjectID, task.MonitorPort, task.TraCIPort)
	if err := cmd.Run(); err != nil {
		s.failTask(task, err)
		return
	}

	finishedAt := time.Now()
	if err := config.DB.Model(&models.SimulationTask{}).
		Where("id = ?", task.ID).
		Updates(map[string]any{
			"status":     "finished",
			"ended_at":   finishedAt,
			"updated_at": finishedAt,
		}).Error; err != nil {
		log.Printf("update finished task %d failed: %v", task.ID, err)
	}

	if err := config.DB.Model(&models.SimulationProject{}).
		Where("id = ?", task.ProjectID).
		Updates(map[string]any{
			"status":    "finished",
			"update_at": finishedAt,
		}).Error; err != nil {
		log.Printf("update finished project %d failed: %v", task.ProjectID, err)
	}
}

func (s *scheduler) failTask(task models.SimulationTask, runErr error) {
	finishedAt := time.Now()
	message := runErr.Error()

	if err := config.DB.Model(&models.SimulationTask{}).
		Where("id = ?", task.ID).
		Updates(map[string]any{
			"status":     "error",
			"last_error": message,
			"ended_at":   finishedAt,
			"updated_at": finishedAt,
		}).Error; err != nil {
		log.Printf("update failed task %d failed: %v", task.ID, err)
	}

	if err := config.DB.Model(&models.SimulationProject{}).
		Where("id = ?", task.ProjectID).
		Updates(map[string]any{
			"status":    "error",
			"update_at": finishedAt,
		}).Error; err != nil {
		log.Printf("update failed project %d failed: %v", task.ProjectID, err)
	}

	log.Printf("task=%d failed: %v", task.ID, runErr)
}

func (s *scheduler) startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/tasks/summary", s.handleSummary)
	mux.HandleFunc("/tasks/running", s.handleRunningTasks)

	if err := http.ListenAndServe(s.httpAddr, mux); err != nil {
		log.Printf("scheduler monitor server stopped: %v", err)
	}
}

func (s *scheduler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"service":   "simulation-scheduler",
		"timestamp": time.Now(),
	})
}

func (s *scheduler) handleSummary(w http.ResponseWriter, _ *http.Request) {
	type summary struct {
		Status string
		Count  int64
	}

	var rows []summary
	if err := config.DB.Model(&models.SimulationTask{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&rows).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	statusCounts := make(map[string]int64, len(rows))
	for _, row := range rows {
		statusCounts[row.Status] = row.Count
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tasks":        statusCounts,
		"runningCount": s.runningCount(),
		"maxWorkers":   s.maxWorkers,
	})
}

func (s *scheduler) handleRunningTasks(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	items := make([]runtimeTask, 0, len(s.running))
	for _, item := range s.running {
		items = append(items, item)
	}
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{"running": items})
}

func (s *scheduler) runningCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.running)
}

func (s *scheduler) setRunning(item runtimeTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running[item.TaskID] = item
}

func (s *scheduler) clearRunning(taskID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.running, taskID)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func intFromEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func stringFromEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
