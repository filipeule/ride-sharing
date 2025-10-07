package main

import (
	"encoding/json"
	"net/http"
	grpcclient "ride-sharing/services/api-gateway/grpc_client"
	"ride-sharing/shared/contracts"
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
		http.Error(w, "failed to get trip preview from trip service: " + err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "error parsing trip start body: " + err.Error(), http.StatusBadRequest)
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
		http.Error(w, "error getting trip service client: " + err.Error(), http.StatusInternalServerError)
		return
	}
	defer tripService.Close()

	tripStartRes, err := tripService.Client.CreateTrip(r.Context(), reqBody.ToProto())
	if err != nil {
		http.Error(w, "error creating trip: " + err.Error(), http.StatusInternalServerError)
		return
	}

	res := contracts.APIResponse{
		Data: tripStartRes,
	}

	writeJSON(w, http.StatusCreated, res)
}