package logs

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"logmojo/internal/config"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nxadm/tail"
)

// LogResult represents a search result
type LogResult struct {
	App       string    `json:"app"`
	File      string    `json:"file"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Regex for common log levels (both bracketed and unbracketed)
var levelRegex = regexp.MustCompile(`(?i)\[?(INFO|WARN|ERROR|DEBUG|FATAL|TRACE)\]?`)

// Comprehensive timestamp patterns for various log formats
var timestampPatterns = []struct {
	regex   *regexp.Regexp
	layouts []string
}{
	// ISO 8601 and RFC 3339 formats
	{regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{3,9})?(?:Z|[+-]\d{2}:?\d{2})?`), []string{
		"2006-01-02T15:04:05Z07:00", "2006-01-02T15:04:05.000Z07:00", "2006-01-02T15:04:05", "2006-01-02T15:04:05.000",
	}},
	// Standard date-time formats
	{regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d{3,9})?`), []string{
		"2006-01-02 15:04:05.000", "2006-01-02 15:04:05",
	}},
	// US date formats
	{regexp.MustCompile(`\d{1,2}/\d{1,2}/\d{4} \d{1,2}:\d{2}:\d{2}(?: [AP]M)?`), []string{
		"1/2/2006 3:04:05 PM", "01/02/2006 15:04:05", "1/2/2006 15:04:05",
	}},
	// European date formats
	{regexp.MustCompile(`\d{1,2}\.\d{1,2}\.\d{4} \d{2}:\d{2}:\d{2}`), []string{
		"02.01.2006 15:04:05", "2.1.2006 15:04:05",
	}},
	// Nginx/Apache log format
	{regexp.MustCompile(`\[\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}\]`), []string{
		"[02/Jan/2006:15:04:05 -0700]",
	}},
	// Syslog format (with year)
	{regexp.MustCompile(`\w{3}\s+\d{1,2}\s+\d{4}\s+\d{2}:\d{2}:\d{2}`), []string{
		"Jan 2 2006 15:04:05", "Jan  2 2006 15:04:05",
	}},
	// Syslog format (current year)
	{regexp.MustCompile(`\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}`), []string{
		"Jan 2 15:04:05", "Jan  2 15:04:05",
	}},
	// macOS full timestamp format (Tue Dec  9 00:47:11.259)
	{regexp.MustCompile(`\w{3}\s+\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}\.\d{3}`), []string{
		"Mon Jan 2 15:04:05.000", "Mon Jan  2 15:04:05.000",
	}},
	// macOS system log format truncated (Tue .682)
	{regexp.MustCompile(`\w{3}\s+\.\d{3}`), []string{"macos_syslog"}},
	// Unix timestamp (seconds)
	{regexp.MustCompile(`\b1[0-9]{9}\b`), []string{"unix"}},
	// Unix timestamp (milliseconds)
	{regexp.MustCompile(`\b1[0-9]{12}\b`), []string{"unix_ms"}},
	// Time only formats
	{regexp.MustCompile(`\[?\d{2}:\d{2}:\d{2}(?:\.\d{3})?\]?`), []string{
		"[15:04:05.000]", "[15:04:05]", "15:04:05.000", "15:04:05",
	}},
}

func parseLevel(line string) string {
	matches := levelRegex.FindStringSubmatch(line)
	if len(matches) > 1 {
		return strings.ToUpper(matches[1]) // Return the captured group without brackets
	}
	return "INFO" // Default
}

func parseTimestamp(line string) (time.Time, string) {
	// First, handle macOS syslog formats directly (with optional leading space from grep)
	// Handle truncated format: Tue .007
	macOSTruncated := regexp.MustCompile(`^\s*\w{3}\s+\.\d{3}\s+`)
	if macOSTruncated.MatchString(line) {
		now := time.Now()
		cleanedMessage := macOSTruncated.ReplaceAllString(line, "")
		cleanedMessage = strings.TrimSpace(cleanedMessage)
		return now, cleanedMessage
	}

	// Handle full format: Tue Dec  9 00:47:11.259
	macOSFull := regexp.MustCompile(`^\s*\w{3}\s+\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}\.\d{3}\s+`)
	if macOSFull.MatchString(line) {
		now := time.Now()
		cleanedMessage := macOSFull.ReplaceAllString(line, "")
		cleanedMessage = strings.TrimSpace(cleanedMessage)
		return now, cleanedMessage
	}

	// Try each timestamp pattern and return both timestamp and cleaned message
	for _, pattern := range timestampPatterns {
		if loc := pattern.regex.FindStringIndex(line); loc != nil {
			tsStr := strings.TrimSpace(line[loc[0]:loc[1]])

			// Handle Unix timestamps
			for _, layout := range pattern.layouts {
				var parsedTime time.Time
				var err error

				if layout == "unix" {
					if ts, parseErr := strconv.ParseInt(tsStr, 10, 64); parseErr == nil {
						parsedTime = time.Unix(ts, 0)
					}
				} else if layout == "unix_ms" {
					if ts, parseErr := strconv.ParseInt(tsStr, 10, 64); parseErr == nil {
						parsedTime = time.Unix(ts/1000, (ts%1000)*1000000)
					}
				} else {
					parsedTime, err = time.Parse(layout, tsStr)
				}

				if err == nil && !parsedTime.IsZero() {
					// Handle year-less formats (syslog)
					if parsedTime.Year() == 0 {
						now := time.Now()
						parsedTime = time.Date(now.Year(), parsedTime.Month(), parsedTime.Day(), parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), parsedTime.Nanosecond(), now.Location())
					}
					// Handle time-only formats
					if parsedTime.Year() == 0 && parsedTime.Month() == 1 && parsedTime.Day() == 1 {
						now := time.Now()
						parsedTime = time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), parsedTime.Nanosecond(), now.Location())
					}

					// Remove the timestamp from the message and clean up
					cleanedMessage := strings.TrimSpace(line[:loc[0]] + line[loc[1]:])

					// Remove duplicate log levels like [INFO] or INFO
					cleanedMessage = regexp.MustCompile(`^\s*\[?(INFO|WARN|ERROR|DEBUG|FATAL|TRACE)\]?\s*`).ReplaceAllString(cleanedMessage, "")
					cleanedMessage = strings.TrimSpace(cleanedMessage)
					return parsedTime, cleanedMessage
				}
			}
		}
	}

	// If no timestamp found, return zero time and original message
	return time.Time{}, line
}

// Search searches logs using grep/zgrep
func Search(query, appFilter, logFilter, specificFile, levelFilter string, limit int) ([]LogResult, error) {
	var filePaths []string
	pathMap := make(map[string]string) // path -> app name

	// 1. Resolve Files
	if specificFile != "" {
		filePaths = []string{specificFile}
		pathMap[specificFile] = appFilter
	} else {
		// Find all allowed files based on filters
		for _, app := range config.AppConfigData.Apps {
			if appFilter != "" && app.Name != appFilter {
				continue
			}
			for _, l := range app.Logs {
				if logFilter != "" && l.Name != logFilter {
					continue
				}

				// Resolve actual files for this entry
				files, _ := ListFiles(app.Name, l.Name)
				for _, f := range files {
					// Add all files (including archives) for search
					filePaths = append(filePaths, f.Path)
					pathMap[f.Path] = app.Name
				}
			}
		}
	}

	if len(filePaths) == 0 {
		log.Printf("[LOGS] No files found for app=%s, log=%s", appFilter, logFilter)
		return []LogResult{}, nil
	}

	log.Printf("[LOGS] Searching %d files for query=%s, level=%s", len(filePaths), query, levelFilter)

	// 2. Optimize file selection for performance
	// Limit to most recent files if too many
	if len(filePaths) > 5 {
		log.Printf("[LOGS] Too many files (%d), limiting to 3 most recent", len(filePaths))
		// Get file info and sort by modification time
		type fileInfo struct {
			path    string
			modTime time.Time
		}
		var fileInfos []fileInfo
		for _, p := range filePaths {
			if info, err := os.Stat(p); err == nil {
				fileInfos = append(fileInfos, fileInfo{p, info.ModTime()})
			}
		}
		// Sort by modification time (newest first)
		sort.Slice(fileInfos, func(i, j int) bool {
			return fileInfos[i].modTime.After(fileInfos[j].modTime)
		})
		// Keep only 3 most recent
		filePaths = nil
		for i := 0; i < 3 && i < len(fileInfos); i++ {
			filePaths = append(filePaths, fileInfos[i].path)
		}
	}

	// 3. Build Command - select appropriate grep tool
	cmdName := "grep"
	hasCompressed := false

	// Check for compressed files and select appropriate tool
	for _, p := range filePaths {
		lower := strings.ToLower(p)
		if strings.HasSuffix(lower, ".gz") {
			cmdName = "zgrep"
			hasCompressed = true
			break
		} else if strings.HasSuffix(lower, ".bz2") {
			cmdName = "bzgrep"
			hasCompressed = true
			break
		} else if strings.HasSuffix(lower, ".xz") {
			cmdName = "xzgrep"
			hasCompressed = true
			break
		} else if strings.HasSuffix(lower, ".lz4") {
			cmdName = "lz4grep"
			hasCompressed = true
			break
		}
	}

	// Fallback to zgrep if we have mixed compressed files
	if hasCompressed && cmdName == "grep" {
		cmdName = "zgrep"
	}

	// 4. Build optimized grep arguments
	args := []string{"-iH", "--text", "-m", "1000"} // Limit to 1000 matches per file

	// Add pattern
	if query == "" {
		args = append(args, ".")
	} else {
		args = append(args, "-E", query)
	}

	args = append(args, filePaths...)

	cmd := exec.Command(cmdName, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C") // Set locale for consistent behavior
	log.Printf("[LOGS] Executing: %s with %d files", cmdName, len(filePaths))

	// Check if the command exists, fallback to grep if not
	if _, err := exec.LookPath(cmdName); err != nil && cmdName != "grep" {
		log.Printf("[LOGS] %s not found, falling back to grep", cmdName)
		// Filter out compressed files that grep can't handle
		var compatibleFiles []string
		for _, p := range filePaths {
			lower := strings.ToLower(p)
			if !strings.HasSuffix(lower, ".gz") && !strings.HasSuffix(lower, ".bz2") &&
				!strings.HasSuffix(lower, ".xz") && !strings.HasSuffix(lower, ".lz4") {
				compatibleFiles = append(compatibleFiles, p)
			}
		}
		if len(compatibleFiles) > 0 {
			// Rebuild args with compatible files
			newArgs := make([]string, 0, len(args)-len(filePaths)+len(compatibleFiles))
			newArgs = append(newArgs, args[:len(args)-len(filePaths)]...)
			newArgs = append(newArgs, compatibleFiles...)
			cmd = exec.Command("grep", newArgs...)
			log.Printf("[LOGS] Using grep with %d compatible files", len(compatibleFiles))
		} else {
			return []LogResult{}, fmt.Errorf("no compatible files found for search")
		}
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[LOGS] StdoutPipe error: %v", err)
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[LOGS] Command start error: %v", err)
		return nil, err
	}

	// Add shorter timeout for better UX
	go func() {
		time.Sleep(10 * time.Second)
		if cmd.Process != nil {
			log.Printf("[LOGS] Command timeout after 10s, killing process")
			cmd.Process.Kill()
		}
	}()

	scanner := bufio.NewScanner(stdout)

	// Increase buffer size
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var collected []LogResult
	maxScan := 2000 // Reduced for faster response
	count := 0

	for scanner.Scan() {
		if count >= maxScan {
			break
		}

		line := scanner.Text()
		count++

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		path := parts[0]
		content := parts[1]

		// Find app name from path
		appName := pathMap[path]
		if appName == "" {
			// Try to match by checking all paths
			for p, a := range pathMap {
				if strings.Contains(path, p) || strings.Contains(p, path) {
					appName = a
					break
				}
			}
		}

		lvl := parseLevel(content)
		if levelFilter != "" && lvl != levelFilter {
			continue
		}

		ts, cleanedContent := parseTimestamp(content)
		if ts.IsZero() {
			ts = time.Now().AddDate(-1, 0, 0) // Use old date for sorting
			cleanedContent = content
		}

		// Debug: log the original content and cleaned content
		if strings.Contains(content, "Tue .") {
			log.Printf("[DEBUG] Original: %s", content)
			log.Printf("[DEBUG] Cleaned: %s", cleanedContent)
		}

		collected = append(collected, LogResult{
			App:       appName,
			File:      path,
			Level:     lvl,
			Message:   cleanedContent,
			Timestamp: ts,
		})
	}

	cmd.Wait()

	log.Printf("[LOGS] Collected %d results before filtering", len(collected))

	// Sort logs by timestamp (Newest first)
	sort.Slice(collected, func(i, j int) bool {
		return collected[i].Timestamp.After(collected[j].Timestamp)
	})

	if len(collected) > limit {
		collected = collected[:limit]
	}

	log.Printf("[LOGS] Returning %d results", len(collected))
	return collected, nil
}

// StreamLog tails a file and sends lines to a channel (for live view)
func StreamLog(ctx context.Context, path string, out chan<- string) error {
	t, err := tail.TailFile(path, tail.Config{
		Follow:   true,
		ReOpen:   true,
		Poll:     true,
		Location: &tail.SeekInfo{Offset: 0, Whence: 2}, // Start at end
	})
	if err != nil {
		fmt.Printf("tail error: %v\n", err)
		return err
	}

	go func() {
		<-ctx.Done()
		t.Stop()
	}()

	for line := range t.Lines {
		select {
		case <-ctx.Done():
			return nil
		case out <- line.Text:
		}
	}
	return nil
}
