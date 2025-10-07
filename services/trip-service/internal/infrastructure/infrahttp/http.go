package infrahttp

import (
	"encoding/json"
	"net/http"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/types"
)

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

type HttpHandler struct {
	Svc *service.Service
}

func (h *HttpHandler) HandleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "failed to parse request body", http.StatusBadRequest)
		return
	}

	if reqBody.UserID == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	t, err := h.Svc.GetRoute(ctx, &reqBody.Pickup, &reqBody.Destination)
	if err != nil {
		http.Error(w, "error creating trip", http.StatusInternalServerError)
		return
	}

	resp := contracts.APIResponse{
		Data: t,
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}
