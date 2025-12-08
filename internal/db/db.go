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
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE,
			password_hash TEXT
		);`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
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
