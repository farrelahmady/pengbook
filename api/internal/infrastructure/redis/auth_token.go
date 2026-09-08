package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"pengbook/api/internal/module/auth"
)

const refreshTokenPrefix = "refresh_token:"

type AuthTokenRepository struct {
	client *redis.Client
}

func NewAuthTokenRepository(client *redis.Client) auth.TokenRepository {
	return &AuthTokenRepository{client: client}
}

func (r *AuthTokenRepository) StoreRefreshToken(ctx context.Context, token string, data auth.RefreshTokenData, expiry time.Duration) error {
	key := refreshTokenPrefix + token
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token data: %w", err)
	}
	return r.client.Set(ctx, key, jsonData, expiry).Err()
}

func (r *AuthTokenRepository) GetRefreshToken(ctx context.Context, token string) (*auth.RefreshTokenData, error) {
	key := refreshTokenPrefix + token
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, auth.ErrTokenRevoked
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	var data auth.RefreshTokenData
	if err := json.Unmarshal(val, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal refresh token data: %w", err)
	}

	if time.Now().After(data.ExpiresAt) {
		r.client.Del(ctx, key)
		return nil, auth.ErrTokenExpired
	}

	return &data, nil
}

func (r *AuthTokenRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	key := refreshTokenPrefix + token
	return r.client.Del(ctx, key).Err()
}
