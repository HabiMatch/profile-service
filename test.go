package main

import (
	"fmt"
	"log"
	"os"

	"github.com/HabiMatch/profile-service/utils"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func test() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize PostgreSQL connection
	db, err := utils.InitDB()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer func() {
		conn, _ := db.DB()
		_ = conn.Close()
	}()

	// Initialize tables and seed data
	// if err := initTablesAndData(db); err != nil {
	// 	log.Fatalf("Error initializing tables and data: %v", err)
	// }

	if err := fetchData(db); err != nil {
		log.Fatalf("Error fetching data: %v", err)
	}
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis;")
	port := os.Getenv("PORT")
	log.Printf("Server running on port %s", port)
}
func fetchData(db *gorm.DB) error {
	// Define table structures
	type User struct {
		ID       uint `gorm:"primaryKey"`
		Name     string
		Email    string
		Password string
	}

	type Profile struct {
		ID      uint `gorm:"primaryKey"`
		UserID  uint
		Bio     string
		Website string
	}

	// Fetch all users
	var users []User
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %v", err)
	}

	// Fetch all profiles
	var profiles []Profile
	if err := db.Find(&profiles).Error; err != nil {
		return fmt.Errorf("failed to fetch profiles: %v", err)
	}

	// Print fetched data
	fmt.Println("Users:")
	for _, user := range users {
		fmt.Printf("ID: %d, Name: %s, Email: %s\n", user.ID, user.Name, user.Email)
	}

	fmt.Println("\nProfiles:")
	for _, profile := range profiles {
		fmt.Printf("ID: %d, UserID: %d, Bio: %s, Website: %s\n", profile.ID, profile.UserID, profile.Bio, profile.Website)
	}

	return nil
}

// initTablesAndData creates tables and inserts initial data
func initTablesAndData(db *gorm.DB) error {
	// Define table structures
	type User struct {
		ID       uint   `gorm:"primaryKey"`
		Name     string `gorm:"size:255;not null"`
		Email    string `gorm:"size:255;unique;not null"`
		Password string `gorm:"size:255;not null"`
	}

	type Profile struct {
		ID      uint   `gorm:"primaryKey"`
		UserID  uint   `gorm:"not null"`
		Bio     string `gorm:"size:500"`
		Website string `gorm:"size:255"`
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(&User{}, &Profile{}); err != nil {
		return err
	}

	// Seed initial data
	users := []User{
		{Name: "John Doe", Email: "john@example.com", Password: "password123"},
		{Name: "Jane Smith", Email: "jane@example.com", Password: "securepassword"},
	}

	profiles := []Profile{
		{UserID: 1, Bio: "Software engineer from NY.", Website: "https://johndoe.com"},
		{UserID: 2, Bio: "Digital marketer from LA.", Website: "https://janesmith.com"},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	for _, profile := range profiles {
		if err := db.Create(&profile).Error; err != nil {
			return err
		}
	}

	log.Println("Tables created and data inserted successfully!")
	return nil
}
