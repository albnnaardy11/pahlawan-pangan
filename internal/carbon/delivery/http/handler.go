package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/albnnaardy11/pahlawan-pangan/internal/carbon/service"
)

type CarbonHandler struct {
	Service *service.CarbonService
}

func NewCarbonHandler(svc *service.CarbonService) *CarbonHandler {
	return &CarbonHandler{Service: svc}
}

// GET /api/v1/carbon/certificate/{vendor_id}?year=2024
func (h *CarbonHandler) GetESGCertificate(w http.ResponseWriter, r *http.Request) {
	vendorID := chi.URLParam(r, "vendor_id")
	yearStr := r.URL.Query().Get("year")
	year := time.Now().Year() // Default to current year

	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	report, err := h.Service.GenerateESGReport(r.Context(), vendorID, year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}

func (h *CarbonHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/certificate/{vendor_id}", h.GetESGCertificate)
	return r
}
