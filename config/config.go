package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string

	DatabaseURL string

	ContextRecentChunks  int
	ContextWorkerInterval time.Duration

	AIProvider string
	AIAPIKey   string

	WebSocketMaxConnections int
}

var AppConfig *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env file not found, reading from environment")
	}

	AppConfig = &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		ContextRecentChunks:  getEnvInt("CONTEXT_RECENT_CHUNKS", 5),
		ContextWorkerInterval: getEnvDuration("CONTEXT_WORKER_INTERVAL", "1s"),

		AIProvider: getEnv("AI_PROVIDER", "mock"),
		AIAPIKey:   getEnv("AI_API_KEY", ""),

		WebSocketMaxConnections: getEnvInt("WEBSOCKET_MAX_CONNECTIONS", 100),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback string) time.Duration {
	val := getEnv(key, fallback)
	d, err := time.ParseDuration(val)
	if err != nil {
		log.Printf("[config] invalid duration for %s: %s, using fallback %s", key, val, fallback)
		d, _ = time.ParseDuration(fallback)
	}
	return d
}

