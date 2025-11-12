package main

import (
	"VoteGolang/internals/app"
	"VoteGolang/internals/app/logging"
	"fmt"
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

	if err := kafkaLogger.Log("INFO", fmt.Sprintf("Kafka logger initialized with broker %s", kafkaBroker)); err != nil {
		log.Printf("Failed to send log to Kafka: %v", err)
	}

	appInstance, authUseCase, tokenManager, rdb, esClient, err := app.NewApp(kafkaLogger)
	if err != nil {
		kafkaLogger.Log("ERROR", fmt.Sprintf("Failed to initialize application: %v", err))
		log.Fatalf("Error initializing app: %v", err)
	}

	defer rdb.Close()

	if err := kafkaLogger.Log("INFO", "App started"); err != nil {
		log.Printf("Failed to send log to Kafka: %v", err)
	}

	appInstance.Run(authUseCase, tokenManager, kafkaLogger, rdb, esClient)

	if err := kafkaLogger.Log("INFO", "Application stopped gracefully"); err != nil {
		log.Printf("Failed to send log to Kafka: %v", err)
	}
}
