package structs

import (
    "time"
)

type Message struct {
    ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    TGID        int       `gorm:"type:integer;not_null" json:"tgid"`
    GroupID       string    `gorm:"type:varchar(36);column:group_id" json:"group_id,omitempty"`
    SenderID      string    `gorm:"type:varchar(36);column:sender_id;not null" json:"sender_id"`
    Message       string    `gorm:"type:text;column:message" json:"message"`
    Media         []string  `gorm:"type:jsonb;serializer:json;column:media" json:"media"` 
    CreatedAt     time.Time `gorm:"column:created_at;index" json:"created_at"`
    ReplyTo       string    `gorm:"column:reply_to_id" json:"reply_to_id,omitempty"`  
    IsPinned      bool      `gorm:"type:bool;column:is_pinned;default:false" json:"is_pinned"`      
    ParentMessage string    `gorm:"column:parent_message" json:"parent_message"`
}

func (Message) TableName() string {
    return "messages"
}

