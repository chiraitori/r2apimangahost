package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	MongoURI        string
	DBName          string
	R2AccountID     string
	R2AccessKeyID   string
	R2SecretKey     string
	R2BucketName    string
	R2PublicDomain  string // e.g. "https://pub-xxx.r2.dev" or custom CDN "https://cdn.domain.com"
	AdminUsername   string
	AdminPassword   string
	JWTSecret       string
	AllowedOrigins  string
}

func LoadConfig() *Config {
	// Try to load .env if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] No .env file found or error reading, using system environment variables")
	}

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		MongoURI:       getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DBName:         getEnv("MONGODB_DB_NAME", "mangahost"),
		R2AccountID:    getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:  getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretKey:    getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:   getEnv("R2_BUCKET_NAME", "mangahost"),
		R2PublicDomain: getEnv("R2_PUBLIC_DOMAIN", ""),
		AdminUsername:  getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", "admin123456"),
		JWTSecret:      getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}
