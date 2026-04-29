package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/event"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/model"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/repository"
	"github.com/google/uuid"
)

var (
	missing_fullname = errors.New("full name is required")
	missing_email    = errors.New("email is required")
	in_use_email     = errors.New("email already in use")
)

type Doctor_usecase struct {
	repository repository.Doctor_repository
	publisher  event.EventPublisher
}

func New_doctor_usecase(repo repository.Doctor_repository, pub event.EventPublisher) *Doctor_usecase {
	return &Doctor_usecase{repository: repo, publisher: pub}
}

func (usecase *Doctor_usecase) Create_doctor(fullName, specialization, email string) (*model.Doctor, error) {
	if fullName == "" {
		return nil, missing_fullname
	}
	if email == "" {
		return nil, missing_email
	}
	if usecase.repository.Exists_by_email(email) {
		return nil, in_use_email
	}

	doctor := &model.Doctor{
		ID:             uuid.New().String(),
		FullName:       fullName,
		Specialization: specialization,
		Email:          email,
	}
	err := usecase.repository.Save(doctor)
	if err != nil {
		return nil, err
	}

	evt := event.DoctorCreatedEvent{
		EventType:      "doctors.created",
		OccurredAt:     time.Now().UTC(),
		ID:             doctor.ID,
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}
	if publisherError := usecase.publisher.Publish("doctors.created", evt); publisherError != nil {
		fmt.Printf("ERROR: failed to publish doctors.created event: %v\n", publisherError)
	}

	return doctor, nil
}

func (usecase *Doctor_usecase) Get_doctor(id string) (*model.Doctor, error) {
	return usecase.repository.Find_by_ID(id)
}

func (usecase *Doctor_usecase) List_doctors() ([]*model.Doctor, error) {
	return usecase.repository.Find_all()
}
