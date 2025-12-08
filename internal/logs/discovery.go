package logs

import (
	"fmt"
	"local-monitor/internal/config"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type LogFile struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"mod_time"`
	IsArchive bool      `json:"is_archive"`
}

// ListFiles returns all log files associated with a specific configured log entry
func ListFiles(appName, logName string) ([]LogFile, error) {
	var targetPath string
	found := false

	// Find config
	for _, app := range config.AppConfigData.Apps {
		if app.Name == appName {
			for _, l := range app.Logs {
				if l.Name == logName {
					targetPath = l.Path
					found = true
					break
				}
			}
		}
	}

	if !found {
		log.Printf("[DISCOVERY] Config not found for app=%s, log=%s", appName, logName)
		return nil, fmt.Errorf("log configuration not found")
	}

	log.Printf("[DISCOVERY] Found path: %s for app=%s, log=%s", targetPath, appName, logName)

	info, err := os.Stat(targetPath)
	if os.IsNotExist(err) {
		// File might not exist yet, but maybe archives do?
		// If explicitly a directory path that doesn't exist, return empty
		return []LogFile{}, nil
	}
	if err != nil {
		return nil, err
	}

	var files []LogFile

	if info.IsDir() {
		entries, err := os.ReadDir(targetPath)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			fInfo, _ := e.Info()
			fullPath := filepath.Join(targetPath, e.Name())
			name := strings.ToLower(e.Name())
			
			if !strings.Contains(name, ".log") &&
				!strings.HasSuffix(name, ".txt") &&
				!strings.HasSuffix(name, ".gz") &&
				!strings.HasSuffix(name, ".out") &&
				!strings.HasSuffix(name, ".err") {
				continue
			}

			files = append(files, LogFile{
				Name:      e.Name(),
				Path:      fullPath,
				Size:      fInfo.Size(),
				ModTime:   fInfo.ModTime(),
				IsArchive: isArchive(e.Name()),
			})
		}
	} else {
		// If file, add itself
		files = append(files, LogFile{
			Name:      filepath.Base(targetPath),
			Path:      targetPath,
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			IsArchive: false,
		})

		// Look for rotated siblings
		dir := filepath.Dir(targetPath)
		base := filepath.Base(targetPath)

		siblings, _ := os.ReadDir(dir)
		for _, s := range siblings {
			if s.IsDir() || s.Name() == base {
				continue
			}
			if strings.HasPrefix(s.Name(), base) {
				sInfo, _ := s.Info()
				files = append(files, LogFile{
					Name:      s.Name(),
					Path:      filepath.Join(dir, s.Name()),
					Size:      sInfo.Size(),
					ModTime:   sInfo.ModTime(),
					IsArchive: true, // Assume siblings are archives
				})
			}
		}
	}

	// Sort by ModTime DESC (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	log.Printf("[DISCOVERY] Found %d files for app=%s, log=%s", len(files), appName, logName)
	return files, nil
}

func isArchive(name string) bool {
	return strings.HasSuffix(name, ".gz") ||
		strings.HasSuffix(name, ".zip") ||
		strings.HasSuffix(name, ".1") ||
		strings.HasSuffix(name, ".2") ||
		strings.HasSuffix(name, ".old")
}
