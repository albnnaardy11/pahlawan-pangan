package http

import (
	"encoding/json"
	"net/http"

	"github.com/albnnaardy11/pahlawan-pangan/internal/domain"
	"github.com/go-chi/chi/v5"
)

type SurplusHandler struct {
	Usecase domain.SurplusUsecase
}

func NewSurplusHandler(r *chi.Mux, us domain.SurplusUsecase) {
	handler := &SurplusHandler{
		Usecase: us,
	}

	r.Route("/api/v1/surplus", func(r chi.Router) {
		r.Post("/", handler.PostSurplus)
		r.Get("/marketplace", handler.GetMarketplace)
	})
}

func (h *SurplusHandler) PostSurplus(w http.ResponseWriter, r *http.Request) {
	var item domain.SurplusItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Usecase.PostSurplus(r.Context(), &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *SurplusHandler) GetMarketplace(w http.ResponseWriter, r *http.Request) {
	// Logic to call usecase
}
