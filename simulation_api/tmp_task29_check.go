package main

import (
  "fmt"
  "config"
  "models"
)

func main() {
  config.InitDB()
  var task models.SimulationTask
  if err := config.DB.Preload("Config").First(&task, 29).Error; err != nil { panic(err) }
  workerID := uint(0)
  if task.WorkerID != nil { workerID = *task.WorkerID }
  fmt.Printf("task29: status=%s progress=%d worker=%d updated=%v lastError=%s\n", task.Status, task.Progress, workerID, task.UpdatedAt, task.LastError)
}
