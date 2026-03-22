package repository

import (
	"time"

	"github.com/parassoni1899/appointment-booking/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	SaveCoachAvailability(availability *models.CoachAvailability) error
	GetCoachAvailability(coachID uint, dayOfWeek string) (*models.CoachAvailability, error)
	SaveBooking(booking *models.Booking) error
	GetBookingsByCoachAndDate(coachID uint, startOfDay, endOfDay time.Time) ([]models.Booking, error)
	GetBookingsByUser(userID uint) ([]models.Booking, error)
	DeleteBooking(bookingID uint, userID uint) error
	InitDB() error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) InitDB() error {
	// AutoMigrate the schemas
	err := r.db.AutoMigrate(
		&models.User{},
		&models.Coach{},
		&models.CoachAvailability{},
		&models.Booking{},
	)
	if err != nil {
		return err
	}
	
	// Seed some dummy users and coaches to make testing easier
	var count int64
	r.db.Model(&models.User{}).Count(&count)
	if count == 0 {
		r.db.Create(&models.User{Name: "Alice"})
		r.db.Create(&models.User{Name: "Bob"})
	}
	
	r.db.Model(&models.Coach{}).Count(&count)
	if count == 0 {
		r.db.Create(&models.Coach{Name: "Coach John"})
		r.db.Create(&models.Coach{Name: "Coach Sarah"})
	}

	return nil
}

func (r *repository) SaveCoachAvailability(availability *models.CoachAvailability) error {
	// We might want to use FirstOrCreate or just Create
	// For simplicity, we just delete the previous availability for that day and create a new one,
	// or we can use Clause to update. Let's do a simple save.
	return r.db.Save(availability).Error
}

func (r *repository) GetCoachAvailability(coachID uint, dayOfWeek string) (*models.CoachAvailability, error) {
	var availability models.CoachAvailability
	err := r.db.Where("coach_id = ? AND day_of_week = ?", coachID, dayOfWeek).First(&availability).Error
	if err != nil {
		return nil, err
	}
	return &availability, nil
}

func (r *repository) SaveBooking(booking *models.Booking) error {
	return r.db.Create(booking).Error
}

func (r *repository) GetBookingsByCoachAndDate(coachID uint, startOfDay, endOfDay time.Time) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.Where("coach_id = ? AND slot_time >= ? AND slot_time < ?", coachID, startOfDay, endOfDay).Find(&bookings).Error
	return bookings, err
}

func (r *repository) GetBookingsByUser(userID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.Preload("Coach").Where("user_id = ?", userID).Order("slot_time asc").Find(&bookings).Error
	return bookings, err
}

func (r *repository) DeleteBooking(bookingID uint, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", bookingID, userID).Delete(&models.Booking{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
