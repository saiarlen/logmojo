package db

import (
	"database/sql"
	"fmt"
	"time"
)

type AlertRule struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	Type         string    `json:"type" db:"type"`
	Condition    string    `json:"condition" db:"condition"`
	Threshold    float64   `json:"threshold" db:"threshold"`
	Severity     string    `json:"severity" db:"severity"`
	Enabled      bool      `json:"enabled" db:"enabled"`
	EmailEnabled bool      `json:"email_enabled" db:"email_enabled"`
	LogPattern   string    `json:"log_pattern" db:"log_pattern"`
	AppFilter    string    `json:"app_filter" db:"app_filter"`
	LogFilter    string    `json:"log_filter" db:"log_filter"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	LastTriggered *time.Time `json:"last_triggered" db:"last_triggered"`
}

type Alert struct {
	ID        int       `json:"id" db:"id"`
	RuleID    string    `json:"rule_id" db:"rule_id"`
	Type      string    `json:"type" db:"type"`
	Severity  string    `json:"severity" db:"severity"`
	Message   string    `json:"message" db:"message"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Resolved  bool      `json:"resolved" db:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at" db:"resolved_at"`
}

func RecordAlertWithRule(alert Alert) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := DB.Exec(`INSERT INTO alerts (rule_id, timestamp, type, severity, message, resolved) 
					 VALUES (?, ?, ?, ?, ?, ?)`, 
					 alert.RuleID, alert.Timestamp, alert.Type, alert.Severity, alert.Message, alert.Resolved)
	return err
}

func GetAlertRules() ([]AlertRule, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	rows, err := DB.Query(`SELECT id, name, description, type, condition, threshold, severity, 
							 enabled, email_enabled, log_pattern, app_filter, log_filter, 
							 created_at, updated_at, last_triggered 
						 FROM alert_rules ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var rules []AlertRule
	for rows.Next() {
		var rule AlertRule
		var lastTriggered sql.NullTime
		err := rows.Scan(&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Condition,
			&rule.Threshold, &rule.Severity, &rule.Enabled, &rule.EmailEnabled,
			&rule.LogPattern, &rule.AppFilter, &rule.LogFilter,
			&rule.CreatedAt, &rule.UpdatedAt, &lastTriggered)
		if err != nil {
			return nil, err
		}
		if lastTriggered.Valid {
			rule.LastTriggered = &lastTriggered.Time
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func CreateAlertRule(rule AlertRule) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := DB.Exec(`INSERT INTO alert_rules (id, name, description, type, condition, threshold, 
						 severity, enabled, email_enabled, log_pattern, app_filter, log_filter, 
						 created_at, updated_at) 
					 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rule.ID, rule.Name, rule.Description, rule.Type, rule.Condition, rule.Threshold,
		rule.Severity, rule.Enabled, rule.EmailEnabled, rule.LogPattern, rule.AppFilter,
		rule.LogFilter, rule.CreatedAt, rule.UpdatedAt)
	return err
}

func UpdateAlertRule(rule AlertRule) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := DB.Exec(`UPDATE alert_rules SET name=?, description=?, type=?, condition=?, threshold=?, 
						 severity=?, enabled=?, email_enabled=?, log_pattern=?, app_filter=?, log_filter=?, 
						 updated_at=? WHERE id=?`,
		rule.Name, rule.Description, rule.Type, rule.Condition, rule.Threshold,
		rule.Severity, rule.Enabled, rule.EmailEnabled, rule.LogPattern, rule.AppFilter,
		rule.LogFilter, rule.UpdatedAt, rule.ID)
	return err
}

func DeleteAlertRule(id string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := DB.Exec("DELETE FROM alert_rules WHERE id=?", id)
	return err
}

func UpdateAlertRuleLastTriggered(id string, timestamp time.Time) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := DB.Exec("UPDATE alert_rules SET last_triggered=? WHERE id=?", timestamp, id)
	return err
}

func GetAlertHistory() ([]Alert, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	rows, err := DB.Query(`SELECT id, rule_id, type, severity, message, timestamp, resolved, resolved_at 
						 FROM alerts ORDER BY timestamp DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var alerts []Alert
	for rows.Next() {
		var alert Alert
		var ruleID sql.NullString
		var resolvedAt sql.NullTime
		err := rows.Scan(&alert.ID, &ruleID, &alert.Type, &alert.Severity, &alert.Message,
			&alert.Timestamp, &alert.Resolved, &resolvedAt)
		if err != nil {
			return nil, err
		}
		if ruleID.Valid {
			alert.RuleID = ruleID.String
		}
		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}
		alerts = append(alerts, alert)
	}
	return alerts, nil
}

func ResolveAlert(id int) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := DB.Exec("UPDATE alerts SET resolved=1, resolved_at=? WHERE id=?", time.Now(), id)
	return err
}