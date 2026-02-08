package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/albnnaardy11/pahlawan-pangan/internal/community/domain"
)

type CommunityHandler struct {
	Usecase domain.CommunityUsecase
}

func NewCommunityHandler(uc domain.CommunityUsecase) *CommunityHandler {
	return &CommunityHandler{Usecase: uc}
}

func (h *CommunityHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	var review domain.Review
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	review.UserID = "user-123" // Mock auth context
	review.CreatedAt = time.Now()
	review.IsVerified = true // Mock transaction check

	if err := h.Usecase.SubmitReview(r.Context(), review); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Review submitted"})
}

func (h *CommunityHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "target_id")
	reviews, err := h.Usecase.GetReviews(r.Context(), targetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reviews)
}

func (h *CommunityHandler) GetReputation(w http.ResponseWriter, r *http.Request) {
	vendorID := chi.URLParam(r, "id")
	summary, err := h.Usecase.GetVendorReputation(r.Context(), vendorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

func (h *CommunityHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/reviews", h.CreateReview)
	r.Get("/reviews/{target_id}", h.GetReviews)
	r.Get("/vendor/{id}/reputation", h.GetReputation)
	return r
}
