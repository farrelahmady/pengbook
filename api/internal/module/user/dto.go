package user

// CreateUserRequest is the DTO payload for the POST /users endpoint.
// The `validate` tags are used by pkg/validator during request validation.
type CreateUserRequest struct {
	Name     string `json:"name" example:"John Doe" validate:"required,min=3,max=100"`
	Username string `json:"username" example:"johndoe" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email" example:"john.doe@example.com" validate:"required,email"`
	Password string `json:"password" example:"password" validate:"required,min=8,max=72"`
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}
