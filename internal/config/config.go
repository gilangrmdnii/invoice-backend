package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort      string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBURL        string
	JWTSecret    string
	JWTExpiryHrs int
}

func Load() (*Config, error) {
	if os.Getenv("PORT") == "" {
		_ = godotenv.Load()
	}

	expiryHrs, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))

	return &Config{
		AppPort:      getEnv("APP_PORT", "3000"),
		DBHost:       getEnvMulti([]string{"DB_HOST", "MYSQLHOST"}, "localhost"),
		DBPort:       getEnvMulti([]string{"DB_PORT", "MYSQLPORT"}, "3306"),
		DBUser:       getEnvMulti([]string{"DB_USER", "MYSQLUSER"}, "root"),
		DBPassword:   getEnvMulti([]string{"DB_PASSWORD", "MYSQLPASSWORD"}, ""),
		DBName:       getEnvMulti([]string{"DB_NAME", "MYSQL_DATABASE"}, "invoice_db"),
		DBURL:        getEnv("MYSQL_URL", ""),
		JWTSecret:    getEnv("JWT_SECRET", ""),
		JWTExpiryHrs: expiryHrs,
	}, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func getEnvMulti(keys []string, fallback string) string {
	for _, key := range keys {
		if val, ok := os.LookupEnv(key); ok && val != "" {
			return val
		}
	}
	return fallback
}
