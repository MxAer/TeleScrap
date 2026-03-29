package structs

import (
	"time"
)

type Message struct {
	TGID          int64     `gorm:"type:bigint;not null" json:"tgid"`
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GroupID       string    `gorm:"index" json:"group_id,omitempty"`
	SenderID      string    `gorm:"type:varchar(36);not null;index" json:"sender_id"`
	Message       string    `gorm:"type:text" json:"message"`
	Media         []string  `gorm:"type:jsonb;serializer:json" json:"media"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
	ReplyTo       string    `gorm:"column:reply_to_id" json:"reply_to_id,omitempty"`
	IsPinned      bool      `gorm:"default:false" json:"is_pinned"`
	ParentMessage string    `gorm:"column:parent_message" json:"parent_message"`
}

func (Message) TableName() string {
	return "messages"
}
