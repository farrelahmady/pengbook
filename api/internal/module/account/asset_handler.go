package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"pengbook/api/internal/middleware"
	"pengbook/api/pkg/response"
	"pengbook/api/pkg/validator"
)

// GetAssetSummary handles GET /api/v1/accounts/assets/summary
func (h *Handler) GetAssetSummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	summary, err := h.service.GetAssetSummary(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get asset summary")
		return
	}

	response.Success(w, http.StatusOK, summary)
}

// GetAssetGroups handles GET /api/v1/accounts/assets/groups
func (h *Handler) GetAssetGroups(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groups, err := h.service.GetAssetGroups(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get asset groups")
		return
	}

	response.Success(w, http.StatusOK, groups)
}

// ExportAssetExcel handles GET /api/v1/accounts/assets/export
// Generates an Excel file of the user's asset posting accounts ordered by
// code, with group name and cached balance.
func (h *Handler) ExportAssetExcel(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	export, err := h.service.ExportAssetExcel(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate export")
		return
	}

	// Filename date follows the client's timezone like the other exports.
	filenameDate := time.Now().In(middleware.LocationFromContext(r.Context())).Format("2006-01-02")

	// Set headers for Excel file download
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=aset-"+filenameDate+".xlsx")
	w.Header().Set("Content-Length", strconv.Itoa(len(export)))

	w.Write(export)
}

// CreateAsset handles POST /api/v1/accounts/assets
func (h *Handler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	created, err := h.service.CreateAsset(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrParentNotFound):
			response.Error(w, http.StatusNotFound, err.Error())
			return
		case errors.Is(err, ErrParentNotAsset), errors.Is(err, ErrParentNotLevel2):
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, ErrCodeExists):
			response.Error(w, http.StatusConflict, err.Error())
			return
		case errors.Is(err, ErrCodeExhausted):
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create asset")
		return
	}

	response.Success(w, http.StatusCreated, created)
}

// AdjustAssetBalance handles POST /api/v1/accounts/assets/{id}/adjust-balance
func (h *Handler) AdjustAssetBalance(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	assetID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid asset id")
		return
	}

	var req AdjustBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	result, err := h.service.AdjustAssetBalance(r.Context(), userID, assetID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "asset not found")
			return
		}
		if errors.Is(err, ErrAssetNotPosting) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to adjust asset balance")
		return
	}

	response.Success(w, http.StatusCreated, result)
}
