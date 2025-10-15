package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config defines all configurable parameters for GoWeather.
type Config struct {
	City         string `mapstructure:"city"`
	Hours        int    `mapstructure:"hours"`
	Emoji        bool   `mapstructure:"emoji"`
	Color        string `mapstructure:"color"`
	Verbose      bool   `mapstructure:"verbose"`
	LogPath      string `mapstructure:"log_path"`
	ForecastMode string `mapstructure:"forecast_mode"` // "hourly" or "current"
}

// Load loads configuration from file, environment, and defaults.
// Priority: defaults < config.yaml < environment variables.
func Load() (*Config, error) {
	v := viper.New()

	// 1️⃣ Basic setup
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".") // project root

	// 2️⃣ Default values
	v.SetDefault("city", "belgrade")
	v.SetDefault("hours", 6)
	v.SetDefault("emoji", true)
	v.SetDefault("color", "auto")
	v.SetDefault("verbose", false)
	v.SetDefault("forecast_mode", "hourly")

	cacheDir, _ := os.UserCacheDir()
	defaultLogPath := filepath.Join(cacheDir, "goweather", "logs")
	v.SetDefault("log_path", defaultLogPath)

	// 3️⃣ Environment variable support
	// Environment vars are automatically uppercased and prefixed
	// e.g. GOWEATHER_CITY, GOWEATHER_HOURS, etc.
	v.SetEnvPrefix("GOWEATHER")
	v.AutomaticEnv()

	// 4️⃣ Read from config.yaml if present
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// create default config file if missing
			_ = saveDefaultConfig()
			fmt.Println("Created default config.yaml.")
		} else {
			return nil, fmt.Errorf("config read error: %v", err)
		}
	}

	// 5️⃣ Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal error: %v", err)
	}

	return &cfg, nil
}

// saveDefaultConfig writes a new config.yaml with defaults.
func saveDefaultConfig() error {
	content := []byte(`# GoWeather configuration file
# You can override any of these values with environment variables, e.g.:
#   export GOWEATHER_CITY=rome
#   export GOWEATHER_HOURS=12

city: "belgrade"
hours: 6
emoji: true
color: "auto"
verbose: false
log_path: ""
forecast_mode: "hourly"
`)
	return os.WriteFile("config.yaml", content, 0644)
}
