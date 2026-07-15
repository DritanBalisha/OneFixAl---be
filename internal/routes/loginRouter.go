package router

import (
	"OneFixAL/internal/api"
	"OneFixAL/internal/middleware"
	"OneFixAL/internal/websocket"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// ── CORS ─────────────────────────────────────────────────────
	// Must be registered BEFORE every route
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"https://one-fix-al-fe.vercel.app",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ── WEBSOCKET ────────────────────────────────────────────────
	r.GET("/ws", websocket.WebSocketHandler)

	// ── TEST ROUTE ───────────────────────────────────────────────
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// ── PUBLIC ROUTES ────────────────────────────────────────────
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

	// ── PROTECTED ROUTES ─────────────────────────────────────────
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// User
		protected.GET("/me", api.GetProfile)
		protected.POST("/set-role", api.SetRole)

		// Technician profile
		protected.PUT(
			"/technician/profile",
			middleware.RoleMiddleware("technician"),
			api.CreateOrUpdateTechnicianProfile,
		)

		protected.GET(
			"/technician/profile",
			middleware.RoleMiddleware("technician"),
			api.GetTechnicianProfile,
		)

		// Availability
		protected.POST(
			"/availability",
			middleware.RoleMiddleware("technician"),
			api.CreateAvailability,
		)

		protected.GET("/availability", api.GetAvailability)

		protected.PUT(
			"/availability/:id",
			middleware.RoleMiddleware("technician"),
			api.UpdateAvailability,
		)

		protected.DELETE(
			"/availability/:id",
			middleware.RoleMiddleware("technician"),
			api.DeleteAvailability,
		)

		// Bookings
		protected.POST(
			"/bookings",
			middleware.RoleMiddleware("client"),
			api.CreateBooking,
		)

		protected.GET("/my-bookings", api.GetMyBookings)

		protected.GET(
			"/tech/bookings",
			middleware.RoleMiddleware("technician"),
			api.GetTechnicianBookings,
		)
		// Booking price flow
protected.POST("/bookings/:id/set-price",
    middleware.RoleMiddleware("technician"),
    api.SetBookingPrice,
)
protected.POST("/bookings/:id/accept-price",
    middleware.RoleMiddleware("client"),
    api.AcceptBookingPrice,
)

		protected.PUT("/bookings/:id/status", api.UpdateBookingStatus)
	}

	return r
}
