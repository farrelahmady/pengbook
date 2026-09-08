package user

// CreateUserRequest is the DTO payload for the POST /users endpoint.
// The `validate` tags are used by pkg/validator during request validation.
type CreateUserRequest struct {
	Name     string `json:"name" example:"John Doe" validate:"required,min=3,max=100"`    // Required, 3–100 characters
	Email    string `json:"email" example:"john.doe@example.com" validate:"required,email"`           // Required, valid email format
	Password string `json:"password" example:"password" validate:"required,min=8,max=72"` // Required, 8–72 characters (72 = bcrypt limit)
}

// UserResponse is the user response DTO (no sensitive fields like password).
type UserResponse struct {
	ID        int64  `json:"id"`         // User ID
	Name      string `json:"name"`       // User name
	Email     string `json:"email"`      // User email
	CreatedAt string `json:"createdAt"` // Creation time (RFC3339 format)
}
