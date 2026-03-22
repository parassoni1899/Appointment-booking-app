# Appointment Booking API

A RESTful backend service built in Go to handle appointment bookings between coaches and users. 

## Features
- **Coach Availability:** Coaches can set their weekly availability (e.g., Monday 09:00 - 14:00).
- **Timezones:** Handles dynamic timezone offsets for specific coach locations, compiling standardized UTC slots.
- **Available Slots:** System dynamically generates 30-minute available slots, automatically excluding slots that are already booked.
- **Book an Appointment:** Users can book an available slot with race-condition prevention (double booking natively prevented by a database-level composite unique index).
- **View Appointments:** Users can view their booked appointments.
- **Booking Cancellation:** Allows users to cancel their bookings securely.

## Tech Stack
- **Go** (Golang)
- **Gin** (HTTP framework)
- **PostgreSQL** (Database)
- **GORM** (ORM)

## Setup and Running Locally

### 1. Prerequisites
- [Go 1.22+](https://go.dev/dl/)
- PostgreSQL database

### 2. Configure Database
By default, the application connects to a local postgres database with the following credentials:
`host=localhost user=postgres password=postgres dbname=booking port=5432 sslmode=disable TimeZone=UTC`

Create a database named `booking`:
```sql
CREATE DATABASE booking;
```

Alternatively, you can set the `DATABASE_URL` and `PORT` environment variables in a `.env` file at the root of the project:
```env
DATABASE_URL=host=localhost user=youruser password=yourpassword dbname=booking port=5432 sslmode=disable TimeZone=UTC
PORT=8080
```

### 3. Install Dependencies and Run
```bash
go mod tidy
go run cmd/main.go
```
*Note: The application will automatically run GORM migrations to create the required tables and seed the database with 2 dummy Users (ID: 1, 2) and 2 dummy Coaches (ID: 1, 2).*

---

## API Documentation

### 1. Set Coach Availability
Allows a coach to set their weekly availability for a specific day.
- **URL**: `/coaches/availability`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "coach_id": 1,
    "day": "Monday",
    "timezone": "America/New_York",
    "start_time": "09:00",
    "end_time": "14:00"
  }
  ```
- **Response** (200 OK):
  ```json
  {
    "message": "Availability saved successfully"
  }
  ```

### 2. Get Available Slots
Fetches all available 30-minute slots for a given coach on a specific day.
- **URL**: `/users/slots?coach_id=1&date=2025-10-28`
- **Method**: `GET`
- **Query Params**:
  - `coach_id` (integer)
  - `date` (string, format YYYY-MM-DD)
- **Response** (200 OK):
  ```json
  [
    "2025-10-28T09:00:00Z",
    "2025-10-28T09:30:00Z",
    "2025-10-28T10:00:00Z"
  ]
  ```

### 3. Book an Appointment
Allows a user to book an available 30-minute slot.
- **URL**: `/users/bookings`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "user_id": 1,
    "coach_id": 1,
    "datetime": "2025-10-28T09:30:00Z"
  }
  ```
- **Response** (201 Created):
  ```json
  {
    "booking": {
      "id": 1,
      "user_id": 1,
      "coach_id": 1,
      "slot_time": "2025-10-28T09:30:00Z",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    },
    "message": "Booking successful"
  }
  ```
- **Response** (409 Conflict): returned if the exact slot is already booked.

### 4. Get User Bookings
Fetches all appointments for a given user.
- **URL**: `/users/bookings?user_id=1`
- **Method**: `GET`
- **Query Params**:
  - `user_id` (integer)
- **Response** (200 OK):
  ```json
  [
    {
      "id": 1,
      "user_id": 1,
      "coach_id": 1,
      "slot_time": "2025-10-28T09:30:00Z",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
  ```

### 5. Cancel Booking
Allows a user to safely cancel their existing appointment.
- **URL**: `/users/bookings/:id?user_id=1`
- **Method**: `DELETE`
- **Query Params**:
  - `user_id` (integer) - the owner of the booking
- **Path Params**:
  - `id` (integer) - the booking ID
- **Response** (200 OK):
  ```json
  {
    "message": "Booking cancelled successfully"
  }
  ```
