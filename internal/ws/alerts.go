package ws

import (
	"encoding/json"
	"local-monitor/internal/db"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
)

type AlertUpdate struct {
	Type    string      `json:"type"`    // "new_alert", "rule_updated", "alert_resolved"
	Alert   *db.Alert   `json:"alert,omitempty"`
	Rule    *db.AlertRule `json:"rule,omitempty"`
	Message string      `json:"message"`
}

var alertClients = make(map[*websocket.Conn]bool)

func AlertsHandler(c *websocket.Conn) {
	defer func() {
		delete(alertClients, c)
		c.Close()
	}()

	// Add client to the list
	alertClients[c] = true
	log.Printf("[WS] Alert client connected, total: %d", len(alertClients))

	// Send initial data
	sendInitialAlertData(c)

	// Heartbeat ticker
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Handle messages and heartbeat
	for {
		select {
		case <-pingTicker.C:
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[WS] Alert client ping failed: %v", err)
				return
			}
		default:
			_, _, err := c.ReadMessage()
			if err != nil {
				log.Printf("[WS] Alert client disconnected: %v", err)
				return
			}
		}
	}
}

func sendInitialAlertData(c *websocket.Conn) {
	// Send current alert rules
	rules, err := db.GetAlertRules()
	if err == nil {
		for _, rule := range rules {
			update := AlertUpdate{
				Type: "rule_updated",
				Rule: &rule,
			}
			sendToClient(c, update)
		}
	}

	// Send recent alerts
	alerts, err := db.GetAlertHistory()
	if err == nil {
		for i, alert := range alerts {
			if i >= 10 { // Send only last 10 alerts
				break
			}
			update := AlertUpdate{
				Type:  "initial_alert",
				Alert: &alert,
			}
			sendToClient(c, update)
		}
	}
}

func sendToClient(c *websocket.Conn, update AlertUpdate) {
	data, err := json.Marshal(update)
	if err != nil {
		return
	}
	c.WriteMessage(websocket.TextMessage, data)
}

// BroadcastNewAlert sends new alert to all connected clients
func BroadcastNewAlert(alert db.Alert) {
	update := AlertUpdate{
		Type:    "new_alert",
		Alert:   &alert,
		Message: "New alert triggered",
	}
	broadcastAlertUpdate(update)
}

// BroadcastRuleUpdate sends rule update to all connected clients
func BroadcastRuleUpdate(rule db.AlertRule) {
	update := AlertUpdate{
		Type:    "rule_updated",
		Rule:    &rule,
		Message: "Alert rule updated",
	}
	broadcastAlertUpdate(update)
}

// BroadcastAlertResolved sends alert resolution to all connected clients
func BroadcastAlertResolved(alertId int) {
	update := AlertUpdate{
		Type:    "alert_resolved",
		Message: "Alert resolved",
	}
	broadcastAlertUpdate(update)
}

func broadcastAlertUpdate(update AlertUpdate) {
	data, err := json.Marshal(update)
	if err != nil {
		log.Printf("[WS] Failed to marshal alert update: %v", err)
		return
	}

	// Send to all connected clients
	for client := range alertClients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("[WS] Failed to send alert update to client: %v", err)
			delete(alertClients, client)
			client.Close()
		}
	}

	if len(alertClients) > 0 {
		log.Printf("[WS] Broadcasted alert update to %d clients", len(alertClients))
	}
}