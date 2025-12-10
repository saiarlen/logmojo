package api

import (
	"local-monitor/internal/auth"
	"local-monitor/internal/config"
	"local-monitor/internal/db"
	"local-monitor/internal/logs"
	"local-monitor/internal/metrics"
	"local-monitor/internal/processes"
	"local-monitor/internal/services"
	"local-monitor/internal/ws"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func Setup(app *fiber.App) {
	app.Use(logger.New())
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
		return c.Render("dashboard", fiber.Map{
			"current_page": "dashboard",
		})
	})

	app.Get("/logs", func(c *fiber.Ctx) error {
		return c.Render("logs", fiber.Map{
			"current_page": "logs",
		})
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
		return c.Render("processes", fiber.Map{
			"current_page": "processes",
		})
	})

	app.Get("/services", func(c *fiber.Ctx) error {
		return c.Render("services", fiber.Map{
			"current_page": "services",
		})
	})

	app.Get("/alerts", func(c *fiber.Ctx) error {
		return c.Render("alerts", fiber.Map{
			"current_page": "alerts",
		})
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
			PID int32 `json:"pid"`
		}
		var req KillReq
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}
		if err := processes.KillProcess(req.PID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "Process killed successfully"})
	})

	api.Get("/alerts/history", func(c *fiber.Ctx) error {
		rows, err := db.DB.Query("SELECT id, timestamp, type, message, resolved FROM alerts ORDER BY timestamp DESC LIMIT 50")
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		defer rows.Close()

		var history []map[string]interface{}
		for rows.Next() {
			var id int
			var ts time.Time
			var typ, msg string
			var res bool
			rows.Scan(&id, &ts, &typ, &msg, &res)
			history = append(history, map[string]interface{}{
				"id":        id,
				"timestamp": ts,
				"type":      typ,
				"message":   msg,
				"resolved":  res,
			})
		}
		return c.JSON(history)
	})

	api.Post("/alerts/test", func(c *fiber.Ctx) error {
		db.RecordAlert("test", "This is a test alert triggered by user")
		return c.JSON(fiber.Map{"status": "ok"})
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
}
