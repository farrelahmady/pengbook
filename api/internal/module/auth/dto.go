package auth

type RegisterRequest struct {
	Name     string `json:"name" example:"John Doe" validate:"required,min=3,max=100"`
	Username string `json:"username" example:"johndoe" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email" example:"john.doe@example.com" validate:"required,email"`
	Password string `json:"password" example:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" example:"johndoe or john@example.com" validate:"required"`
	Password   string `json:"password" example:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
	TokenType   string `json:"tokenType"`
}
