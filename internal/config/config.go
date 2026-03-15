//go:build ignore
// +build ignore

package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv    string
	Port      string
	DBHost    string
	DBPort    int
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string
}

func Load() Config {
	return Config{
		AppEnv:    getEnv("APP_ENV", "development"),
		Port:      getEnv("PORT", "8080"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnvInt("DB_PORT", 5432),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPass:    getEnv("DB_PASSWORD", "postgres"),
		DBName:    getEnv("DB_NAME", "taskdb"),
		DBSSLMode: getEnv("DB_SSLMODE", "disable"),
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
//go:build ignore
// +build ignore

package config






















































}	return n	}		return fallback	if err != nil {	n, err := strconv.Atoi(v)	}		return fallback	if v == "" {	v := getEnv(key, "")func getEnvInt(key string, fallback int) int {}	return fallback	}		return value	if value, ok := os.LookupEnv(key); ok && value != "" {func getEnv(key, fallback string) string {}	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)func (c Config) DatabaseURL() string {}	}		DBSSLMode: getEnv("DB_SSLMODE", "disable"),		DBName:   getEnv("DB_NAME", "taskdb"),		DBPass:   getEnv("DB_PASSWORD", "postgres"),		DBUser:   getEnv("DB_USER", "postgres"),		DBPort:   getEnvInt("DB_PORT", 5432),		DBHost:   getEnv("DB_HOST", "localhost"),		Port:     getEnv("PORT", "8080"),		AppEnv:   getEnv("APP_ENV", "development"),	return Config{func Load() Config {}	DBSSLMode string	DBName   string	DBPass   string	DBUser   string	DBPort   int	DBHost   string	Port     string	AppEnv   stringtype Config struct {)	"strconv"	"os"	"fmt"import (package config