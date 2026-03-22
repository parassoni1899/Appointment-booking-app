package service_test

import (
	"testing"
	"time"

	"github.com/parassoni1899/appointment-booking/internal/models"
	"github.com/parassoni1899/appointment-booking/internal/service"
	"gorm.io/gorm"
)

type mockRepo struct {
	availability *models.CoachAvailability
	bookings     []models.Booking
}

func (m *mockRepo) SaveCoachAvailability(availability *models.CoachAvailability) error {
	m.availability = availability
	return nil
}

func (m *mockRepo) GetCoachAvailability(coachID uint, dayOfWeek string) (*models.CoachAvailability, error) {
	if m.availability != nil && m.availability.CoachID == coachID && m.availability.DayOfWeek == dayOfWeek {
		return m.availability, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRepo) SaveBooking(booking *models.Booking) error {
	m.bookings = append(m.bookings, *booking)
	return nil
}

func (m *mockRepo) GetBookingsByCoachAndDate(coachID uint, startOfDay, endOfDay time.Time) ([]models.Booking, error) {
	var result []models.Booking
	for _, b := range m.bookings {
		if b.CoachID == coachID && !b.SlotTime.Before(startOfDay) && b.SlotTime.Before(endOfDay) {
			result = append(result, b)
		}
	}
	return result, nil
}

func (m *mockRepo) GetBookingsByUser(userID uint) ([]models.Booking, error) {
	return nil, nil
}

func (m *mockRepo) DeleteBooking(bookingID uint, userID uint) error {
	return nil
}

func (m *mockRepo) InitDB() error {
	return nil
}

func TestGetAvailableSlots_Timezone(t *testing.T) {
	repo := &mockRepo{
		availability: &models.CoachAvailability{
			CoachID:   1,
			DayOfWeek: "Tuesday",
			Timezone:  "America/New_York",
			StartTime: "09:00",
			EndTime:   "11:00",
		},
		bookings: []models.Booking{},
	}
	
	loc, _ := time.LoadLocation("America/New_York")
	date, _ := time.ParseInLocation("2006-01-02", "2025-10-28", loc) // 2025-10-28 is a Tuesday
	bookedTime := time.Date(date.Year(), date.Month(), date.Day(), 9, 30, 0, 0, loc).UTC()

	// Simulate one booked slot exactly at 9:30 AM New York time (which is implicitly validated in UTC)
	repo.bookings = append(repo.bookings, models.Booking{CoachID: 1, SlotTime: bookedTime})

	svc := service.NewService(repo)

	slots, err := svc.GetAvailableSlots(1, "2025-10-28")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 09:00 to 11:00 should have 4 slots: 09:00, 09:30, 10:00, 10:30
	// But 09:30 is booked, so 3 slots remaining
	if len(slots) != 3 {
		t.Fatalf("expected 3 slots, got %d", len(slots))
	}

	expected0900 := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, loc).UTC().Format(time.RFC3339)
	if slots[0] != expected0900 {
		t.Errorf("expected first slot %s, got %s", expected0900, slots[0])
	}
	
	expected1000 := time.Date(date.Year(), date.Month(), date.Day(), 10, 0, 0, 0, loc).UTC().Format(time.RFC3339)
	if slots[1] != expected1000 {
		t.Errorf("expected second slot %s, got %s", expected1000, slots[1])
	}
}
