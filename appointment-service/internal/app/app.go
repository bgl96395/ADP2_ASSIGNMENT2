package app

import (
	"appointment-service/internal/repository"
	transport "appointment-service/internal/transport/http"
	"appointment-service/internal/usecase"
	"database/sql"
	"log"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

func Run() {
	database, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=appointment sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect Postgres: ", err)
	}
	defer database.Close()

	err = database.Ping()
	if err != nil {
		log.Fatal("database is not reachable: ", err)
	}

	err = migrate(database)
	if err != nil {
		log.Fatal("migration failed: ", err)
	}

	repository := repository.New_postgres_appointment_repository(database)
	doctor_client := transport.New_HTTP_doctor_client("http://localhost:8080")
	usecase := usecase.New_appointment_usecase(repository, doctor_client)
	handler := transport.New_appointment_handler(usecase)

	register := gin.Default()
	handler.Register_routes(register)
	register.Run(":8081")
}

func migrate(database *sql.DB) error {
	_, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS appointments (
			id serial PRIMARY KEY,
			title varchar NOT NULL,
			description varchar,
			doctor_id integer NOT NULL,
			status varchar NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)
	`)
	return err
}
