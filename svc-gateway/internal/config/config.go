package config

import (
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

func Load() *Config {
	v := viper.New()

	// Set Defaults
	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("port", "8080")

	// Read from Environment Variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Config file support (optional)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		slog.Info("no config file found, using defaults and environment variables", "err", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		slog.Error("unable to decode config into struct", "err", err)
		// Return defaults if unmarshal fails but v was ok
	}

	return &cfg
}

func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}
