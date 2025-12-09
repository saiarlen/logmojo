package config

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Security  SecurityConfig  `mapstructure:"security"`
	Apps      []AppConfig     `mapstructure:"apps"`
	Alerts    AlertsConfig    `mapstructure:"alerts"`
	Notifiers NotifiersConfig `mapstructure:"notifiers"`
	DevMode   bool            `mapstructure:"dev_mode"`
}

type ServerConfig struct {
	ListenAddr string `mapstructure:"listen_addr"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type SecurityConfig struct {
	JWTSecret      string `mapstructure:"jwt_secret"`
	SessionTimeout string `mapstructure:"session_timeout"`
}

type AppConfig struct {
	Name        string      `mapstructure:"name" json:"name"`
	ServiceName string      `mapstructure:"service_name" json:"service_name"`
	Logs        []LogConfig `mapstructure:"logs" json:"logs"`
}

type LogConfig struct {
	Name string `mapstructure:"name" json:"name"`
	Path string `mapstructure:"path" json:"path"`
}

type AlertsConfig struct {
	CPU  AlertRule `mapstructure:"cpu_high"`
	Disk AlertRule `mapstructure:"disk_low"`
}

type AlertRule struct {
	Enabled              bool    `mapstructure:"enabled"`
	Threshold            float64 `mapstructure:"threshold"`
	ThresholdPercentFree float64 `mapstructure:"threshold_percent_free"`
}

type NotifiersConfig struct {
	Email   EmailConfig   `mapstructure:"email"`
	Webhook WebhookConfig `mapstructure:"webhook"`
}

type EmailConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	SMTPHost string   `mapstructure:"smtp_host"`
	SMTPPort int      `mapstructure:"smtp_port"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	To       []string `mapstructure:"to"`
}

type WebhookConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	URL     string `mapstructure:"url"`
}

var AppConfigData Config

func Load() error {
	// Load .env file if it exists (ignore errors if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Set up Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/local-monitor/")
	viper.AddConfigPath(".")

	// Environment variable configuration
	viper.SetEnvPrefix("MONITOR")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using environment variables and defaults: %v", err)
	}

	return viper.Unmarshal(&AppConfigData)
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.listen_addr", "0.0.0.0:7005")

	// Database defaults
	viper.SetDefault("database.path", "./monitor.db")

	// Security defaults
	viper.SetDefault("security.jwt_secret", "default-jwt-secret")
	viper.SetDefault("security.session_timeout", "24h")

	// Alert defaults
	viper.SetDefault("alerts.cpu_high.enabled", true)
	viper.SetDefault("alerts.cpu_high.threshold", 80.0)
	viper.SetDefault("alerts.disk_low.enabled", true)
	viper.SetDefault("alerts.disk_low.threshold_percent_free", 10.0)

	// Notifier defaults
	viper.SetDefault("notifiers.email.enabled", false)
	viper.SetDefault("notifiers.webhook.enabled", false)

	// Development mode
	viper.SetDefault("dev_mode", false)
}
