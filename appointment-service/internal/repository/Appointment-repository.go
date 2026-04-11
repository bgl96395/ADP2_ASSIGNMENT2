package repository

import (
	"appointment-service/internal/model"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

type Appointment_repository interface {
	Save(appointment *model.Appointment) error
	Find_by_ID(id string) (*model.Appointment, error)
	Find_all() ([]*model.Appointment, error)
	Update(appointment *model.Appointment) error
}

type Postgres_appointment_repository struct {
	database *sql.DB
}

func New_postgres_appointment_repository(database *sql.DB) *Postgres_appointment_repository {
	return &Postgres_appointment_repository{database: database}
}

func (repository *Postgres_appointment_repository) Save(appointment *model.Appointment) error {
	_, err := repository.database.Exec(`INSERT INTO appointments (title,description,doctor_id,status,created_at,updated_at) VALUES($1, $2, $3, $4, $5, $6)`, appointment.Title, appointment.Description, appointment.DoctorID, appointment.Status, appointment.CreatedAt, appointment.UpdatedAt)
	return err
}

func (repository *Postgres_appointment_repository) Find_by_ID(id string) (*model.Appointment, error) {
	row := repository.database.QueryRow(`SELECT id,title,description,doctor_id,status,created_at,updated_at FROM appointments WHERE id=$1`, id)
	appointment := &model.Appointment{}
	err := row.Scan(&appointment.ID, &appointment.Title, &appointment.Description, &appointment.DoctorID, &appointment.Status, &appointment.CreatedAt, &appointment.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("appointment not found")
	}
	if err != nil {
		return nil, err
	}
	return appointment, nil
}

func (repository *Postgres_appointment_repository) Find_all() ([]*model.Appointment, error) {
	rows, err := repository.database.Query(`SELECT * FROM appointments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Appointment
	for rows.Next() {
		appointment := &model.Appointment{}
		err := rows.Scan(&appointment.ID, &appointment.Title, &appointment.Description, &appointment.DoctorID, &appointment.Status, &appointment.CreatedAt, &appointment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, appointment)
	}
	return list, nil
}

func (repository *Postgres_appointment_repository) Update(appointment *model.Appointment) error {
	result, err := repository.database.Exec(`UPDATE appointments SET status = $1, updated_at = $2 WHERE id=$3`, &appointment.Status, &appointment.UpdatedAt, &appointment.ID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("appointment not found")
	}
	return nil
}
