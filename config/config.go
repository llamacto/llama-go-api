package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// GlobalConfig stores the global configuration
var GlobalConfig *Config

// ErrConfigCacheNotFound indicates there is no cached configuration available.
var ErrConfigCacheNotFound = errors.New("config cache not found")

const defaultCachePath = "storage/framework/cache/config.json"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
	OpenAI   OpenAIConfig
	R2       R2Config
	Email    EmailConfig
	App      AppConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port           int    `json:"port"`
	Mode           string `json:"mode"`
	ReadTimeout    int    `json:"read_timeout"`
	WriteTimeout   int    `json:"write_timeout"`
	MaxHeaderBytes int    `json:"max_header_bytes"`
}

type CORSConfig struct {
	AllowOrigins     []string `json:"allow_origins"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
}

type DatabaseConfig struct {
	Driver          string `json:"driver"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"-"` // 敏感信息不序列化
	DBName          string `json:"dbname"`
	SSLMode         string `json:"sslmode"`
	Timezone        string `json:"timezone"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	MaxOpenConns    int    `json:"max_open_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
	Enabled         bool   `json:"enabled"`
}

type RedisConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Password     string `json:"-"` // 敏感信息不序列化
	DB           int    `json:"db"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
}

type JWTConfig struct {
	Secret         string        `json:"-"` // 敏感信息不序列化
	ExpireDays     int           `json:"expire_days"`
	ExpireDuration time.Duration `json:"-"`
}

type LogConfig struct {
	Level      string `json:"level"`
	Filename   string `json:"filename"`
	MaxSize    int    `json:"max_size"`
	MaxAge     int    `json:"max_age"`
	MaxBackups int    `json:"max_backups"`
	Compress   bool   `json:"compress"`
}

type OpenAIConfig struct {
	APIKey string `json:"-"` // 敏感信息不序列化
}

type R2Config struct {
	AccessKeyID     string `json:"-"` // 敏感信息不序列化
	SecretAccessKey string `json:"-"` // 敏感信息不序列化
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	Endpoint        string `json:"endpoint"`
	PublicURL       string `json:"public_url"`
	PublicDomain    string `json:"public_domain"`
}

type EmailConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"-"` // 敏感信息不序列化
	From         string `json:"from"`
	ResendAPIKey string `json:"-"` // 敏感信息不序列化
}

type AppConfig struct {
	Name      string        `json:"name"`
	Version   string        `json:"version"`
	Secret    string        `json:"-"` // 敏感信息不序列化
	JWTSecret string        `json:"-"` // 敏感信息不序列化
	JWTExpire time.Duration `json:"jwt_expire"`
}

// Load loads configuration, preferring cached values if available.
func Load() (*Config, error) {
	return loadInternal(false)
}

// LoadFresh loads configuration ignoring any cached values.
func LoadFresh() (*Config, error) {
	return loadInternal(true)
}

func loadInternal(forceReload bool) (*Config, error) {
	if !forceReload {
		cfg, err := loadCachedConfig()
		if err == nil {
			GlobalConfig = cfg
			return cfg, nil
		}
		if err != nil && !errors.Is(err, ErrConfigCacheNotFound) {
			return nil, err
		}
	}

	resolvedEnv := loadEnvHierarchy()

	mode := strings.ToLower(os.Getenv("SERVER_MODE"))
	if mode == "" {
		mode = strings.ToLower(resolvedEnv)
	}
	if mode == "" {
		mode = strings.ToLower(os.Getenv("APP_ENV"))
	}
	if mode == "" {
		mode = "development"
	}

	if mode == "production" || mode == "release" {
		fmt.Println("Running in production mode, using system environment variables with highest priority")
	}

	config := &Config{}

	if err := loadServerConfig(config); err != nil {
		return nil, err
	}

	if err := loadDatabaseConfig(config); err != nil {
		return nil, err
	}

	if err := loadRedisConfig(config); err != nil {
		return nil, err
	}

	if err := loadJWTConfig(config); err != nil {
		return nil, err
	}

	if err := loadLogConfig(config); err != nil {
		return nil, err
	}

	if err := loadOpenAIConfig(config); err != nil {
		return nil, err
	}

	if err := loadR2Config(config); err != nil {
		return nil, err
	}

	if err := loadEmailConfig(config); err != nil {
		return nil, err
	}

	if err := loadCORSConfig(config); err != nil {
		return nil, err
	}

	if err := loadAppConfig(config); err != nil {
		return nil, err
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	GlobalConfig = config
	return config, nil
}

func loadServerConfig(config *Config) error {
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "6066"))
	if err != nil {
		return fmt.Errorf("invalid SERVER_PORT: %v", err)
	}

	readTimeout, err := strconv.Atoi(getEnv("SERVER_READ_TIMEOUT", "60"))
	if err != nil {
		return fmt.Errorf("invalid SERVER_READ_TIMEOUT: %v", err)
	}

	writeTimeout, err := strconv.Atoi(getEnv("SERVER_WRITE_TIMEOUT", "60"))
	if err != nil {
		return fmt.Errorf("invalid SERVER_WRITE_TIMEOUT: %v", err)
	}

	maxHeaderBytes, err := strconv.Atoi(getEnv("SERVER_MAX_HEADER_BYTES", "1048576"))
	if err != nil {
		return fmt.Errorf("invalid SERVER_MAX_HEADER_BYTES: %v", err)
	}

	mode := os.Getenv("SERVER_MODE")
	if mode == "" {
		mode = getEnv("APP_ENV", "debug")
	}

	config.Server = ServerConfig{
		Port:           port,
		Mode:           mode,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	return nil
}

func loadEnvHierarchy() string {
	seen := make(map[string]struct{})
	loadFile := func(path string) {
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		if _, err := os.Stat(path); err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("warning: unable to access %s: %v\n", path, err)
			}
			return
		}
		if err := godotenv.Overload(path); err != nil {
			fmt.Printf("warning: failed to load %s: %v\n", path, err)
			return
		}
		seen[path] = struct{}{}
	}

	// Base configuration
	loadFile(".env")

	// Determine active environment after base load
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("SERVER_MODE")
	}
	if env == "" {
		env = os.Getenv("GO_ENV")
	}

	// Environment-specific overrides (e.g., .env.production)
	if env != "" {
		loadFile(fmt.Sprintf(".env.%s", env))
	}

	// Local overrides shared across environments
	loadFile(".env.local")

	// Environment-specific local overrides (e.g., .env.production.local)
	if env != "" {
		loadFile(fmt.Sprintf(".env.%s.local", env))
	}

	return env
}

// cacheFilePath resolves the config cache file path.
func cacheFilePath() string {
	if path := os.Getenv("CONFIG_CACHE_PATH"); path != "" {
		return path
	}
	return filepath.FromSlash(defaultCachePath)
}

// CacheFilePath returns the current config cache file path.
func CacheFilePath() string {
	return cacheFilePath()
}

// CacheConfig writes the provided configuration to the cache file.
func CacheConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("cannot cache nil config")
	}
	serialized := newCachedConfig(cfg)
	data, err := json.MarshalIndent(serialized, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config cache: %w", err)
	}

	path := cacheFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to prepare cache directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config cache: %w", err)
	}
	return nil
}

// ClearCache removes the cached configuration file if it exists.
func ClearCache() error {
	path := cacheFilePath()
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to stat cache file: %w", err)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove config cache: %w", err)
	}
	return nil
}

func loadCachedConfig() (*Config, error) {
	path := cacheFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigCacheNotFound
		}
		return nil, fmt.Errorf("failed to read config cache: %w", err)
	}

	var cached cachedConfig
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("failed to decode config cache: %w", err)
	}

	return cached.toConfig(), nil
}

type cachedConfig struct {
	Server   cachedServerConfig   `json:"server"`
	Database cachedDatabaseConfig `json:"database"`
	Redis    cachedRedisConfig    `json:"redis"`
	JWT      cachedJWTConfig      `json:"jwt"`
	Log      cachedLogConfig      `json:"log"`
	OpenAI   cachedOpenAIConfig   `json:"openai"`
	R2       cachedR2Config       `json:"r2"`
	Email    cachedEmailConfig    `json:"email"`
	App      cachedAppConfig      `json:"app"`
}

type cachedServerConfig struct {
	Port           int    `json:"port"`
	Mode           string `json:"mode"`
	ReadTimeout    int    `json:"read_timeout"`
	WriteTimeout   int    `json:"write_timeout"`
	MaxHeaderBytes int    `json:"max_header_bytes"`
}

type cachedDatabaseConfig struct {
	Driver          string `json:"driver"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	DBName          string `json:"dbname"`
	SSLMode         string `json:"sslmode"`
	Timezone        string `json:"timezone"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	MaxOpenConns    int    `json:"max_open_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
	Enabled         bool   `json:"enabled"`
}

type cachedRedisConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
}

type cachedJWTConfig struct {
	Secret     string `json:"secret"`
	ExpireDays int    `json:"expire_days"`
}

type cachedLogConfig struct {
	Level      string `json:"level"`
	Filename   string `json:"filename"`
	MaxSize    int    `json:"max_size"`
	MaxAge     int    `json:"max_age"`
	MaxBackups int    `json:"max_backups"`
	Compress   bool   `json:"compress"`
}

type cachedOpenAIConfig struct {
	APIKey string `json:"api_key"`
}

type cachedR2Config struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	Endpoint        string `json:"endpoint"`
	PublicURL       string `json:"public_url"`
	PublicDomain    string `json:"public_domain"`
}

type cachedEmailConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	From         string `json:"from"`
	ResendAPIKey string `json:"resend_api_key"`
}

type cachedAppConfig struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Secret        string `json:"secret"`
	JWTSecret     string `json:"jwt_secret"`
	JWTExpireDays int    `json:"jwt_expire_days"`
}

func newCachedConfig(cfg *Config) cachedConfig {
	return cachedConfig{
		Server: cachedServerConfig{
			Port:           cfg.Server.Port,
			Mode:           cfg.Server.Mode,
			ReadTimeout:    cfg.Server.ReadTimeout,
			WriteTimeout:   cfg.Server.WriteTimeout,
			MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
		},
		Database: cachedDatabaseConfig{
			Driver:          cfg.Database.Driver,
			Host:            cfg.Database.Host,
			Port:            cfg.Database.Port,
			Username:        cfg.Database.Username,
			Password:        cfg.Database.Password,
			DBName:          cfg.Database.DBName,
			SSLMode:         cfg.Database.SSLMode,
			Timezone:        cfg.Database.Timezone,
			MaxIdleConns:    cfg.Database.MaxIdleConns,
			MaxOpenConns:    cfg.Database.MaxOpenConns,
			ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
			Enabled:         cfg.Database.Enabled,
		},
		Redis: cachedRedisConfig{
			Host:         cfg.Redis.Host,
			Port:         cfg.Redis.Port,
			Password:     cfg.Redis.Password,
			DB:           cfg.Redis.DB,
			PoolSize:     cfg.Redis.PoolSize,
			MinIdleConns: cfg.Redis.MinIdleConns,
		},
		JWT: cachedJWTConfig{
			Secret:     cfg.JWT.Secret,
			ExpireDays: cfg.JWT.ExpireDays,
		},
		Log: cachedLogConfig{
			Level:      cfg.Log.Level,
			Filename:   cfg.Log.Filename,
			MaxSize:    cfg.Log.MaxSize,
			MaxAge:     cfg.Log.MaxAge,
			MaxBackups: cfg.Log.MaxBackups,
			Compress:   cfg.Log.Compress,
		},
		OpenAI: cachedOpenAIConfig{
			APIKey: cfg.OpenAI.APIKey,
		},
		R2: cachedR2Config{
			AccessKeyID:     cfg.R2.AccessKeyID,
			SecretAccessKey: cfg.R2.SecretAccessKey,
			Bucket:          cfg.R2.Bucket,
			Region:          cfg.R2.Region,
			Endpoint:        cfg.R2.Endpoint,
			PublicURL:       cfg.R2.PublicURL,
			PublicDomain:    cfg.R2.PublicDomain,
		},
		Email: cachedEmailConfig{
			Host:         cfg.Email.Host,
			Port:         cfg.Email.Port,
			Username:     cfg.Email.Username,
			Password:     cfg.Email.Password,
			From:         cfg.Email.From,
			ResendAPIKey: cfg.Email.ResendAPIKey,
		},
		App: cachedAppConfig{
			Name:          cfg.App.Name,
			Version:       cfg.App.Version,
			Secret:        cfg.App.Secret,
			JWTSecret:     cfg.App.JWTSecret,
			JWTExpireDays: int(cfg.App.JWTExpire / (24 * time.Hour)),
		},
	}
}

func (c cachedConfig) toConfig() *Config {
	cfg := &Config{}

	cfg.Server = ServerConfig{
		Port:           c.Server.Port,
		Mode:           c.Server.Mode,
		ReadTimeout:    c.Server.ReadTimeout,
		WriteTimeout:   c.Server.WriteTimeout,
		MaxHeaderBytes: c.Server.MaxHeaderBytes,
	}

	cfg.Database = DatabaseConfig{
		Driver:          c.Database.Driver,
		Host:            c.Database.Host,
		Port:            c.Database.Port,
		Username:        c.Database.Username,
		Password:        c.Database.Password,
		DBName:          c.Database.DBName,
		SSLMode:         c.Database.SSLMode,
		Timezone:        c.Database.Timezone,
		MaxIdleConns:    c.Database.MaxIdleConns,
		MaxOpenConns:    c.Database.MaxOpenConns,
		ConnMaxLifetime: c.Database.ConnMaxLifetime,
		Enabled:         c.Database.Enabled,
	}

	cfg.Redis = RedisConfig{
		Host:         c.Redis.Host,
		Port:         c.Redis.Port,
		Password:     c.Redis.Password,
		DB:           c.Redis.DB,
		PoolSize:     c.Redis.PoolSize,
		MinIdleConns: c.Redis.MinIdleConns,
	}

	cfg.JWT = JWTConfig{
		Secret:         c.JWT.Secret,
		ExpireDays:     c.JWT.ExpireDays,
		ExpireDuration: time.Duration(c.JWT.ExpireDays) * 24 * time.Hour,
	}

	cfg.Log = LogConfig{
		Level:      c.Log.Level,
		Filename:   c.Log.Filename,
		MaxSize:    c.Log.MaxSize,
		MaxAge:     c.Log.MaxAge,
		MaxBackups: c.Log.MaxBackups,
		Compress:   c.Log.Compress,
	}

	cfg.OpenAI = OpenAIConfig{APIKey: c.OpenAI.APIKey}

	cfg.R2 = R2Config{
		AccessKeyID:     c.R2.AccessKeyID,
		SecretAccessKey: c.R2.SecretAccessKey,
		Bucket:          c.R2.Bucket,
		Region:          c.R2.Region,
		Endpoint:        c.R2.Endpoint,
		PublicURL:       c.R2.PublicURL,
		PublicDomain:    c.R2.PublicDomain,
	}

	cfg.Email = EmailConfig{
		Host:         c.Email.Host,
		Port:         c.Email.Port,
		Username:     c.Email.Username,
		Password:     c.Email.Password,
		From:         c.Email.From,
		ResendAPIKey: c.Email.ResendAPIKey,
	}

	cfg.App = AppConfig{
		Name:      c.App.Name,
		Version:   c.App.Version,
		Secret:    c.App.Secret,
		JWTSecret: c.App.JWTSecret,
		JWTExpire: time.Duration(c.App.JWTExpireDays) * 24 * time.Hour,
	}

	return cfg
}

func loadDatabaseConfig(config *Config) error {
	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return fmt.Errorf("invalid DB_PORT: %v", err)
	}

	maxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "10"))
	if err != nil {
		return fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %v", err)
	}

	maxOpenConns, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "100"))
	if err != nil {
		return fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %v", err)
	}

	connMaxLifetime, err := strconv.Atoi(getEnv("DB_CONN_MAX_LIFETIME", "3600"))
	if err != nil {
		return fmt.Errorf("invalid DB_CONN_MAX_LIFETIME: %v", err)
	}

	enabled, err := strconv.ParseBool(getEnv("DB_ENABLED", "true"))
	if err != nil {
		return fmt.Errorf("invalid DB_ENABLED: %v", err)
	}

	config.Database = DatabaseConfig{
		Driver:          getEnv("DB_DRIVER", "postgres"),
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            port,
		Username:        getEnv("DB_USERNAME", "postgres"),
		Password:        getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "gin-kit"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		Timezone:        getEnv("DB_TIMEZONE", "Asia/Shanghai"),
		MaxIdleConns:    maxIdleConns,
		MaxOpenConns:    maxOpenConns,
		ConnMaxLifetime: connMaxLifetime,
		Enabled:         enabled,
	}

	return nil
}

func loadRedisConfig(config *Config) error {
	port, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return fmt.Errorf("invalid REDIS_PORT: %v", err)
	}

	db, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return fmt.Errorf("invalid REDIS_DB: %v", err)
	}

	poolSize, err := strconv.Atoi(getEnv("REDIS_POOL_SIZE", "10"))
	if err != nil {
		return fmt.Errorf("invalid REDIS_POOL_SIZE: %v", err)
	}

	minIdleConns, err := strconv.Atoi(getEnv("REDIS_MIN_IDLE_CONNS", "5"))
	if err != nil {
		return fmt.Errorf("invalid REDIS_MIN_IDLE_CONNS: %v", err)
	}

	config.Redis = RedisConfig{
		Host:         getEnv("REDIS_HOST", "localhost"),
		Port:         port,
		Password:     getEnv("REDIS_PASSWORD", ""),
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
	}

	return nil
}

func loadJWTConfig(config *Config) error {
	expireDays, err := strconv.Atoi(getEnv("JWT_EXPIRE_DAYS", "7"))
	if err != nil {
		return fmt.Errorf("invalid JWT_EXPIRE_DAYS: %v", err)
	}

	config.JWT = JWTConfig{
		Secret:         getEnv("JWT_SECRET", ""),
		ExpireDays:     expireDays,
		ExpireDuration: time.Duration(expireDays) * 24 * time.Hour,
	}

	return nil
}

func loadLogConfig(config *Config) error {
	maxSize, err := strconv.Atoi(getEnv("LOG_MAX_SIZE", "100"))
	if err != nil {
		return fmt.Errorf("invalid LOG_MAX_SIZE: %v", err)
	}

	maxAge, err := strconv.Atoi(getEnv("LOG_MAX_AGE", "30"))
	if err != nil {
		return fmt.Errorf("invalid LOG_MAX_AGE: %v", err)
	}

	maxBackups, err := strconv.Atoi(getEnv("LOG_MAX_BACKUPS", "7"))
	if err != nil {
		return fmt.Errorf("invalid LOG_MAX_BACKUPS: %v", err)
	}

	compress, err := strconv.ParseBool(getEnv("LOG_COMPRESS", "true"))
	if err != nil {
		return fmt.Errorf("invalid LOG_COMPRESS: %v", err)
	}

	config.Log = LogConfig{
		Level:      getEnv("LOG_LEVEL", "debug"),
		Filename:   getEnv("LOG_FILENAME", "logs/app.log"),
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		Compress:   compress,
	}

	return nil
}

func loadOpenAIConfig(config *Config) error {
	config.OpenAI = OpenAIConfig{
		APIKey: getEnv("OPENAI_API_KEY", ""),
	}
	return nil
}

func loadR2Config(config *Config) error {
	config.R2 = R2Config{
		AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		Bucket:          getEnv("R2_BUCKET", ""),
		Region:          getEnv("R2_REGION", "auto"),
		Endpoint:        getEnv("R2_ENDPOINT", ""),
		PublicURL:       getEnv("R2_PUBLIC_URL", ""),
		PublicDomain:    getEnv("R2_PUBLIC_DOMAIN", ""),
	}
	return nil
}

func loadEmailConfig(config *Config) error {
	port, err := strconv.Atoi(getEnv("EMAIL_PORT", "587"))
	if err != nil {
		return fmt.Errorf("invalid EMAIL_PORT: %v", err)
	}

	config.Email = EmailConfig{
		Host:         getEnv("EMAIL_HOST", "smtp.gmail.com"),
		Port:         port,
		Username:     getEnv("EMAIL_USERNAME", ""),
		Password:     getEnv("EMAIL_PASSWORD", ""),
		From:         getEnv("EMAIL_FROM", ""),
		ResendAPIKey: getEnv("EMAIL_RESEND_API_KEY", ""),
	}
	return nil
}

func loadAppConfig(config *Config) error {
	expireDays, err := strconv.Atoi(getEnv("APP_JWT_EXPIRE_DAYS", "7"))
	if err != nil {
		return fmt.Errorf("invalid APP_JWT_EXPIRE_DAYS: %v", err)
	}

	config.App = AppConfig{
		Name:      getEnv("APP_NAME", "Llamabase"),
		Version:   getEnv("APP_VERSION", "1.0.0"),
		Secret:    getEnv("APP_SECRET", ""),
		JWTSecret: getEnv("APP_JWT_SECRET", ""),
		JWTExpire: time.Duration(expireDays) * 24 * time.Hour,
	}
	return nil
}

func loadCORSConfig(config *Config) error {
	// Parse allowed origins from environment variable (comma-separated)
	originsStr := getEnv("CORS_ALLOW_ORIGINS", "http://localhost:3000,http://localhost:3001")
	var origins []string
	if originsStr != "" {
		origins = strings.Split(originsStr, ",")
		// Trim whitespace from each origin
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
	}

	// Parse allowed methods from environment variable (comma-separated)
	methodsStr := getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,DELETE,OPTIONS")
	var methods []string
	if methodsStr != "" {
		methods = strings.Split(methodsStr, ",")
		for i, method := range methods {
			methods[i] = strings.TrimSpace(method)
		}
	}

	// Parse allowed headers from environment variable (comma-separated)
	headersStr := getEnv("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization")
	var headers []string
	if headersStr != "" {
		headers = strings.Split(headersStr, ",")
		for i, header := range headers {
			headers[i] = strings.TrimSpace(header)
		}
	}

	// Parse exposed headers from environment variable (comma-separated)
	exposeHeadersStr := getEnv("CORS_EXPOSE_HEADERS", "Content-Length")
	var exposeHeaders []string
	if exposeHeadersStr != "" {
		exposeHeaders = strings.Split(exposeHeadersStr, ",")
		for i, header := range exposeHeaders {
			exposeHeaders[i] = strings.TrimSpace(header)
		}
	}

	// Parse allow credentials from environment variable
	allowCredentials, err := strconv.ParseBool(getEnv("CORS_ALLOW_CREDENTIALS", "true"))
	if err != nil {
		return fmt.Errorf("invalid CORS_ALLOW_CREDENTIALS: %v", err)
	}

	config.CORS = CORSConfig{
		AllowOrigins:     origins,
		AllowMethods:     methods,
		AllowHeaders:     headers,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: allowCredentials,
	}

	return nil
}

func validateConfig(config *Config) error {
	// Validate required fields
	if config.Database.Enabled && config.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
