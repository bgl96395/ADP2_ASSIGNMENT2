package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/usecase"

	pb "github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DoctorGRPCHandler struct {
	pb.UnimplementedDoctorServiceServer
	usecase *usecase.Doctor_usecase
}

func NewDoctorGRPCHandler(uc *usecase.Doctor_usecase) *DoctorGRPCHandler {
	return &DoctorGRPCHandler{usecase: uc}
}

func (h *DoctorGRPCHandler) CreateDoctor(ctx context.Context, req *pb.CreateDoctorRequest) (*pb.DoctorResponse, error) {
	doctor, err := h.usecase.Create_doctor(req.FullName, req.Specialization, req.Email)
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
		Id:             fmt.Sprintf("%d", doctor.ID),
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}, nil
}

func (h *DoctorGRPCHandler) GetDoctor(ctx context.Context, req *pb.GetDoctorRequest) (*pb.DoctorResponse, error) {
	doctor, err := h.usecase.Get_doctor(req.Id)
	if err != nil {
		if errors.Is(err, errors.New("doctor not found")) || err.Error() == "doctor not found" {
			return nil, status.Error(codes.NotFound, "doctor not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DoctorResponse{
		Id:             fmt.Sprintf("%d", doctor.ID),
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}, nil
}

func (h *DoctorGRPCHandler) ListDoctors(ctx context.Context, req *pb.ListDoctorsRequest) (*pb.ListDoctorsResponse, error) {
	list, err := h.usecase.List_doctors()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var doctors []*pb.DoctorResponse
	for _, d := range list {
		doctors = append(doctors, &pb.DoctorResponse{
			Id:             fmt.Sprintf("%d", d.ID),
			FullName:       d.FullName,
			Specialization: d.Specialization,
			Email:          d.Email,
		})
	}
	return &pb.ListDoctorsResponse{Doctors: doctors}, nil
}
