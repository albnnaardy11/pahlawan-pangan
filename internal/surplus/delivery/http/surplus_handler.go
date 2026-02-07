package http

import (
	"encoding/json"
	"net/http"

	"github.com/albnnaardy11/pahlawan-pangan/internal/domain"
	"github.com/albnnaardy11/pahlawan-pangan/pkg/errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type SurplusHandler struct {
	Usecase  domain.SurplusUsecase
	validate *validator.Validate
}

func NewSurplusHandler(r *chi.Mux, us domain.SurplusUsecase) {
	handler := &SurplusHandler{
		Usecase:  us,
		validate: validator.New(),
	}

	r.Route("/api/v1/surplus", func(r chi.Router) {
		r.Post("/", handler.PostSurplus)
		r.Get("/marketplace", handler.GetMarketplace)
	})
}

func (h *SurplusHandler) PostSurplus(w http.ResponseWriter, r *http.Request) {
	var item domain.SurplusItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		h.respondWithError(w, errors.ErrBadRequest)
		return
	}

	// Validate using go-playground/validator v10
	if err := h.validate.Struct(item); err != nil {
		h.respondWithError(w, errors.NewAppError("ERR-VALIDATION-FAILED", err.Error(), http.StatusUnprocessableEntity))
		return
	}

	if err := h.Usecase.PostSurplus(r.Context(), &item); err != nil {
		h.respondWithError(w, errors.ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

func (h *SurplusHandler) GetMarketplace(w http.ResponseWriter, r *http.Request) {
	// Logic to call usecase
}

func (h *SurplusHandler) respondWithError(w http.ResponseWriter, err *errors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	_ = json.NewEncoder(w).Encode(err)
}
