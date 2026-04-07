package handlers

import (
	"SumoConfig"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"utils"

	"github.com/gin-gonic/gin"
)

func ConvertUploadedOSM(c *gin.Context) {
	osmFile, err := c.FormFile("osmFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "osmFile is required"})
		return
	}

	mapName := c.PostForm("mapName")
	if mapName == "" {
		mapName, _ = utils.GenerateRandomString(8)
	}

	workDir := filepath.Join("SimulationConfig", "public", mapName)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create workdir failed: " + err.Error()})
		return
	}

	osmPath := filepath.Join(workDir, "map.osm")
	netPath := filepath.Join(workDir, "map.net.xml")

	src, err := osmFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "open osm failed: " + err.Error()})
		return
	}
	defer src.Close()

	dst, err := os.Create(osmPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create osm file failed: " + err.Error()})
		return
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save osm failed: " + err.Error()})
		return
	}

	if err := dst.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "close osm file failed: " + err.Error()})
		return
	}

	cmd := exec.Command("netconvert", "--osm-files", osmPath, "--output-file", netPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "netconvert failed: " + err.Error(),
			"detail": string(output),
		})
		return
	}

	var geojson any
	if err := json.Unmarshal(SumoConfig.ToGeojson(netPath), &geojson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode geojson failed: " + err.Error()})
		return
	}

	netXML, err := os.ReadFile(netPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read net.xml failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mapName": mapName,
		"osmFile": osmPath,
		"netFile": netPath,
		"geojson": geojson,
		"netXml":  string(netXML),
	})
}
