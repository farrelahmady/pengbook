package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"pengbook/api/internal/middleware"
	"pengbook/api/pkg/response"
	"pengbook/api/pkg/validator"
)

// Handler is the HTTP handler for the account module.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns the chi router specific to accounts.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/summary", h.GetSummary)
	r.Get("/posting", h.GetPostingAccounts)
	r.Post("/", h.Create)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// GetSummary handles GET /api/v1/accounts/summary
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	summary, err := h.service.GetSummary(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get summary")
		return
	}

	response.Success(w, http.StatusOK, summary)
}

// GetPostingAccounts handles GET /api/v1/accounts/posting
func (h *Handler) GetPostingAccounts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accounts, err := h.service.GetPostingAccounts(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get posting accounts")
		return
	}

	response.Success(w, http.StatusOK, accounts)
}

// Create handles POST /api/v1/accounts
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	created, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrCodeExists) {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	response.Success(w, http.StatusCreated, created)
}

// Update handles PUT /api/v1/accounts/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accountID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid account id")
		return
	}

	var req UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	updated, err := h.service.Update(r.Context(), userID, accountID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "account not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update account")
		return
	}

	response.Success(w, http.StatusOK, updated)
}

// Delete handles DELETE /api/v1/accounts/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accountID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid account id")
		return
	}

	err = h.service.Delete(r.Context(), userID, accountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "account not found")
			return
		}
		if errors.Is(err, ErrHasChildren) {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, ErrHasJournalLines) {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete account")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "account deleted"})
}
