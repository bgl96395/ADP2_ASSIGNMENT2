package usecase

import (
	"errors"

	"doctor-service/internal/model"
	"doctor-service/internal/repository"
)

var (
	missing_fullname = errors.New("full name is required")
	missing_email    = errors.New("email is required")
	in_use_email     = errors.New("email already in use")
)

type Doctor_usecase struct {
	repository repository.Doctor_repository
}

func New_doctor_usecase(repo repository.Doctor_repository) *Doctor_usecase {
	return &Doctor_usecase{repository: repo}
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
		FullName:       fullName,
		Specialization: specialization,
		Email:          email,
	}
	err := usecase.repository.Save(doctor)
	if err != nil {
		return nil, err
	}
	return doctor, nil
}

func (usecase *Doctor_usecase) Get_doctor(id string) (*model.Doctor, error) {
	return usecase.repository.Find_by_ID(id)
}

func (usecase *Doctor_usecase) List_doctors() ([]*model.Doctor, error) {
	return usecase.repository.Find_all()
}
