package conf

import (
	"VoteGolang/internals/app/logging"
	"fmt"
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

func LoadConfig(kafkaLogger *logging.KafkaLogger) *Config {
	// Загружаем .env файл, если он есть
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
		kafkaLogger.Log("WARN", "No .env file found, using system environment variables")
	} else {
		log.Println(".env file FOUND!")
		kafkaLogger.Log("INFO", ".env file successfully loaded")
	}

	cfg := &Config{
		JWTSecret: getEnv("JWT_SECRET", "defaultsecret", kafkaLogger),
		DBHost:    getEnv("DB_HOST", "localhost", kafkaLogger),
		DBPort:    getEnv("DB_PORT", "3306", kafkaLogger),
		DBUser:    getEnv("DB_USER", "root", kafkaLogger),
		DBPass:    getEnv("DB_PASS", "$F00tba11!", kafkaLogger),
		DBName:    getEnv("DB_NAME", "vote_database", kafkaLogger),
	}

	kafkaLogger.Log("INFO", fmt.Sprintf("Configuration loaded successfully for DB %s:%s", cfg.DBHost, cfg.DBPort))
	return cfg
}

func getEnv(key, fallback string, kafkaLogger *logging.KafkaLogger) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		msg := fmt.Sprintf("Environment variable %s not set, using default value", key)
		log.Println(msg)
		kafkaLogger.Log("WARN", msg)
		return fallback
	}
	kafkaLogger.Log("DEBUG", fmt.Sprintf("Environment variable %s loaded", key))
	return value
}
