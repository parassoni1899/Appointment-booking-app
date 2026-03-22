package service

import (
	"errors"
	"time"

	"github.com/parassoni1899/appointment-booking/internal/models"
	"github.com/parassoni1899/appointment-booking/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrAvailabilityNotFound = errors.New("availability not found for the given day")
	ErrSlotAlreadyBooked    = errors.New("slot is already booked")
	ErrInvalidTimeFormat    = errors.New("invalid time format")
)

type Service interface {
	SetCoachAvailability(coachID uint, dayOfWeek, timezone, startTime, endTime string) error
	GetAvailableSlots(coachID uint, dateStr string) ([]string, error)
	BookSlot(userID, coachID uint, slotTimeStr string) (*models.Booking, error)
	GetUserBookings(userID uint) ([]models.Booking, error)
	CancelBooking(userID, bookingID uint) error
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) SetCoachAvailability(coachID uint, dayOfWeek, timezone, startTime, endTime string) error {
	// Parse times to ensure they are valid "15:04" format
	_, err := time.Parse("15:04", startTime)
	if err != nil {
		return ErrInvalidTimeFormat
	}
	_, err = time.Parse("15:04", endTime)
	if err != nil {
		return ErrInvalidTimeFormat
	}
	// Validate Timezone
	if timezone == "" {
		timezone = "UTC"
	}
	_, err = time.LoadLocation(timezone)
	if err != nil {
		return errors.New("invalid timezone identifier")
	}

	// For simplicity, find if exists and update, or create
	availability, err := s.repo.GetCoachAvailability(coachID, dayOfWeek)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		availability = &models.CoachAvailability{
			CoachID:   coachID,
			DayOfWeek: dayOfWeek,
		}
	}

	availability.Timezone = timezone
	availability.StartTime = startTime
	availability.EndTime = endTime

	return s.repo.SaveCoachAvailability(availability)
}

func (s *service) GetAvailableSlots(coachID uint, dateStr string) ([]string, error) {
	// 1. Fetch coach availability first to know their timezone
	// We parse the date into standard time to extract DayOfWeek initially
	genericDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("invalid date format, must be YYYY-MM-DD")
	}

	dayOfWeek := genericDate.Weekday().String()

	availability, err := s.repo.GetCoachAvailability(coachID, dayOfWeek)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []string{}, nil // No availability
		}
		return nil, err
	}

	// Parse timezone
	loc, err := time.LoadLocation(availability.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// Reparse the date firmly in the coach's timezone
	date, _ := time.ParseInLocation("2006-01-02", dateStr, loc)

	// 3. Parse Start and End times
	startTime, _ := time.Parse("15:04", availability.StartTime)
	endTime, _ := time.Parse("15:04", availability.EndTime)

	// Combine date with parsed times in specific location
	startDateTime := time.Date(date.Year(), date.Month(), date.Day(), startTime.Hour(), startTime.Minute(), 0, 0, loc)
	endDateTime := time.Date(date.Year(), date.Month(), date.Day(), endTime.Hour(), endTime.Minute(), 0, 0, loc)

	// 4. Fetch bookings for the date mapped to pure UTC bounds
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc).UTC()
	endOfDay := startOfDay.Add(24 * time.Hour)
	bookings, err := s.repo.GetBookingsByCoachAndDate(coachID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	// Map of booked slots
	bookedMap := make(map[time.Time]bool)
	for _, b := range bookings {
		// Normalise to UTC for comparison 
		bookedMap[b.SlotTime.UTC()] = true
	}

	// 5. Generate slots
	var availableSlots []string
	currentSlot := startDateTime
	for currentSlot.Before(endDateTime) || currentSlot.Equal(endDateTime) {
		// Stop if currentSlot + 30m goes beyond endDateTime
		nextSlot := currentSlot.Add(30 * time.Minute)
		if nextSlot.After(endDateTime) {
			break
		}

		// Check if it's not booked against UTC bounds
		if !bookedMap[currentSlot.UTC()] {
			availableSlots = append(availableSlots, currentSlot.UTC().Format(time.RFC3339))
		}

		currentSlot = nextSlot
	}

	return availableSlots, nil
}

func (s *service) BookSlot(userID, coachID uint, slotTimeStr string) (*models.Booking, error) {
	slotTime, err := time.Parse(time.RFC3339, slotTimeStr)
	if err != nil {
		return nil, errors.New("invalid slot time format, must be RFC3339")
	}

	// Ensure the slot time is exactly on the hour or half-hour (optional but good for consistency)
	if slotTime.Minute() % 30 != 0 || slotTime.Second() != 0 {
		return nil, errors.New("slot time must be in 30-minute increments")
	}
	
	// Normalize to UTC for reliable comparison
	slotTime = slotTime.UTC()

	booking := &models.Booking{
		UserID:   userID,
		CoachID:  coachID,
		SlotTime: slotTime,
	}

	err = s.repo.SaveBooking(booking)
	if err != nil {
		// Check for unique constraint violation
		if isUniqueConstraintError(err) {
			return nil, ErrSlotAlreadyBooked
		}
		return nil, err
	}

	return booking, nil
}

func (s *service) GetUserBookings(userID uint) ([]models.Booking, error) {
	return s.repo.GetBookingsByUser(userID)
}

func (s *service) CancelBooking(userID, bookingID uint) error {
	return s.repo.DeleteBooking(bookingID, userID)
}

// Helper to determine if the error is a PostgreSQL unique constraint violation
func isUniqueConstraintError(err error) bool {
	// GORM pg driver returns SQLSTATE 23505 for unique_violation
	return err != nil && (err.Error() == "ERROR: duplicate key value violates unique constraint \"idx_coach_slot\" (SQLSTATE 23505)" || 
	// Or sometimes just contains the state
	len(err.Error()) > 0 && (contains(err.Error(), "23505") || contains(err.Error(), "duplicate key")))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s != "" && substr != ""
}
