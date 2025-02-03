package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string
	GRPCPort    int
	Environment string
	ServiceName string
	LogLevel    string
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("GRPC_PORT", "8080"))

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://explore:explore@localhost:5432/explore?sslmode=disable"),
		GRPCPort:    port,
		Environment: getEnv("ENVIRONMENT", "development"),
		ServiceName: "explore-service",
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
