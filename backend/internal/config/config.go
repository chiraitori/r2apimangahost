package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	MongoURI              string
	DBName                string
	R2AccountID           string
	S3Endpoint            string
	S3Region              string
	R2AccessKeyID         string
	R2SecretKey           string
	R2BucketName          string
	R2PublicDomain        string // e.g. "https://pub-xxx.r2.dev" or custom CDN "https://dl.chiraitori.dev"
	AdminUsername         string
	AdminPassword         string
	JWTSecret             string
	AllowedOrigins        string
	TrustProxyHeaders     bool
	RateLimitGeneralRPM   int // Requests per minute for public endpoints (default: 300)
	RateLimitGeneralBurst int // Max burst for public endpoints (default: 60)
	RateLimitLoginRPM     int // Requests per minute for login attempts (default: 20)
	RateLimitLoginBurst   int // Max burst for login attempts (default: 10)
	RateLimitAdminRPM     int // Requests per minute for authenticated admin endpoints (default: 120)
	RateLimitAdminBurst   int // Max burst for authenticated admin endpoints (default: 20)
}

func LoadConfig() *Config {
	// Try to load .env if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] No .env file found or error reading, using system environment variables")
	}

	cfg := &Config{
		Port:                  getEnv("PORT", "8080"),
		MongoURI:              getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DBName:                getEnv("MONGODB_DB_NAME", "mangahost"),
		R2AccountID:           getEnv("R2_ACCOUNT_ID", ""),
		S3Endpoint:            strings.TrimRight(getEnv("S3_ENDPOINT", ""), "/"),
		S3Region:              getEnv("S3_REGION", "auto"),
		R2AccessKeyID:         getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretKey:           getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:          getEnv("R2_BUCKET_NAME", "mangahost"),
		R2PublicDomain:        getEnv("R2_PUBLIC_DOMAIN", ""),
		AdminUsername:         getEnv("ADMIN_USERNAME", ""),
		AdminPassword:         getEnv("ADMIN_PASSWORD", ""),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		AllowedOrigins:        getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
		TrustProxyHeaders:     getEnvBool("TRUST_PROXY_HEADERS", true),
		RateLimitGeneralRPM:   getEnvInt("RATE_LIMIT_GENERAL_RPM", 300),
		RateLimitGeneralBurst: getEnvInt("RATE_LIMIT_GENERAL_BURST", 60),
		RateLimitLoginRPM:     getEnvInt("RATE_LIMIT_LOGIN_RPM", 20),
		RateLimitLoginBurst:   getEnvInt("RATE_LIMIT_LOGIN_BURST", 10),
		RateLimitAdminRPM:     getEnvInt("RATE_LIMIT_ADMIN_RPM", 120),
		RateLimitAdminBurst:   getEnvInt("RATE_LIMIT_ADMIN_BURST", 20),
	}

	return cfg
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.AdminUsername) == "" {
		return fmt.Errorf("ADMIN_USERNAME is required")
	}
	if len(c.AdminPassword) < 12 {
		return fmt.Errorf("ADMIN_PASSWORD must contain at least 12 characters")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}
	if len(c.AllowedOriginList()) == 0 {
		return fmt.Errorf("ALLOWED_ORIGINS must contain at least one frontend origin")
	}
	if c.RateLimitGeneralRPM <= 0 || c.RateLimitGeneralBurst <= 0 ||
		c.RateLimitLoginRPM <= 0 || c.RateLimitLoginBurst <= 0 ||
		c.RateLimitAdminRPM <= 0 || c.RateLimitAdminBurst <= 0 {
		return fmt.Errorf("all rate-limit values must be greater than zero")
	}
	return nil
}

func (c *Config) AllowedOriginList() []string {
	parts := strings.Split(c.AllowedOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins = append(origins, strings.TrimRight(origin, "/"))
		}
	}
	return origins
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}
