package http

import (
	"encoding/json"
	"net/http"

	"github.com/albnnaardy11/pahlawan-pangan/internal/domain"
	"github.com/go-chi/chi/v5"
)

type DisputeHandler struct {
	Usecase domain.DisputeUsecase
}

func NewDisputeHandler(r chi.Router, us domain.DisputeUsecase) {
	handler := &DisputeHandler{
		Usecase: us,
	}

	r.Route("/api/v1/dispute", func(r chi.Router) {
		r.Post("/", handler.RaiseDispute)
	})
}

func (h *DisputeHandler) RaiseDispute(w http.ResponseWriter, r *http.Request) {
	var dispute domain.Dispute
	if err := json.NewDecoder(r.Body).Decode(&dispute); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Usecase.RaiseDispute(r.Context(), &dispute); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Dispute submitted. Our legal team and AI will review the freshness quality.",
		"status":  "PENDING_REVIEW",
	})
}
