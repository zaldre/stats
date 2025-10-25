package api

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/snowpea/stats/internal/config"
	"github.com/snowpea/stats/internal/models"
)

func GetSABQueueSize(cfg *config.Config) (int64, error) {
	var url string
	// Check if sabHost already includes a port
	if strings.Contains(cfg.SabHost, ":") {
		url = fmt.Sprintf("%s/sabnzbd/api?mode=queue&output=json&apikey=%s",
			cfg.SabHost, cfg.SabAPIKey)
	} else {
		url = fmt.Sprintf("%s:%d/sabnzbd/api?mode=queue&output=json&apikey=%s",
			cfg.SabHost, cfg.SabPort, cfg.SabAPIKey)
	}
	if cfg.LogLevel == "Debug" {
		fmt.Println(url)
	}

	client := &http.Client{Timeout: time.Duration(cfg.WebTimeout) * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var sabQueue models.SABQueue
	if err := json.NewDecoder(resp.Body).Decode(&sabQueue); err != nil {
		return 0, err
	}

	// Convert mbleft to float, then to bytes
	mbLeft, err := strconv.ParseFloat(sabQueue.Queue.MBLeft, 64)
	if err != nil {
		return 0, err
	}

	mbLeft = math.Round(mbLeft*100) / 100 // Round to 2 decimal places
	sizeBytes := int64(mbLeft * 1024 * 1024)

	return sizeBytes, nil
}
