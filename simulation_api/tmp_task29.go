package main

import (
  "fmt"
  "config"
  "models"
)

func main() {
  config.InitDB()
  var task models.SimulationTask
  if err := config.DB.Preload("Config").First(&task, 29).Error; err != nil {
    panic(err)
  }
  workerID := uint(0)
  if task.WorkerID != nil {
    workerID = *task.WorkerID
  }
  fmt.Printf("task29: status=%s progress=%d worker=%d project=%d config=%d configPath=%s created=%v started=%v updated=%v lastError=%s log=%s\n", task.Status, task.Progress, workerID, task.ProjectID, task.ConfigID, task.Config.ConfigPath, task.CreatedAt, task.StartedAt, task.UpdatedAt, task.LastError, task.LogURL)

  var suggestion models.SignalOptimizationSuggestion
  if err := config.DB.Where("rerun_task_id = ?", 29).First(&suggestion).Error; err == nil {
    fmt.Printf("suggestion: id=%d taskID=%d status=%s rerunTaskID=%v appliedConfig=%s appliedSignal=%s\n", suggestion.ID, suggestion.TaskID, suggestion.Status, suggestion.RerunTaskID, suggestion.AppliedConfigPath, suggestion.AppliedSignalFile)
  } else {
    fmt.Printf("suggestion lookup for rerun task 29 failed: %v\n", err)
  }
}
