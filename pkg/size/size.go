package size

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/snowpea/stats/internal/config"
	"github.com/snowpea/stats/internal/logger"
)

func GetMediaSize(cfg *config.Config) (int64, error) {
	var totalSize int64

	for _, mediaDir := range cfg.MediaDirs {
		cmd := exec.Command("du", "-sb", mediaDir)
		output, err := cmd.Output()
		if err != nil {
			logger.OutLog(fmt.Sprintf("Error getting size for %s: %v", mediaDir, err), nil)
			continue
		}

		parts := strings.Split(string(output), "\t")
		if len(parts) < 1 {
			continue
		}

		size, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			logger.OutLog(fmt.Sprintf("Error parsing size for %s: %v", mediaDir, err), nil)
			continue
		}

		totalSize += size
	}

	return totalSize, nil
}
