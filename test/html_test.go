package test

import (
	"strings"
	"testing"

	"github.com/snowpea/stats/pkg/html"
)

func TestGenerateHTML(t *testing.T) {
	tests := []struct {
		name         string
		mediaSize    string
		downloadSize string
		maintenance  string
		expected     []string // strings that should be present in HTML
	}{
		{
			name:         "Basic HTML generation",
			mediaSize:    "1.5 GB",
			downloadSize: "500 MB",
			maintenance:  "No maintenance scheduled",
			expected: []string{
				"<!DOCTYPE html>",
				"<html lang=\"en\">",
				"<title>Server Statistics</title>",
				"1.5 GB",
				"500 MB",
				"No maintenance scheduled",
				"Plex - Watch TV/Movies",
				"https://app.plex.tv",
			},
		},
		{
			name:         "Empty maintenance",
			mediaSize:    "2.0 TB",
			downloadSize: "0 Bytes",
			maintenance:  "",
			expected: []string{
				"2.0 TB",
				"0 Bytes",
				"<strong>Maintenance:</strong> ",
			},
		},
		{
			name:         "Large values",
			mediaSize:    "10.5 TB",
			downloadSize: "1.2 GB",
			maintenance:  "Scheduled maintenance on 2024-01-01",
			expected: []string{
				"10.5 TB",
				"1.2 GB",
				"Scheduled maintenance on 2024-01-01",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlContent := html.GenerateHTML(tt.mediaSize, tt.downloadSize, tt.maintenance, "https://example.com")

			// Check that all expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(htmlContent, expected) {
					t.Errorf("GenerateHTML() should contain '%s', but it doesn't", expected)
				}
			}

			// Check that HTML is properly structured
			if !strings.Contains(htmlContent, "<!DOCTYPE html>") {
				t.Errorf("GenerateHTML() should start with DOCTYPE declaration")
			}

			if !strings.Contains(htmlContent, "</html>") {
				t.Errorf("GenerateHTML() should end with </html> tag")
			}

			// Check for specific sections
			if !strings.Contains(htmlContent, "<div class=\"container\">") {
				t.Errorf("GenerateHTML() should contain container div")
			}

			if !strings.Contains(htmlContent, "<div class=\"stats-grid\">") {
				t.Errorf("GenerateHTML() should contain stats grid")
			}

			// Check that Plex link is present
			if !strings.Contains(htmlContent, "https://app.plex.tv") {
				t.Errorf("GenerateHTML() should contain Plex link")
			}
		})
	}
}

func TestGenerateHTMLStructure(t *testing.T) {
	htmlContent := html.GenerateHTML("1 GB", "500 MB", "Test maintenance", "https://example.com")

	// Check for required HTML structure
	requiredElements := []string{
		"<!DOCTYPE html>",
		"<html lang=\"en\">",
		"<head>",
		"<meta charset=\"UTF-8\">",
		"<meta name=\"viewport\"",
		"<title>Server Statistics</title>",
		"<style>",
		"</style>",
		"<body>",
		"<div class=\"container\">",
		"<a href=\"https://app.plex.tv\"",
		"<div class=\"stats-grid\">",
		"<div class=\"stat-item\">",
		"<strong>Total:</strong>",
		"<strong>Downloads:</strong>",
		"<strong>Maintenance:</strong>",
		"</div>",
		"</body>",
		"</html>",
	}

	for _, element := range requiredElements {
		if !strings.Contains(htmlContent, element) {
			t.Errorf("GenerateHTML() should contain '%s'", element)
		}
	}
}

func TestGenerateHTMLCSS(t *testing.T) {
	htmlContent := html.GenerateHTML("1 GB", "500 MB", "Test maintenance", "https://example.com")

	// Check for CSS classes and styles
	cssElements := []string{
		"body {",
		"font-family:",
		"background-color: #1a1b1e",
		"color: #e4e5e7",
		".container {",
		"background: #2c2e33",
		"border-radius: 12px",
		".stats-grid {",
		".stat-item {",
		".plex-link {",
		"background: #e5a00d",
		"@media (max-width: 600px)",
	}

	for _, element := range cssElements {
		if !strings.Contains(htmlContent, element) {
			t.Errorf("GenerateHTML() should contain CSS element '%s'", element)
		}
	}
}
