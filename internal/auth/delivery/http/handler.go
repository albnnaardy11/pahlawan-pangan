package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/albnnaardy11/pahlawan-pangan/internal/auth/domain"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	Usecase domain.AuthUsecase
}

func NewAuthHandler(uc domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{Usecase: uc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string          `json:"email"`
		Password string          `json:"password"`
		FullName string          `json:"full_name"`
		Role     domain.UserRole `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user := &domain.User{
		Email:    req.Email,
		FullName: req.FullName,
		Role:     req.Role,
	}

	if err := h.Usecase.Register(r.Context(), user, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "user registered with Argon2id protection"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	res, err := h.Usecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	
	if err := h.Usecase.Logout(r.Context(), token); err != nil {
		http.Error(w, "failed to logout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.Usecase.RequestOTP(r.Context(), req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "OTP sent via NATS background worker"})
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Code   string `json:"code"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	success, err := h.Usecase.VerifyOTP(r.Context(), req.UserID, req.Code)
	if err != nil || !success {
		http.Error(w, "invalid or expired OTP", http.StatusUnauthorized)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "OTP verified successfully"})
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Post("/otp/request", h.RequestOTP)
	r.Post("/otp/verify", h.VerifyOTP)
	return r
}
