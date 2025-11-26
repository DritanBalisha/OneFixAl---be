package models

import (
	"time"

	"gorm.io/gorm"
)

type Availability struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	TechnicianID uint           `json:"technician_id"`
	Technician   User           `gorm:"foreignKey:TechnicianID" json:"technician"`
	DayOfWeek    int            `json:"dayOfWeek"` // 0=Sunday, 6=Saturday
	StartTime    string         `json:"startTime"` // format: "09:00"
	EndTime      string         `json:"endTime"`   // format: "17:00"
	IsBooked     bool           `json:"is_booked"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
