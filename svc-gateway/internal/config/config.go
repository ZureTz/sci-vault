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
	Storage         StorageConfig  `mapstructure:"storage"`
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

type StorageConfig struct {
	Endpoint         string `mapstructure:"endpoint"`
	// PresignEndpoint is the S3 endpoint used when generating presigned URLs.
	// It must match the Host header that the reverse proxy forwards to RustFS,
	// because presigned URLs are signed over the Host header.
	// Defaults to Endpoint when empty.
	PresignEndpoint  string `mapstructure:"presign_endpoint"`
	AccessKey        string `mapstructure:"access_key"`
	SecretKey        string `mapstructure:"secret_key"`
	PrivateBucket    string `mapstructure:"private_bucket"`
	PublicBucket     string `mapstructure:"public_bucket"`
	UseSSL           bool   `mapstructure:"use_ssl"`
	PublicProxyPath  string `mapstructure:"public_proxy_path"`
	PrivateProxyPath string `mapstructure:"private_proxy_path"`
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

func Load(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
		slog.Info("loading config from specified path", "path", configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		slog.Info("loading config from default path", "path", "./config.yaml")
	}

	if err := v.ReadInConfig(); err != nil {
		slog.Error("no config file found", "err", err)
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		slog.Error("unable to decode config into struct", "err", err)
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}
