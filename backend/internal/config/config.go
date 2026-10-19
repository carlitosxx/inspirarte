package config

import "os"

// Config centraliza las variables de entorno del servicio.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	AdminEmail  string
	AdminPass   string
}

func Load() Config {
	return Config{
		Port:        env("PORT", "8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://inspirarte:inspirarte@localhost:5432/inspirarte?sslmode=disable"),
		JWTSecret:   env("JWT_SECRET", "cambia-esto-en-produccion"),
		AdminEmail:  env("ADMIN_EMAIL", "admin@inspirarte.local"),
		AdminPass:   env("ADMIN_PASSWORD", "admin123"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
