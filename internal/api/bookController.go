package api

import (
	"net/http"
	"time"

	"OneFixAL/internal/db"
	"OneFixAL/internal/models"

	"github.com/gin-gonic/gin"
)

type CreateBookingRequest struct {
	TechnicianID   uint      `json:"technician_id"`
	AvailabilityID uint      `json:"availability_id"`
	Timeslot       time.Time `json:"timeslot"`
	LocationLat    float64   `json:"location_lat"`
	LocationLon    float64   `json:"location_lon"`
	Description    string    `json:"description"`
	BookingFee     int64     `json:"booking_fee"`
}

func CreateBooking(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 🔍 Check if slot exists
	var slot models.Availability
	if err := db.DB.First(&slot, req.AvailabilityID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Slot not found"})
		return
	}

	// 🚫 Prevent double-booking
	if slot.IsBooked {
		c.JSON(http.StatusConflict, gin.H{"error": "This time slot is already booked"})
		return
	}
	booking := models.Booking{
		CustomerID:     userID.(uint),
		TechnicianID:   req.TechnicianID,
		AvailabilityID: req.AvailabilityID,
		Timeslot:       req.Timeslot,
		LocationLat:    req.LocationLat,
		LocationLon:    req.LocationLon,
		Description:    req.Description,
		BookingFee:     req.BookingFee,
		Status:         "pending",
		CreatedAt:      time.Now(),
	}

	if err := db.DB.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	// 🔒 Mark slot as booked
	slot.IsBooked = true
	db.DB.Save(&slot)

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking created successfully",
		"booking": booking,
	})
}
