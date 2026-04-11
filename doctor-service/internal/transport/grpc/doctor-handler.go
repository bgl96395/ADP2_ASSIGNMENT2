package grpc

import (
	"context"
	"strconv"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/usecase"
	pb "github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/proto/doctorpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Doctor_handler struct {
	usecase *usecase.Doctor_usecase
	pb.UnimplementedDoctorServiceServer
}

func New_doctor_handler(usecase *usecase.Doctor_usecase) *Doctor_handler {
	return &Doctor_handler{usecase: usecase}
}

func (handler *Doctor_handler) CreateDoctor(ctx context.Context, req *pb.CreateDoctorRequest) (*pb.DoctorResponse, error) {
	doctor, err := handler.usecase.Create_doctor(req.FullName, req.Specialization, req.Email)
	if err != nil {
		if err.Error() == "full name is required" || err.Error() == "email is required" {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if err.Error() == "email already in use" {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DoctorResponse{
		Id:             strconv.Itoa(doctor.ID),
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}, nil
}

func (handler *Doctor_handler) GetDoctor(ctx context.Context, req *pb.GetDoctorRequest) (*pb.DoctorResponse, error) {
	doctor, err := handler.usecase.Get_doctor(req.Id)
	if err != nil {
		if err.Error() == "doctor not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DoctorResponse{
		Id:             strconv.Itoa(doctor.ID),
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}, nil
}

func (handler *Doctor_handler) ListDoctors(ctx context.Context, req *pb.ListDoctorsRequest) (*pb.ListDoctorsResponse, error) {
	list, err := handler.usecase.List_doctors()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var doctors []*pb.DoctorResponse
	for _, doctor := range list {
		doctors = append(doctors, &pb.DoctorResponse{
			Id:             strconv.Itoa(doctor.ID),
			FullName:       doctor.FullName,
			Specialization: doctor.Specialization,
			Email:          doctor.Email,
		})
	}
	return &pb.ListDoctorsResponse{Doctors: doctors}, nil
}
