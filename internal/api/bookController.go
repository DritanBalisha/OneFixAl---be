package api

import (
	"fmt"
	"net/http"
	"time"

	"OneFixAL/internal/db"
	"OneFixAL/internal/models"
	"github.com/gin-gonic/gin"
)

const PlatformFeePercent = 2  // your 2% cut
const BookingFeePercent  = 10 // 10% deposit from client

type CreateBookingRequest struct {
	TechnicianID   uint      `json:"technician_id"`
	AvailabilityID uint      `json:"availability_id"`
	Timeslot       time.Time `json:"timeslot"`
	LocationLat    float64   `json:"location_lat"`
	LocationLon    float64   `json:"location_lon"`
	Description    string    `json:"description"`
	JobPrice       int64     `json:"job_price"` // client enters estimated job price
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

	// Prevent self-booking
	if userID.(uint) == req.TechnicianID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot book your own service"})
		return
	}

	// Validate job price
	if req.JobPrice <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job price must be greater than 0"})
		return
	}

	// Check slot exists
	var slot models.Availability
	if err := db.DB.First(&slot, req.AvailabilityID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Slot not found"})
		return
	}

	// Prevent double-booking
	if slot.IsBooked {
		c.JSON(http.StatusConflict, gin.H{"error": "This time slot is already booked"})
		return
	}

	// ✅ Calculate fees automatically
	bookingFee  := (req.JobPrice * BookingFeePercent) / 100  // 10% deposit
	platformFee := (req.JobPrice * PlatformFeePercent) / 100 // your 2%

	booking := models.Booking{
		CustomerID:     userID.(uint),
		TechnicianID:   req.TechnicianID,
		AvailabilityID: req.AvailabilityID,
		Timeslot:       req.Timeslot,
		LocationLat:    req.LocationLat,
		LocationLon:    req.LocationLon,
		Description:    req.Description,
		JobPrice:       req.JobPrice,
		BookingFee:     bookingFee,
		PlatformFee:    platformFee,
		TotalAmount:    req.JobPrice,
		Status:         "pending",
		PaymentStatus:  "unpaid",
		CreatedAt:      time.Now(),
	}

	if err := db.DB.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	// Mark slot as booked
	slot.IsBooked = true
	db.DB.Save(&slot)

	// Load customer and technician for emails
	var customer models.User
	var technician models.User
	db.DB.First(&customer, booking.CustomerID)
	db.DB.First(&technician, booking.TechnicianID)

	// ✅ Send confirmation emails (non-blocking)
	go sendBookingConfirmationEmail(customer, technician, booking)

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking created successfully",
		"booking": booking,
		"fee_breakdown": gin.H{
			"job_price":    fmt.Sprintf("%d LEK", booking.JobPrice),
			"booking_fee":  fmt.Sprintf("%d LEK (10%% deposit due now)", booking.BookingFee),
			"platform_fee": fmt.Sprintf("%d LEK (2%% OneFixAL fee)", booking.PlatformFee),
			"total":        fmt.Sprintf("%d LEK", booking.TotalAmount),
		},
	})
}
func sendBookingConfirmationEmail(customer, technician models.User, booking models.Booking) {
	// Email to customer
	customerBody := fmt.Sprintf(`Your booking has been submitted successfully!

Booking Details:
────────────────────────────
Technician:   %s
Date/Time:    %s
Description:  %s

Fee Breakdown:
────────────────────────────
Job Price:    %d LEK
Booking Fee:  %d LEK (10%% deposit — pay to confirm)
Platform Fee: %d LEK (included)
Total:        %d LEK
────────────────────────────

Payment Instructions:
Please pay the booking deposit of %d LEK to confirm your booking.
The technician will contact you shortly.

Thank you for using OneFixAL!`,
		technician.Name,
		booking.Timeslot.Format("02 Jan 2006 15:04"),
		booking.Description,
		booking.JobPrice,
		booking.BookingFee,
		booking.PlatformFee,
		booking.TotalAmount,
		booking.BookingFee,
	)

	// Email to technician
	techBody := fmt.Sprintf(`You have a new booking request!

Booking Details:
────────────────────────────
Client:       %s
Phone:        %s
Date/Time:    %s
Description:  %s
Location:     %.4f, %.4f

Job Price:    %d LEK
────────────────────────────

Please log in to OneFixAL to confirm or decline this booking.`,
		customer.Name,
		customer.Phone,
		booking.Timeslot.Format("02 Jan 2006 15:04"),
		booking.Description,
		booking.LocationLat,
		booking.LocationLon,
		booking.JobPrice,
	)

	sendEmail(customer.Email, "OneFixAL - Booking Confirmation", customerBody)
	sendEmail(technician.Email, "OneFixAL - New Booking Request", techBody)
}

// Generic email sender using SendGrid
func sendEmail(toEmail, subject, body string) {
	apiKey   := os.Getenv("SENDGRID_API_KEY")
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
