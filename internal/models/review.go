package models

import "time"

type Review struct {
	ID           uint `gorm:"primaryKey"`
	BookingID    uint
	CustomerID   uint
	TechnicianID uint
	Rating       int // 1-5 stars
	Comment      string
	CreatedAt    time.Time
}
