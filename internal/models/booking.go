package models

import "time"

type Booking struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	CustomerID     uint         `json:"customer_id"`
	Customer       User         `gorm:"foreignKey:CustomerID" json:"customer"`
	TechnicianID   uint         `json:"technician_id"`
	Technician     User         `gorm:"foreignKey:TechnicianID" json:"technician"`
	AvailabilityID uint         `json:"availability_id"`
	Availability   Availability `gorm:"foreignKey:AvailabilityID" json:"availability"`
	Timeslot       time.Time    `json:"timeslot"`
	LocationLat    float64      `json:"location_lat"`
	LocationLon    float64      `json:"location_lon"`
	Description    string       `json:"description"`
	JobPrice       int64        `json:"job_price"`       
	BookingFee     int64        `json:"booking_fee"`     
	PlatformFee    int64        `json:"platform_fee"`   
	TotalAmount    int64        `json:"total_amount"`   
	Status         string       `json:"status"`       
	PaymentStatus  string       `json:"payment_status"`
	CreatedAt      time.Time    `json:"created_at"`
}
