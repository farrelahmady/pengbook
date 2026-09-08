package auth

type RegisterRequest struct {
	Name     string `json:"name" example:"John Doe" validate:"required,min=3,max=100"`
	Email    string `json:"email" example:"john.doe@example.com" validate:"required,email"`
	Password string `json:"password" example:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"john.doe@example.com" validate:"required,email"`
	Password string `json:"password" example:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}
