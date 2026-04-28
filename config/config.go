/*
config.go 负责从环境变量和 .env 文件加载运行配置。
*/
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config 聚合服务运行所需的配置。
type Config struct {
	Server   ServerConfig
	MySQL    MySQLConfig
	Redis    RedisConfig
	Session  SessionConfig
	Security SecurityConfig
}

// ServerConfig 定义 HTTP 服务监听地址。
type ServerConfig struct {
	Address string
}

// MySQLConfig 定义 MySQL 连接配置。
type MySQLConfig struct {
	DSN string
}

// RedisConfig 定义 Redis 连接配置。
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// SessionConfig 定义会话配置。
type SessionConfig struct {
	Expire time.Duration
}

// SecurityConfig 定义安全相关配置。
type SecurityConfig struct {
	LoginRateLimit LoginRateLimitConfig
}

// LoginRateLimitConfig 定义登录限流配置。
type LoginRateLimitConfig struct {
	MaxAttempts   int
	Window        time.Duration
	BlockDuration time.Duration
}

// New 加载并校验应用配置。
func New() (Config, error) {
	// 优先加载环境变量，环境变量不存在时才读取 .env 文件
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	// 从环境变量构建配置对象，并进行必要的类型转换和校验
	cfg := Config{
		Server: ServerConfig{
			Address: strings.TrimSpace(os.Getenv("APP_ADDR")),
		},
		MySQL: MySQLConfig{
			DSN: strings.TrimSpace(os.Getenv("MYSQL_DSN")),
		},
		Redis: RedisConfig{
			Addr:     strings.TrimSpace(os.Getenv("REDIS_ADDR")),
			Password: os.Getenv("REDIS_PASSWORD"),
		},
	}

	redisDB, err := getEnvInt("REDIS_DB", false, 0)
	if err != nil {
		return Config{}, err
	}
	cfg.Redis.DB = redisDB

	sessionExpireMinutes, err := getEnvInt("SESSION_EXPIRE_MINUTES", true, 0)
	if err != nil {
		return Config{}, err
	}
	cfg.Session.Expire = time.Duration(sessionExpireMinutes) * time.Minute

	loginRateLimitMaxAttempts, err := getEnvInt("LOGIN_RATE_LIMIT_MAX_ATTEMPTS", false, 5)
	if err != nil {
		return Config{}, err
	}
	loginRateLimitWindowMinutes, err := getEnvInt("LOGIN_RATE_LIMIT_WINDOW_MINUTES", false, 10)
	if err != nil {
		return Config{}, err
	}
	loginRateLimitBlockMinutes, err := getEnvInt("LOGIN_RATE_LIMIT_BLOCK_MINUTES", false, 15)
	if err != nil {
		return Config{}, err
	}
	cfg.Security.LoginRateLimit = LoginRateLimitConfig{
		MaxAttempts:   loginRateLimitMaxAttempts,
		Window:        time.Duration(loginRateLimitWindowMinutes) * time.Minute,
		BlockDuration: time.Duration(loginRateLimitBlockMinutes) * time.Minute,
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// validate 校验关键配置是否可用。
func validate(cfg Config) error {
	var missing []string
	if cfg.Server.Address == "" {
		missing = append(missing, "APP_ADDR")
	}
	if cfg.MySQL.DSN == "" {
		missing = append(missing, "MYSQL_DSN")
	}
	if cfg.Redis.Addr == "" {
		missing = append(missing, "REDIS_ADDR")
	}
	if cfg.Session.Expire <= 0 {
		return errors.New("SESSION_EXPIRE_MINUTES must be greater than 0")
	}
	if cfg.Security.LoginRateLimit.MaxAttempts <= 0 {
		return errors.New("LOGIN_RATE_LIMIT_MAX_ATTEMPTS must be greater than 0")
	}
	if cfg.Security.LoginRateLimit.Window <= 0 {
		return errors.New("LOGIN_RATE_LIMIT_WINDOW_MINUTES must be greater than 0")
	}
	if cfg.Security.LoginRateLimit.BlockDuration <= 0 {
		return errors.New("LOGIN_RATE_LIMIT_BLOCK_MINUTES must be greater than 0")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
	}
	return nil
}

// getEnvInt 读取整数环境变量。
func getEnvInt(key string, required bool, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		if required {
			return 0, fmt.Errorf("missing required config: %s", key)
		}
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", key)
	}
	return parsed, nil
}

// loadDotEnv 读取当前目录下的 .env 文件。
func loadDotEnv(path string) error {
	if filepath.Base(path) != path {
		return fmt.Errorf("env path must be a file name inside current directory: %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("invalid line in .env: %s", line)
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			return fmt.Errorf("empty env key in .env: %s", line)
		}

		// 已存在的环境变量优先于 .env 中的值。
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}
