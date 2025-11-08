package main

import (
	"VoteGolang/internals/app"
	"VoteGolang/internals/app/logging"
	"log"
	"os"
)

// @title Online Election Vote
// @version 1.0
// @description This is the backend API for the Online Election system.
// @termsOfService http://dayus.kz
// @contact.name API Support
// @contact.email support@dauys.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092" // fallback
	}

	kafkaLogger := logging.NewKafkaLogger(kafkaBroker, "app-logs", "vote-golang-api")
	defer kafkaLogger.Close()

	appInstance, authUseCase, tokenManager, rdb, esClient, err := app.NewApp()
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}

	defer rdb.Close()

	if err := kafkaLogger.Log("INFO", "App started"); err != nil {
		log.Printf("Failed to send log to Kafka: %v", err)
	}

	appInstance.Run(authUseCase, tokenManager, kafkaLogger, rdb, esClient)
}
