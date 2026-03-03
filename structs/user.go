package structs

import("time")

type User struct {
    ID int `gorm:"type:integer;not null;primaryKey;" json:"id"`
    Username string `gorm:"type:varchar(255);not null" json:"username"`
    FirstName string `gorm:"type:varchar(255);not null" json:"firstname"`
    LastName string `gorm:"type:varchar(255)" json:"lastname"`
    PhoneNumber string `gorm:"type:varchar(15)" json:"phone_number,omitempty"`
    Description string `gorm:"type:text" json:"description,omitempty"`
    LinkedChannel string `gorm:"type:varchar(255)" json:"linked_channel,omitempty"`
    CreatedAt time.Time `gorm:"column:created_at;index" json:"created_at"`    
}