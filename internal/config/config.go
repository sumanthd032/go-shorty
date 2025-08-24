package config

import (
	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	Server ServerConfig
}

// ServerConfig holds the configuration for the HTTP server.
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (config Config, err error) {
	// --- Set up Viper ---

	// AddConfigPath tells viper where to look for the config file.
	// "." means the current directory (the root of our project).
	viper.AddConfigPath(".")
	// SetConfigName tells viper the name of the config file (without the extension).
	viper.SetConfigName("config")
	// SetConfigType tells viper the type of the config file.
	viper.SetConfigType("yaml")

	// --- Set up Environment Variable handling ---
	// This allows us to override config file settings with environment variables.
	// e.g., SERVER_PORT=8888 go run ./cmd/server/main.go
	viper.AutomaticEnv()

	// --- Read the config ---
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	// --- Unmarshal the config into our struct ---
	// Unmarshaling is the process of converting the read configuration
	// into our strongly-typed Go struct.
	err = viper.Unmarshal(&config)
	return
}