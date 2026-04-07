package handlers

import (
	"config"
	"models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetSimulationOverview(c *gin.Context) {
	username := c.GetString("username")
	var projectCount int64
	var simulationCount int64
	var mapCount int64
	var recentProjects []models.SimulationProject
	var recentProjectsToWebs []models.SimulationProjectToWeb

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	config.DB.Model(&models.SimulationProject{}).Where("tenant_id = ?", user.ID).Count(&projectCount)
	config.DB.Model(&models.Simulation{}).Where("tenant_id = ?", user.ID).Count(&simulationCount)
	config.DB.Model(&models.MapData{}).Where("tenant_id = ?", user.ID).Count(&mapCount)

	config.DB.Where("tenant_id = ?", user.ID).
		Order("update_at desc").
		Limit(5).
		Find(&recentProjects)

	for _, project := range recentProjects {
		recentProjectsToWebs = append(recentProjectsToWebs, models.SimulationProjectToWeb{
			ID:       project.ID,
			Name:     project.Name,
			Status:   project.Status,
			UpdateAt: project.UpdateAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"project_count":    projectCount,
		"simulation_count": simulationCount,
		"map_count":        mapCount,
		"recent_projects":  recentProjectsToWebs,
	})
}
