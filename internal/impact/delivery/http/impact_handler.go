package http

import (
	"encoding/json"
	"net/http"

	"github.com/albnnaardy11/pahlawan-pangan/internal/domain"
	"github.com/go-chi/chi/v5"
)

type ImpactHandler struct {
	Usecase domain.ImpactUsecase
}

func NewImpactHandler(r chi.Router, us domain.ImpactUsecase) {
	handler := &ImpactHandler{
		Usecase: us,
	}

	r.Route("/api/v1/impact", func(r chi.Router) {
		r.Get("/leaderboard", handler.GetNationalLeaderboard)
		r.Get("/user/{id}", handler.GetUserImpact)
		r.Get("/share/{claim_id}", handler.GenerateShareCard)
	})
}

func (h *ImpactHandler) GetNationalLeaderboard(w http.ResponseWriter, r *http.Request) {
	res, err := h.Usecase.GetNationalLeaderboard(r.Context(), "Indonesia")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *ImpactHandler) GetUserImpact(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	res, err := h.Usecase.GetUserImpact(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *ImpactHandler) GenerateShareCard(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "claim_id")
	url, err := h.Usecase.GenerateShareCard(r.Context(), claimID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"share_url": url})
}
