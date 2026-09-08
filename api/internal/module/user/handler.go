package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"pengbook/api/pkg/response"
	"pengbook/api/pkg/validator"
)

// Handler is the HTTP handler for the user module.
// Responsibilities: decode request → validate → call Service → send response.
type Handler struct {
	service Service // business logic (interface, not concrete)
}

// NewHandler creates a user handler instance.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns the chi router specific to user.
// It is mounted at `/api/v1/users` in the HTTP server.
//
//	POST  /     → create a new user
//	GET   /{id} → fetch a user by id
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	return r
}

// Create handles POST /api/v1/users.
//
// Flow: decode JSON → validate DTO → call service.Create → respond 201 with
// the user data, or 400/500 on failure.
// @Summary Create a new user
// @Description Creates a new user with name, email and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User creation data"
// @Success 201 {object} response.body{success=bool,data=UserResponse}
// @Failure 400 {object} response.body{success=bool,message=string,errors=[]string}
// @Failure 500 {object} response.body{success=bool,message=string}
// @Router /api/v1/users [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	created, err := h.service.Create(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	response.Success(w, http.StatusCreated, created)
}

// GetByID handles GET /api/v1/users/{id}.
//
// The id is read from the URL path and parsed to int64. Returns 404 when not
// found (via ErrNotFound), otherwise 200 or 500.
// @Summary Get user by ID
// @Description Fetch a user by their ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.body{success=bool,data=UserResponse}
// @Failure 400 {object} response.body{success=bool,message=string}
// @Failure 404 {object} response.body{success=bool,message=string}
// @Failure 500 {object} response.body{success=bool,message=string}
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	u, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	response.Success(w, http.StatusOK, u)
}
