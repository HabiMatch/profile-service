package utils

import (
	"fmt"
	"log"

	"github.com/HabiMatch/profile-service/models"
	"gorm.io/gorm"
)

// Store geolocation as a point.
func StoreGeolocation(db *gorm.DB, userid string, latitude, longitude float64) error {

	location := fmt.Sprintf("POINT(%f %f)", longitude, latitude)
	geolocation := models.Geolocation{
		UserID:    userid,
		Location:  location,
		Latitude:  latitude,
		Longitude: longitude,
	}

	if result := db.Create(&geolocation); result.Error != nil {
		log.Println("Failed to store geolocation:", result.Error)
		return result.Error
	}

	return nil
}

func UpdateGeolocation(db *gorm.DB, userid string, newLatitude, newLongitude float64) error {
	println("Updating geolocation for user: ", userid)
	newLocation := fmt.Sprintf("POINT(%f %f)", newLongitude, newLatitude)
	result := db.Model(&models.Geolocation{}).
		Where("user_id = ?", userid).
		Updates(models.Geolocation{
			Location:  newLocation,
			Latitude:  newLatitude,
			Longitude: newLongitude,
		})

	if result.Error != nil {
		log.Println("Failed to update geolocation:", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.Println("No record found for the given userid to update geolocation")
		return fmt.Errorf("no record found for userid: %s", userid)
	}
	return nil
}

// (distance-based query).
func GetProfilesWithinRadius(db *gorm.DB, latitude, longitude, radius float64) ([]models.Profile, error) {
	query := `
		SELECT p.* 
		FROM profiles p
		JOIN geolocations g ON p.id = g.profile_id
		WHERE ST_DWithin(
			g.location::geography, 
			ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography, 
			?
		)
	`
	var profiles []models.Profile
	if err := db.Raw(query, longitude, latitude, radius).Scan(&profiles).Error; err != nil {
		log.Println("Failed to fetch profiles within radius:", err)
		return nil, err
	}

	return profiles, nil
}
