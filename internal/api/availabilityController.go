package api

import (
	"net/http"

	"OneFixAL/internal/db"
	"OneFixAL/internal/models"

	"github.com/gin-gonic/gin"
)

type AvailabilityInput struct {
	DayOfWeek int    `json:"dayOfWeek" binding:"required"`
	StartTime string `json:"startTime" binding:"required"`
	EndTime   string `json:"endTime" binding:"required"`
}

func GetAvailabilityByTechnicianID(c *gin.Context) {
	technicianID := c.Param("id")

	var availabilities []models.Availability
	if err := db.DB.Where("technician_id = ?", technicianID).Find(&availabilities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch availability"})
		return
	}

	if len(availabilities) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No availability found"})
		return
	}

	c.JSON(http.StatusOK, availabilities)
}

// fetch all availabilities for that technician.
func GetTechnicianAvailability(c *gin.Context) {
	technicianID := c.Param("technician_id")

	var availabilities []models.Availability
	if err := db.DB.Where("technician_id = ?", technicianID).Find(&availabilities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching availability"})
		return
	}

	c.JSON(http.StatusOK, availabilities)
}

// POST /availability
func CreateAvailability(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input AvailabilityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	availability := models.Availability{
		TechnicianID: userID.(uint),
		DayOfWeek:    input.DayOfWeek,
		StartTime:    input.StartTime,
		EndTime:      input.EndTime,
	}

	if err := db.DB.Create(&availability).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save availability"})
		return
	}

	c.JSON(http.StatusCreated, availability)
}

// GET /availability
func GetAvailability(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var availabilities []models.Availability
	if err := db.DB.Where("technician_id = ?", userID.(uint)).Find(&availabilities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load availability"})
		return
	}

	c.JSON(http.StatusOK, availabilities)
}

// PUT /availability/:id
func UpdateAvailability(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")
	var input models.Availability
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var availability models.Availability
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&availability).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Availability not found"})
		return
	}

	if err := db.DB.Model(&availability).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update availability"})
		return
	}

	c.JSON(http.StatusOK, availability)
}

// DELETE /availability/:id
func DeleteAvailability(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)

	var availability models.Availability
	if err := db.DB.Where("id = ? AND technician_id = ?", id, userID).First(&availability).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Availability not found"})
		return
	}

	if err := db.DB.Delete(&availability).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete availability"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}
