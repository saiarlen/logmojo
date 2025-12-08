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
					logPath = l.Path
					break
				}
			}
		}
	}

	if logPath == "" {
		c.WriteMessage(websocket.TextMessage, []byte("Log not found"))
		c.Close()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lines := make(chan string)
	go func() {
		if err := logs.StreamLog(ctx, logPath, lines); err != nil {
			log.Println("Error streaming log:", err)
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
