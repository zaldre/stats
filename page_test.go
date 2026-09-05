package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func samplePage() PageData {
	return PageData{
		PlexURL:        "https://app.plex.tv",
		UptimeImageURL: "https://example.invalid/button.png",
		MediaSize:      "46.54 TB",
		DownloadSize:   "1.21 GB",
		Maintenance:    "",
		Uploaded:       "19.53 TB (42.0%)",
		Remaining:      "27.02 TB",
	}
}

func TestRenderPage(t *testing.T) {
	html, err := RenderPage(samplePage())
	require.NoError(t, err)

	t.Run("renders every figure with its label", func(t *testing.T) {
		for _, want := range []string{
			"<strong>Total:</strong> 46.54 TB",
			"<strong>Downloads:</strong> 1.21 GB",
			"<strong>Uploaded:</strong> 19.53 TB (42.0%)",
			"<strong>Remaining:</strong> 27.02 TB",
		} {
			assert.Contains(t, html, want)
		}
	})

	t.Run("links the favicon", func(t *testing.T) {
		assert.Contains(t, html, `<link rel="icon" href="favicon.ico"`)
	})

	t.Run("ships both colour schemes", func(t *testing.T) {
		assert.Contains(t, html, "color-scheme: light dark")
		assert.Contains(t, html, "@media (prefers-color-scheme: dark)")
		// The light scheme lives on bare :root so it is also the no-preference
		// fallback; the dark one only overrides the custom properties.
		assert.Contains(t, html, "--page-bg: #f4f5f7")
		assert.Contains(t, html, "--page-bg: #1a1b1e")
	})

	t.Run("no hard-coded colours survive outside the palette", func(t *testing.T) {
		body := html[strings.Index(html, "body {"):]
		assert.Contains(t, body, "background-color: var(--page-bg)")
		assert.Contains(t, body, "color: var(--text)")
	})
}

func TestRenderPageEscapesMaintenanceNotice(t *testing.T) {
	page := samplePage()
	// Operator-supplied free text on an internet-facing page.
	page.Maintenance = `<script>alert("x")</script>`

	html, err := RenderPage(page)
	require.NoError(t, err)

	assert.NotContains(t, html, "<script>alert")
	assert.Contains(t, html, "&lt;script&gt;")
}

func TestWriteFavicon(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favicon.ico")
	require.NoError(t, WriteFavicon(path))

	written, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.NotEmpty(t, written)
	// ICO header: reserved 0, type 1 (icon).
	assert.Equal(t, []byte{0x00, 0x00, 0x01, 0x00}, written[:4])
	assert.Equal(t, faviconICO, written)
}
