package models

import (
	"time"

	"gorm.io/datatypes"
)

type Profile struct {
	ID                 uint           `gorm:"primaryKey;autoIncrement"`
	UserID             string         `json:"userid" gorm:"type:varchar(255);not null;unique;index"` // Firebase UserID
	Name               string         `json:"name" gorm:"type:varchar(100);not null"`
	Smoking            bool           `json:"smoking"`
	Drinking           bool           `json:"drinking"`
	Cleanliness        int            `json:"cleanliness" gorm:"type:int;not null"` // Scale: 1-5
	WorkingHours       datatypes.JSON `json:"working_hours" gorm:"type:json"`       // JSON field
	Hobbies            datatypes.JSON `json:"hobbies" gorm:"type:json"`             // JSON field
	FavoriteActivities datatypes.JSON `json:"favorite_activities" gorm:"type:json"` // JSON field
	MusicPreferences   datatypes.JSON `json:"music_preferences" gorm:"type:json"`   // JSON field
	GenderPreference   datatypes.JSON `json:"gender_preference" gorm:"type:json"`
	PetFriendly        bool           `json:"pet_friendly"`
	FoodHabits         datatypes.JSON `json:"food_habits" gorm:"type:json"` // JSON field
	Location           string         `json:"location" gorm:"type:varchar(255)"`
	ProfilePicture     string         `json:"profile_picture" gorm:"type:varchar(255)"`
	Description        string         `json:"description" gorm:"type:text"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	Latitude           float64        `json:"latitude"`
	Longitude          float64        `json:"longitude"`
	Geolocation        *Geolocation   `gorm:"foreignKey:UserID;references:UserID"` // One-to-One relationship
}
