package router

import (
	"OneFixAL/internal/api"
	"OneFixAL/internal/middleware"
	"OneFixAL/internal/websocket"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// ── CORS ─────────────────────────────────────────────────────
	// Must be registered FIRST before any routes
	
	// ✅ CORS must be FIRST — before any routes
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // cache preflight for 12h
	}))
	// ── WEBSOCKET ─────────────────────────────────────────────────
	r.GET("/ws", websocket.WebSocketHandler)

	// ── HEALTH CHECK ──────────────────────────────────────────────
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// ── PUBLIC ROUTES ─────────────────────────────────────────────
	public := r.Group("/")
	{
		public.POST("/signup", api.Signup)
		public.POST("/login", api.Login)

		// OTP password reset
		public.POST("/auth/forgot-password", api.ForgotPassword)
		public.POST("/auth/verify-otp", api.VerifyOTPAndResetPassword)

		// Public reads
		public.GET("/users/:id", api.GetUserByID)
		public.GET("/technicians", api.GetTechniciansHandler)
		public.GET("/availability/:id", api.GetAvailabilityByTechnicianID)
	}

	// ── PROTECTED ROUTES ──────────────────────────────────────────
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// User
		protected.GET("/me", api.GetProfile)
		protected.POST("/set-role", api.SetRole)

		// Technician profile
		protected.PUT("/technician/profile",
			middleware.RoleMiddleware("technician"),
			api.CreateOrUpdateTechnicianProfile,
		)
		protected.GET("/technician/profile",
			middleware.RoleMiddleware("technician"),
			api.GetTechnicianProfile,
		)

		// Availability
		protected.POST("/availability",
			middleware.RoleMiddleware("technician"),
			api.CreateAvailability,
		)
		protected.GET("/availability", api.GetAvailability)
		protected.PUT("/availability/:id",
			middleware.RoleMiddleware("technician"),
			api.UpdateAvailability,
		)
		protected.DELETE("/availability/:id",
			middleware.RoleMiddleware("technician"),
			api.DeleteAvailability,
		)

		// Bookings
		protected.POST("/bookings",
			middleware.RoleMiddleware("client"),
			api.CreateBooking,
		)
		protected.GET("/my-bookings", api.GetMyBookings)
		protected.GET("/tech/bookings",
			middleware.RoleMiddleware("technician"),
			api.GetTechnicianBookings,
		)
		protected.PUT("/bookings/:id/status", api.UpdateBookingStatus)
	}

	return r
}
