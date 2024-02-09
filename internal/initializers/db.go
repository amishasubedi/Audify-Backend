package initializers

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

/*
 * Connects to the database.
 */
func ConnectDatabase() {
	dsn := os.Getenv("DB_URL")

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		fmt.Println("Failed to connect to the database:", err)
		return
	}

	fmt.Println("Successfully connected to the database")
	DB = database
}
