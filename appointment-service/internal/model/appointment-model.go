package model

import "time"

type Status string

const (
	Status_new         = "new"
	Status_in_progress = "in_progress"
	Status_done        = "done"
)

type Appointment struct {
	ID          int
	Title       string
	Description string
	DoctorID    int
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
