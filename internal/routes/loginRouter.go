package router

import (
	"OneFixAL/internal/api"
	"OneFixAL/internal/middleware"
	"OneFixAL/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ws", websocket.WebSocketHandler)

	// Allow CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.POST("/signup", api.Signup)
	r.POST("/login", api.Login)
	r.PUT("/technician/profile", middleware.AuthMiddleware(), api.CreateOrUpdateTechnicianProfile)
	r.GET("/technician/profile", middleware.AuthMiddleware(), api.GetProfile)

	r.POST("/set-role", middleware.AuthMiddleware(), api.SetRole)

	r.GET("/users/:id", api.GetUserByID)

	r.GET("/technicians", api.GetTechniciansHandler)

	//availability RoUTERs
	r.POST("/availability", middleware.AuthMiddleware(), api.CreateAvailability)
	r.GET("/availability", middleware.AuthMiddleware(), api.GetAvailability)
	r.GET("/availability/:id", api.GetAvailabilityByTechnicianID)
	r.PUT("/availability/:id", middleware.AuthMiddleware(), api.UpdateAvailability)
	r.DELETE("/availability/:id", middleware.AuthMiddleware(), api.DeleteAvailability)

	//BOOK ROUTES
	r.POST("/bookings", middleware.AuthMiddleware(), api.CreateBooking)
	r.GET("/tech/bookings", middleware.AuthMiddleware(), api.GetTechnicianBookings) //technician

	r.PUT("/bookings/:id/status", middleware.AuthMiddleware(), api.UpdateBookingStatus)
	r.GET("/my-bookings", middleware.AuthMiddleware(), api.GetMyBookings) // route for technicians to se wich client  booked

	// r.HandleFunc("/bookings", api.CreateBooking(db)).Methods("POST")

	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/me", api.GetProfile)
	}

	// SetupAvailabilityRoutes(r)

	return r
}
