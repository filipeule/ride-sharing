package main

import (
	"context"
	pb "ride-sharing/shared/proto/driver"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	Service *Service
	pb.UnimplementedDriverServiceServer
}

func NewGrpcHandler(s *grpc.Server, service *Service) {
	handler := &grpcHandler{
		Service: service,
	}
	pb.RegisterDriverServiceServer(s, handler)
}

func (h *grpcHandler) RegisterDriver(ctx context.Context, r *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driverRes, err := h.Service.RegisterDriver(r.GetDriverID(), r.GetPackageSlug())
	if err != nil {
		return nil, status.Errorf(codes.Unimplemented, "failed to register driver: %v", err)
	}

	return &pb.RegisterDriverResponse{
		Driver: driverRes,
	}, nil
}

func (h *grpcHandler) UnregisterDriver(ctx context.Context, r *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	h.Service.UnregisterDriver(r.GetDriverID())

	return &pb.RegisterDriverResponse{
		Driver: &pb.Driver{
			Id: r.GetDriverID(),
		},
	}, nil
}