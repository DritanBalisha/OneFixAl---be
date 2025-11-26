package models

import "time"

type Payment struct {
	ID            uint `gorm:"primaryKey"`
	BookingID     uint
	Amount        int64  // cents
	Currency      string // e.g., "EUR", "USD"
	Status        string // pending, paid, failed, refunded
	Provider      string // e.g., "Stripe", "PayPal"
	TransactionID string // from payment provider
	CreatedAt     time.Time
}
