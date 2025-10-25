package main

import (
	"fmt"
	"os"

	"github.com/snowpea/stats/internal/api"
	"github.com/snowpea/stats/internal/config"
	"github.com/snowpea/stats/internal/logger"
	"github.com/snowpea/stats/internal/utils"
	"github.com/snowpea/stats/pkg/html"
	"github.com/snowpea/stats/pkg/size"
)

func Logic(cfg *config.Config) {
	logger.SetLogLevel(cfg.LogLevel)
	logger.OutLog("Querying SabNZBD for queue size and remaining MB", nil)

	downloadSizeBytes, err := api.GetSABQueueSize(cfg)
	if err != nil {
		logger.OutLog(fmt.Sprintf("Error querying SabNZBD: %v", err), nil)
		os.Exit(1)
	}

	logger.OutLog("Finalising processing for download queue size", nil)
	downloadSizeStr, err := utils.CalcSize(downloadSizeBytes)
	if err != nil {
		logger.OutLog(fmt.Sprintf("Unable to calculate download size: %v", err), nil)
		os.Exit(1)
	}

	logger.OutLog("Getting media directory stats", nil)
	totalMediaSize, err := size.GetMediaSize(cfg)
	if err != nil {
		logger.OutLog(fmt.Sprintf("Error getting media size: %v", err), nil)
		os.Exit(1)
	}

	mediaSizeStr, err := utils.CalcSize(totalMediaSize)
	if err != nil {
		logger.OutLog(fmt.Sprintf("Error: Unable to calculate size of media size dir: %v", err), nil)
		os.Exit(1)
	}

	if cfg.LogLevel == "Debug" {
		logger.OutLog(fmt.Sprintf("MediaSize: %s", mediaSizeStr), nil)
		logger.OutLog(fmt.Sprintf("DownloadSize: %s", downloadSizeStr), nil)
	}

	maintenance, err := utils.GetMaintenanceNotice(cfg)

	logger.OutLog("Creating HTML", nil)
	htmlContent := html.GenerateHTML(mediaSizeStr, downloadSizeStr, maintenance, cfg.Uptime)

	if cfg.LogLevel == "debug" {
		logger.OutLog(htmlContent, nil)
	}

	if err := os.WriteFile(cfg.StatsFile, []byte(htmlContent), 0644); err != nil {
		logger.OutLog(fmt.Sprintf("Error writing HTML file: %v", err), nil)
		os.Exit(1)
	}

	logger.OutLog("HTML file created successfully", nil)
}
