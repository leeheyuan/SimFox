package main

import (
  "fmt"
  "config"
  "models"
)

func main() {
  config.InitDB()
  var tasks []models.SimulationTask
  config.DB.Order("id desc").Limit(8).Find(&tasks)
  fmt.Println("tasks:")
  for _, t := range tasks {
    fmt.Printf("id=%d status=%s progress=%d worker=%v log=%s lastError=%s\n", t.ID, t.Status, t.Progress, t.WorkerID, t.LogURL, t.LastError)
  }
  var workers []models.WorkerNode
  config.DB.Order("id asc").Find(&workers)
  fmt.Println("workers:")
  for _, w := range workers {
    fmt.Printf("id=%d status=%s running=%d heartbeat=%v address=%s\n", w.ID, w.Status, w.RunningTasks, w.LastHeartbeat, w.Address)
  }
}
