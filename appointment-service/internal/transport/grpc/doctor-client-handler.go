package grpc

import (
	"context"

	doctorpb "github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/proto/doctorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GRPC_doctor_client struct {
	client doctorpb.DoctorServiceClient
}

func New_gRPC_doctor_client(addr string) (*GRPC_doctor_client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPC_doctor_client{client: doctorpb.NewDoctorServiceClient(conn)}, nil
}

func (c *GRPC_doctor_client) Doctor_exists(doctorID string) (bool, error) {
	_, err := c.client.GetDoctor(context.Background(), &doctorpb.GetDoctorRequest{Id: doctorID})
	if err != nil {
		status_of, _ := status.FromError(err)
		if status_of.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
