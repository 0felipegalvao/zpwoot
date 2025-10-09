package config

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	ServerHost string
	ServerURL  string
	LogLevel   string
	LogFormat  string
	LogOutput  string
	APIKey     string

	Database DatabaseConfig
	Cache    CacheConfig

	Postgres PostgresConfig

	WALogLevel string

	GlobalWebhookURL string

	Environment string
}

type DatabaseConfig struct {
	URL string
}

type PostgresConfig struct {
	DB       string
	User     string
	Password string
	Port     int
}

type CacheConfig struct {
	RedisEnabled bool `env:"REDIS_ENABLED" default:"false"`
	Type         string

	URL      string `env:"REDIS_URL" default:"redis://localhost:6379/0"`
	Password string `env:"REDIS_PASSWORD" default:""`
	DB       int    `env:"REDIS_DB" default:"0"`

	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	PoolSize     int
	MinIdleConns int
	MaxIdleConns int

	DefaultTTL time.Duration
	SessionTTL time.Duration
	WebhookTTL time.Duration
	ProxyTTL   time.Duration
	ListTTL    time.Duration

	KeyPrefix string

	EnableCompression bool
	EnableMetrics     bool
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {

		_ = err
	}

	cfg := &Config{

		Port:       getEnv("PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerURL:  getEnv("SERVER_URL", ""),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		LogFormat:  getEnv("LOG_FORMAT", "console"),
		LogOutput:  getEnv("LOG_OUTPUT", "stdout"),
		APIKey:     getEnv("ZP_API_KEY", ""),

		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},

		Cache: loadCacheConfig(),

		Postgres: PostgresConfig{
			DB:       getEnv("POSTGRES_DB", ""),
			User:     getEnv("POSTGRES_USER", ""),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
		},

		WALogLevel: getEnv("WA_LOG_LEVEL", "INFO"),

		GlobalWebhookURL: getEnv("GLOBAL_WEBHOOK_URL", ""),

		Environment: getEnv("NODE_ENV", "development"),
	}

	if err := cfg.Validate(); err != nil {
		panic("Configuration validation failed: " + err.Error())
	}

	return cfg
}

func (c *Config) Validate() error {
	if c.APIKey == "" {
		return errors.New("ZP_API_KEY is required")
	}

	if c.Database.URL == "" {
		return errors.New("DATABASE_URL is required")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}

	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}

	return fallback
}

func loadCacheConfig() CacheConfig {
	redisEnabled := getEnvAsBool("REDIS_ENABLED", false)

	cacheType := "memory"
	if redisEnabled {
		cacheType = "redis"
	}

	config := CacheConfig{
		RedisEnabled: redisEnabled,
		Type:         cacheType,
		URL:          getEnv("REDIS_URL", "redis://localhost:6379/0"),
		Password:     getEnv("REDIS_PASSWORD", ""),
		DB:           getEnvAsInt("REDIS_DB", 0),

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		PoolSize:     10,
		MinIdleConns: 2,
		MaxIdleConns: 5,

		DefaultTTL: 5 * time.Minute,
		SessionTTL: 5 * time.Minute,
		WebhookTTL: 10 * time.Minute,
		ProxyTTL:   15 * time.Minute,
		ListTTL:    1 * time.Minute,

		KeyPrefix:         "zpwoot",
		EnableCompression: false,
		EnableMetrics:     true,
	}

	return config
}

func (c *Config) GetServerAddress() string {
	return c.ServerHost + ":" + c.Port
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
