package utils

import (
	"fmt"
	"math"
)

func CalcSize(byteCount int64) (string, error) {
	if byteCount < 0 {
		return "", fmt.Errorf("byte count cannot be negative: %d", byteCount)
	}

	if byteCount == 0 {
		return "0 Bytes", nil
	}

	sizeUnits := []string{"Bytes", "KB", "MB", "GB", "TB"}
	unitIndex := int(math.Floor(math.Log(float64(byteCount)) / math.Log(1024)))

	// Handle edge case where unitIndex exceeds available units
	if unitIndex >= len(sizeUnits) {
		// Either cap it or return an error
		unitIndex = len(sizeUnits) - 1
		// Or: return "", fmt.Errorf("byte count too large: %d", byteCount)
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%d Bytes", byteCount), nil
	}

	size := float64(byteCount) / math.Pow(1024, float64(unitIndex))
	return fmt.Sprintf("%.2f %s", size, sizeUnits[unitIndex]), nil
}
