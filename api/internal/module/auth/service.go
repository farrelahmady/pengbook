package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"pengbook/api/internal/database"
	"pengbook/api/internal/module/user"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
	JWTSecretKey        = "your-secret-key-change-in-production"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*TokenResponse, error)
	Login(ctx context.Context, req LoginRequest) (*TokenResponse, error)
	Refresh(ctx context.Context, req RefreshRequest) (*RefreshTokenResponse, error)
	Logout(ctx context.Context, req LogoutRequest) error
	Me(ctx context.Context, accessToken string) (*user.User, error)
}

type service struct {
	userRepo     user.Repository
	tokenRepo    TokenRepository
	tx           database.TxManager
	jwtSecret    []byte
}

func NewService(userRepo user.Repository, tokenRepo TokenRepository, tx database.TxManager, jwtSecret string) Service {
	return &service{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		tx:        tx,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*TokenResponse, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	existingUsername, err := s.userRepo.FindByEmailOrUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing username: %w", err)
	}
	if existingUsername != nil {
		return nil, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &user.User{
		Name:         req.Name,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.generateTokenPair(ctx, newUser.ID, newUser.Email)
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	existingUser, err := s.userRepo.FindByEmailOrUsername(ctx, req.Identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if existingUser == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(ctx, existingUser.ID, existingUser.Email)
}

func (s *service) Refresh(ctx context.Context, req RefreshRequest) (*RefreshTokenResponse, error) {
	claims, err := s.validateToken(req.RefreshToken, RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	tokenData, err := s.tokenRepo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrTokenRevoked) || errors.Is(err, ErrTokenExpired) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get token from storage: %w", err)
	}

	if tokenData.UserID != claims.UserID {
		return nil, ErrInvalidToken
	}

	accessToken, err := s.generateAccessToken(tokenData.UserID, tokenData.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(AccessTokenDuration.Seconds()),
		TokenType:   "Bearer",
	}, nil
}

func (s *service) Logout(ctx context.Context, req LogoutRequest) error {
	_, err := s.validateToken(req.RefreshToken, RefreshToken)
	if err != nil {
		return ErrInvalidToken
	}

	if err := s.tokenRepo.RevokeRefreshToken(ctx, req.RefreshToken); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

func (s *service) Me(ctx context.Context, accessToken string) (*user.User, error) {
	claims, err := s.validateToken(accessToken, AccessToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *service) generateTokenPair(ctx context.Context, userID int64, email string) (*TokenResponse, error) {
	accessToken, err := s.generateAccessToken(userID, email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(userID, email)
	if err != nil {
		return nil, err
	}

	tokenData := RefreshTokenData{
		UserID:    userID,
		Email:     email,
		ExpiresAt: time.Now().Add(RefreshTokenDuration),
	}

	if err := s.tokenRepo.StoreRefreshToken(ctx, refreshToken, tokenData, RefreshTokenDuration); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(AccessTokenDuration.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *service) generateAccessToken(userID int64, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":    userID,
		"email":      email,
		"token_type": AccessToken,
		"iat":        now.Unix(),
		"exp":        now.Add(AccessTokenDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *service) generateRefreshToken(userID int64, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":    userID,
		"email":      email,
		"token_type": RefreshToken,
		"iat":        now.Unix(),
		"exp":        now.Add(RefreshTokenDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *service) validateToken(tokenString string, expectedType TokenType) (*TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	tokenType := TokenType(claims["token_type"].(string))
	if tokenType != expectedType {
		return nil, ErrInvalidToken
	}

	userID := int64(claims["user_id"].(float64))
	email := claims["email"].(string)
	exp := time.Unix(int64(claims["exp"].(float64)), 0)
	iat := time.Unix(int64(claims["iat"].(float64)), 0)

	if time.Now().After(exp) {
		return nil, ErrTokenExpired
	}

	return &TokenClaims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		ExpiresAt: exp,
		IssuedAt:  iat,
	}, nil
}
