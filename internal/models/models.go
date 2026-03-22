package models

import "time"

// User represents a user who can book an appointment
type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Coach represents a coach who can be booked
type Coach struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CoachAvailability represents the weekly availability of a coach
type CoachAvailability struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CoachID   uint      `json:"coach_id"`
	DayOfWeek string    `json:"day_of_week"` // e.g., "Monday"
	Timezone  string    `json:"timezone"`    // e.g., "America/New_York"
	StartTime string    `json:"start_time"`  // Format: "15:04" (24-hour)
	EndTime   string    `json:"end_time"`    // Format: "15:04"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Coach Coach `gorm:"foreignKey:CoachID" json:"-"`
}

// Booking represents a 30-minute booked appointment slot
type Booking struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `json:"user_id"`
	CoachID   uint      `json:"coach_id"`
	SlotTime  time.Time `gorm:"uniqueIndex:idx_coach_slot" json:"slot_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User  User  `gorm:"foreignKey:UserID" json:"-"`
	Coach Coach `gorm:"foreignKey:CoachID" json:"-"`
}
