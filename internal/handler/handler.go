package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/parassoni1899/appointment-booking/internal/service"
)

type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

type SetAvailabilityRequest struct {
	CoachID   uint   `json:"coach_id" binding:"required"`
	DayOfWeek string `json:"day" binding:"required"`
	Timezone  string `json:"timezone"` // e.g., "America/New_York"
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

func (h *Handler) SetCoachAvailability(c *gin.Context) {
	var req SetAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err := h.svc.SetCoachAvailability(req.CoachID, req.DayOfWeek, req.Timezone, req.StartTime, req.EndTime)
	if err != nil {
		if err == service.ErrInvalidTimeFormat {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time format (use 15:04)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Availability saved successfully"})
}

func (h *Handler) GetAvailableSlots(c *gin.Context) {
	coachIDStr := c.Query("coach_id")
	dateStr := c.Query("date")

	if coachIDStr == "" || dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "coach_id and date are required query parameters"})
		return
	}

	coachID, err := strconv.ParseUint(coachIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coach_id"})
		return
	}

	slots, err := h.svc.GetAvailableSlots(uint(coachID), dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, slots)
}

type BookSlotRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	CoachID  uint   `json:"coach_id" binding:"required"`
	Datetime string `json:"datetime" binding:"required"` // e.g. "2025-10-28T09:30:00Z"
}

func (h *Handler) BookSlot(c *gin.Context) {
	var req BookSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	booking, err := h.svc.BookSlot(req.UserID, req.CoachID, req.Datetime)
	if err != nil {
		if err == service.ErrSlotAlreadyBooked {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Booking successful",
		"booking": booking,
	})
}

func (h *Handler) GetUserBookings(c *gin.Context) {
	userIDStr := c.Query("user_id")

	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is a required query parameter"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	bookings, err := h.svc.GetUserBookings(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *Handler) CancelBooking(c *gin.Context) {
	bookingIDStr := c.Param("id")
	userIDStr := c.Query("user_id")

	if bookingIDStr == "" || userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking id in path and user_id in query are required"})
		return
	}

	bookingID, err := strconv.ParseUint(bookingIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking id"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	err = h.svc.CancelBooking(uint(userID), uint(bookingID))
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found or not owned by user"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
}
