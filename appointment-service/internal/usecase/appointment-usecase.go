package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/model"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/repository"
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
}

func New_appointment_usecase(repo repository.Appointment_repository, doctor_client Doctor_client) *Appointment_usecase {
	return &Appointment_usecase{repo: repo, doctor_client: doctor_client}
}

func (usecase *Appointment_usecase) Create_appointment(title, description string, doctorID int) (*model.Appointment, error) {
	if title == "" {
		return nil, missing_title
	}
	if doctorID == 0 {
		return nil, missing_doctor_id
	}

	exists, err := usecase.doctor_client.Doctor_exists(fmt.Sprintf("%d", doctorID))
	if err != nil {
		return nil, E_doctor_service_unavailable
	}
	if !exists {
		return nil, doctor_not_found
	}

	current_time := time.Now()
	appoinment := &model.Appointment{
		Title:       title,
		Description: description,
		DoctorID:    doctorID,
		Status:      model.Status_new,
		CreatedAt:   current_time,
		UpdatedAt:   current_time,
	}

	err = usecase.repo.Save(appoinment)
	if err != nil {
		return nil, err
	}
	return appoinment, nil
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

	appointment.Status = new_status
	appointment.UpdatedAt = time.Now()
	err = usecase.repo.Update(appointment)
	if err != nil {
		return nil, err
	}
	return appointment, nil
}
