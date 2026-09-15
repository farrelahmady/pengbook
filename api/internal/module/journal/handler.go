package journal

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	r.Get("/template", h.DownloadTemplate)
	r.Get("/export", h.Export)
	r.Post("/", h.Create)
	r.Post("/upload", h.Upload)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// parseListRequest extracts cursor pagination + filter query parameters shared
// by the scroll-view list and the export (export ignores cursor/limit).
func parseListRequest(r *http.Request) ListRequest {
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

	return ListRequest{
		Limit:      limit,
		Cursor:     &cursor,
		StartDate:  startDate,
		EndDate:    endDate,
		AccountIDs: accountIDs,
	}
}

// GetAllScrollView handles GET /api/v1/journals
func (h *Handler) GetAllScrollView(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	req := parseListRequest(r)

	result, err := h.service.GetAllScrollView(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidFilter) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get journals")
		return
	}

	response.Success(w, http.StatusOK, result)
}

// GetTotalSummary handles GET /api/v1/journals/summary?month=RFC3339
// The month instant carries the client timezone in its ISO offset; month
// bounds are derived in that zone.
func (h *Handler) GetTotalSummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	month := r.URL.Query().Get("month")

	summary, err := h.service.GetTotalSummary(r.Context(), userID, month)
	if err != nil {
		if errors.Is(err, ErrInvalidMonth) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
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

// DownloadTemplate handles GET /api/v1/journals/template
// Generates an Excel template with user's posting accounts.
func (h *Handler) DownloadTemplate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	template, err := h.service.GenerateTemplate(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate template")
		return
	}

	// Set headers for Excel file download
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=template-jurnal.xlsx")
	w.Header().Set("Content-Length", strconv.Itoa(len(template)))

	w.Write(template)
}

// Export handles GET /api/v1/journals/export
// Generates an Excel file of the user's journals (honoring the same filters
// as the list view) that can be re-uploaded as-is.
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	req := parseListRequest(r)

	export, err := h.service.ExportExcel(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidFilter) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to generate export")
		return
	}

	// Filename date follows the client's timezone like the export content.
	filenameDate := time.Now().In(middleware.LocationFromContext(r.Context())).Format("2006-01-02")

	// Set headers for Excel file download
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=jurnal-"+filenameDate+".xlsx")
	w.Header().Set("Content-Length", strconv.Itoa(len(export)))

	w.Write(export)
}

// Upload handles POST /api/v1/journals/upload
// Accepts a CSV or Excel file via multipart form.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse multipart form (max 5MB)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "failed to parse form: file is required")
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// Validate file extension
	filename := strings.ToLower(header.Filename)
	isCSV := strings.HasSuffix(filename, ".csv")
	isExcel := strings.HasSuffix(filename, ".xlsx") || strings.HasSuffix(filename, ".xls")

	if !isCSV && !isExcel {
		response.Error(w, http.StatusBadRequest, "file must be a CSV (.csv) or Excel (.xlsx)")
		return
	}

	// Parse file based on extension. Bare calendar dates in the file are read
	// at midnight in the client's timezone (X-Timezone header); storage stays
	// UTC. No extra form field needed: the header rides along automatically.
	loc := middleware.LocationFromContext(r.Context())
	var entries []CreateJournalRequest

	if isExcel {
		entries, err = ParseExcel(file, loc)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid Excel: "+err.Error())
			return
		}
	} else {
		entries, err = ParseCSV(file, loc)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid CSV: "+err.Error())
			return
		}
	}

	if len(entries) == 0 {
		response.Error(w, http.StatusBadRequest, "no valid entries found in file")
		return
	}

	// Process entries
	count, err := h.service.CreateBulk(r.Context(), userID, CreateBulkJournalRequest{Entries: entries})
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
		response.Error(w, http.StatusInternalServerError, "failed to upload journals")
		return
	}

	response.Success(w, http.StatusCreated, map[string]int64{"count": count})
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

// Delete handles DELETE /api/v1/journals/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.Delete(r.Context(), userID, entryID); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "journal not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete journal")
		return
	}

	response.Success(w, http.StatusOK, map[string]int64{"id": entryID})
}
