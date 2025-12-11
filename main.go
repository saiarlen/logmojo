package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"logmojo/internal/alerts"
	"logmojo/internal/api"
	"logmojo/internal/config"
	"logmojo/internal/db"
	"logmojo/internal/logger"
	"logmojo/internal/metrics"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/jet/v2"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Parse command line flags
	userAction := flag.String("user", "", "User management: create, delete, update, list")
	username := flag.String("username", "", "Username for user operations")
	password := flag.String("password", "", "Password for user operations")
	dbPath := flag.String("db", "", "Database path (optional)")
	flag.Parse()

	// Handle user management commands
	if *userAction != "" {
		handleUserCommand(*userAction, *username, *password, *dbPath)
		return
	}

	// Normal server startup
	startServer()
}

func handleUserCommand(action, username, password, dbPath string) {
	// Load config first
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use provided db path or config default
	if dbPath == "" {
		dbPath = config.AppConfigData.Database.Path
		if dbPath == "" {
			dbPath = "monitor.db"
		}
	}

	if err := db.Init(dbPath); err != nil {
		fmt.Printf("❌ Failed to open database: %v\n", err)
		os.Exit(1)
	}

	switch action {
	case "create":
		if username == "" || password == "" {
			fmt.Println("❌ Usage: --user=create --username=USERNAME --password=PASSWORD")
			os.Exit(1)
		}
		createUser(username, password)

	case "delete":
		if username == "" {
			fmt.Println("❌ Usage: --user=delete --username=USERNAME")
			os.Exit(1)
		}
		deleteUser(username)

	case "update":
		if username == "" || password == "" {
			fmt.Println("❌ Usage: --user=update --username=USERNAME --password=NEW_PASSWORD")
			os.Exit(1)
		}
		updatePassword(username, password)

	case "list":
		listUsers()

	default:
		fmt.Println("Logmojo - User Management")
		fmt.Println("\nUsage:")
		fmt.Println("  --user=create --username=USERNAME --password=PASSWORD")
		fmt.Println("  --user=delete --username=USERNAME")
		fmt.Println("  --user=update --username=USERNAME --password=NEW_PASSWORD")
		fmt.Println("  --user=list")
		fmt.Println("\nOptions:")
		fmt.Println("  --db=PATH    Database file path (optional)")
		os.Exit(1)
	}
}

func startServer() {
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

func createUser(username, password string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("❌ Failed to hash password: %v\n", err)
		os.Exit(1)
	}

	if err := db.CreateUser(username, string(hash)); err != nil {
		fmt.Printf("❌ Failed to create user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ User '%s' created successfully\n", username)
}

func deleteUser(username string) {
	result, err := db.DB.Exec("DELETE FROM users WHERE username = ?", username)
	if err != nil {
		fmt.Printf("❌ Failed to delete user: %v\n", err)
		os.Exit(1)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		fmt.Printf("⚠️  User '%s' not found\n", username)
		os.Exit(1)
	}

	fmt.Printf("✅ User '%s' deleted successfully\n", username)
}

func updatePassword(username, newPassword string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("❌ Failed to hash password: %v\n", err)
		os.Exit(1)
	}

	result, err := db.DB.Exec("UPDATE users SET password_hash = ? WHERE username = ?", string(hash), username)
	if err != nil {
		fmt.Printf("❌ Failed to update password: %v\n", err)
		os.Exit(1)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		fmt.Printf("⚠️  User '%s' not found\n", username)
		os.Exit(1)
	}

	fmt.Printf("✅ Password updated for user '%s'\n", username)
}

func listUsers() {
	rows, err := db.DB.Query("SELECT username FROM users ORDER BY username")
	if err != nil {
		fmt.Printf("❌ Failed to list users: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("📋 Users:")
	count := 0
	for rows.Next() {
		var username string
		rows.Scan(&username)
		fmt.Printf("  - %s\n", username)
		count++
	}

	if count == 0 {
		fmt.Println("  (no users found)")
	}
}
