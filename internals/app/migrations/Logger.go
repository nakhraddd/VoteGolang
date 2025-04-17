package migrations

import (
	"gorm.io/gorm/logger"
	"log"
	"os"
	"path/filepath"
	"time"
)

func SetupDatabaseLogger() logger.Interface {

	dir := "../../internals/app/migrations/result"
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatalf("could not create migration directory: %v", err)
	}
	logFilePath := filepath.Join(dir, "gorm_queries.log")

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("could not open log file: %v", err)
	}

	newLogger := logger.New(
		log.New(logFile, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
	return newLogger
}
