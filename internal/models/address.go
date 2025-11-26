package models

type Address struct {
	ID      uint `gorm:"primaryKey"`
	UserID  uint
	Street  string
	City    string
	State   string
	ZipCode string
	Country string
}
