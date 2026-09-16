package account

import (
	"net/http"

	"pengbook/api/internal/middleware"
	"pengbook/api/pkg/response"
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
