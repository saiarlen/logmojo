package services

import (
	"bufio"
	"fmt"
	"local-monitor/internal/config"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ServiceStatus struct {
	Name        string    `json:"name"`
	ServiceName string    `json:"service_name"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Active      bool      `json:"active"`
	Loaded      bool      `json:"loaded"`
	Uptime      string    `json:"uptime"`
	MemoryUsage string    `json:"memory_usage"`
	CPUUsage    string    `json:"cpu_usage"`
	PID         int       `json:"pid"`
	ConfigPath  string    `json:"config_path"`
	LogPath     string    `json:"log_path"`
	LastRestart time.Time `json:"last_restart"`
	Version     string    `json:"version"`
}

func GetAllServices() ([]ServiceStatus, error) {
	var services []ServiceStatus
	
	for _, app := range config.AppConfigData.Apps {
		for _, svc := range app.Services {
			if !svc.Enabled {
				continue
			}
			
			status := ServiceStatus{
				Name:        svc.Name,
				ServiceName: svc.ServiceName,
				Enabled:     svc.Enabled,
				Description: svc.Description,
				ConfigPath:  svc.ConfigPath,
				LogPath:     svc.LogPath,
			}
			
			// Get systemctl status
			if err := getServiceStatus(&status); err != nil {
				status.Status = "unknown"
				status.Active = false
				status.Loaded = false
			}
			
			// Get service version
			status.Version = getServiceVersion(svc.ServiceName)
			
			services = append(services, status)
		}
	}
	
	return services, nil
}

func getServiceStatus(status *ServiceStatus) error {
	cmd := exec.Command("systemctl", "status", status.ServiceName)
	output, err := cmd.Output()
	if err != nil {
		// Service might not exist or be inactive
		status.Status = "inactive"
		status.Active = false
		status.Loaded = false
		return nil
	}
	
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Parse Active status
		if strings.Contains(line, "Active:") {
			if strings.Contains(line, "active (running)") {
				status.Status = "running"
				status.Active = true
			} else if strings.Contains(line, "inactive") {
				status.Status = "inactive"
				status.Active = false
			} else if strings.Contains(line, "failed") {
				status.Status = "failed"
				status.Active = false
			}
		}
		
		// Parse Loaded status
		if strings.Contains(line, "Loaded:") {
			status.Loaded = !strings.Contains(line, "not-found")
		}
		
		// Parse PID
		if strings.Contains(line, "Main PID:") {
			re := regexp.MustCompile(`Main PID: (\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if pid, err := strconv.Atoi(matches[1]); err == nil {
					status.PID = pid
				}
			}
		}
	}
	
	// Get memory and CPU usage if PID is available
	if status.PID > 0 {
		getProcessMetrics(status)
	}
	
	return nil
}

func getProcessMetrics(status *ServiceStatus) {
	cmd := exec.Command("ps", "-p", strconv.Itoa(status.PID), "-o", "pid,pcpu,pmem,etime")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return
	}
	
	fields := strings.Fields(lines[1])
	if len(fields) >= 4 {
		status.CPUUsage = fields[1] + "%"
		status.MemoryUsage = fields[2] + "%"
		status.Uptime = fields[3]
	}
}



func RestartService(serviceName string) error {
	cmd := exec.Command("systemctl", "restart", serviceName)
	return cmd.Run()
}

func StartService(serviceName string) error {
	cmd := exec.Command("systemctl", "start", serviceName)
	return cmd.Run()
}

func StopService(serviceName string) error {
	cmd := exec.Command("systemctl", "stop", serviceName)
	return cmd.Run()
}

func EnableService(serviceName string) error {
	cmd := exec.Command("systemctl", "enable", serviceName)
	return cmd.Run()
}

func DisableService(serviceName string) error {
	cmd := exec.Command("systemctl", "disable", serviceName)
	return cmd.Run()
}

func GetServiceLogs(serviceName string, lines int) ([]string, error) {
	cmd := exec.Command("journalctl", "-u", serviceName, "-n", strconv.Itoa(lines), "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var logLines []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		logLines = append(logLines, scanner.Text())
	}
	
	return logLines, nil
}

func CheckConfigFile(configPath string) (bool, error) {
	if configPath == "" {
		return false, nil
	}
	
	// Handle glob patterns
	if strings.Contains(configPath, "*") {
		// For now, just check if directory exists
		dir := strings.Split(configPath, "*")[0]
		if _, err := os.Stat(dir); err != nil {
			return false, err
		}
		return true, nil
	}
	
	if _, err := os.Stat(configPath); err != nil {
		return false, err
	}
	return true, nil
}

func getServiceVersion(serviceName string) string {
	// Try common version commands for different services
	versionCommands := map[string][]string{
		"nginx":      {"nginx", "-v"},
		"apache2":    {"apache2", "-v"},
		"httpd":      {"httpd", "-v"},
		"mysql":      {"mysql", "--version"},
		"mysqld":     {"mysqld", "--version"},
		"postgresql": {"postgres", "--version"},
		"postgres":   {"postgres", "--version"},
		"redis":      {"redis-server", "--version"},
		"docker":     {"docker", "--version"},
		"ssh":        {"ssh", "-V"},
		"sshd":       {"sshd", "-V"},
	}
	
	// Try service-specific version command
	if cmd, exists := versionCommands[serviceName]; exists {
		if version := tryVersionCommand(cmd); version != "" {
			return version
		}
	}
	
	// Try generic version commands
	genericCommands := [][]string{
		{serviceName, "--version"},
		{serviceName, "-v"},
		{serviceName, "version"},
	}
	
	for _, cmd := range genericCommands {
		if version := tryVersionCommand(cmd); version != "" {
			return version
		}
	}
	
	return "N/A"
}

func tryVersionCommand(cmdArgs []string) string {
	if len(cmdArgs) < 2 {
		return ""
	}
	
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	
	outputStr := strings.TrimSpace(string(output))
	lines := strings.Split(outputStr, "\n")
	
	// Extract version from first line (most common)
	if len(lines) > 0 {
		line := lines[0]
		// Common version patterns
		versionPatterns := []*regexp.Regexp{
			regexp.MustCompile(`version\s+([\d\.]+[\w\-\.]*)`),
			regexp.MustCompile(`([\d]+\.[\d]+[\w\-\.]*)`),
			regexp.MustCompile(`v([\d\.]+[\w\-\.]*)`),
		}
		
		for _, pattern := range versionPatterns {
			if matches := pattern.FindStringSubmatch(strings.ToLower(line)); len(matches) > 1 {
				return matches[1]
			}
		}
		
		// If no pattern matches, return first 50 chars of output
		if len(line) > 50 {
			return line[:50] + "..."
		}
		return line
	}
	
	return ""
}

func ValidateServiceConfig(serviceName string) ([]string, error) {
	var issues []string
	
	// Try to validate the service configuration
	cmd := exec.Command("systemctl", "is-valid", serviceName)
	if err := cmd.Run(); err != nil {
		issues = append(issues, fmt.Sprintf("Service configuration is invalid: %v", err))
	}
	
	return issues, nil
}
