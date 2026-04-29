package grpc

import (
	"context"

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

func New_appointment_handler(uc *usecase.Appointment_usecase) *Appointment_handler {
	return &Appointment_handler{usecase: uc}
}

func (handler *Appointment_handler) CreateAppointment(ctx context.Context, req *appointmentpb.CreateAppointmentRequest) (*appointmentpb.AppointmentResponse, error) {
	appointment, err := handler.usecase.Create_appointment(req.Title, req.Description, req.DoctorId)
	if err != nil {
		if err == usecase.E_doctor_service_unavailable {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
		if err.Error() == "doctor not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toResponse(appointment), nil
}

func (handler *Appointment_handler) GetAppointment(ctx context.Context, req *appointmentpb.GetAppointmentRequest) (*appointmentpb.AppointmentResponse, error) {
	appointment, err := handler.usecase.Get_appointment(req.Id)
	if err != nil {
		if err.Error() == "appointment not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toResponse(appointment), nil
}

func (handler *Appointment_handler) ListAppointments(ctx context.Context, req *appointmentpb.ListAppointmentsRequest) (*appointmentpb.ListAppointmentsResponse, error) {
	list, err := handler.usecase.List_appoinments()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var result []*appointmentpb.AppointmentResponse
	for _, a := range list {
		result = append(result, toResponse(a))
	}
	return &appointmentpb.ListAppointmentsResponse{Appointments: result}, nil
}

func (handler *Appointment_handler) UpdateAppointmentStatus(ctx context.Context, req *appointmentpb.UpdateStatusRequest) (*appointmentpb.AppointmentResponse, error) {
	appointment, err := handler.usecase.Update_status(req.Id, model.Status(req.Status))
	if err != nil {
		if err.Error() == "appointment not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toResponse(appointment), nil
}

func toResponse(a *model.Appointment) *appointmentpb.AppointmentResponse {
	return &appointmentpb.AppointmentResponse{
		Id:          a.ID,
		Title:       a.Title,
		Description: a.Description,
		DoctorId:    a.DoctorID,
		Status:      string(a.Status),
		CreatedAt:   a.CreatedAt.String(),
		UpdatedAt:   a.UpdatedAt.String(),
	}
}
