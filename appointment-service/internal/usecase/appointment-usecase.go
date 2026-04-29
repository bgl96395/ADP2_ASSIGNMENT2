package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/event"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/model"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/repository"
	"github.com/google/uuid"
)

var (
	missing_title                = errors.New("title is required")
	missing_doctor_id            = errors.New("doctor id is required")
	doctor_not_found             = errors.New("doctor not found")
	E_doctor_service_unavailable = errors.New("doctor service unavailable")
	invalid_status               = errors.New("invalid status")
	invalid_transition           = errors.New("cannot transition from done back to new")
)

type Appointment_usecase struct {
	repo          repository.Appointment_repository
	doctor_client Doctor_client
	publisher     event.EventPublisher
}

func New_appointment_usecase(repo repository.Appointment_repository, doctor_client Doctor_client, pub event.EventPublisher) *Appointment_usecase {
	return &Appointment_usecase{repo: repo, doctor_client: doctor_client, publisher: pub}
}

func (usecase *Appointment_usecase) Create_appointment(title, description string, doctorID string) (*model.Appointment, error) {
	if title == "" {
		return nil, missing_title
	}
	if doctorID == "" {
		return nil, missing_doctor_id
	}

	exists, err := usecase.doctor_client.Doctor_exists(doctorID)
	if err != nil {
		return nil, E_doctor_service_unavailable
	}
	if !exists {
		return nil, doctor_not_found
	}

	now := time.Now()
	appointment := &model.Appointment{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		DoctorID:    doctorID,
		Status:      model.Status_new,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = usecase.repo.Save(appointment)
	if err != nil {
		return nil, err
	}

	evt := event.AppointmentCreatedEvent{
		EventType:  "appointments.created",
		OccurredAt: time.Now().UTC(),
		ID:         appointment.ID,
		Title:      appointment.Title,
		DoctorID:   appointment.DoctorID,
		Status:     string(appointment.Status),
	}
	publisherError := usecase.publisher.Publish("appointments.created", evt)
	if publisherError != nil {
		fmt.Printf("ERROR: failed to publish appointments.created: %v\n", publisherError)
	}

	return appointment, nil
}

func (usecase *Appointment_usecase) Get_appointment(id string) (*model.Appointment, error) {
	return usecase.repo.Find_by_ID(id)
}

func (usecase *Appointment_usecase) List_appoinments() ([]*model.Appointment, error) {
	return usecase.repo.Find_all()
}

func (usecase *Appointment_usecase) Update_status(id string, new_status model.Status) (*model.Appointment, error) {
	if new_status != model.Status_new && new_status != model.Status_in_progress && new_status != model.Status_done {
		return nil, invalid_status
	}
	appointment, err := usecase.repo.Find_by_ID(id)
	if err != nil {
		return nil, err
	}

	if appointment.Status == model.Status_done && new_status == model.Status_new {
		return nil, invalid_transition
	}

	oldStatus := string(appointment.Status)
	appointment.Status = new_status
	appointment.UpdatedAt = time.Now()

	err = usecase.repo.Update(appointment)
	if err != nil {
		return nil, err
	}

	evt := event.AppointmentStatusUpdatedEvent{
		EventType:  "appointments.status_updated",
		OccurredAt: time.Now().UTC(),
		ID:         appointment.ID,
		OldStatus:  oldStatus,
		NewStatus:  string(new_status),
	}
	publisherError := usecase.publisher.Publish("appointments.status_updated", evt)
	if publisherError != nil {
		fmt.Printf("ERROR: failed to publish appointments.status_updated: %v\n", publisherError)
	}

	return appointment, nil
}
