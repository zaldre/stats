package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

// getEnvInt retrieves an integer environment variable or returns a default value
func getEnvInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// Load startup vars
var (
	mediaDirs = []string{"/mnt/core/pub/cloud/tv", "/mnt/core/pub/cloud/movies"}
)

var sabAPIKey string = getEnv("SABAPIKEY", "YOURKEY")
var sabHost string = getEnv("SABHOST", "https://sab.zaldre.com")
var uptime string = getEnv("UPTIME", "https://app.statuscake.com/button/index.php?Track=lmmBTReo4c&Days=30&Design=2")
var logLevel string = getEnv("LOGLEVEL", "Normal") // None, Normal, Debug
var webTimeout int = getEnvInt("WEBTIMEOUT", 15)   // seconds)
var sabPort int = getEnvInt("SABPORT", 443)
var statsFile string = getEnv("STATSFILE", "/container/data/stats/index.html")

// SABQueue represents the SabNZBD queue response
type SABQueue struct {
	Queue struct {
		MBLeft string `json:"mbleft"`
	} `json:"queue"`
}

func outLog(text string, obj interface{}) {
	if logLevel == "None" {
		return
	}
	timestamp := time.Now().Format("02.01.2006 15:04:05:")
	if text != "" {
		fmt.Printf("%s %s\n", timestamp, text)
	}
	if obj != nil {
		fmt.Println(timestamp)
		fmt.Println(obj)
	}
}

func calcSize(byteCount int64) string {
	if byteCount == 0 {
		return "0 Bytes"
	}

	sizeUnits := []string{"Bytes", "KB", "MB", "GB", "TB"}
	unitIndex := int(math.Floor(math.Log(float64(byteCount)) / math.Log(1024)))

	if unitIndex == 0 {
		return fmt.Sprintf("%d Bytes", byteCount)
	}

	size := float64(byteCount) / math.Pow(1024, float64(unitIndex))
	return fmt.Sprintf("%.2f %s", size, sizeUnits[unitIndex])
}

func getSABQueueSize() (int64, error) {
	url := fmt.Sprintf("%s:%d/sabnzbd/api?mode=queue&output=json&apikey=%s",
		sabHost, sabPort, sabAPIKey)
	if logLevel == "Debug" {
		fmt.Println(url)
	}

	client := &http.Client{Timeout: time.Duration(webTimeout) * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var sabQueue SABQueue
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

func getMediaSize() (int64, error) {
	var totalSize int64

	for _, mediaDir := range mediaDirs {
		cmd := exec.Command("du", "-sb", mediaDir)
		output, err := cmd.Output()
		if err != nil {
			outLog(fmt.Sprintf("Error getting size for %s: %v", mediaDir, err), nil)
			continue
		}

		parts := strings.Split(string(output), "\t")
		if len(parts) < 1 {
			continue
		}

		size, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			outLog(fmt.Sprintf("Error parsing size for %s: %v", mediaDir, err), nil)
			continue
		}

		totalSize += size
	}

	return totalSize, nil
}

func getMaintenanceNotice() string {
	scriptDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}

	maintenanceFile := filepath.Join(scriptDir, "maintenance.txt")
	content, err := os.ReadFile(maintenanceFile)
	if err != nil {
		return ""
	}

	outLog("Maintenance notice found, Appending this to HTML output", nil)
	notice := strings.TrimSpace(string(content))
	if logLevel == "debug" {
		outLog(notice, nil)
	}
	fmt.Println(notice)
	return notice
}

func generateHTML(mediaSize, downloadSize string, maintenance string) string {
	plex := "https://app.plex.tv"

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Server Statistics</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
            background-color: #1a1b1e;
            color: #e4e5e7;
        }

        .container {
            background: #2c2e33;
            border-radius: 12px;
            padding: 2rem;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
        }

        .stats-grid {
            display: grid;
            gap: 1rem;
            margin: 2rem 0;
        }

        .stat-item {
            background: #373a40;
            padding: 1rem;
            border-radius: 8px;
            border: 1px solid #4a4d52;
        }

        .plex-link {
            display: inline-block;
            background: #e5a00d;
            color: #1a1b1e;
            text-decoration: none;
            padding: 0.75rem 1.5rem;
            border-radius: 6px;
            margin-bottom: 1.5rem;
            transition: background-color 0.2s;
            font-weight: 500;
        }

        .plex-link:hover {
            background: #f5b025;
        }

        .status-section {
            margin-top: 2rem;
            text-align: center;
        }

        .status-section img {
            border-radius: 4px;
        }

        strong {
            color: #e5a00d;
        }

        @media (max-width: 600px) {
            body {
                padding: 1rem;
            }
            
            .container {
                padding: 1rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <a href="%s" class="plex-link">Plex - Watch TV/Movies</a>
        <br>
        <img src="%s" alt="Server Status">
        <br>
        <div class="stats-grid">
            <div class="stat-item">
                <strong>Total:</strong> %s
            </div>
            <div class="stat-item">
                <strong>Downloads:</strong> %s
            </div>
            <div class="stat-item">
                <strong>Maintenance:</strong> %s
            </div>

        </div>

    </div>
</body>
</html>`, plex, uptime, mediaSize, downloadSize, maintenance)
}

func main() {
	outLog("Querying SabNZBD for queue size and remaining MB", nil)

	downloadSizeBytes, err := getSABQueueSize()
	if err != nil {
		log.Fatalf("Error querying SabNZBD: %v", err)
	}

	outLog("Finalising processing for download queue size", nil)
	downloadSizeStr := calcSize(downloadSizeBytes)

	outLog("Getting media directory stats", nil)
	totalMediaSize, err := getMediaSize()
	if err != nil {
		log.Fatalf("Error getting media size: %v", err)
	}

	mediaSizeStr := calcSize(totalMediaSize)

	if logLevel == "Debug" {
		outLog(fmt.Sprintf("MediaSize: %s", mediaSizeStr), nil)
		outLog(fmt.Sprintf("DownloadSize: %s", downloadSizeStr), nil)
	}

	maintenance := getMaintenanceNotice()

	outLog("Creating HTML", nil)
	html := generateHTML(mediaSizeStr, downloadSizeStr, maintenance)

	if logLevel == "debug" {
		outLog(html, nil)
	}

	if err := os.WriteFile(statsFile, []byte(html), 0644); err != nil {
		log.Fatalf("Error writing HTML file: %v", err)
	}

	outLog("HTML file created successfully", nil)
}
