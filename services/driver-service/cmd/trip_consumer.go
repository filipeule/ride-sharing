package main

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type tripConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  *Service
}

func NewTripConsumer(rabbit *messaging.RabbitMQ, service *Service) *tripConsumer {
	return &tripConsumer{
		rabbitmq: rabbit,
		service:  service,
	}
}

func (c *tripConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessage(
		messaging.FindAvailableDriversQueue,
		func(ctx context.Context, msg amqp091.Delivery) error {
			var tripEvent contracts.AmqpMessage
			if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
				log.Printf("error unmarshaling message: %v\n", err)
				return err
			}

			var payload messaging.TripEventData
			if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
				log.Printf("error unmarshaling message: %v\n", err)
				return err
			}

			log.Printf("driver received message: %+v\n", payload)

			switch msg.RoutingKey {
			case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
				return c.handleFindAndNotifyDrivers(ctx, payload)
			}

			log.Printf("unknown trip event: %+v\n", payload)

			return nil
		})
}

func (c *tripConsumer) handleFindAndNotifyDrivers(ctx context.Context, payload messaging.TripEventData) error {
	suitableIDs := c.service.FindAvailableDrivers(payload.Trip.SelectedFare.PackageSlug)

	log.Printf("found suitable drivers: %v\n", len(suitableIDs))

	if len(suitableIDs) == 0 {
		// notify the user that no drivers are available
		if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventNoDriversFound, contracts.AmqpMessage{
			OwnerID: payload.Trip.UserID,
		}); err != nil {
			log.Printf("failed to publish message to exchange: %v\n", err)
			return err
		}

		return nil
	}

	suitableDriverID := suitableIDs[0]

	marshalledEvent, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// notify the driver about a potential trip
	if err := c.rabbitmq.PublishMessage(ctx, contracts.DriverCmdTripRequest, contracts.AmqpMessage{
		OwnerID: suitableDriverID,
		Data: marshalledEvent,
	}); err != nil {
		log.Printf("failed to publish message to exchange: %v\n", err)
		return err
	}

	return nil
}
