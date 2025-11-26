// booking status Controller
package api

import (
	"net/http"

	"OneFixAL/internal/db"
	"OneFixAL/internal/models"
	"OneFixAL/internal/websocket"

	"github.com/gin-gonic/gin"
)

type UpdateBookingStatusInput struct {
	Status string `json:"status"` // confirmed, cancelled

}

func UpdateBookingStatus(c *gin.Context) {
	id := c.Param("id") // booking ID
	var input UpdateBookingStatusInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var booking models.Booking
	if err := db.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	booking.Status = input.Status
	db.DB.Save(&booking)

	// 🔓 If cancelled, free the slot again
	if input.Status == "cancelled" {
		db.DB.Model(&models.Availability{}).
			Where("id = ?", booking.AvailabilityID).
			Update("is_booked", false)
	}

	// if Confirmed
	if input.Status == "confirmed" {
		websocket.SendNotification(booking.CustomerID, "Your booking has been confirmed!")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking status updated",
		"booking": booking,
	})
}

func GetMyBookings(c *gin.Context) {
	userID, _ := c.Get("userID")
	role, _ := c.Get("role") // You can add role in your AuthMiddleware

	var bookings []models.Booking

	query := db.DB.Model(&models.Booking{})

	if role == "technician" {
		query = query.Where("technician_id = ?", userID)
	} else {
		query = query.Where("customer_id = ?", userID)
	}

	if err := query.Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load bookings"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func GetTechnicianBookings(c *gin.Context) {
	technicianID := c.GetInt("userID")

	var bookings []models.Booking

	err := db.DB.
		Preload("Customer").
		Preload("Availability").
		Where("technician_id = ?", technicianID).
		Find(&bookings).Error

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, bookings)
}

// Notifications- -- - -- -

func GetNotifications(c *gin.Context) {
	userID := c.GetInt("userID")

	var notifs []models.Notification
	db.DB.Where("user_id = ? AND seen = ?", userID, false).Find(&notifs)

	c.JSON(200, notifs)
}

func MarkNotificationSeen(c *gin.Context) {
	id := c.Param("id")
	db.DB.Model(&models.Notification{}).Where("id = ?", id).Update("seen", true)
	c.JSON(200, gin.H{"message": "Notification marked as seen"})
}
