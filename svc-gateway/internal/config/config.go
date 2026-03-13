package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

type Config struct {
	Host            string         `mapstructure:"host"`
	Port            string         `mapstructure:"port"`
	RecommenderAddr string         `mapstructure:"recommender_addr"`
	Log             LogConfig      `mapstructure:"log"`
	Redis           RedisConfig    `mapstructure:"redis"`
	Database        DatabaseConfig `mapstructure:"database"`
	Mailer          MailerConfig   `mapstructure:"mailer"`
	JWT             JWTConfig      `mapstructure:"jwt"`
}

type JWTConfig struct {
	Secret  string `mapstructure:"secret"`
	Timeout int    `mapstructure:"timeout"` // in hours
}

type MailerConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
	TimeZone string `mapstructure:"timezone"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func Load() *Config {
	v := viper.New()

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
