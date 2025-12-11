package api

import (
	"fmt"
	"log"
	"logmojo/internal/alerts"
	"logmojo/internal/auth"
	"logmojo/internal/config"
	"logmojo/internal/db"
	"logmojo/internal/logger"
	"logmojo/internal/logs"
	"logmojo/internal/metrics"
	"logmojo/internal/processes"
	"logmojo/internal/services"
	"logmojo/internal/version"
	"logmojo/internal/ws"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
)

func Setup(app *fiber.App) {
	app.Use(fiberlogger.New(fiberlogger.Config{
		Output: logger.GetLogWriter(),
		Format: "[${time}] ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
	}))
	app.Use(cors.New())

	// Serve static files
	app.Static("/public", "./public")

	// Auth
	auth.CreateDefaultUser()
	app.Use(auth.RequireLogin)

	app.Get("/login", func(c *fiber.Ctx) error {
		token := c.Cookies("auth_token")
		if token != "" {
			return c.Redirect("/")
		}
		return c.Render("login", nil)
	})
	app.Post("/login", auth.LoginHandler)
	app.Get("/logout", auth.LogoutHandler)

	// Page Routes
	app.Get("/", func(c *fiber.Ctx) error {
		settings, _ := db.GetAppSettings()
		data := fiber.Map{
			"current_page":   "dashboard",
			"app_name":       "Logmojo Monitor",
			"copyright_text": "© 2024 Logmojo",
			"logo_type":      "text",
			"version":        version.Version,
		}
		if settings != nil {
			data["app_name"] = settings["app_name"]
			data["copyright_text"] = settings["copyright_text"]
			data["logo_type"] = settings["logo_type"]
		}
		return c.Render("dashboard", data)
	})

	app.Get("/logs", func(c *fiber.Ctx) error {
		settings, _ := db.GetAppSettings()
		data := fiber.Map{
			"current_page":   "logs",
			"app_name":       "Logmojo Monitor",
			"copyright_text": "© 2024 Logmojo",
			"logo_type":      "text",
			"version":        version.Version,
		}
		if settings != nil {
			data["app_name"] = settings["app_name"]
			data["copyright_text"] = settings["copyright_text"]
			data["logo_type"] = settings["logo_type"]
		}
		return c.Render("logs", data)
	})

	// API Group
	apiGroup := app.Group("/api")

	apiGroup.Get("/logs/files", func(c *fiber.Ctx) error {
		appName := c.Query("app")
		logName := c.Query("log")
		if appName == "" || logName == "" {
			return c.Status(400).JSON(fiber.Map{"error": "app and log params required"})
		}

		files, err := logs.ListFiles(appName, logName)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		if files == nil {
			files = []logs.LogFile{}
		}

		return c.JSON(files)
	})

	apiGroup.Get("/logs/search", func(c *fiber.Ctx) error {
		q := c.Query("q")
		app := c.Query("app")
		logName := c.Query("log")
		file := c.Query("file")
		level := c.Query("level")
		limit := c.QueryInt("limit", 500)

		if app == "" || logName == "" {
			return c.Status(400).JSON(fiber.Map{"error": "app and log parameters required"})
		}

		results, err := logs.Search(q, app, logName, file, level, limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		if results == nil {
			results = []logs.LogResult{}
		}

		return c.JSON(results)
	})

	app.Get("/processes", func(c *fiber.Ctx) error {
		settings, _ := db.GetAppSettings()
		data := fiber.Map{
			"current_page":   "processes",
			"app_name":       "Logmojo Monitor",
			"copyright_text": "© 2024 Logmojo",
			"logo_type":      "text",
			"version":        version.Version,
		}
		if settings != nil {
			data["app_name"] = settings["app_name"]
			data["copyright_text"] = settings["copyright_text"]
			data["logo_type"] = settings["logo_type"]
		}
		return c.Render("processes", data)
	})

	app.Get("/services", func(c *fiber.Ctx) error {
		settings, _ := db.GetAppSettings()
		data := fiber.Map{
			"current_page":   "services",
			"app_name":       "Logmojo Monitor",
			"copyright_text": "© 2024 Logmojo",
			"logo_type":      "text",
			"version":        version.Version,
		}
		if settings != nil {
			data["app_name"] = settings["app_name"]
			data["copyright_text"] = settings["copyright_text"]
			data["logo_type"] = settings["logo_type"]
		}
		return c.Render("services", data)
	})

	app.Get("/alerts", func(c *fiber.Ctx) error {
		settings, _ := db.GetAppSettings()
		data := fiber.Map{
			"current_page":   "alerts",
			"app_name":       "Logmojo Monitor",
			"copyright_text": "© 2024 Logmojo",
			"logo_type":      "text",
			"version":        version.Version,
		}
		if settings != nil {
			data["app_name"] = settings["app_name"]
			data["copyright_text"] = settings["copyright_text"]
			data["logo_type"] = settings["logo_type"]
		}
		return c.Render("alerts", data)
	})

	app.Get("/settings", func(c *fiber.Ctx) error {
		settings, _ := db.GetAppSettings()
		data := fiber.Map{
			"current_page":   "settings",
			"app_name":       "Logmojo Monitor",
			"copyright_text": "© 2024 Logmojo",
			"logo_type":      "text",
			"version":        version.Version,
		}
		if settings != nil {
			data["app_name"] = settings["app_name"]
			data["copyright_text"] = settings["copyright_text"]
			data["logo_type"] = settings["logo_type"]
		}
		return c.Render("settings", data)
	})

	api := app.Group("/api")

	api.Get("/apps", func(c *fiber.Ctx) error {
		return c.JSON(config.AppConfigData.Apps)
	})

	api.Get("/metrics/host", func(c *fiber.Ctx) error {
		m, err := metrics.GetHostMetrics()
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(m)
	})

	api.Get("/metrics/history", func(c *fiber.Ctx) error {
		metricType := c.Query("type", "disk")
		rangeStr := c.Query("range", "1h") // 1h, 6h, 24h

		var tableName string
		var valCol string
		switch metricType {
		case "cpu":
			tableName = "cpu_history"
			valCol = "usage_percent"
		case "ram":
			tableName = "ram_history"
			valCol = "usage_percent"
		default:
			tableName = "disk_history"
			valCol = "used_percent"
		}

		// Calculate time cutoff
		limit := time.Now().Add(-1 * time.Hour)
		if rangeStr == "6h" {
			limit = time.Now().Add(-6 * time.Hour)
		} else if rangeStr == "24h" {
			limit = time.Now().Add(-24 * time.Hour)
		}

		query := "SELECT timestamp, " + valCol + " FROM " + tableName + " WHERE timestamp > ? ORDER BY timestamp ASC"
		rows, err := db.DB.Query(query, limit)
		if err != nil {
			// Handle table not found or other db errors gracefully
			return c.Status(500).SendString(err.Error())
		}
		defer rows.Close()

		var history []map[string]interface{}
		for rows.Next() {
			var ts time.Time
			var val float64
			rows.Scan(&ts, &val)
			history = append(history, map[string]interface{}{
				"timestamp": ts,
				"value":     val,
			})
		}
		return c.JSON(history)
	})

	api.Get("/metrics/disk-history", func(c *fiber.Ctx) error {
		return c.Redirect("/api/metrics/history?type=disk")
	})

	api.Get("/processes", func(c *fiber.Ctx) error {
		p, err := processes.ListProcesses()
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(p)
	})

	api.Post("/processes/kill", func(c *fiber.Ctx) error {
		type KillReq struct {
			PID  int32  `json:"pid"`
			Name string `json:"name"`
		}
		var req KillReq
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}
		logger.LogEvent("PROCESS_KILL", c.Locals("username").(string), fmt.Sprintf("PID: %d, Name: %s", req.PID, req.Name))
		if err := processes.KillProcess(req.PID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "Process killed successfully"})
	})

	api.Get("/alerts/history", func(c *fiber.Ctx) error {
		alerts, err := db.GetAlertHistory()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(alerts)
	})

	api.Post("/alerts/test", func(c *fiber.Ctx) error {
		db.RecordAlert("test", "This is a test alert triggered by user")
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Alert Rules API
	api.Get("/alerts/rules", func(c *fiber.Ctx) error {
		rules, err := db.GetAlertRules()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(rules)
	})

	api.Post("/alerts/rules", func(c *fiber.Ctx) error {
		var rule db.AlertRule
		if err := c.BodyParser(&rule); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		rule.ID = fmt.Sprintf("rule_%d", time.Now().UnixNano())
		rule.CreatedAt = time.Now()
		rule.UpdatedAt = time.Now()

		if err := db.CreateAlertRule(rule); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Reload alert rules cache
		alerts.ReloadAlertRules()

		// Broadcast rule creation to WebSocket clients
		ws.BroadcastRuleUpdate(rule)

		return c.JSON(rule)
	})

	api.Put("/alerts/rules/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var rule db.AlertRule
		if err := c.BodyParser(&rule); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		// Get existing rule to preserve CreatedAt
		existingRules, err := db.GetAlertRules()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		var existingRule *db.AlertRule
		for _, r := range existingRules {
			if r.ID == id {
				existingRule = &r
				break
			}
		}

		if existingRule == nil {
			return c.Status(404).JSON(fiber.Map{"error": "Rule not found"})
		}

		rule.ID = id
		rule.CreatedAt = existingRule.CreatedAt
		rule.UpdatedAt = time.Now()

		if err := db.UpdateAlertRule(rule); err != nil {
			log.Printf("[API] Failed to update alert rule %s: %v", id, err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Reload alert rules cache
		alerts.ReloadAlertRules()

		// Broadcast rule update to WebSocket clients
		ws.BroadcastRuleUpdate(rule)

		return c.JSON(rule)
	})

	api.Delete("/alerts/rules/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := db.DeleteAlertRule(id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Reload alert rules cache
		alerts.ReloadAlertRules()

		return c.JSON(fiber.Map{"status": "deleted"})
	})

	api.Post("/alerts/rules/:id/toggle", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var req struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		// Get current rule
		rules, err := db.GetAlertRules()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		var rule *db.AlertRule
		for _, r := range rules {
			if r.ID == id {
				rule = &r
				break
			}
		}

		if rule == nil {
			return c.Status(404).JSON(fiber.Map{"error": "Rule not found"})
		}

		rule.Enabled = req.Enabled
		rule.UpdatedAt = time.Now()

		if err := db.UpdateAlertRule(*rule); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Reload alert rules cache
		alerts.ReloadAlertRules()

		return c.JSON(fiber.Map{"status": "updated"})
	})

	api.Post("/alerts/:id/resolve", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid alert ID"})
		}

		if err := db.ResolveAlert(id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Broadcast alert resolution to WebSocket clients
		ws.BroadcastAlertResolved(id)

		return c.JSON(fiber.Map{"status": "resolved"})
	})

	// Settings API
	api.Post("/settings/password", func(c *fiber.Ctx) error {
		type PasswordReq struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		var req PasswordReq
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		username := c.Locals("username").(string)
		if !auth.Login(username, req.CurrentPassword) {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid current password"})
		}

		logger.LogEvent("PASSWORD_CHANGE", username, "User changed password")
		if err := auth.UpdatePassword(username, req.NewPassword); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update password"})
		}

		return c.JSON(fiber.Map{"status": "success"})
	})

	api.Post("/settings/app", func(c *fiber.Ctx) error {
		appName := c.FormValue("app_name")
		copyrightText := c.FormValue("copyright_text")
		logoType := c.FormValue("logo_type")

		logger.LogEvent("SETTINGS_UPDATE", c.Locals("username").(string), fmt.Sprintf("App: %s, Logo: %s", appName, logoType))

		// Handle logo upload
		if logoFile, err := c.FormFile("logo"); err == nil {
			logoPath := "./public/images/logo." + strings.Split(logoFile.Filename, ".")[1]
			if err := c.SaveFile(logoFile, logoPath); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to save logo"})
			}
			logger.LogEvent("LOGO_UPLOAD", c.Locals("username").(string), logoFile.Filename)
		}

		// Handle favicon upload
		if faviconFile, err := c.FormFile("favicon"); err == nil {
			faviconPath := "./public/images/favicon." + strings.Split(faviconFile.Filename, ".")[1]
			if err := c.SaveFile(faviconFile, faviconPath); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to save favicon"})
			}
			logger.LogEvent("FAVICON_UPLOAD", c.Locals("username").(string), faviconFile.Filename)
		}

		// Save settings to database
		if err := db.SaveAppSettings(appName, copyrightText, logoType); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save settings"})
		}

		return c.JSON(fiber.Map{"status": "success"})
	})

	api.Get("/settings/app", func(c *fiber.Ctx) error {
		settings, err := db.GetAppSettings()
		if err != nil {
			return c.JSON(fiber.Map{
				"app_name":       "Logmojo Monitor",
				"copyright_text": "© 2024 Logmojo",
				"logo_type":      "text",
			})
		}
		return c.JSON(settings)
	})

	api.Get("/system/info", func(c *fiber.Ctx) error {
		m, err := metrics.GetHostMetrics()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"os":        "macOS",
			"platform":  "darwin",
			"arch":      "amd64",
			"hostname":  "localhost",
			"uptime":    m.Uptime,
			"version":   version.Version,
			"commit":    version.Commit,
			"buildDate": version.BuildDate,
		})
	})

	// Services API
	api.Get("/services", func(c *fiber.Ctx) error {
		serviceList, err := services.GetAllServices()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(serviceList)
	})

	api.Post("/services/:action", func(c *fiber.Ctx) error {
		action := c.Params("action")
		type ServiceReq struct {
			ServiceName string `json:"service_name"`
		}
		var req ServiceReq
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		logger.LogEvent("SERVICE_"+strings.ToUpper(action), c.Locals("username").(string), fmt.Sprintf("Service: %s", req.ServiceName))

		var err error
		switch action {
		case "start":
			err = services.StartService(req.ServiceName)
		case "stop":
			err = services.StopService(req.ServiceName)
		case "restart":
			err = services.RestartService(req.ServiceName)
		case "enable":
			err = services.EnableService(req.ServiceName)
		case "disable":
			err = services.DisableService(req.ServiceName)
		default:
			return c.Status(400).JSON(fiber.Map{"error": "Invalid action"})
		}

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success", "action": action})
	})

	api.Get("/services/:service/logs", func(c *fiber.Ctx) error {
		serviceName := c.Params("service")
		lines := c.QueryInt("lines", 50)

		logs, err := services.GetServiceLogs(serviceName, lines)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"logs": logs})
	})

	// WebSocket
	app.Use("/api/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			token := c.Cookies("auth_token")
			if token == "" {
				return fiber.ErrUnauthorized
			}
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/api/ws/logs", websocket.New(ws.Handler))
	app.Get("/api/ws/metrics", websocket.New(ws.MetricsHandler))
	app.Get("/api/ws/processes", websocket.New(ws.ProcessesHandler))
	app.Get("/api/ws/alerts", websocket.New(ws.AlertsHandler))

	// Version API
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version":   version.Version,
			"commit":    version.Commit,
			"buildDate": version.BuildDate,
		})
	})

}
