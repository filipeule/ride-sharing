package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	grpcclient "ride-sharing/services/api-gateway/grpc_client"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// validation
	if reqBody.UserID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	tripService, err := grpcclient.NewTripServiceClient()
	if err != nil {
		http.Error(w, "failed to get trip service grpc client", http.StatusInternalServerError)
		return
	}
	defer tripService.Close()

	tripPreview, err := tripService.Client.PreviewTrip(r.Context(), reqBody.ToProto())
	if err != nil {
		http.Error(w, "failed to get trip preview from trip service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res := contracts.APIResponse{
		Data: tripPreview,
	}

	writeJSON(w, http.StatusCreated, res)
}

func handleTripStart(w http.ResponseWriter, r *http.Request) {
	var reqBody startTripRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "error parsing trip start body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// validate userID
	if reqBody.UserID == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	tripService, err := grpcclient.NewTripServiceClient()
	if err != nil {
		http.Error(w, "error getting trip service client: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tripService.Close()

	tripStartRes, err := tripService.Client.CreateTrip(r.Context(), reqBody.ToProto())
	if err != nil {
		http.Error(w, "error creating trip: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res := contracts.APIResponse{
		Data: tripStartRes,
	}

	writeJSON(w, http.StatusCreated, res)
}

func handleStripeWebhook(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	webhookKey := env.GetString("STRIPE_WEBHOOK_KEY", "")
	if webhookKey == "" {
		log.Printf("webhook key is required")
		return
	}

	event, err := webhook.ConstructEventWithOptions(
		body,
		r.Header.Get("Stripe-Signature"),
		webhookKey,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		log.Printf("error verifying webhook signature: %v", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	log.Printf("received Stripe event: %v", event)

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession

		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("error parsing webhook JSON: %v", err)
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		payload := messaging.PaymentStatusUpdateData{
			TripID:   session.Metadata["trip_id"],
			UserID:   session.Metadata["user_id"],
			DriverID: session.Metadata["driver_id"],
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("error marshalling payload: %v", err)
			http.Error(w, "failed to marshal payload", http.StatusInternalServerError)
			return
		}

		message := contracts.AmqpMessage{
			OwnerID: session.Metadata["user_id"],
			Data:    payloadBytes,
		}

		if err := rb.PublishMessage(
			r.Context(),
			contracts.PaymentEventSuccess,
			message,
		); err != nil {
			log.Printf("error publishing payment event: %v", err)
			http.Error(w, "failed to publish payment event", http.StatusInternalServerError)
			return
		}
	}
}
