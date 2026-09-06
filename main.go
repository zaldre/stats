package main

import (
	"context"
	"log"
	"time"
)

// VERSION is bumped whenever a change becomes eligible for commit.
const VERSION = "0.5.0"

func main() {
	if err := run(); err != nil {
		// Fatal is reserved for the unrecoverable: by this point no page could be
		// produced at all. Every individual figure degrades to "Unavailable"
		// instead of reaching here.
		log.Fatalf("stats %s: %v", VERSION, err)
	}
}

func run() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	logger := NewLogger(config.LogLevel)
	logger.Infof("stats %s starting", VERSION)

	ctx := context.Background()
	page := PageData{
		PlexURL:        config.PlexURL,
		UptimeImageURL: config.UptimeImageURL,
		MediaSize:      Unavailable,
		DownloadSize:   Unavailable,
		Uploaded:       Unavailable,
		Remaining:      Unavailable,
	}

	logger.Infof("Querying SabNZBD for queue size and remaining MB")
	downloadBytes, err := FetchDownloadSize(ctx, config)
	if err != nil {
		logger.Errorf("Could not read the download queue: %v", err)
	} else if page.DownloadSize, err = FormatBytes(downloadBytes); err != nil {
		logger.Errorf("Could not format the download queue size: %v", err)
		page.DownloadSize = Unavailable
	}

	page.Maintenance, err = ReadMaintenanceNotice(config.MaintenanceFile)
	if err != nil {
		// An unreadable banner should not cost the rest of the page.
		logger.Errorf("Could not read the maintenance notice: %v", err)
	} else if page.Maintenance != "" {
		logger.Infof("Maintenance notice found, appending it to the HTML output")
		logger.Debugf("Maintenance notice: %s", page.Maintenance)
	}

	// One measurement of each side feeds all three figures: the local tree is
	// the total, and the remote is what of it has already gone up. Measuring the
	// total separately is what used to let it disagree with the other two.
	cloudCtx, cancel := context.WithTimeout(ctx, config.CloudTimeout)
	defer cancel()
	page.MediaSize, page.Uploaded, page.Remaining = DescribeCloudProgress(
		CollectCloudProgress(cloudCtx, config, logger), time.Now())

	logger.Debugf("Total: %s", page.MediaSize)
	logger.Debugf("Downloads: %s", page.DownloadSize)
	logger.Debugf("Uploaded: %s", page.Uploaded)
	logger.Debugf("Remaining: %s", page.Remaining)

	logger.Infof("Creating HTML")
	html, err := RenderPage(page)
	if err != nil {
		return err
	}
	logger.Debugf("%s", html)

	if err := WritePage(config.StatsFile, html); err != nil {
		return err
	}
	logger.Infof("HTML file created successfully")

	// The icon is cosmetic: failing to place it should not fail a run that has
	// already written a good page.
	if err := WriteFavicon(config.FaviconFile); err != nil {
		logger.Errorf("Could not write the favicon: %v", err)
	}

	return nil
}
