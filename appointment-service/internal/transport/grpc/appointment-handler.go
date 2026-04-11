package grpc

import (
	"context"
	"strconv"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/model"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/usecase"
	appointmentpb "github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/proto/appointmentpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Appointment_handler struct {
	appointmentpb.UnimplementedAppointmentServiceServer
	usecase *usecase.Appointment_usecase
}

func New_appointment_handler(ucecase *usecase.Appointment_usecase) *Appointment_handler {
	return &Appointment_handler{usecase: ucecase}
}

func (handler *Appointment_handler) CreateAppointment(ctx context.Context, req *appointmentpb.CreateAppointmentRequest) (*appointmentpb.AppointmentResponse, error) {
	doctorID, err := strconv.Atoi(req.DoctorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid doctor_id")
	}
	appointment, err := handler.usecase.Create_appointment(req.Title, req.Description, doctorID)
	if err != nil {
		if err == usecase.E_doctor_service_unavailable {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &appointmentpb.AppointmentResponse{
		Id:          strconv.Itoa(appointment.ID),
		Title:       appointment.Title,
		Description: appointment.Description,
		DoctorId:    strconv.Itoa(appointment.DoctorID),
		Status:      string(appointment.Status),
		CreatedAt:   appointment.CreatedAt.String(),
		UpdatedAt:   appointment.UpdatedAt.String(),
	}, nil
}

func (handler *Appointment_handler) GetAppointment(ctx context.Context, req *appointmentpb.GetAppointmentRequest) (*appointmentpb.AppointmentResponse, error) {
	appointment, err := handler.usecase.Get_appointment(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &appointmentpb.AppointmentResponse{
		Id:          strconv.Itoa(appointment.ID),
		Title:       appointment.Title,
		Description: appointment.Description,
		DoctorId:    strconv.Itoa(appointment.DoctorID),
		Status:      string(appointment.Status),
		CreatedAt:   appointment.CreatedAt.String(),
		UpdatedAt:   appointment.UpdatedAt.String(),
	}, nil
}

func (handler *Appointment_handler) ListAppointments(ctx context.Context, req *appointmentpb.ListAppointmentsRequest) (*appointmentpb.ListAppointmentsResponse, error) {
	list, err := handler.usecase.List_appoinments()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var result []*appointmentpb.AppointmentResponse
	for _, appointment := range list {
		result = append(result, &appointmentpb.AppointmentResponse{
			Id:          strconv.Itoa(appointment.ID),
			Title:       appointment.Title,
			Description: appointment.Description,
			DoctorId:    strconv.Itoa(appointment.DoctorID),
			Status:      string(appointment.Status),
			CreatedAt:   appointment.CreatedAt.String(),
			UpdatedAt:   appointment.UpdatedAt.String(),
		})
	}
	return &appointmentpb.ListAppointmentsResponse{Appointments: result}, nil
}

func (handler *Appointment_handler) UpdateAppointmentStatus(ctx context.Context, req *appointmentpb.UpdateStatusRequest) (*appointmentpb.AppointmentResponse, error) {
	appointment, err := handler.usecase.Update_status(req.Id, model.Status(req.Status))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &appointmentpb.AppointmentResponse{
		Id:          strconv.Itoa(appointment.ID),
		Title:       appointment.Title,
		Description: appointment.Description,
		DoctorId:    strconv.Itoa(appointment.DoctorID),
		Status:      string(appointment.Status),
		CreatedAt:   appointment.CreatedAt.String(),
		UpdatedAt:   appointment.UpdatedAt.String(),
	}, nil
}
