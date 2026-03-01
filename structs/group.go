package structs

type Group struct { 
    ID          string `gorm:"type:varchar(25);not null;primaryKey;" json:"id"`
    Name        string `gorm:"type:varchar(255);not null" json:"name"`
    Description string `gorm:"type:text" json:"description,omitempty"`
    Subscribers int    `gorm:"type:int;not null" json:"subscribers"`
}