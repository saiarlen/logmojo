package ws

import (
	"log"
	"time"

	"logmojo/internal/metrics"

	"github.com/gofiber/contrib/websocket"
)

func MetricsHandler(c *websocket.Conn) {
	defer c.Close()

	ticker := time.NewTicker(1 * time.Second)
	pingTicker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer pingTicker.Stop()

	// Handle client messages (including pong)
	go func() {
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			m, err := metrics.GetHostMetrics()
			if err != nil {
				log.Println("Error getting metrics:", err)
				continue
			}
			if err := c.WriteJSON(m); err != nil {
				return
			}
		case <-pingTicker.C:
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
