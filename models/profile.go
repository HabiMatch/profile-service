package models

import (
	"time"
)

type Profile struct {
	ID             uint         `gorm:"primaryKey;autoIncrement"`
	UserID         string       `json:"userid" gorm:"type:varchar(255);not null;unique;index"` // Firebase UserID
	FirstName      string       `json:"firstname" gorm:"type:varchar(100);not null"`
	LastName       string       `json:"lastname" gorm:"type:varchar(100);not null"`
	Gender         bool         `json:"gender" gorm:"type:boolean;not null"`
	Occupation     string       `json:"occupation" gorm:"type:varchar(100);not null"`
	Address        string       `json:"address" gorm:"type:varchar(255);not null"`
	Contactno      string       `json:"contactno" gorm:"type:varchar(20);not null"`
	ProfilePicture string       `json:"profile_picture" gorm:"type:varchar(255)"`
	Description    string       `json:"description" gorm:"type:text"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	Latitude       float64      `json:"latitude"`
	Longitude      float64      `json:"longitude"`
	Geolocation    *Geolocation `gorm:"foreignKey:UserID;references:UserID"` // One-to-One relationship
	Selftags       []string     `json:"selftags" gorm:"type:varchar(255)[]"`
}
