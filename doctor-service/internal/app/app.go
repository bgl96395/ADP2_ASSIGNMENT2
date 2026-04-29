package app

import (
	"database/sql"
	"log"
	"net"
	"os"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/event"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/repository"
	transportgrpc "github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/transport/grpc"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/usecase"
	pb "github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/proto/doctorpb"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	googlegrpc "google.golang.org/grpc"
)

func Run() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "host=localhost port=5432 user=postgres password=postgres dbname=doctor sslmode=disable"
	}

	database, err := sql.Open("postgres", databaseURL)
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

	// Connect to NATS (best-effort)
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	var publisher event.EventPublisher
	natsPublisher, err := event.NewNATSPublisher(natsURL)
	if err != nil {
		log.Printf("WARN: failed to connect to NATS broker: %v. Events will be dropped.", err)
		publisher = &event.NoopPublisher{}
	} else {
		log.Println("Connected to NATS broker")
		publisher = natsPublisher
		defer natsPublisher.Close()
	}

	repo := repository.New_postgre_doctor_repository(database)
	uc := usecase.New_doctor_usecase(repo, publisher)
	handler := transportgrpc.New_doctor_handler(uc)

	grpcServer := googlegrpc.NewServer()
	pb.RegisterDoctorServiceServer(grpcServer, handler)

	listen, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("Doctor Service running on port :50051")
	if err = grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
