package config

import "os"

type Config struct {
	Env string
}

func Load() Config {
	return Config{
		Env: getEnv("APP_ENV", "dev"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
