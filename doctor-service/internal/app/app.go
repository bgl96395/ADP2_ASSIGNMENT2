package app

import (
	"database/sql"
	"log"
	"net"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/repository"
	transporthttp "github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/transport/grpc"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/usecase"

	pb "github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/proto/doctorpb"
	_ "github.com/lib/pq"
	googlegrpc "google.golang.org/grpc"
)

func Run() {
	database, err := sql.Open("postgres", "host= localhost port=5432 user=postgres password=postgres dbname=doctor sslmode=disable")
	if err != nil {
		log.Fatal("Failde to connect Postgresql error: ", err)
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

	repository := repository.New_postgre_doctor_repository(database)
	usecase := usecase.New_doctor_usecase(repository)
	handler := transporthttp.New_doctor_handler(usecase)

	grpc_server := googlegrpc.NewServer()
	pb.RegisterDoctorServiceServer(grpc_server, handler)

	listen, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("Failed to listen", err)
	}

	log.Println("Server running on port :50051")
	err = grpc_server.Serve(listen)
	if err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

func migrate(database *sql.DB) error {
	_, err := database.Exec(`
			CREATE TABLE IF NOT EXISTS doctors (
				id serial PRIMARY KEY,
				full_name varchar NOT NULL,
				specialization varchar,
				email varchar NOT NULL UNIQUE
			)
		`)
	return err
}
