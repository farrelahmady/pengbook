package journal

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"pengbook/api/internal/middleware"
	"pengbook/api/pkg/response"
	"pengbook/api/pkg/validator"
)

// Handler is the HTTP handler for the journal module.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns the chi router specific to journals.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.GetAllScrollView)
	r.Get("/summary", h.GetTotalSummary)
	r.Post("/", h.Create)
	r.Put("/{id}", h.Update)
	return r
}

// GetAllScrollView handles GET /api/v1/journals
func (h *Handler) GetAllScrollView(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	cursor := r.URL.Query().Get("cursor")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	var accountIDs []int64
	if idsStr := r.URL.Query().Get("accountIds"); idsStr != "" {
		for _, idStr := range strings.Split(idsStr, ",") {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil {
				accountIDs = append(accountIDs, id)
			}
		}
	}

	req := ListRequest{
		Limit:      limit,
		Cursor:     &cursor,
		StartDate:  startDate,
		EndDate:    endDate,
		AccountIDs: accountIDs,
	}

	result, err := h.service.GetAllScrollView(r.Context(), userID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get journals")
		return
	}

	response.Success(w, http.StatusOK, result)
}

// GetTotalSummary handles GET /api/v1/journals/summary
func (h *Handler) GetTotalSummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	summary, err := h.service.GetTotalSummary(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get summary")
		return
	}

	response.Success(w, http.StatusOK, summary)
}

// Create handles POST /api/v1/journals
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateJournalRequest
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
		if errors.Is(err, ErrNotBalanced) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrInvalidLines) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrNotPostingAccount) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create journal")
		return
	}

	response.Success(w, http.StatusCreated, created)
}

// Update handles PUT /api/v1/journals/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entryID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid journal id")
		return
	}

	var req UpdateJournalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	updated, err := h.service.Update(r.Context(), userID, entryID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "journal not found")
			return
		}
		if errors.Is(err, ErrNotBalanced) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrInvalidLines) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrNotPostingAccount) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update journal")
		return
	}

	response.Success(w, http.StatusOK, updated)
}
