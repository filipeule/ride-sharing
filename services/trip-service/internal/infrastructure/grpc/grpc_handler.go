package grpc

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedTripServiceServer
	service   domain.TripService
	publisher *events.TripEventPublisher
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService, publisher *events.TripEventPublisher) *grpcHandler {
	handler := &grpcHandler{
		service:   service,
		publisher: publisher,
	}

	pb.RegisterTripServiceServer(server, handler)

	return handler
}

func (gh *grpcHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickupGrpc := req.StartLocation
	destinationGrpc := req.EndLocation

	pickup := &types.Coordinate{
		Latitude:  pickupGrpc.Latitude,
		Longitude: pickupGrpc.Longitude,
	}

	destination := &types.Coordinate{
		Latitude:  destinationGrpc.Latitude,
		Longitude: destinationGrpc.Longitude,
	}

	route, err := gh.service.GetRoute(ctx, pickup, destination)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	userID := req.GetUserID()

	estimatedFares := gh.service.EstimatePackagesPriceWithRoute(route)
	fares, err := gh.service.GenerateTripFares(ctx, estimatedFares, userID, route)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate the ride fares: %v", err)
	}

	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}

func (gh *grpcHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	fareID := req.GetRideFareID()
	userID := req.GetUserID()

	rideFare, err := gh.service.GetAndValidateFare(ctx, fareID, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error validating fare: %v", err)
	}

	tripModel, err := gh.service.CreateTrip(ctx, rideFare)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create trip: %v", err)
	}

	if err := gh.publisher.PublishTripCreated(ctx, tripModel); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish trip created event: %v", err)
	}

	return &pb.CreateTripResponse{
		TripID: tripModel.ID.Hex(),
	}, nil
}
