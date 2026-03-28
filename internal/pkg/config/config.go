package config

import (
	"os"
)

// Config holds all the environment variables required to run the application.
type Config struct {
	ServerPort       string
	BankSimulatorURL string
}

// Load reads environment variables and provides sensible local defaults.
func Load() *Config {
	return &Config{
		ServerPort:       getEnvOrDefault("SERVER_PORT", "8090"),
		BankSimulatorURL: getEnvOrDefault("BANK_SIMULATOR_URL", "http://localhost:8080"),
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}