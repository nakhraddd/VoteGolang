package connect

import (
	"VoteGolang/conf"
	"VoteGolang/internals/app/logging"
	"VoteGolang/internals/app/migrations"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDB(config *conf.Config, kafkaLogger *logging.KafkaLogger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DBUser, config.DBPass, config.DBHost, config.DBPort, config.DBName)
	kafkaLogger.Log("INFO", fmt.Sprintf("Attempting to connect to database %s:%s/%s", config.DBHost, config.DBPort, config.DBName))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: migrations.SetupDatabaseLogger()})
	if err != nil {
		kafkaLogger.Log("ERROR", fmt.Sprintf("Failed to connect to database: %v", err))
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}

	kafkaLogger.Log("INFO", fmt.Sprintf("Successfully connected to database: %s", config.DBName))
	log.Println("Connected to database")
	return db, nil
}
