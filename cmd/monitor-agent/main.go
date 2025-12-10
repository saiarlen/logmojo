package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"local-monitor/internal/alerts"
	"local-monitor/internal/api"
	"local-monitor/internal/config"
	"local-monitor/internal/db"
	"local-monitor/internal/logger"
	"local-monitor/internal/metrics"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/jet/v2"
)

func main() {
	// 1. Load Config
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Init DB
	dbPath := config.AppConfigData.Database.Path
	if dbPath == "" {
		dbPath = "monitor.db" // fallback
	}
	if err := db.Init(dbPath); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	// 2.5. Init Logger
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to init Logger: %v", err)
	}
	logger.LogEvent("SYSTEM_START", "system", "Application started")

	// 3. Start Background Tasks
	metrics.StartHistoryRecorder()
	alerts.StartAlertEngine()
	// logs.StartBackgroundIngestion() removed as we use direct file search

	// 4. Setup Fiber
	engine := jet.New("./views", ".jet.html")

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		Views:                 engine,
	})

	api.Setup(app)

	// 5. Start Server
	go func() {
		addr := config.AppConfigData.Server.ListenAddr
		fmt.Printf("Server listening on %s\n", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Shutting down...")
	_ = app.Shutdown()
}
