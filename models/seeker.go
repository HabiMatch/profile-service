package models

type Seeker struct {
	ID          uint     `gorm:"primaryKey;autoIncrement"`
	UserID      string   `json:"userid" gorm:"type:varchar(255);not null;unique;index"` // Firebase UserID
	LookingFor  string   `json:"lookingfor" gorm:"type:varchar(100);not null"`
	Highlights  []string `json:"highlights" gorm:"type:varchar(255)[]"`
	Description string   `json:"description" gorm:"type:text"`
}
