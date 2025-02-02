package models

type UserCredentials struct {
	ID     uint   `gorm:"primaryKey;autoIncrement"`
	UserID string `json:"userid" gorm:"type:varchar(255);not null;unique;index"` // Firebase UserID
}
