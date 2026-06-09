package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	ServerPort string
	ServerHost string

	JWTSecret  string
	JWTExpiry  string

	LogLevel  string
	LogFormat string
	AppEnv    string

	// S3 Configuration
	S3Region          string
	S3BucketName      string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3Endpoint        string
	S3ForcePathStyle  bool
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Don't return error if .env file doesn't exist
		fmt.Println("No .env file found, using environment variables")
	}

	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}

	// Parse S3 force path style boolean
	s3ForcePathStyle := false
	if forcePathStyle := getEnv("S3_FORCE_PATH_STYLE", "false"); forcePathStyle == "true" {
		s3ForcePathStyle = true
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     port,
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "pay_your_dues"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "localhost"),

		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-here"),
		JWTExpiry: getEnv("JWT_EXPIRY", "24h"),

		LogLevel:  getEnv("LOG_LEVEL", "debug"),
		LogFormat: getEnv("LOG_FORMAT", "console"),
		AppEnv:    getEnv("APP_ENV", "production"),

		// S3 Configuration
		S3Region:          getEnv("S3_REGION", "us-east-1"),
		S3BucketName:      getEnv("S3_BUCKET_NAME", ""),
		S3AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", ""),
		S3Endpoint:        getEnv("S3_ENDPOINT", ""),
		S3ForcePathStyle:  s3ForcePathStyle,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
} 