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

// cssBlock returns the body of the rule introduced by selector.
func cssBlock(t *testing.T, html, selector string) string {
	t.Helper()

	start := strings.Index(html, selector+" {")
	require.GreaterOrEqual(t, start, 0, "no rule for %s", selector)

	end := strings.Index(html[start:], "}")
	require.Greater(t, end, 0, "unterminated rule for %s", selector)

	return html[start : start+end]
}

func TestThemesAreDistinct(t *testing.T) {
	require.Len(t, Themes, 5, "the picker is documented as offering five themes")

	seen := map[string]bool{}
	for _, theme := range Themes {
		assert.NotEmpty(t, theme.Label)
		assert.False(t, seen[theme.Value], "duplicate theme %s", theme.Value)
		seen[theme.Value] = true
	}
	assert.True(t, seen["plex"], "the original look must remain on offer")
}

func TestRenderPage(t *testing.T) {
	html, err := RenderPage(samplePage())
	require.NoError(t, err)

	t.Run("renders every figure with its label, in order", func(t *testing.T) {
		// Maintenance sits last: it is usually empty, and a blank tile between
		// two figures reads as a missing number rather than as no notice.
		ordered := []string{
			"<strong>Total:</strong> 46.54 TB",
			"<strong>Downloads:</strong> 1.21 GB",
			"<strong>Uploaded:</strong> 19.53 TB (42.0%)",
			"<strong>Remaining:</strong> 27.02 TB",
			"<strong>Maintenance:</strong>",
		}

		previous := -1
		for _, want := range ordered {
			at := strings.Index(html, want)
			require.GreaterOrEqual(t, at, 0, "missing %s", want)
			assert.Greater(t, at, previous, "%s is out of order", want)
			previous = at
		}
	})

	t.Run("links the favicon", func(t *testing.T) {
		assert.Contains(t, html, `<link rel="icon" href="favicon.ico"`)
	})

	t.Run("ships every theme in the picker", func(t *testing.T) {
		assert.Contains(t, html, "@media (prefers-color-scheme: light)")
		for _, theme := range Themes {
			assert.Contains(t, html, `:root[data-theme="`+theme.Value+`"]`,
				"palette for %s", theme.Value)
			assert.Contains(t, html, `<option value="`+theme.Value+`">`+theme.Label+`</option>`,
				"picker entry for %s", theme.Value)
		}
	})

	t.Run("every theme defines the whole palette", func(t *testing.T) {
		// A palette missing a property inherits it from :root, which silently
		// leaves one Plex colour stranded in an otherwise different theme.
		properties := []string{
			"--page-bg", "--panel-bg", "--tile-bg", "--tile-border", "--text",
			"--label", "--button-bg", "--button-bg-hover", "--button-text", "--shadow",
		}
		for _, theme := range Themes {
			block := cssBlock(t, html, `:root[data-theme="`+theme.Value+`"]`)
			for _, property := range properties {
				assert.Contains(t, block, property+":", "%s is missing %s", theme.Value, property)
			}
			assert.Contains(t, block, "color-scheme:", "%s is missing color-scheme", theme.Value)
		}
	})

	t.Run("an explicit choice overrides the system preference", func(t *testing.T) {
		// Without the :not(), a system-light reader who picked a dark theme
		// would still get the light palette from the media query.
		assert.Contains(t, html, ":root:not([data-theme])")
		// The explicit blocks tie with the media query on specificity, so they
		// have to come after it in source order to win.
		assert.Greater(t, strings.Index(html, `:root[data-theme="plex"]`),
			strings.Index(html, "@media (prefers-color-scheme: light)"))
	})

	t.Run("offers an Auto option alongside the named themes", func(t *testing.T) {
		assert.Contains(t, html, `<option value="auto">Auto</option>`)
	})

	t.Run("the picker starts hidden", func(t *testing.T) {
		// Hidden until the script unhides it, so a reader without JavaScript is
		// not shown a dead control.
		assert.Contains(t, html, `id="theme-picker" for="theme-select" hidden`)
		assert.Contains(t, html, "picker.hidden = false")
	})

	t.Run("applies a stored choice before the body renders", func(t *testing.T) {
		head := html[:strings.Index(html, "</head>")]
		assert.Contains(t, head, "localStorage.getItem('theme')")
		assert.Contains(t, head, "setAttribute('data-theme', stored)")
	})

	t.Run("storage access is guarded", func(t *testing.T) {
		// Blocked storage throws on access rather than returning null, so both
		// the read and the write have to sit inside a try.
		assert.GreaterOrEqual(t, strings.Count(html, "try {"), 2)
		assert.GreaterOrEqual(t, strings.Count(html, "catch (error)"), 2)
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
