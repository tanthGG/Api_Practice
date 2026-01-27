package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	App   AppConfig   `mapstructure:"app"`
	Log   LogConfig   `mapstructure:"log"`
	DbSQL DbSQLConfig `mapstructure:"db-sql"`
	Redis RedisConfig `mapstructure:"redis"`
}

// AppConfig represents the application configuration
type AppConfig struct {
	Name          string `mapstructure:"name"`
	Port          string `mapstructure:"port"`
	Version       string `mapstructure:"version"`
	MobileVersion string `mapstructure:"mobile-version"`
	ProjectID     string `mapstructure:"project_id"`
	Env           string `mapstructure:"env"`
}

// LogConfig represents the logging configuration
type LogConfig struct {
	Env   string `mapstructure:"env"`
	Level string `mapstructure:"level"`
}

// DbSQLConfig represents the database configuration
type DbSQLConfig struct {
	DBName             string        `mapstructure:"dbname"`
	Host               string        `mapstructure:"host"` // e.g. "(localhost:3306)" or "tcp(localhost:3306)" or "mysql"
	Username           string        `mapstructure:"username"`
	Password           string        `mapstructure:"password"`
	MaxIdleConns       int           `mapstructure:"maxIdleConns"`
	MaxOpenConns       int           `mapstructure:"maxOpenConns"`
	MaxLifeTimeMinutes time.Duration `mapstructure:"maxLifeTimeMinutes"`
}

// RedisConfig represents Redis connection settings
type RedisConfig struct {
	Host     string `mapstructure:"host"`     // e.g. "127.0.0.1" or "recent-merchant-redis"
	Port     string `mapstructure:"port"`     // e.g. "6379"
	Password string `mapstructure:"password"` // optional
	DB       int    `mapstructure:"db"`       // optional, default 0
}

// Addr returns "host:port" with sensible defaults
func (r RedisConfig) Addr() string {
	h := r.Host
	if h == "" {
		h = "127.0.0.1"
	}
	p := r.Port
	if p == "" {
		p = "6379"
	}
	return fmt.Sprintf("%s:%s", h, p)
}

// GetDSN returns the database connection string
func (c *DbSQLConfig) GetDSN() string {
	host := c.Host
	// Handle Docker environment where host is just the service name
	if host == "mysql" {
		host = "tcp(mysql:3306)"
	} else if host != "" && host[0] != '(' {
		// If host doesn't contain protocol, wrap it with tcp()
		host = "tcp(" + host + ")"
	}

	if c.Password != "" {
		return c.Username + ":" + c.Password + "@" + host + "/" + c.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	}
	return c.Username + "@" + host + "/" + c.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "dev" || c.App.Env == "development"
}

// IsProduction returns true if the application is running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Env == "prod" || c.App.Env == "production"
}

// AppInfo holds basic application information
type AppInfo struct {
	Name          string
	Version       string
	MobileVersion string
	ProjectID     string
	Environment   string
}

// GetAppInfo returns application information from the config
func (c *Config) GetAppInfo() AppInfo {
	return AppInfo{
		Name:          c.App.Name,
		Version:       c.App.Version,
		MobileVersion: c.App.MobileVersion,
		ProjectID:     c.App.ProjectID,
		Environment:   c.App.Env,
	}
}

// LoadConfig loads the configuration from file and environment variables
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Environment variables
	viper.SetEnvPrefix("MARKETPLACE")
	viper.AutomaticEnv()

	// Allow overriding via common envs (useful in Docker)
	viper.BindEnv("db-sql.host", "DB_HOST")
	viper.BindEnv("db-sql.dbname", "DB_NAME")
	viper.BindEnv("db-sql.username", "DB_USER")
	viper.BindEnv("db-sql.password", "DB_PASSWORD")
	viper.BindEnv("app.port", "APP_PORT")

	// Redis envs
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	// Defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	// App
	viper.SetDefault("app.name", "marketplace-merchant-api")
	viper.SetDefault("app.port", "80")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.env", "dev")

	// Log
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.env", "dev")

	// SQL
	viper.SetDefault("db-sql.host", "tcp(localhost:3306)")
	viper.SetDefault("db-sql.dbname", "marketplace_db")
	viper.SetDefault("db-sql.username", "marketplace")
	viper.SetDefault("db-sql.password", "password")
	viper.SetDefault("db-sql.maxIdleConns", 32)
	viper.SetDefault("db-sql.maxOpenConns", 64)
	viper.SetDefault("db-sql.maxLifeTimeMinutes", "5m")

}
