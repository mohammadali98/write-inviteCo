package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL     string
	Port            string
	ResendAPIKey    string
	ResendFromEmail string
	AdminEmail      string
	AdminUser       string
	AdminPass       string
}

func Load() Config {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		host := strings.TrimSpace(os.Getenv("DB_HOST"))
		port := getEnv("DB_PORT", "5432")
		user := strings.TrimSpace(os.Getenv("DB_USER"))
		password := os.Getenv("DB_PASSWORD")
		name := strings.TrimSpace(os.Getenv("DB_NAME"))
		sslmode := getEnv("DB_SSLMODE", "disable")

		if host != "" && user != "" && name != "" {
			dsn = buildDatabaseURL(user, password, host, port, name, sslmode)
		}
	}

	return Config{
		DatabaseURL:     dsn,
		Port:            getEnv("PORT", "8080"),
		ResendAPIKey:    getEnv("RESEND_API_KEY", ""),
		ResendFromEmail: getEnv("RESEND_FROM_EMAIL", "onboarding@resend.dev"),
		AdminEmail:      strings.TrimSpace(os.Getenv("ADMIN_EMAIL")),
		AdminUser:       strings.TrimSpace(os.Getenv("ADMIN_USER")),
		AdminPass:       strings.TrimSpace(os.Getenv("ADMIN_PASS")),
	}
}

func buildDatabaseURL(user string, password string, host string, port string, name string, sslmode string) string {
	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   "/" + strings.TrimPrefix(name, "/"),
	}

	switch {
	case user != "" && password != "":
		u.User = url.UserPassword(user, password)
	case user != "":
		u.User = url.User(user)
	}

	query := u.Query()
	query.Set("sslmode", sslmode)
	u.RawQuery = query.Encode()

	return u.String()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
