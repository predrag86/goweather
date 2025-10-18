package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	City          string        `yaml:"city"`
	Hours         int           `yaml:"hours"`
	Emoji         bool          `yaml:"emoji"`
	Color         string        `yaml:"color"`
	Verbose       bool          `yaml:"verbose"`
	ForecastMode  string        `yaml:"forecast_mode"`
	LogPath       string        `yaml:"log_path"`
	CacheDuration time.Duration `yaml:"cache_duration"`
	TimeZone      string        `yaml:"time_zone"` // 🆕 added
}

// Load reads configuration from config.yaml (or sets defaults)
func Load() (*Config, error) {
	cfg := &Config{
		City:          "belgrade",
		Hours:         12,
		Emoji:         true,
		Color:         "auto",
		Verbose:       false,
		ForecastMode:  "current",
		LogPath:       "",
		CacheDuration: 10 * time.Minute,
		TimeZone:      "local", // 🆕 default (system local)
	}

	file, err := os.ReadFile("config.yaml")
	if err == nil {
		if err := yaml.Unmarshal(file, cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}
