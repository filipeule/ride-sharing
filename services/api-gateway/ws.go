package main

import (
	"encoding/json"
	"log"
	"net/http"
	grpcclient "ride-sharing/services/api-gateway/grpc_client"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
	pb "ride-sharing/shared/proto/driver"
)

var (
	connManager = messaging.NewConnectionManager()
)

func handleRidersWebSocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("websocket upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("no user ID provided")
		return
	}

	// add connection to manager
	connManager.Add(userID, conn)
	defer connManager.Remove(userID)

	// initialize queue consumers
	queues := []string{
		messaging.NotifyDriverNoDriversFoundQueue,
		messaging.NotifyDriverAssignQueue,
		messaging.NotifyPaymentSessionCreatedQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("failed to start consumer for queue: %s: %v\n", q, err)
		}
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading message: %v\n", err)
			break
		}

		log.Printf("received message %s\n", msg)
	}
}

func handleDriversWebSocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	ctx := r.Context()

	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("websocket upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("no user ID provided")
		return
	}

	// add connection to manager
	connManager.Add(userID, conn)

	packageSlug := r.URL.Query().Get("packageSlug")
	if packageSlug == "" {
		log.Println("no package slug provided")
		return
	}

	driverClient, err := grpcclient.NewDriverServiceClient()
	if err != nil {
		log.Printf("error creating driver service client: %v\n", err)
		return
	}

	driverStrct := &pb.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	}

	defer func() {
		connManager.Remove(userID)

		_, unErr := driverClient.Client.UnregisterDriver(r.Context(), driverStrct)
		if unErr != nil {
			log.Printf("error unregistering driver: %v\n", unErr)
		}

		driverClient.Close()
		log.Printf("driver unregistered: %s\n", driverStrct.DriverID)
	}()

	driverRes, err := driverClient.Client.RegisterDriver(r.Context(), driverStrct)
	if err != nil {
		log.Printf("error registering driver: %v\n", err)
		return
	}

	if err := connManager.SendMessage(userID, contracts.WSMessage{
		Type: contracts.DriverCmdRegister,
		Data: driverRes.Driver,
	}); err != nil {
		log.Printf("error sending message: %v", err)
		return
	}

	// initialize queue consumers
	queues := []string{
		messaging.DriverCmdTripRequestQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("failed to start consumer for queue: %s: %v\n", q, err)
		}
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading message: %v\n", err)
			break
		}

		type DriverMessage struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		var driverMessage DriverMessage
		if err := json.Unmarshal(message, &driverMessage); err != nil{
			log.Printf("error unmarshaling driver message: %v\n", err)
			continue
		}

		// handle the difference message type
		switch driverMessage.Type {
		case contracts.DriverCmdLocation:
			// handle driver location update in the future
			continue
		case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
			// foward the message to rabbitmq
			if err := rb.PublishMessage(ctx, driverMessage.Type, contracts.AmqpMessage{
				OwnerID: userID,
				Data: driverMessage.Data,
			}); err != nil {
				log.Printf("error publishing message to rabbitmq: %v\n", err)
			}
		default:
			log.Printf("unknown message type: %v\n", driverMessage.Type)
		}
	}
}
