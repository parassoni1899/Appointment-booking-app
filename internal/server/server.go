package server

import (
	"github.com/gin-gonic/gin"
	"github.com/parassoni1899/appointment-booking/internal/handler"
)

func NewRouter(h *handler.Handler) *gin.Engine {
	r := gin.Default()

	// Coaches endpoints
	coachesRef := r.Group("/coaches")
	{
		coachesRef.POST("/availability", h.SetCoachAvailability)
	}

	// Users endpoints
	usersRef := r.Group("/users")
	{
		usersRef.GET("/slots", h.GetAvailableSlots)
		usersRef.POST("/bookings", h.BookSlot)
		usersRef.GET("/bookings", h.GetUserBookings)
		usersRef.DELETE("/bookings/:id", h.CancelBooking)
	}

	return r
}
