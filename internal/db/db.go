package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(path string) error {
	var err error
	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	return migrate()
}

func migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS disk_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME,
			used_percent REAL
		);`,
		`CREATE TABLE IF NOT EXISTS cpu_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME,
			usage_percent REAL
		);`,
		`CREATE TABLE IF NOT EXISTS ram_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME,
			usage_percent REAL,
			used_bytes INTEGER
		);`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME,
			type TEXT,
			message TEXT,
			resolved BOOLEAN DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS alert_rules (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL,
			condition TEXT,
			threshold REAL DEFAULT 0,
			severity TEXT DEFAULT 'medium',
			enabled BOOLEAN DEFAULT 1,
			email_enabled BOOLEAN DEFAULT 0,
			log_pattern TEXT,
			app_filter TEXT,
			log_filter TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			last_triggered DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE,
			password_hash TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS processed_log_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entry_hash TEXT UNIQUE,
			processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_processed_entries_hash ON processed_log_entries(entry_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_processed_entries_time ON processed_log_entries(processed_at);`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
	}
	
	// Add new columns to existing alerts table
	migrationQueries := []string{
		`ALTER TABLE alerts ADD COLUMN rule_id TEXT;`,
		`ALTER TABLE alerts ADD COLUMN severity TEXT DEFAULT 'medium';`,
		`ALTER TABLE alerts ADD COLUMN resolved_at DATETIME;`,
	}
	
	for _, q := range migrationQueries {
		DB.Exec(q) // Ignore errors for existing columns
	}
	
	return nil
}

func RecordMetricsHistory(cpu, ramPercent float64, ramUsed uint64, diskPercent float64) {
	if DB == nil {
		return
	}
	t := time.Now()
	_, _ = DB.Exec("INSERT INTO cpu_history (timestamp, usage_percent) VALUES (?, ?)", t, cpu)
	_, _ = DB.Exec("INSERT INTO ram_history (timestamp, usage_percent, used_bytes) VALUES (?, ?, ?)", t, ramPercent, ramUsed)
	_, _ = DB.Exec("INSERT INTO disk_history (timestamp, used_percent) VALUES (?, ?)", t, diskPercent)

	// Cleanup old records (keep last 24h)
	cutoff := t.Add(-24 * time.Hour)
	_, _ = DB.Exec("DELETE FROM cpu_history WHERE timestamp < ?", cutoff)
	_, _ = DB.Exec("DELETE FROM ram_history WHERE timestamp < ?", cutoff)
	_, _ = DB.Exec("DELETE FROM disk_history WHERE timestamp < ?", cutoff)
}

func RecordAlert(alertType, message string) {
	if DB == nil {
		return
	}
	_, err := DB.Exec("INSERT INTO alerts (timestamp, type, message) VALUES (?, ?, ?)", time.Now(), alertType, message)
	if err != nil {
		log.Println("Error recording alert:", err)
	}
}

func GetUser(username string) (string, error) {
	var hash string
	err := DB.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&hash)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CreateUser(username, hash string) error {
	_, err := DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, hash)
	return err
}

func UserExists() bool {
	var count int
	_ = DB.QueryRow("SELECT count(*) FROM users").Scan(&count)
	return count > 0
}
