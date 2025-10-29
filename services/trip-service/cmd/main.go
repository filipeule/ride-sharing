package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var (
	GrpcAddr = env.GetString("GRPC_ADDR", ":9093")
)

func main() {
	// initializing tracing
	traceCfg := tracing.Config{
		ServiceName: "trip-service",
		Environment: env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}

	sh, err := tracing.InitTracing(traceCfg)
	if err != nil {
		log.Fatalf("failed to initialize the tracer: %v\n", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}
	defer lis.Close()

	// connect to rabbitmq
	rabbitmqConn := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	rabbitmq, err := messaging.NewRabbitMQ(rabbitmqConn)
	if err != nil {
		log.Println(err)
		return
	}
	defer rabbitmq.Close()

	log.Println("starting rabbitmq connection")

	publisher := events.NewTripEventPublisher(rabbitmq)

	driverConsumer := events.NewDriverConsumer(rabbitmq, svc)
	go driverConsumer.Listen()

	// starting payment consumer
	paymentConsumer := events.NewPaymentConsumer(rabbitmq, svc)
	go paymentConsumer.Listen()

	// starting the grpc server
	grpcServer := grpcserver.NewServer()
	grpc.NewGRPCHandler(grpcServer, svc, publisher)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	log.Printf("starting grpc server trip service on port %s\n", lis.Addr())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve: %v\n", err)
			cancel()
		}
	}()

	// wait for the shutdown signal
	<-ctx.Done()
	log.Println("shutting down the server")
	grpcServer.GracefulStop()
}
