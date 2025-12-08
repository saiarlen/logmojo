package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"local-monitor/internal/config"
	"local-monitor/internal/db"
	"local-monitor/internal/metrics"
	"net/http"
	"net/smtp"
	"strconv"
	"time"
)

func StartAlertEngine() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			checkAlerts()
		}
	}()
}

func checkAlerts() {
	m, err := metrics.GetHostMetrics()
	if err != nil {
		return
	}

	cfg := config.AppConfigData.Alerts

	// Check CPU
	if cfg.CPU.Enabled && m.CPUPercent > cfg.CPU.Threshold {
		triggerAlert("cpu_high", fmt.Sprintf("CPU usage is %.2f%% (Threshold: %.2f%%)", m.CPUPercent, cfg.CPU.Threshold))
	}

	// Check Disk
	// DiskPercent is used %, so free is 100 - used
	free := 100 - m.DiskPercent
	if cfg.Disk.Enabled && free < cfg.Disk.ThresholdPercentFree {
		triggerAlert("disk_low", fmt.Sprintf("Disk free space is %.2f%% (Threshold: %.2f%%)", free, cfg.Disk.ThresholdPercentFree))
	}
}

func triggerAlert(alertType, message string) {
	// Record to DB
	db.RecordAlert(alertType, message)

	// Notify
	go sendNotifications(alertType, message)
}

func sendNotifications(subject, body string) {
	notifiers := config.AppConfigData.Notifiers

	if notifiers.Email.Enabled {
		sendEmail(notifiers.Email, subject, body)
	}
	if notifiers.Webhook.Enabled {
		sendWebhook(notifiers.Webhook, subject, body)
	}
}

func sendEmail(cfg config.EmailConfig, subject, body string) {
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
	msg := []byte("To: " + cfg.To[0] + "\r\n" +
		"Subject: Alert: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")
	addr := cfg.SMTPHost + ":" + strconv.Itoa(cfg.SMTPPort)
	_ = smtp.SendMail(addr, auth, cfg.Username, cfg.To, msg)
}

func sendWebhook(cfg config.WebhookConfig, subject, body string) {
	payload := map[string]string{
		"alert":   subject,
		"message": body,
	}
	data, _ := json.Marshal(payload)
	_, _ = http.Post(cfg.URL, "application/json", bytes.NewBuffer(data))
}
