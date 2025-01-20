package models

type Keeper struct {
	ID             uint     `gorm:"primaryKey;autoIncrement"`
	UserID         string   `json:"userid" gorm:"type:varchar(255);not null;unique;index"` // Firebase UserID
	RentPerPerson  int      `json:"rentperperson" gorm:"type:int;not null"`
	LookingFor     string   `json:"lookingfor" gorm:"type:varchar(100);not null"`
	FlatImages     []string `json:"flatimages" gorm:"type:varchar(255)[]"`
	FlatHighlights []string `json:"highlights" gorm:"type:varchar(255)[]"`
	Amenities      []string `json:"amenities" gorm:"type:varchar(255)[]"`
	Address        string   `json:"address" gorm:"type:varchar(255);not null"`
	Description    string   `json:"description" gorm:"type:text"`
}
