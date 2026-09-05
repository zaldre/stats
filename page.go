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

// Themes are the palettes the reader can pick between, in the order the picker
// offers them. The value is the data-theme attribute; the label is what the
// reader sees.
//
// Plex is the page's original look and stays the default. Auto is not in this
// list because it is the absence of a choice rather than a palette of its own:
// it resolves to Plex or Light from the system preference.
var Themes = []struct {
	Value string
	Label string
}{
	{"plex", "Plex"},
	{"light", "Light"},
	{"nord", "Nord"},
	{"dracula", "Dracula"},
	{"solarized", "Solarized"},
}

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

// templateData is what actually reaches the template: the page figures plus the
// theme list, which is fixed rather than collected.
type templateData struct {
	PageData
	Themes []struct {
		Value string
		Label string
	}
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
    <script>
        // Runs before the body renders so a stored choice never flashes another
        // theme on the way in.
        (function () {
            try {
                var stored = localStorage.getItem('theme');
                if (stored && stored !== 'auto') {
                    document.documentElement.setAttribute('data-theme', stored);
                }
            } catch (error) {
                // Storage can be blocked outright in a private window. Falling
                // through leaves the system preference in charge, which is the
                // right default anyway.
            }
        })();
    </script>
    <style>
        /* Plex, the page's original look, is the default and so lives on bare
           :root. Every palette sets the same custom properties; nothing below
           this block refers to a literal colour. */
        :root {
            color-scheme: dark;
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

        /* With no explicit choice, a reader whose system asks for light gets the
           light palette. The :not() guard is what lets an explicit pick of a
           dark theme survive on a light system. */
        @media (prefers-color-scheme: light) {
            :root:not([data-theme]) {
                color-scheme: light;
                --page-bg: #f4f5f7;
                --panel-bg: #ffffff;
                --tile-bg: #f0f1f4;
                --tile-border: #d9dce1;
                --text: #1a1b1e;
                /* Plex amber is too light for text on white, so the label
                   darkens while the button keeps the brand tone. */
                --label: #8a5a00;
                --button-bg: #e5a00d;
                --button-bg-hover: #cf8f08;
                --button-text: #1a1b1e;
                --shadow: rgba(0, 0, 0, 0.08);
            }
        }

        /* Explicit choices. These come last so they win the cascade against the
           media query above, which they tie with on specificity. */
        :root[data-theme="plex"] {
            color-scheme: dark;
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

        :root[data-theme="light"] {
            color-scheme: light;
            --page-bg: #f4f5f7;
            --panel-bg: #ffffff;
            --tile-bg: #f0f1f4;
            --tile-border: #d9dce1;
            --text: #1a1b1e;
            --label: #8a5a00;
            --button-bg: #e5a00d;
            --button-bg-hover: #cf8f08;
            --button-text: #1a1b1e;
            --shadow: rgba(0, 0, 0, 0.08);
        }

        :root[data-theme="nord"] {
            color-scheme: dark;
            --page-bg: #2e3440;
            --panel-bg: #3b4252;
            --tile-bg: #434c5e;
            --tile-border: #4c566a;
            --text: #eceff4;
            --label: #88c0d0;
            --button-bg: #88c0d0;
            --button-bg-hover: #8fbcbb;
            --button-text: #2e3440;
            --shadow: rgba(0, 0, 0, 0.3);
        }

        :root[data-theme="dracula"] {
            color-scheme: dark;
            --page-bg: #282a36;
            --panel-bg: #343746;
            --tile-bg: #44475a;
            --tile-border: #6272a4;
            --text: #f8f8f2;
            --label: #ff79c6;
            --button-bg: #bd93f9;
            --button-bg-hover: #d0aeff;
            --button-text: #282a36;
            --shadow: rgba(0, 0, 0, 0.35);
        }

        :root[data-theme="solarized"] {
            color-scheme: light;
            --page-bg: #eee8d5;
            --panel-bg: #fdf6e3;
            --tile-bg: #f7f1de;
            --tile-border: #ddd6c1;
            --text: #586e75;
            --label: #b58900;
            --button-bg: #268bd2;
            --button-bg-hover: #1f7ab8;
            --button-text: #fdf6e3;
            --shadow: rgba(88, 110, 117, 0.15);
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

        .theme-picker {
            float: right;
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            font-size: 0.85rem;
            color: var(--text);
        }

        .theme-picker select {
            background: var(--tile-bg);
            color: var(--text);
            border: 1px solid var(--tile-border);
            border-radius: 6px;
            padding: 0.35rem 0.5rem;
            font: inherit;
            font-size: 0.85rem;
            cursor: pointer;
        }

        .theme-picker select:hover {
            border-color: var(--label);
        }

        .theme-picker select:focus-visible {
            outline: 2px solid var(--label);
            outline-offset: 2px;
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
        <label class="theme-picker" id="theme-picker" for="theme-select" hidden>
            Theme
            <select id="theme-select">
                <option value="auto">Auto</option>
                {{- range .Themes }}
                <option value="{{ .Value }}">{{ .Label }}</option>
                {{- end }}
            </select>
        </label>
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
                <strong>Uploaded:</strong> {{ .Uploaded }}
            </div>
            <div class="stat-item">
                <strong>Remaining:</strong> {{ .Remaining }}
            </div>
            <div class="stat-item">
                <strong>Maintenance:</strong> {{ .Maintenance }}
            </div>

        </div>

    </div>
    <script>
        (function () {
            var root = document.documentElement;
            var picker = document.getElementById('theme-picker');
            var select = document.getElementById('theme-select');

            function apply(theme) {
                if (theme === 'auto') {
                    root.removeAttribute('data-theme');
                } else {
                    root.setAttribute('data-theme', theme);
                }

                try {
                    if (theme === 'auto') {
                        localStorage.removeItem('theme');
                    } else {
                        localStorage.setItem('theme', theme);
                    }
                } catch (error) {
                    // The choice still applies to this page view; it just will
                    // not survive a reload.
                }
            }

            // Reflect whatever the pre-paint script settled on. Reading the
            // attribute rather than storage keeps the two in step even when
            // storage is unavailable.
            var active = root.getAttribute('data-theme') || 'auto';
            // A stored theme this build no longer ships would leave the select
            // blank, so fall back rather than show an empty control.
            select.value = active;
            if (!select.value) {
                select.value = 'auto';
                apply('auto');
            }

            // Revealed only once the script is running, so a reader without
            // JavaScript is not shown a control that cannot do anything.
            picker.hidden = false;

            select.addEventListener('change', function () {
                apply(select.value);
            });
        })();
    </script>
</body>
</html>`))

// RenderPage produces the finished HTML.
func RenderPage(data PageData) (string, error) {
	var rendered strings.Builder
	if err := pageTemplate.Execute(&rendered, templateData{PageData: data, Themes: Themes}); err != nil {
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
