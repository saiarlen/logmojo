package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10MB
	maxFiles    = 10
	logDir      = "./logs"
	logFileName = "app.log"
)

var (
	appLogger *log.Logger
	logFile   *os.File
)

func Init() error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	
	logPath := filepath.Join(logDir, logFileName)
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	
	appLogger = log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags)
	return nil
}

func LogEvent(event, user, details string) {
	if appLogger == nil {
		return
	}
	
	// Check file size and rotate if needed
	rotateIfNeeded()
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("[%s] EVENT: %s | USER: %s | DETAILS: %s", timestamp, event, user, details)
	appLogger.Println(message)
}

func rotateIfNeeded() {
	logPath := filepath.Join(logDir, logFileName)
	
	info, err := os.Stat(logPath)
	if err != nil || info.Size() < maxFileSize {
		return
	}
	
	// Rotate files
	for i := maxFiles - 1; i > 0; i-- {
		oldPath := filepath.Join(logDir, fmt.Sprintf("%s.%d", logFileName, i))
		newPath := filepath.Join(logDir, fmt.Sprintf("%s.%d", logFileName, i+1))
		
		if i == maxFiles-1 {
			os.Remove(newPath) // Remove oldest file
		}
		
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}
	
	// Move current log to .1
	rotatedPath := filepath.Join(logDir, fmt.Sprintf("%s.1", logFileName))
	os.Rename(logPath, rotatedPath)
	
	// Create new log file
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	
	appLogger.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

func GetLogFiles() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(logDir, "app.log*"))
	if err != nil {
		return nil, err
	}
	
	sort.Slice(files, func(i, j int) bool {
		// Sort by modification time, newest first
		info1, _ := os.Stat(files[i])
		info2, _ := os.Stat(files[j])
		return info1.ModTime().After(info2.ModTime())
	})
	
	// Return just filenames
	var result []string
	for _, file := range files {
		result = append(result, filepath.Base(file))
	}
	
	return result, nil
}

func ReadLogFile(filename string) (string, error) {
	if !strings.HasPrefix(filename, "app.log") {
		return "", fmt.Errorf("invalid log file")
	}
	
	logPath := filepath.Join(logDir, filename)
	content, err := os.ReadFile(logPath)
	if err != nil {
		return "", err
	}
	
	return string(content), nil
}

func GetLogWriter() io.Writer {
	if logFile != nil {
		return io.MultiWriter(os.Stdout, logFile)
	}
	return os.Stdout
}