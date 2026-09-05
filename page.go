package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"strings"
)

// The icon is compiled in so the page and its favicon always ship together;
// there is no build step that could deploy one without the other.
//
//go:embed favicon.ico
var faviconICO []byte

// PageData is everything the template renders. Values are pre-formatted
// strings: the template decides layout, not units.
type PageData struct {
	PlexURL        string
	UptimeImageURL string
	MediaSize      string
	DownloadSize   string
	Maintenance    string
	Uploaded       string
	Remaining      string
}

// html/template escapes every field on the way out. The maintenance notice is
// operator-supplied free text on an internet-facing page, and the previous
// Sprintf-based build would have let a stray angle bracket break the markup.
var pageTemplate = template.Must(template.New("stats").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Server Statistics</title>
    <link rel="icon" href="favicon.ico" sizes="any">
    <style>
        /* Light scheme. Declared on :root so it is also the fallback for any
           browser that does not report a preference. */
        :root {
            color-scheme: light dark;
            --page-bg: #f4f5f7;
            --panel-bg: #ffffff;
            --tile-bg: #f0f1f4;
            --tile-border: #d9dce1;
            --text: #1a1b1e;
            /* Plex amber is too light for text on white, so the label colour
               darkens while the button keeps the brand tone. */
            --label: #8a5a00;
            --button-bg: #e5a00d;
            --button-bg-hover: #cf8f08;
            --button-text: #1a1b1e;
            --shadow: rgba(0, 0, 0, 0.08);
        }

        /* Dark scheme, matching the original page. */
        @media (prefers-color-scheme: dark) {
            :root {
                --page-bg: #1a1b1e;
                --panel-bg: #2c2e33;
                --tile-bg: #373a40;
                --tile-border: #4a4d52;
                --text: #e4e5e7;
                --label: #e5a00d;
                --button-bg: #e5a00d;
                --button-bg-hover: #f5b025;
                --button-text: #1a1b1e;
                --shadow: rgba(0, 0, 0, 0.3);
            }
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
            background-color: var(--page-bg);
            color: var(--text);
        }

        .container {
            background: var(--panel-bg);
            border-radius: 12px;
            padding: 2rem;
            box-shadow: 0 4px 6px var(--shadow);
        }

        .stats-grid {
            display: grid;
            gap: 1rem;
            margin: 2rem 0;
        }

        .stat-item {
            background: var(--tile-bg);
            padding: 1rem;
            border-radius: 8px;
            border: 1px solid var(--tile-border);
        }

        .plex-link {
            display: inline-block;
            background: var(--button-bg);
            color: var(--button-text);
            text-decoration: none;
            padding: 0.75rem 1.5rem;
            border-radius: 6px;
            margin-bottom: 1.5rem;
            transition: background-color 0.2s;
            font-weight: 500;
        }

        .plex-link:hover {
            background: var(--button-bg-hover);
        }

        .status-section {
            margin-top: 2rem;
            text-align: center;
        }

        .status-section img {
            border-radius: 4px;
        }

        strong {
            color: var(--label);
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
        <a href="{{ .PlexURL }}" class="plex-link">Plex - Watch TV/Movies</a>
        <br>
        <img src="{{ .UptimeImageURL }}" alt="Server Status">
        <br>
        <div class="stats-grid">
            <div class="stat-item">
                <strong>Total:</strong> {{ .MediaSize }}
            </div>
            <div class="stat-item">
                <strong>Downloads:</strong> {{ .DownloadSize }}
            </div>
            <div class="stat-item">
                <strong>Maintenance:</strong> {{ .Maintenance }}
            </div>
            <div class="stat-item">
                <strong>Uploaded:</strong> {{ .Uploaded }}
            </div>
            <div class="stat-item">
                <strong>Remaining:</strong> {{ .Remaining }}
            </div>

        </div>

    </div>
</body>
</html>`))

// RenderPage produces the finished HTML.
func RenderPage(data PageData) (string, error) {
	var rendered strings.Builder
	if err := pageTemplate.Execute(&rendered, data); err != nil {
		return "", fmt.Errorf("rendering page: %w", err)
	}
	return rendered.String(), nil
}

// WriteFavicon places the embedded icon next to the page.
//
// It is rewritten on every run rather than only when missing: that keeps a
// deploy of a new icon from needing any manual step on the volume.
func WriteFavicon(path string) error {
	if err := os.WriteFile(path, faviconICO, 0o644); err != nil {
		return fmt.Errorf("writing favicon to %s: %w", path, err)
	}
	return nil
}

// WritePage writes the rendered HTML to disk.
func WritePage(path, html string) error {
	if err := os.WriteFile(path, []byte(html), 0o644); err != nil {
		return fmt.Errorf("writing page to %s: %w", path, err)
	}
	return nil
}
