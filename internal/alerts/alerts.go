package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"local-monitor/internal/config"
	"local-monitor/internal/db"
	"local-monitor/internal/logs"
	"local-monitor/internal/metrics"
	"local-monitor/internal/ws"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

var (
	alertRules = make(map[string]*db.AlertRule)
	lastLogCheck = time.Now()
	lastRuleReload = time.Now()
	// Removed in-memory tracking - now using persistent database storage
)

func StartAlertEngine() {
	// Load alert rules from database
	loadAlertRules()
	
	// Initial cleanup of old processed entries
	go func() {
		deleted, err := db.CleanupOldProcessedEntries(24)
		if err != nil {
			log.Printf("[ALERTS] Failed initial cleanup: %v", err)
		} else {
			log.Printf("[ALERTS] Initial cleanup: removed %d old processed entries", deleted)
		}
		
		if count, err := db.GetProcessedEntriesCount(); err == nil {
			log.Printf("[ALERTS] Starting with %d processed entries in database", count)
		}
	}()
	
	// Start monitoring goroutines
	go systemMetricsMonitor()
	go logPatternMonitor()
	go exceptionDetectionMonitor()
}

func loadAlertRules() {
	rules, err := db.GetAlertRules()
	if err != nil {
		log.Printf("[ALERTS] Failed to load alert rules: %v", err)
		return
	}
	
	alertRules = make(map[string]*db.AlertRule)
	for _, rule := range rules {
		alertRules[rule.ID] = &rule
	}
}

// ReloadAlertRules forces a reload of alert rules from database
// Call this when rules are created, updated, or deleted
func ReloadAlertRules() {
	loadAlertRules()
}

func systemMetricsMonitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		checkSystemMetrics()
	}
}

func logPatternMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		checkLogPatterns()
	}
}

func exceptionDetectionMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		checkExceptions()
	}
}

func checkSystemMetrics() {
	m, err := metrics.GetHostMetrics()
	if err != nil {
		return
	}

	for _, rule := range alertRules {
		if !rule.Enabled || rule.Type != "system_metric" {
			continue
		}

		var triggered bool
		var message string

		switch rule.Condition {
		case "cpu_high":
			if m.CPUPercent > rule.Threshold {
				triggered = true
				message = fmt.Sprintf("CPU usage is %.2f%% (threshold: %.2f%%)", m.CPUPercent, rule.Threshold)
			}
		case "memory_high":
			if m.RAMPercent > rule.Threshold {
				triggered = true
				message = fmt.Sprintf("Memory usage is %.2f%% (threshold: %.2f%%)", m.RAMPercent, rule.Threshold)
			}
		case "disk_high":
			if m.DiskPercent > rule.Threshold {
				triggered = true
				message = fmt.Sprintf("Disk usage is %.2f%% (threshold: %.2f%%)", m.DiskPercent, rule.Threshold)
			}
		case "disk_low":
			free := 100 - m.DiskPercent
			if free < rule.Threshold {
				triggered = true
				message = fmt.Sprintf("Disk free space is %.2f%% (threshold: %.2f%%)", free, rule.Threshold)
			}
		}

		if triggered {
			triggerAlert(rule, message)
		}
	}
}

func checkLogPatterns() {
	for _, rule := range alertRules {
		if !rule.Enabled || rule.Type != "log_pattern" {
			continue
		}

		// Search logs for pattern matches (limit results for performance)
		results, err := logs.Search(rule.LogPattern, rule.AppFilter, rule.LogFilter, "", "", 500)
		if err != nil {
			continue
		}

		// Filter results to only NEW entries from recent time window (not previously processed)
		recentWindow := 10 * time.Minute // TODO: Make this configurable
		recentThreshold := time.Now().Add(-recentWindow)
		var newMatches []logs.LogResult
		for _, result := range results {
			// Skip entries older than 10 minutes
			if result.Timestamp.Before(recentThreshold) {
				continue
			}
			
			// Create hash-based key for this log entry (shorter and more efficient)
			entryHash := db.HashLogEntry(result.File, result.Message, result.Timestamp.Unix())
			
			// Only include if we haven't processed this entry before
			if !db.IsEntryProcessed(entryHash) {
				newMatches = append(newMatches, result)
				// Mark as processed in database
				db.MarkEntryProcessed(entryHash)
			}
		}

		if len(newMatches) > 0 {
			// Batch alerts for performance - don't send individual alerts for each match
			message := fmt.Sprintf("Found %d new log pattern matches for '%s'", len(newMatches), rule.LogPattern)
			if len(newMatches) > 0 {
				message += fmt.Sprintf(". Latest: %s", truncateString(newMatches[0].Message, 100))
			}
			// Add sample of other matches if many
			if len(newMatches) > 1 {
				message += fmt.Sprintf(". First few: %s", truncateString(newMatches[len(newMatches)-1].Message, 50))
			}
			triggerAlert(rule, message)
		}
	}
	lastLogCheck = time.Now()
	
	// Cleanup old processed entries (time-based cleanup every 10 minutes)
	if time.Now().Minute()%10 == 0 {
		go func() {
			db.CleanupOldProcessedEntries(24)
		}()
	}
}

func checkExceptions() {
	// Simplified exception patterns for better performance
	exceptionPatterns := []string{
		`Traceback`,
		`Exception`,
		`TypeError`,
		`ValueError`,
		`KeyError`,
		`NullPointerException`,
		`RuntimeException`,
		`Fatal error`,
		`Uncaught`,
	}

	for _, rule := range alertRules {
		if !rule.Enabled || rule.Type != "exception_detection" {
			continue
		}

		// Use custom pattern or default exception patterns
		pattern := rule.LogPattern
		if pattern == "" {
			// Use simplified exception patterns
			pattern = strings.Join(exceptionPatterns, "|")
		}

		// Search for exceptions using rule's log filter (empty means all logs)
		results, err := logs.Search(pattern, rule.AppFilter, rule.LogFilter, "", "", 50)
		if err != nil {
			continue
		}

		// Filter NEW exceptions from recent time window (not previously processed)
		recentWindow := 10 * time.Minute // TODO: Make this configurable
		recentThreshold := time.Now().Add(-recentWindow)
		var newExceptions []logs.LogResult
		for _, result := range results {
			// Skip entries older than 10 minutes
			if result.Timestamp.Before(recentThreshold) {
				continue
			}
			
			// Create hash-based key for this log entry
			entryHash := db.HashLogEntry(result.File, result.Message, result.Timestamp.Unix())
			
			// Only include if we haven't processed this entry before
			if !db.IsEntryProcessed(entryHash) {
				newExceptions = append(newExceptions, result)
				// Mark as processed in database
				db.MarkEntryProcessed(entryHash)
			}
		}

		if len(newExceptions) > 0 {
			message := fmt.Sprintf("Detected %d new exceptions in logs", len(newExceptions))
			if len(newExceptions) > 0 {
				message += fmt.Sprintf(". Latest: %s", truncateString(newExceptions[0].Message, 100))
			}
			triggerAlert(rule, message)
		}
	}
}

func triggerAlert(rule *db.AlertRule, message string) {
	// No cooldown needed - duplicate detection handles spam prevention

	// Record alert in database
	alert := db.Alert{
		RuleID:    rule.ID,
		Type:      rule.Name,
		Severity:  rule.Severity,
		Message:   message,
		Timestamp: time.Now(),
		Resolved:  false,
	}
	
	if err := db.RecordAlertWithRule(alert); err != nil {
		log.Printf("[ALERTS] Failed to record alert: %v", err)
		return
	}

	// Update rule's last triggered time
	now := time.Now()
	rule.LastTriggered = &now
	db.UpdateAlertRuleLastTriggered(rule.ID, now)
	
	// Update the in-memory rule as well
	alertRules[rule.ID] = rule

	// Send notifications
	if rule.EmailEnabled {
		go sendNotifications(rule.Name, message, rule.Severity)
	}

	log.Printf("[ALERTS] Triggered: %s - %s", rule.Name, message)
	
	// Broadcast alert to WebSocket clients
	ws.BroadcastNewAlert(alert)
	
	// Broadcast rule update to refresh "Last triggered" time on frontend
	ws.BroadcastRuleUpdate(*rule)
}

func sendNotifications(subject, body, severity string) {
	notifiers := config.AppConfigData.Notifiers

	if notifiers.Email.Enabled {
		sendEmail(notifiers.Email, subject, body, severity)
	}
	if notifiers.Webhook.Enabled {
		sendWebhook(notifiers.Webhook, subject, body, severity)
	}
}

func sendEmail(cfg config.EmailConfig, subject, body, severity string) {
	if cfg.SMTPHost == "" || cfg.Username == "" || len(cfg.To) == 0 {
		log.Printf("[ALERTS] Email configuration incomplete")
		return
	}

	// Use From field if specified, otherwise use Username
	fromAddr := cfg.From
	if fromAddr == "" {
		fromAddr = cfg.Username
	}

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
	
	// Create HTML email with severity styling
	severityColor := getSeverityColor(severity)
	htmlBody := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif;">
			<div style="background-color: %s; color: white; padding: 10px; border-radius: 5px; margin-bottom: 20px;">
				<h2 style="margin: 0;">🚨 Alert: %s</h2>
				<p style="margin: 5px 0 0 0;">Severity: %s</p>
			</div>
			<div style="padding: 20px; background-color: #f9f9f9; border-radius: 5px;">
				<p><strong>Message:</strong></p>
				<p>%s</p>
				<hr>
				<p><small>Timestamp: %s</small></p>
				<p><small>Generated by Logger EMP</small></p>
			</div>
		</body>
		</html>
	`, severityColor, subject, strings.ToUpper(severity), body, time.Now().Format("2006-01-02 15:04:05"))

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: [%s] Alert: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", fromAddr, cfg.To[0], strings.ToUpper(severity), subject, htmlBody))

	addr := cfg.SMTPHost + ":" + strconv.Itoa(cfg.SMTPPort)
	if err := smtp.SendMail(addr, auth, fromAddr, cfg.To, msg); err != nil {
		log.Printf("[ALERTS] Failed to send email: %v", err)
	} else {
		log.Printf("[ALERTS] Email sent successfully from %s to %s", fromAddr, cfg.To[0])
	}
}

func sendWebhook(cfg config.WebhookConfig, subject, body, severity string) {
	payload := map[string]interface{}{
		"alert":     subject,
		"message":   body,
		"severity":  severity,
		"timestamp": time.Now().Unix(),
		"source":    "logger-emp",
	}
	
	data, _ := json.Marshal(payload)
	resp, err := http.Post(cfg.URL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("[ALERTS] Failed to send webhook: %v", err)
	} else {
		resp.Body.Close()
		log.Printf("[ALERTS] Webhook sent successfully")
	}
}

func getSeverityColor(severity string) string {
	colors := map[string]string{
		"low":      "#17a2b8",
		"medium":   "#ffc107",
		"high":     "#fd7e14",
		"critical": "#dc3545",
	}
	if color, ok := colors[severity]; ok {
		return color
	}
	return "#6c757d"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Legacy function for backward compatibility
func checkAlerts() {
	checkSystemMetrics()
}