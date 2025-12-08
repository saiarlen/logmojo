package logs

import (
	"bufio"
	"context"
	"fmt"
	"local-monitor/internal/config"
	"log"
	"os/exec"
	"regexp"
	"sort"
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

// Regex for common log levels
var levelRegex = regexp.MustCompile(`(?i)(INFO|WARN|ERROR|DEBUG|FATAL|TRACE)`)

// Common timestamp formats
var timeRegexes = []*regexp.Regexp{
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`), // YYYY-MM-DD HH:MM:SS
	regexp.MustCompile(`^\w{3} \s+\d{1,2} \d{2}:\d{2}:\d{2}`),  // Mon Jan 02 15:04:05
	regexp.MustCompile(`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`), // YYYY/MM/DD HH:MM:SS
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`), // ISO8601
}

func parseLevel(line string) string {
	loc := levelRegex.FindStringIndex(line)
	if loc != nil {
		return strings.ToUpper(line[loc[0]:loc[1]])
	}
	return "INFO" // Default
}

func parseTimestamp(line string) time.Time {
	// Try to match common formats
	for _, re := range timeRegexes {
		if loc := re.FindStringIndex(line); loc != nil {
			tsStr := line[loc[0]:loc[1]]
			layouts := []string{
				"2006-01-02 15:04:05",
				"2006/01/02 15:04:05",
				"2006-01-02T15:04:05",
				"Jan 2 15:04:05",
				"Jan  2 15:04:05",
				time.Stamp,
			}
			for _, layout := range layouts {
				if t, err := time.Parse(layout, tsStr); err == nil {
					if t.Year() == 0 {
						t = t.AddDate(time.Now().Year(), 0, 0)
					}
					return t
				}
			}
		}
	}
	return time.Time{}
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
					// Always add non-archive files, add archives only if specific file requested
					if !f.IsArchive {
						filePaths = append(filePaths, f.Path)
						pathMap[f.Path] = app.Name
					}
				}
			}
		}
	}

	if len(filePaths) == 0 {
		log.Printf("[LOGS] No files found for app=%s, log=%s", appFilter, logFilter)
		return []LogResult{}, nil
	}

	log.Printf("[LOGS] Searching %d files for query=%s, level=%s", len(filePaths), query, levelFilter)

	// 2. Build Command - use grep for all files
	cmdName := "grep"
	args := []string{"-iH", "--text"} // -H: always print filename, --text: treat as text

	// Add pattern
	if query == "" {
		args = append(args, ".")
	} else {
		args = append(args, "-E", query)
	}

	args = append(args, filePaths...)

	cmd := exec.Command(cmdName, args...)
	log.Printf("[LOGS] Executing: %s %v", cmdName, args)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[LOGS] StdoutPipe error: %v", err)
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[LOGS] Command start error: %v", err)
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)

	// Increase buffer size
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var collected []LogResult
	maxScan := 5000
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

		ts := parseTimestamp(content)
		if ts.IsZero() {
			ts = time.Now().AddDate(-1, 0, 0) // Use old date for sorting
		}

		collected = append(collected, LogResult{
			App:       appName,
			File:      path,
			Level:     lvl,
			Message:   content,
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
