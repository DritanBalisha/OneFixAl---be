package api

import (
	"OneFixAL/internal/db"
	"OneFixAL/internal/models"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"net/smtp"
	"os"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// var jwtSecret = []byte("your_secret_key") // TODO: load from env

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("your_secret_key")
	}
	return []byte(secret)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWT(userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID, // store userID here
		"role": role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // expires in 24h
		"iat":  time.Now().Unix(),                     // issued at
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func GetUserByID(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "User ID not found in token"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Base response (don’t send password)
	response := gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"phone": user.Phone,
		"role":  user.Role,
	}

	// If technician, also fetch profile
	if user.Role == "technician" {
		var techProfile models.TechnicianProfile
		if err := db.DB.Where("user_id = ?", user.ID).First(&techProfile).Error; err == nil {
			response["technicianProfile"] = techProfile
		}
	}

	c.JSON(200, response)
}

// handler
func GetTechniciansHandler(c *gin.Context) {
	var users []models.User

	// Preload TechnicianProfile so it joins the technician_profiles table
	if err := db.DB.Preload("TechnicianProfile").Where("role = ?", "technician").Find(&users).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch technicians"})
		return
	}

	// Return user + profile
	c.JSON(200, users)
}

func Signup(c *gin.Context) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, _ := HashPassword(input.Password)

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Phone:    input.Phone,
		Password: hashedPassword,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User created successfully"})
}

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "Invalid email or password"})
		return
	}

	if !CheckPasswordHash(input.Password, user.Password) {
		c.JSON(401, gin.H{"error": "Invalid email or password"})
		return
	}

	token, _ := GenerateJWT(user.ID, user.Role)

	c.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"phone": user.Phone,
			"role":  user.Role,
		},
	})
	fmt.Println(user)

}

func SetRole(c *gin.Context) {
	userID, _ := c.Get("userID")

	var body struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if body.Role != "client" && body.Role != "technician" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Role = body.Role
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update role"})
		return
	}

	// ✅ return updated user object like frontend expects
	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"phone": user.Phone,
		"role":  user.Role,
	})
}

func CreateOrUpdateTechnicianProfile(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Ensure correct type (JWT often stores as float64)
	var userID uint
	switch v := userIDVal.(type) {
	case uint:
		userID = v
	case float64:
		userID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
		return
	}

	var input models.TechnicianProfile
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.UserID = userID

	var profile models.TechnicianProfile
	if err := db.DB.Where("user_id = ?", userID).First(&profile).Error; err == nil {
		// ✅ Update existing profile
		if err := db.DB.Model(&profile).Updates(input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}
	} else {
		// ✅ Create new profile
		if err := db.DB.Create(&input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		profile = input
	}

	c.JSON(http.StatusOK, gin.H{"profile": profile})
}

func GetTechnicianProfile(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var userID uint
	switch v := userIDVal.(type) {
	case uint:
		userID = v
	case float64:
		userID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
		return
	}

	var profile models.TechnicianProfile
	if err := db.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// ── OTP HELPERS ─────────────────────────────────────────────────────────────

// generateOTP creates a cryptographically secure random 6-digit string
func generateOTP() (string, error) {
	const digits = "0123456789"
	otp := make([]byte, 6)
	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[num.Int64()]
	}
	return string(otp), nil
}

// sendOTPEmail sends the OTP code to the user via SMTP
func sendOTPEmail(toEmail, otpCode string) error {
	smtpHost := os.Getenv("SMTP_HOST")         // smtp.gmail.com
	smtpPort := os.Getenv("SMTP_PORT")         // 587
	smtpUser := os.Getenv("SMTP_USER")         // your@gmail.com
	smtpPass := os.Getenv("SMTP_PASSWORD")     // Gmail app password

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("missing SMTP environment variables")
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	subject := "Subject: OneFixAL - Password Reset Code\r\n"
	from := fmt.Sprintf("From: OneFixAL <%s>\r\n", smtpUser)
	to := fmt.Sprintf("To: %s\r\n", toEmail)
	mime := "MIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"

	body := fmt.Sprintf(
		"Your OneFixAL password reset code is:\n\n"+
			"  %s\n\n"+
			"This code expires in 10 minutes.\n"+
			"If you did not request this, you can safely ignore this email.",
		otpCode,
	)

	msg := []byte(from + to + subject + mime + body)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	err := smtp.SendMail(addr, auth, smtpUser, []string{toEmail}, msg)
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}
// func sendOTPEmail(toEmail, otpCode string) error {
// 	smtpHost := os.Getenv("SMTP_HOST")     // e.g. smtp.gmail.com
// 	smtpPort := os.Getenv("SMTP_PORT")     // e.g. 587
// 	smtpUser := os.Getenv("SMTP_USER")     // your@gmail.com
// 	smtpPass := os.Getenv("SMTP_PASSWORD") // gmail app password

// 	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

// 	subject := "Subject: OneFixAL - Password Reset Code\n"
// 	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
// 	body := fmt.Sprintf(
// 		"Your OneFixAL password reset code is:\n\n"+
// 			"  %s\n\n"+
// 			"This code expires in 10 minutes.\n"+
// 			"If you did not request this, you can safely ignore this email.",
// 		otpCode,
// 	)

// 	msg := []byte(subject + mime + body)
// 	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
// 	return smtp.SendMail(addr, auth, smtpUser, []string{toEmail}, msg)
// }

// ── OTP HANDLERS ─────────────────────────────────────────────────────────────

// POST /auth/forgot-password
// Body: { "email": "user@example.com" }
func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println("❌ Invalid JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	if input.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	fmt.Println("📩 Forgot password request for:", input.Email)

	// Always return same message if user does not exist
	var user models.User
	if err := db.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		fmt.Println("⚠️ Email not found, but returning OK:", input.Email)

		c.JSON(http.StatusOK, gin.H{
			"message": "If that email exists, a reset code has been sent",
		})
		return
	}

	fmt.Println("✅ User found:", user.Email)

	// Invalidate old unused OTPs
	if err := db.DB.Model(&models.OTPCode{}).
		Where("user_id = ? AND purpose = ? AND used = ?", user.ID, "reset_password", false).
		Update("used", true).Error; err != nil {
		fmt.Println("❌ Could not invalidate old OTPs:", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not process reset request",
		})
		return
	}

	// Generate plain OTP
	otp, err := generateOTP()
	if err != nil {
		fmt.Println("❌ Could not generate OTP:", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not generate reset code",
		})
		return
	}

	fmt.Println("✅ OTP generated")

	// Hash OTP before saving
	hashedOTP, err := HashPassword(otp)
	if err != nil {
		fmt.Println("❌ Could not hash OTP:", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not process reset code",
		})
		return
	}

	// Save OTP record
	otpRecord := models.OTPCode{
		UserID:    user.ID,
		Code:      hashedOTP,
		Purpose:   "reset_password",
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Used:      false,
	}

	if err := db.DB.Create(&otpRecord).Error; err != nil {
		fmt.Println("❌ Could not save OTP:", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not save reset code",
		})
		return
	}

	fmt.Println("✅ OTP saved in database")

	// Send plain OTP by email
	if err := sendOTPEmail(user.Email, otp); err != nil {
		fmt.Println("❌ Email send error:", err)

		// Optional: delete failed OTP so unused codes do not remain
		db.DB.Delete(&otpRecord)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not send email, try again later",
		})
		return
	}

	fmt.Printf("✅ OTP sent to %s\n", user.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "If that email exists, a reset code has been sent",
	})
}

// POST /auth/verify-otp
// Body: { "email": "user@example.com", "otp": "123456", "new_password": "newpass123" }
func VerifyOTPAndResetPassword(c *gin.Context) {
	var input struct {
		Email       string `json:"email"`
		OTP         string `json:"otp"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if input.Email == "" || input.OTP == "" || input.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email, OTP and new password are required"})
		return
	}

	if len(input.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
		return
	}

	// Find user
	var user models.User
	if err := db.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code or email"})
		return
	}

	// Find latest unused OTP for this user
	var otpRecord models.OTPCode
	if err := db.DB.Where(
		"user_id = ? AND purpose = ? AND used = ?",
		user.ID, "reset_password", false,
	).Order("created_at DESC").First(&otpRecord).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active reset code found, please request a new one"})
		return
	}

	// Check expiry
	if time.Now().After(otpRecord.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset code has expired, please request a new one"})
		return
	}

	// Check OTP matches hashed value
	if !CheckPasswordHash(input.OTP, otpRecord.Code) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reset code"})
		return
	}

	// Hash new password
	newHashed, err := HashPassword(input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process new password"})
		return
	}

	// Update password
	if err := db.DB.Model(&user).Update("password", newHashed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update password"})
		return
	}

	// Mark OTP as used — never reusable
	db.DB.Model(&otpRecord).Update("used", true)

	fmt.Printf("✅ Password reset successful for %s\n", user.Email)
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully, you can now log in"})
}
