package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/HabiMatch/profile-service/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	// Database connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// Open the database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Initialize PostGIS extension
	err = db.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error
	if err != nil {
		return nil, err
	}

	// Automatically migrate the models
	err = db.AutoMigrate(&models.Profile{}, &models.Geolocation{}, &models.Keeper{}, &models.Seeker{})
	if err != nil {
		return nil, err
	}

	log.Println("Database connected and models migrated successfully!")
	return db, nil
}
