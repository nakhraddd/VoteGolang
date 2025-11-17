package conf

import (
	"VoteGolang/internals/app/logging"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type BnbConfig struct {
	NodeURL         string
	PrivateKey      string
	ContractAddress string
	ChainID         int64
}

type Config struct {
	JWTSecret string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	BNB       *BnbConfig // Added
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

	return &Config{
		JWTSecret: getEnv("JWT_SECRET", "defaultsecret"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "3306"),
		DBUser:    getEnv("DB_USER", "root"),
		DBPass:    getEnv("DB_PASS", "$F00tba11!"),
		DBName:    getEnv("DB_NAME", "vote_database"),
		BNB: &BnbConfig{
			NodeURL:         os.Getenv("BNB_NODE_URL"),
			PrivateKey:      os.Getenv("BNB_PRIVATE_KEY"),
			ContractAddress: os.Getenv("BNB_CONTRACT_ADDRESS"),
			ChainID:         getEnvAsInt64("BNB_CHAIN")},
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

func getEnvAsInt64(key string) int64 {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Warning: Environment variable %s not set, using default value %d", key)
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		log.Printf("Warning: Invalid format for %s (expected integer, got '%s'), using default value %d. Error: %v", key, valueStr, err)
	}

	return value
}
