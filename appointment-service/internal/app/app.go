package app

import (
	"database/sql"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/cache"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/event"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/middleware"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/repository"
	transportgrpc "github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/transport/grpc"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/usecase"
	pb "github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/proto/appointmentpb"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	googlegrpc "google.golang.org/grpc"
)

func Run() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	database, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect PostgreSQL: %v", err)
	}
	defer database.Close()

	err = database.Ping()
	if err != nil {
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
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations applied successfully")

	natsURL := os.Getenv("NATS_URL")
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

	redisURL := os.Getenv("REDIS_URL")
	var cacheRepo cache.CacheRepository
	redisClient, err := cache.NewRedisClient(redisURL)
	if err != nil {
		log.Printf("WARN: Redis unavailable: %v. Running without cache.", err)
		cacheRepo = &cache.NoopCache{}
	} else {
		log.Println("Connected to Redis")
		cacheRepo = cache.NewRedisCacheRepository(redisClient)
	}

	rpm := 100
	val := os.Getenv("RATE_LIMIT_RPM")
	if val != "" {
		n, err := strconv.Atoi(val)
		if err == nil {
			rpm = n
		}
	}

	doctorAddr := os.Getenv("DOCTOR_SERVICE_ADDR")
	doctorClient, err := transportgrpc.New_gRPC_doctor_client(doctorAddr)
	if err != nil {
		log.Fatalf("Failed to connect to doctor service: %v", err)
	}

	repo := repository.New_postgres_appointment_repository(database)
	uc := usecase.New_appointment_usecase(repo, doctorClient, publisher, cacheRepo)
	handler := transportgrpc.New_appointment_handler(uc)

	rateLimiter := middleware.NewRateLimiterInterceptor(redisClient, rpm)
	grpcServer := googlegrpc.NewServer(googlegrpc.UnaryInterceptor(rateLimiter))
	pb.RegisterAppointmentServiceServer(grpcServer, handler)

	grpcPort := os.Getenv("GRPC_PORT")
	listen, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Appointment Service running on port :%s", grpcPort)
	err = grpcServer.Serve(listen)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
