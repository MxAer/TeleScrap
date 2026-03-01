package structs

import("time")


type Link struct{
    ID string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Link string `gorm:"varchar(255);not null" json:"link"`
    FirstEncounterMessage string  `gorm:"column:tg_message_id" json:"tg_message_id"`
    OtherMessages []string  `gorm:"type:jsonb;serializer:json;column:other_messages" json:"other_messages"` 
    CreatedAt time.Time `gorm:"column:created_at;index" json:"created_at"`   
} 