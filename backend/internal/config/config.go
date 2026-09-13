// Package config 集中解析环境变量配置。
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config 应用配置，全部来自环境变量。
type Config struct {
	AppName     string `env:"APP_NAME" envDefault:"aasplit"`
	Env         string `env:"ENV" envDefault:"development"`
	Port        int    `env:"PORT" envDefault:"8080"`
	DBHost      string `env:"DB_HOST" envDefault:"localhost"`
	DBPort      int    `env:"DB_PORT" envDefault:"44012"`
	DBUser      string `env:"DB_USER" envDefault:"aasplit_user"`
	DBPassword  string `env:"DB_PASSWORD" envDefault:"aasplit_pwd"`
	DBName      string `env:"DB_NAME" envDefault:"aasplit_db"`
	RedisAddr   string `env:"REDIS_ADDR" envDefault:"localhost:46312"`
	RedisPass   string `env:"REDIS_PASSWORD" envDefault:""`
	JWTSecret   string `env:"JWT_SECRET" envDefault:"change_me_to_a_long_random_string"`
	JWTExpireH  int    `env:"JWT_EXPIRE_HOURS" envDefault:"72"`
	RateLimit   int    `env:"RATE_LIMIT_PER_MIN" envDefault:"120"`
	CORSOrigins string `env:"CORS_ORIGINS" envDefault:"*"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
}

// Load 从环境变量加载配置。
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if len(cfg.JWTSecret) < 16 {
		return nil, fmt.Errorf("JWT_SECRET too short: got %d chars, need >= 16", len(cfg.JWTSecret))
	}
	return cfg, nil
}

// DSN 构造 PostgreSQL 连接串。
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

// JWTExpire 返回 JWT 有效期。
func (c *Config) JWTExpire() time.Duration {
	return time.Duration(c.JWTExpireH) * time.Hour
}
