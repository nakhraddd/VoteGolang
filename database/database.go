package database

import (
	"VoteGolang/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// GetDB initializes the database connection and stores it in the global DB variable.
func GetDB() {
	if DB != nil {
		return // If DB is already initialized, no need to reconnect
	}

	// Setup connection string
	dsn := "root:Darhani2004@tcp(localhost:3306)/vote_db?charset=utf8mb4&parseTime=True&loc=Local"

	// Retry logic (3 retries)
	var err error
	for i := 0; i < 3; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			fmt.Println("Database connected successfully!")
			return
		}
		log.Printf("Failed to connect to database (Attempt %d/3): %v\n", i+1, err)
	}
	log.Fatal("Failed to connect to database after 3 attempts:", err)
}

// GetDBInstance returns the initialized database instance.
func GetDBInstance() *gorm.DB {
	// Ensure DB is initialized before use
	if DB == nil {
		GetDB() // Initialize DB if not already done
	}
	return DB
}

func Migrate() {
	// Get DB instance
	db := GetDBInstance()
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Error migrating the database:", err) // Log the migration error
	}
	fmt.Println("Database migration completed successfully!")
}
func GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func CreateUser(user *models.User) error {
	result := DB.Create(user)
	return result.Error
}
