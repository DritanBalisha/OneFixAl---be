package models

import "time"

type Booking struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	CustomerID     uint         `json:"customer_id"`
	Customer       User         `gorm:"foreignKey:CustomerID" json:"customer"` // ✅ add relation
	TechnicianID   uint         `json:"technician_id"`
	Technician     User         `gorm:"foreignKey:TechnicianID" json:"technician"` // ✅ add relation
	AvailabilityID uint         `json:"availability_id"`
	Availability   Availability `gorm:"foreignKey:AvailabilityID" json:"availability"` // ✅ add relation
	Timeslot       time.Time    `json:"timeslot"`
	LocationLat    float64      `json:"location_lat"`
	LocationLon    float64      `json:"location_lon"`
	Description    string       `json:"description"`
	BookingFee     int64        `json:"booking_fee"`
	Status         string       `json:"status"`
	CreatedAt      time.Time    `json:"created_at"`
}

// // internal/models/booking.go
// package models

// import "time"

// type Booking struct {
// 	ID             uint      `gorm:"primaryKey" json:"id"`
// 	CustomerID     uint      `json:"customer_id"`
// 	TechnicianID   uint      `json:"technician_id"`
// 	AvailabilityID uint      `json:"availability_id"`
// 	Timeslot       time.Time `json:"timeslot"`
// 	LocationLat    float64   `json:"location_lat"`
// 	LocationLon    float64   `json:"location_lon"`
// 	Description    string    `json:"description"`
// 	BookingFee     int64     `json:"booking_fee"`
// 	Status         string    `json:"status"` // e.g. pending, confirmed, completed, cancelled
// 	CreatedAt      time.Time `json:"created_at"`

// 	// ✅ For preload (joins)
// 	Customer     User         `gorm:"foreignKey:CustomerID" json:"customer"`
// 	Availability Availability `gorm:"foreignKey:AvailabilityID" json:"availability"`
// }
