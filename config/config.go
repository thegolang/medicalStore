package config

import "os"

type Config struct {
	MongoURI      string
	MongoDB       string
	SessionSecret string
	Port          string
}

func Load() *Config {
	return &Config{
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:       getEnv("MONGO_DB", "medicalstore"),
		SessionSecret: getEnv("SESSION_SECRET", "super-secret-key-dev"),
		Port:          getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
