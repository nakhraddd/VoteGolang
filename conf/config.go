package conf

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
}

func LoadConfig() *Config {
	// Загружаем .env файл, если он есть
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	} else {
		log.Println(".env file FOUND!")
	}

	return &Config{
		JWTSecret: getEnv("JWT_SECRET", "defaultsecret"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "3306"),
		DBUser:    getEnv("DB_USER", "root"),
		DBPass:    getEnv("DB_PASS", "$F00tba11!"),
		DBName:    getEnv("DB_NAME", "vote_database"),
	}
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Warning: Environment variable %s not set, using default value", key)
		return fallback
	}
	return value
}
