package connections

import (
	models2 "VoteGolang/internals/data/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func GetDB() {
	if DB != nil {
		return
	}

	dsn := "root:Darhani2004@tcp(localhost:3306)/vote_db?charset=utf8mb4&parseTime=True&loc=Local"

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

func GetDBInstance() *gorm.DB {
	if DB == nil {
		GetDB()
	}
	return DB
}

func Migrate() {
	db := GetDBInstance()
	err := db.AutoMigrate(
		&models2.User{},
		&models2.Vote{},
		&models2.Deputy{},
		&models2.GeneralNews{},
		&models2.Petition{},
		&models2.President{},
		&models2.SessionDeputy{},
	)
	if err != nil {
		log.Fatal("Error migrating the database:", err)
	}
	fmt.Println("Database migration completed successfully!")
}

func GetUserByUsername(username string) (*models2.User, error) {
	var user models2.User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func CreateUser(user *models2.User) error {
	result := DB.Create(user)
	return result.Error
}
