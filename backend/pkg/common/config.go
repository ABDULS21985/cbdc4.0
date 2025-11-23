package common

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	FabricConfig string
	MSP          string
	CertPath     string
	KeyPath      string
}

func LoadConfig() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		FabricConfig: getEnv("FABRIC_CONFIG", "connection-profile.yaml"),
		MSP:          getEnv("MSP_ID", "CentralBankMSP"),
		CertPath:     getEnv("CERT_PATH", ""),
		KeyPath:      getEnv("KEY_PATH", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return fallback
}
