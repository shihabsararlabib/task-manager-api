package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv            string
	Port              string
	DBHost            string
	DBPort            int
	DBUser            string
	DBPass            string
	DBName            string
	DBSSLMode         string
	JWTSecret         string
	AccessTokenTTLMin int
	RefreshTokenTTLHr int
}

func Load() Config {
	return Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		Port:              getEnv("PORT", "8080"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnvInt("DB_PORT", 5432),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPass:            getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "taskdb"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		JWTSecret:         getEnv("JWT_SECRET", "replace-me-in-production"),
		AccessTokenTTLMin: getEnvInt("ACCESS_TOKEN_TTL_MIN", 30),
		RefreshTokenTTLHr: getEnvInt("REFRESH_TOKEN_TTL_HR", 168),
	}
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := getEnv(key, "")
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
