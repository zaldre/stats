package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowpea/stats/internal/config"
	"github.com/snowpea/stats/internal/logger"
)

func GetMaintenanceNotice(cfg *config.Config) (string, error) {
	scriptDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}

	maintenanceFile := filepath.Join(scriptDir, "maintenance.txt")
	content, err := os.ReadFile(maintenanceFile)
	if err != nil {
		return "", err
	}

	logger.OutLog("Maintenance notice found, Appending this to HTML output", nil)
	notice := strings.TrimSpace(string(content))
	if cfg.LogLevel == "debug" {
		logger.OutLog(notice, nil)
	}
	fmt.Println(notice)
	return notice, nil
}
