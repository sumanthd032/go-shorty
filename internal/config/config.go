package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig 
	Auth     AuthConfig
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// Add this new struct
type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

// ... rest of the file is the same
func LoadConfig() (config Config, err error) {
    // ... no changes needed here, Viper will automatically pick up the new section
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	SessionKey string `mapstructure:"session_key"`
}