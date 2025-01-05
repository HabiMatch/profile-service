package models

type Geolocation struct {
	ID        uint    `gorm:"primaryKey;autoIncrement"`
	UserID    string  `json:"user_id" gorm:"type:varchar(255);not null;unique;index"` // Firebase UserID
	Latitude  float64 `json:"latitude" gorm:"not null"`                               // Precision for mapping
	Longitude float64 `json:"longitude" gorm:"not null"`                              // Precision for mapping
	Location  string  `json:"location" gorm:"type:geography(Point,4326);not null"`    // Geospatial index
}
