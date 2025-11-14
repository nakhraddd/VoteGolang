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

	kafkaLogger := logging.NewKafkaLogger(kafkaBroker, "app-logs")
	defer kafkaLogger.Close()

	appInstance, authUseCase, tokenManager, rdb, esClient, err := app.NewApp(kafkaLogger)
	if err != nil {
		kafkaLogger.Log("ERROR", fmt.Sprintf("Failed to initialize application: %v", err))
		log.Fatalf("Error initializing app: %v", err)
	}

	defer rdb.Close()

	appInstance.Run(authUseCase, tokenManager, kafkaLogger, rdb, esClient)
}
