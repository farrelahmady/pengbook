package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"pengbook/api/internal/module/user"
	"pengbook/api/pkg/response"
	"pengbook/api/pkg/validator"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	r.Get("/me", h.Me)
	return r
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account and returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data"
// @Success 201 {object} response.body{success=bool,data=TokenResponse}
// @Failure 400 {object} response.body{success=bool,message=string,errors=[]string}
// @Failure 409 {object} response.body{success=bool,message=string}
// @Failure 500 {object} response.body{success=bool,message=string}
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	tokens, err := h.service.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			response.Error(w, http.StatusConflict, "user already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	response.Success(w, http.StatusCreated, tokens)
}

// Login godoc
// @Summary Login user
// @Description Authenticates user and returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} response.body{success=bool,data=TokenResponse}
// @Failure 400 {object} response.body{success=bool,message=string,errors=[]string}
// @Failure 401 {object} response.body{success=bool,message=string}
// @Failure 500 {object} response.body{success=bool,message=string}
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	tokens, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to login")
		return
	}

	response.Success(w, http.StatusOK, tokens)
}

// Refresh godoc
// @Summary Refresh access token
// @Description Returns a new access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} response.body{success=bool,data=RefreshTokenResponse}
// @Failure 400 {object} response.body{success=bool,message=string,errors=[]string}
// @Failure 401 {object} response.body{success=bool,message=string}
// @Failure 500 {object} response.body{success=bool,message=string}
// @Router /api/v1/auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	tokens, err := h.service.Refresh(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrTokenRevoked) {
			response.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to refresh token")
		return
	}

	response.Success(w, http.StatusOK, tokens)
}

// Logout godoc
// @Summary Logout user
// @Description Revokes the refresh token to invalidate the session
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Refresh token to revoke"
// @Success 200 {object} response.body{success=bool,message=string}
// @Failure 400 {object} response.body{success=bool,message=string,errors=[]string}
// @Failure 401 {object} response.body{success=bool,message=string}
// @Failure 500 {object} response.body{success=bool,message=string}
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.service.Logout(r.Context(), req); err != nil {
		if errors.Is(err, ErrInvalidToken) {
			response.Error(w, http.StatusUnauthorized, "invalid token")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// Me godoc
// @Summary Get current user
// @Description Returns the authenticated user's profile using the access token from the Authorization header
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.body{success=bool,data=user.UserResponse}
// @Failure 401 {object} response.body{success=bool,message=string}
// @Failure 500 {object} response.body{success=bool,message=string}
// @Router /api/v1/auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	token, err := bearerToken(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	u, err := h.service.Me(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenExpired) {
			response.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	response.Success(w, http.StatusOK, user.UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	})
}

// bearerToken extracts the access token from the `Authorization: Bearer <token>`
// request header. Returns an error when the header is missing or malformed.
func bearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", errors.New("malformed authorization header")
	}

	return parts[1], nil
}
