package main

import (
	"VoteGolang/internals/app"
	"VoteGolang/internals/app/logging"
	"log"
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
	kafkaLogger := logging.NewKafkaLogger("kafka:9092", "app-logs")
	defer kafkaLogger.Close()

	appInstance, authUseCase, tokenManager, err := app.NewApp()
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}

	if err := kafkaLogger.Log("App started"); err != nil {
		log.Printf("Failed to send log to Kafka: %v", err)
	}

	appInstance.Run(authUseCase, tokenManager)
}
