package version

import (
	"logmojo/internal/config"
	"os/exec"
	"strings"
	"time"
)

var (
	Version   string
	Commit    string
	BuildDate string
)

func Init() {
	// Get version from config
	Version = config.AppConfigData.General.Version
	Commit = "none"
	BuildDate = "unknown"

	// If built locally (no ldflags), try reading git info
	if Version == "dev" {
		if v, err := gitDescribe(); err == nil {
			Version = v
		}
	}

	if Commit == "none" {
		if c, err := gitCommit(); err == nil {
			Commit = c
		}
	}

	if BuildDate == "unknown" {
		if d, err := gitCommitDate(); err == nil {
			BuildDate = d
		} else {
			// fallback to local time if git unavailable
			BuildDate = time.Now().UTC().Format(time.RFC3339)
		}
	}
}

func gitDescribe() (string, error) {
	out, err := exec.Command("git", "describe", "--tags", "--always").Output()
	if err != nil {
		return "", err
	}
	version := strings.TrimSpace(string(out))
	// Remove -dirty suffix for cleaner version display
	if strings.HasSuffix(version, "-dirty") {
		version = strings.TrimSuffix(version, "-dirty")
	}
	return version, nil
}

func gitCommit() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitCommitDate() (string, error) {
	out, err := exec.Command("git", "log", "-1", "--format=%cI").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
