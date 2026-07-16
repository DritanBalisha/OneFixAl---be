package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"OneFixAL/internal/db"
	"OneFixAL/internal/models"
	"OneFixAL/internal/websocket"
	"github.com/gin-gonic/gin"
)

const PlatformFeePercent = 2
const BookingFeePercent  = 10

type CreateBookingRequest struct {
	TechnicianID   uint    `json:"technician_id"`
	AvailabilityID uint    `json:"availability_id"`
	LocationLat    float64 `json:"location_lat"`
	LocationLon    float64 `json:"location_lon"`
	Description    string  `json:"description"`
	// ✅ No job_price here anymore — technician sets it
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

	if userID.(uint) == req.TechnicianID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot book your own service"})
		return
	}

	if strings.TrimSpace(req.Description) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please describe your problem"})
		return
	}

	var slot models.Availability
	if err := db.DB.First(&slot, req.AvailabilityID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Slot not found"})
		return
	}

	if slot.IsBooked {
		c.JSON(http.StatusConflict, gin.H{"error": "This time slot is already booked"})
		return
	}

	booking := models.Booking{
		CustomerID:     userID.(uint),
		TechnicianID:   req.TechnicianID,
		AvailabilityID: req.AvailabilityID,
		LocationLat:    req.LocationLat,
		LocationLon:    req.LocationLon,
		Description:    req.Description,
		JobPrice:       0,
		BookingFee:     0,
		PlatformFee:    0,
		TotalAmount:    0,
		PriceSetByTech: false,
		Status:         "pending",
		PaymentStatus:  "unpaid",
		CreatedAt:      time.Now(),
	}

	if err := db.DB.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	slot.IsBooked = true
	db.DB.Save(&slot)

	// Notify technician by email
	var customer models.User
	var technician models.User
	db.DB.First(&customer, booking.CustomerID)
	db.DB.First(&technician, booking.TechnicianID)

	go sendEmail(
		technician.Email,
		"OneFixAL - New Booking Request",
		fmt.Sprintf(
			"New booking from %s!\n\nProblem: %s\n\nLog in to set your price and accept or decline.",
			customer.Name, booking.Description,
		),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking sent! The technician will review and set a price.",
		"booking": booking,
	})
}

// ── NEW: Technician sets price ───────────────────────────────────
// POST /bookings/:id/set-price
// Body: { "job_price": 3000 }
func SetBookingPrice(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	var req struct {
		JobPrice int64 `json:"job_price"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.JobPrice <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid job price is required"})
		return
	}

	var booking models.Booking
	if err := db.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Only the assigned technician can set price
	if booking.TechnicianID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your booking"})
		return
	}

	if booking.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price can only be set on pending bookings"})
		return
	}

	// Calculate fees
	bookingFee  := (req.JobPrice * int64(BookingFeePercent)) / 100
	platformFee := (req.JobPrice * int64(PlatformFeePercent)) / 100

	booking.JobPrice       = req.JobPrice
	booking.BookingFee     = bookingFee
	booking.PlatformFee    = platformFee
	booking.TotalAmount    = req.JobPrice
	booking.PriceSetByTech = true
	booking.Status         = "price_set"

	if err := db.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set price"})
		return
	}

	// Notify client via websocket + email
	var customer models.User
	db.DB.First(&customer, booking.CustomerID)

	websocket.SendNotification(booking.CustomerID, fmt.Sprintf(
		"Technician set a price of %d LEK for your booking. Log in to accept!", booking.JobPrice,
	))

	go sendEmail(
		customer.Email,
		"OneFixAL - Technician Set a Price for Your Booking",
		fmt.Sprintf(
			"Your technician reviewed your request and set a price!\n\n"+
				"Job Price:    %d LEK\n"+
				"Booking Fee:  %d LEK (10%% deposit due now)\n"+
				"Platform Fee: %d LEK (2%% OneFixAL)\n"+
				"Total:        %d LEK\n\n"+
				"Log in to accept or cancel.",
			booking.JobPrice, booking.BookingFee, booking.PlatformFee, booking.TotalAmount,
		),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Price set, client has been notified",
		"booking": booking,
	})
}

// ── NEW: Client accepts price ────────────────────────────────────
// POST /bookings/:id/accept-price
func AcceptBookingPrice(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	var booking models.Booking
	if err := db.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	if booking.CustomerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your booking"})
		return
	}

	if booking.Status != "price_set" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No price to accept yet"})
		return
	}

	booking.Status = "confirmed"
	if err := db.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm booking"})
		return
	}

	// Notify technician
	var technician models.User
	db.DB.First(&technician, booking.TechnicianID)

	websocket.SendNotification(booking.TechnicianID, "Client accepted your price! Job is confirmed.")

	go sendEmail(
		technician.Email,
		"OneFixAL - Client Accepted Your Price!",
		fmt.Sprintf(
			"Great news! The client accepted your price of %d LEK.\n\nJob is confirmed. Please proceed with the work.",
			booking.JobPrice,
		),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking confirmed!",
		"booking": booking,
	})
}

// ── Email helper ─────────────────────────────────────────────────
func sendEmail(toEmail, subject, body string) {
	if toEmail == "" {
		return
	}
	apiKey    := os.Getenv("SENDGRID_API_KEY")
	fromEmail := os.Getenv("SMTP_USER")

	payload := fmt.Sprintf(`{
		"personalizations": [{"to": [{"email": "%s"}]}],
		"from": {"email": "%s", "name": "OneFixAL"},
		"subject": "%s",
		"content": [{"type": "text/plain", "value": %q}]
	}`, toEmail, fromEmail, subject, body)

	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", strings.NewReader(payload))
	if err != nil {
		fmt.Println("❌ Email build error:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ Email send error:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("📨 Email to %s — status: %d\n", toEmail, resp.StatusCode)
}
