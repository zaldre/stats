package html

import "fmt"

func GenerateHTML(mediaSize, downloadSize string, maintenance string, uptime string) string {
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
