package repository

import (
	"database/sql"
	"errors"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/model"
	_ "github.com/lib/pq"
)

type Doctor_repository interface {
	Save(doctor *model.Doctor) error
	Find_by_ID(id string) (*model.Doctor, error)
	Find_all() ([]*model.Doctor, error)
	Exists_by_email(email string) bool
}

type Postgres_doctor_repository struct {
	database *sql.DB
}

func New_postgre_doctor_repository(database *sql.DB) *Postgres_doctor_repository {
	return &Postgres_doctor_repository{database: database}
}

func (repository *Postgres_doctor_repository) Save(doctor *model.Doctor) error {
	_, err := repository.database.Exec(
		`INSERT INTO doctors (id, full_name, specialization, email) VALUES ($1, $2, $3, $4)`,
		doctor.ID, doctor.FullName, doctor.Specialization, doctor.Email,
	)
	return err
}

func (repository *Postgres_doctor_repository) Find_by_ID(id string) (*model.Doctor, error) {
	row := repository.database.QueryRow(
		`SELECT id, full_name, specialization, email FROM doctors WHERE id = $1`, id,
	)
	doctor := &model.Doctor{}
	err := row.Scan(&doctor.ID, &doctor.FullName, &doctor.Specialization, &doctor.Email)
	if err == sql.ErrNoRows {
		return nil, errors.New("doctor not found")
	}
	if err != nil {
		return nil, err
	}
	return doctor, nil
}

func (repository *Postgres_doctor_repository) Find_all() ([]*model.Doctor, error) {
	rows, err := repository.database.Query(`SELECT id, full_name, specialization, email FROM doctors`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Doctor
	for rows.Next() {
		doctor := &model.Doctor{}
		err := rows.Scan(&doctor.ID, &doctor.FullName, &doctor.Specialization, &doctor.Email)
		if err != nil {
			return nil, err
		}
		list = append(list, doctor)
	}
	return list, nil
}

func (repository *Postgres_doctor_repository) Exists_by_email(email string) bool {
	var count int
	repository.database.QueryRow(`SELECT COUNT(*) FROM doctors WHERE email = $1`, email).Scan(&count)
	return count > 0
}
