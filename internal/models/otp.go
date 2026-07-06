package models

import "time"

type OTPCode struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Code      string    `json:"-"`       // hashed, never expose
	Purpose   string    `json:"purpose"` // "reset_password", "verify_email", "2fa"
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `gorm:"default:false"`
	CreatedAt time.Time
}
