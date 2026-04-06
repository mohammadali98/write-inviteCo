package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL     string
	Port            string
	ResendAPIKey    string
	ResendFromEmail string
}

func Load() Config {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "ali")
	name := getEnv("DB_NAME", "inviteandco")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", user, host, port, name, sslmode)

	return Config{
		DatabaseURL:     dsn,
		Port:            getEnv("PORT", "8080"),
		ResendAPIKey:    getEnv("RESEND_API_KEY", ""),
		ResendFromEmail: getEnv("RESEND_FROM_EMAIL", "onboarding@resend.dev"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
