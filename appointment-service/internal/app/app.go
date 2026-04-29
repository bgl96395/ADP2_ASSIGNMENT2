package app

import (
	"database/sql"
	"log"
	"net"
	"os"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/event"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/repository"
	transportgrpc "github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/transport/grpc"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/usecase"
	pb "github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/proto/appointmentpb"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	googlegrpc "google.golang.org/grpc"
)

func Run() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "host=localhost port=5432 user=postgres password=postgres dbname=appointment sslmode=disable"
	}

	database, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect PostgreSQL: %v", err)
	}
	defer database.Close()

	if err = database.Ping(); err != nil {
		log.Fatalf("Database is not reachable: %v", err)
	}

	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migrate driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		log.Fatalf("Failed to init migrations: %v", err)
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations applied successfully")

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	var publisher event.EventPublisher
	natsPublisher, err := event.NewNATSPublisher(natsURL)
	if err != nil {
		log.Printf("WARN: failed to connect to NATS: %v. Events will be dropped.", err)
		publisher = &event.NoopPublisher{}
	} else {
		log.Println("Connected to NATS broker")
		publisher = natsPublisher
		defer natsPublisher.Close()
	}

	doctorAddr := os.Getenv("DOCTOR_SERVICE_ADDR")
	if doctorAddr == "" {
		doctorAddr = "localhost:50051"
	}
	doctorClient, err := transportgrpc.New_gRPC_doctor_client(doctorAddr)
	if err != nil {
		log.Fatalf("Failed to connect to doctor service: %v", err)
	}

	repo := repository.New_postgres_appointment_repository(database)
	uc := usecase.New_appointment_usecase(repo, doctorClient, publisher)
	handler := transportgrpc.New_appointment_handler(uc)

	grpcServer := googlegrpc.NewServer()
	pb.RegisterAppointmentServiceServer(grpcServer, handler)

	listen, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("Appointment Service running on port :50052")
	err = grpcServer.Serve(listen)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
