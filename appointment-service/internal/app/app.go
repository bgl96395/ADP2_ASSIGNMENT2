package app

import (
	"database/sql"
	"log"
	"net"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/repository"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/transport/grpc"
	transportgrpc "github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/transport/grpc"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/usecase"
	pb "github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/proto/appointmentpb"

	_ "github.com/lib/pq"
	google_grpc "google.golang.org/grpc"
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
	doctor_client, err := grpc.New_gRPC_doctor_client("localhost:50051")
	if err != nil {
		log.Fatal("Failed to connect to doctor service: ", err)
	}

	ucecase := usecase.New_appointment_usecase(repository, doctor_client)
	handler := transportgrpc.New_appointment_handler(ucecase)

	grpc_server := google_grpc.NewServer()
	pb.RegisterAppointmentServiceServer(grpc_server, handler)

	listen, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal("Failed to listen: ", err)
	}

	log.Println("Appointment service on port :50052")
	if err := grpc_server.Serve(listen); err != nil {
		log.Fatal("Failed to serve: ", err)
	}
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
