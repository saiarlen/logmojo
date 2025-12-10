package db

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// HashLogEntry creates a short hash for log entry identification
func HashLogEntry(file, message string, timestamp int64) string {
	data := fmt.Sprintf("%s:%s:%d", file, message, timestamp)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes (16 chars)
}

// IsEntryProcessed checks if a log entry has been processed before
func IsEntryProcessed(entryHash string) bool {
	if DB == nil {
		return false
	}
	
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM processed_log_entries WHERE entry_hash = ?", entryHash).Scan(&count)
	return err == nil && count > 0
}

// MarkEntryProcessed marks a log entry as processed
func MarkEntryProcessed(entryHash string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	
	_, err := DB.Exec("INSERT OR IGNORE INTO processed_log_entries (entry_hash) VALUES (?)", entryHash)
	return err
}

// CleanupOldProcessedEntries removes entries older than specified hours
func CleanupOldProcessedEntries(hoursOld int) (int64, error) {
	if DB == nil {
		return 0, fmt.Errorf("database not initialized")
	}
	
	cutoff := time.Now().Add(-time.Duration(hoursOld) * time.Hour)
	result, err := DB.Exec("DELETE FROM processed_log_entries WHERE processed_at < ?", cutoff)
	if err != nil {
		return 0, err
	}
	
	return result.RowsAffected()
}

// GetProcessedEntriesCount returns the total number of processed entries
func GetProcessedEntriesCount() (int, error) {
	if DB == nil {
		return 0, fmt.Errorf("database not initialized")
	}
	
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM processed_log_entries").Scan(&count)
	return count, err
}