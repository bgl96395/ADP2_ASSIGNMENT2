package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/cache"
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
	cache      cache.CacheRepository
}

func New_doctor_usecase(repo repository.Doctor_repository, pub event.EventPublisher, c cache.CacheRepository) *Doctor_usecase {
	return &Doctor_usecase{repository: repo, publisher: pub, cache: c}
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

	ctx := context.Background()
	err = usecase.cache.InvalidateDoctorList(ctx)
	if err != nil {
		log.Printf("ERROR: cache invalidation failed for doctors:list: %v", err)
	}

	evt := event.DoctorCreatedEvent{
		EventType:      "doctors.created",
		OccurredAt:     time.Now().UTC(),
		ID:             doctor.ID,
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}
	publisherError := usecase.publisher.Publish("doctors.created", evt)
	if publisherError != nil {
		fmt.Printf("ERROR: failed to publish doctors.created event: %v\n", publisherError)
	}

	return doctor, nil
}

func (usecase *Doctor_usecase) Get_doctor(id string) (*model.Doctor, error) {
	ctx := context.Background()

	cached, err := usecase.cache.GetDoctor(ctx, id)
	if err != nil {
		log.Printf("WARN: cache get failed for doctor:%s: %v", id, err)
	}
	if cached != nil {
		return cached, nil
	}

	doctor, err := usecase.repository.Find_by_ID(id)
	if err != nil {
		return nil, err
	}

	err = usecase.cache.SetDoctor(ctx, id, doctor, cacheTTL())
	if err != nil {
		log.Printf("ERROR: cache set failed for doctor:%s: %v", id, err)
	}

	return doctor, nil
}

func (usecase *Doctor_usecase) List_doctors() ([]*model.Doctor, error) {
	ctx := context.Background()

	cached, err := usecase.cache.GetDoctorList(ctx)
	if err != nil {
		log.Printf("WARN: cache get failed for doctors:list: %v", err)
	}
	if cached != nil {
		return cached, nil
	}

	doctors, err := usecase.repository.Find_all()
	if err != nil {
		return nil, err
	}

	err = usecase.cache.SetDoctorList(ctx, doctors, cacheTTL())
	if err != nil {
		log.Printf("ERROR: cache set failed for doctors:list: %v", err)
	}

	return doctors, nil
}
