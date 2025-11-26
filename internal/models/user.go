package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint   `json:"id", gorm:"primaryKey"`
	Name      string `json:"name"`
	Email     string `json:"email", gorm:"uniqueIndex"`
	Phone     string `json:"phone",gorm:"uniqueIndex"`
	Role      string `json:"role"` // customer, technician, admin
	Password  string
	CreatedAt time.Time

	TechnicianProfile TechnicianProfile `gorm:"foreignKey:UserID"`
}

// Fetch a single user by ID
func GetUserByID(db *gorm.DB, id interface{}) (*User, error) {
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Fetch all technicians
func GetTechnicians(db *gorm.DB) ([]User, error) {
	var users []User
	if err := db.Where("role = ?", "technician").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
