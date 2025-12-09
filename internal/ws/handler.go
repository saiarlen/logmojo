package ws

import (
	"context"
	"local-monitor/internal/config"
	"local-monitor/internal/logs"
	"log"

	"github.com/gofiber/contrib/websocket"
)

func Handler(c *websocket.Conn) {
	appName := c.Query("app")
	logName := c.Query("log")

	var logPath string
	// Find the log path from config
	for _, app := range config.AppConfigData.Apps {
		if app.Name == appName {
			for _, l := range app.Logs {
				if l.Name == logName {
					// Get the actual file path (handle directories)
					files, err := logs.ListFiles(appName, logName)
					if err == nil && len(files) > 0 {
						// Use the first non-archive file
						for _, f := range files {
							if !f.IsArchive {
								logPath = f.Path
								break
							}
						}
						if logPath == "" && len(files) > 0 {
							logPath = files[0].Path // Fallback to first file
						}
					} else {
						// Direct file path
						logPath = l.Path
					}
					break
				}
			}
		}
	}

	if logPath == "" {
		c.WriteMessage(websocket.TextMessage, []byte("Log file not found or not accessible"))
		c.Close()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lines := make(chan string)
	go func() {
		log.Printf("Starting stream for: %s", logPath)
		if err := logs.StreamLog(ctx, logPath, lines); err != nil {
			log.Printf("Error streaming log %s: %v", logPath, err)
			c.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		}
		close(lines)
	}()

	// Handle client disconnect
	go func() {
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	for line := range lines {
		if err := c.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			break
		}
	}
}
