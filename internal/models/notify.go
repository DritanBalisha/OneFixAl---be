// models/notification.go
package models

import "time"

type Notification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Message   string    `json:"message"`
	Seen      bool      `json:"seen" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}
