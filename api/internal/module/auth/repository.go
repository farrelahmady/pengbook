package auth

import (
	"context"
	"time"
)

type TokenRepository interface {
	StoreRefreshToken(ctx context.Context, token string, data RefreshTokenData, expiry time.Duration) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshTokenData, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}
