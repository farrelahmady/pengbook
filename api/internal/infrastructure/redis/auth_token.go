package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"pengbook/api/internal/module/auth"
	"pengbook/api/pkg/logger"
)

const refreshTokenPrefix = "refresh_token:"

type AuthTokenRepository struct {
	client *redis.Client
}

func NewAuthTokenRepository(client *redis.Client) auth.TokenRepository {
	return &AuthTokenRepository{client: client}
}

func (r *AuthTokenRepository) StoreRefreshToken(ctx context.Context, token string, data auth.RefreshTokenData, expiry time.Duration) error {
	log := logger.FromContext(ctx)
	key := refreshTokenPrefix + token
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("repo: failed to marshal refresh token data", "user_id", data.UserID, "error", err)
		return fmt.Errorf("failed to marshal refresh token data: %w", err)
	}
	if err := r.client.Set(ctx, key, jsonData, expiry).Err(); err != nil {
		log.Error("repo: failed to store refresh token", "user_id", data.UserID, "error", err)
		return err
	}
	return nil
}

func (r *AuthTokenRepository) GetRefreshToken(ctx context.Context, token string) (*auth.RefreshTokenData, error) {
	log := logger.FromContext(ctx)
	key := refreshTokenPrefix + token
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, auth.ErrTokenRevoked
		}
		log.Error("repo: failed to get refresh token", "error", err)
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	var data auth.RefreshTokenData
	if err := json.Unmarshal(val, &data); err != nil {
		log.Error("repo: failed to unmarshal refresh token data", "error", err)
		return nil, fmt.Errorf("failed to unmarshal refresh token data: %w", err)
	}

	if time.Now().After(data.ExpiresAt) {
		r.client.Del(ctx, key)
		return nil, auth.ErrTokenExpired
	}

	return &data, nil
}

func (r *AuthTokenRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	log := logger.FromContext(ctx)
	key := refreshTokenPrefix + token
	if err := r.client.Del(ctx, key).Err(); err != nil {
		log.Error("repo: failed to revoke refresh token", "error", err)
		return err
	}
	return nil
}
