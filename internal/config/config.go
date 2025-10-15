package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Session  SessionConfig  `mapstructure:"session"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
	URL         string `mapstructure:"url"`
	Key         string `mapstructure:"key"`
	Timezone    string `mapstructure:"timezone"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
	Charset  string `mapstructure:"charset"`
	MaxOpen  int    `mapstructure:"max_open"`
	MaxIdle  int    `mapstructure:"max_idle"`
	MaxLife  int    `mapstructure:"max_life"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Driver string `mapstructure:"driver"`
	Host   string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	DB     int    `mapstructure:"db"`
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Driver   string        `mapstructure:"driver"`
	Lifetime time.Duration `mapstructure:"lifetime"`
	Secure   bool          `mapstructure:"secure"`
	HttpOnly bool          `mapstructure:"http_only"`
	SameSite string        `mapstructure:"same_site"`
	Encrypt  bool          `mapstructure:"encrypt"`
	Key      string        `mapstructure:"key"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
	Issuer     string        `mapstructure:"issuer"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string        `mapstructure:"jwt_secret"`
	TokenExpiry   time.Duration `mapstructure:"token_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
	PasswordSalt  string        `mapstructure:"password_salt"`
}

// Load loads configuration from files and environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional
	}

	// Set default values
	setDefaults()

	// Configure viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./configs")

	// Enable reading from environment variables
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config file not found, use defaults and environment variables
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Override with environment variables
	overrideWithEnv(&config)

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "Dolphin Framework")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", true)
	viper.SetDefault("app.url", "http://localhost:8080")
	viper.SetDefault("app.timezone", "UTC")

	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 120)

	// Database defaults
	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.database", "dolphin")
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("database.max_open", 25)
	viper.SetDefault("database.max_idle", 5)
	viper.SetDefault("database.max_life", 300)

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")

	// Cache defaults
	viper.SetDefault("cache.driver", "redis")
	viper.SetDefault("cache.host", "localhost")
	viper.SetDefault("cache.port", 6379)
	viper.SetDefault("cache.db", 0)

	// Session defaults
	viper.SetDefault("session.driver", "cookie")
	viper.SetDefault("session.lifetime", "24h")
	viper.SetDefault("session.secure", false)
	viper.SetDefault("session.http_only", true)
	viper.SetDefault("session.same_site", "Lax")
	viper.SetDefault("session.encrypt", false)

	// JWT defaults
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiration", "24h")
	viper.SetDefault("jwt.issuer", "dolphin-framework")

	// Auth defaults
	viper.SetDefault("auth.jwt_secret", "your-jwt-secret-key")
	viper.SetDefault("auth.token_expiry", "1h")
	viper.SetDefault("auth.refresh_expiry", "7d")
	viper.SetDefault("auth.password_salt", "")
}

// overrideWithEnv overrides configuration with environment variables
func overrideWithEnv(config *Config) {
	// App overrides
	if val := os.Getenv("APP_NAME"); val != "" {
		config.App.Name = val
	}
	if val := os.Getenv("APP_ENV"); val != "" {
		config.App.Environment = val
	}
	if val := os.Getenv("APP_DEBUG"); val != "" {
		if debug, err := strconv.ParseBool(val); err == nil {
			config.App.Debug = debug
		}
	}
	if val := os.Getenv("APP_URL"); val != "" {
		config.App.URL = val
	}
	if val := os.Getenv("APP_KEY"); val != "" {
		config.App.Key = val
	}

	// Server overrides
	if val := os.Getenv("SERVER_HOST"); val != "" {
		config.Server.Host = val
	}
	if val := os.Getenv("SERVER_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.Server.Port = port
		}
	}

	// Database overrides
	if val := os.Getenv("DB_DRIVER"); val != "" {
		config.Database.Driver = val
	}
	if val := os.Getenv("DB_HOST"); val != "" {
		config.Database.Host = val
	}
	if val := os.Getenv("DB_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.Database.Port = port
		}
	}
	if val := os.Getenv("DB_DATABASE"); val != "" {
		config.Database.Database = val
	}
	if val := os.Getenv("DB_USERNAME"); val != "" {
		config.Database.Username = val
	}
	if val := os.Getenv("DB_PASSWORD"); val != "" {
		config.Database.Password = val
	}

	// Log overrides
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.Log.Level = val
	}
	if val := os.Getenv("LOG_FORMAT"); val != "" {
		config.Log.Format = val
	}

	// Cache overrides
	if val := os.Getenv("CACHE_HOST"); val != "" {
		config.Cache.Host = val
	}
	if val := os.Getenv("CACHE_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.Cache.Port = port
		}
	}

	// JWT overrides
	if val := os.Getenv("JWT_SECRET"); val != "" {
		config.JWT.Secret = val
	}

	// Auth overrides
	if val := os.Getenv("AUTH_JWT_SECRET"); val != "" {
		config.Auth.JWTSecret = val
	}
	if val := os.Getenv("AUTH_TOKEN_EXPIRY"); val != "" {
		if expiry, err := time.ParseDuration(val); err == nil {
			config.Auth.TokenExpiry = expiry
		}
	}
	if val := os.Getenv("AUTH_REFRESH_EXPIRY"); val != "" {
		if expiry, err := time.ParseDuration(val); err == nil {
			config.Auth.RefreshExpiry = expiry
		}
	}
	if val := os.Getenv("AUTH_PASSWORD_SALT"); val != "" {
		config.Auth.PasswordSalt = val
	}
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development" || c.App.Environment == "local"
}

// IsTesting returns true if the environment is testing
func (c *Config) IsTesting() bool {
	return c.App.Environment == "testing"
}
