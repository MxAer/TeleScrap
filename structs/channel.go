package structs

import (
	"time"
)

type Channel struct {
	ID          int
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Subscribers int       `gorm:"type:bigint;not null" json:"subscribers"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null" json:"created_at"`
}
