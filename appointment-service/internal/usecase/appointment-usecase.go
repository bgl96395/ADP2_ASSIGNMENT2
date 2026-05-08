package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/cache"
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
	cache         cache.CacheRepository
}

func New_appointment_usecase(repo repository.Appointment_repository, doctor_client Doctor_client, pub event.EventPublisher, c cache.CacheRepository) *Appointment_usecase {
	return &Appointment_usecase{repo: repo, doctor_client: doctor_client, publisher: pub, cache: c}
}

func cacheTTL() time.Duration {
	val := os.Getenv("CACHE_TTL_SECONDS")
	if val != "" {
		secs, err := strconv.Atoi(val)
		if err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return 60 * time.Second
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

	ctx := context.Background()
	err = usecase.cache.InvalidateAppointmentList(ctx)
	if err != nil {
		log.Printf("ERROR: cache invalidation failed for appointments:list: %v", err)
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
	ctx := context.Background()

	cached, err := usecase.cache.GetAppointment(ctx, id)
	if err != nil {
		log.Printf("WARN: cache get failed for appointment:%s: %v", id, err)
	}
	if cached != nil {
		return cached, nil
	}

	appointment, err := usecase.repo.Find_by_ID(id)
	if err != nil {
		return nil, err
	}

	err = usecase.cache.SetAppointment(ctx, id, appointment, cacheTTL())
	if err != nil {
		log.Printf("ERROR: cache set failed for appointment:%s: %v", id, err)
	}

	return appointment, nil
}

func (usecase *Appointment_usecase) List_appoinments() ([]*model.Appointment, error) {
	ctx := context.Background()

	cached, err := usecase.cache.GetAppointmentList(ctx)
	if err != nil {
		log.Printf("WARN: cache get failed for appointments:list: %v", err)
	}
	if cached != nil {
		return cached, nil
	}

	list, err := usecase.repo.Find_all()
	if err != nil {
		return nil, err
	}

	err = usecase.cache.SetAppointmentList(ctx, list, cacheTTL())
	if err != nil {
		log.Printf("ERROR: cache set failed for appointments:list: %v", err)
	}

	return list, nil
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

	ctx := context.Background()
	err = usecase.cache.SetAppointment(ctx, id, appointment, cacheTTL())
	if err != nil {
		log.Printf("ERROR: cache set failed for appointment:%s: %v", id, err)
	}
	err = usecase.cache.InvalidateAppointmentList(ctx)
	if err != nil {
		log.Printf("ERROR: cache invalidation failed for appointments:list: %v", err)
	}

	evt := event.AppointmentStatusUpdatedEvent{
		EventType:  "appointments.status_updated",
		OccurredAt: time.Now().UTC(),
		ID:         appointment.ID,
		OldStatus:  oldStatus,
		NewStatus:  string(new_status),
		DoctorID:   appointment.DoctorID,
	}
	publisherError := usecase.publisher.Publish("appointments.status_updated", evt)
	if publisherError != nil {
		fmt.Printf("ERROR: failed to publish appointments.status_updated: %v\n", publisherError)
	}

	return appointment, nil
}
