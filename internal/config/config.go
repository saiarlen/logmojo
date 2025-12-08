package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Apps      []AppConfig     `mapstructure:"apps"`
	Alerts    AlertsConfig    `mapstructure:"alerts"`
	Notifiers NotifiersConfig `mapstructure:"notifiers"`
}

type ServerConfig struct {
	ListenAddr string `mapstructure:"listen_addr"`
	AuthToken  string `mapstructure:"auth_token"`
}

type AppConfig struct {
	Name        string      `mapstructure:"name"`
	ServiceName string      `mapstructure:"service_name"`
	Logs        []LogConfig `mapstructure:"logs"`
}

type LogConfig struct {
	Name string `mapstructure:"name"`
	Path string `mapstructure:"path"`
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
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/local-monitor/")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("MONITOR")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.Unmarshal(&AppConfigData)
}
