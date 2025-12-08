package ws

import (
	"log"
	"time"

	"local-monitor/internal/metrics"

	"github.com/gofiber/contrib/websocket"
)

func MetricsHandler(c *websocket.Conn) {
	// Upgrade happens in main.go middleware
	defer c.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m, err := metrics.GetHostMetrics()
		if err != nil {
			log.Println("Error getting metrics:", err)
			continue
		}
		if err := c.WriteJSON(m); err != nil {
			// Client disconnected
			return
		}
	}
}
