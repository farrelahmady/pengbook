package user

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"pengbook/api/internal/database"
)

// ErrNotFound indicates the user was not found.
var ErrNotFound = errors.New("user not found")

// Service is the PORT (interface) for user business logic.
// Handlers depend only on this interface.
type Service interface {
	// Create creates a new user plus an audit log within one transaction.
	Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error)

	// GetByID returns the user for the given id.
	// Returns ErrNotFound when it does not exist.
	GetByID(ctx context.Context, id int64) (*UserResponse, error)
}

// service is the implementation of Service.
//
// The struct only holds its dependencies (Repository + TxManager) — there is
// no transaction logic here; it is fully managed by database.TxManager.
type service struct {
	repo Repository         // data access (port, not pgx)
	tx   database.TxManager // shared transaction manager
}

// NewService creates a service instance.
// tx is a *postgres.TxManager (injected from cmd/api/main.go).
func NewService(repo Repository, tx database.TxManager) Service {
	return &service{repo: repo, tx: tx}
}

// Create creates a new user and records an audit log inside ONE transaction.
//
// Flow:
//  1. The password is hashed with bcrypt (never stored as plaintext).
//  2. s.tx.WithTransaction opens a transaction (or reuses one already
//     injected in the context, e.g. from a test — without commit).
//  3. Inside the transaction: insert the user, then insert the audit log.
//     If either fails → everything is rolled back.
func (s *service) Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashed),
	}

	err = s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.Create(ctx, u); err != nil {
			return err
		}
		return s.repo.InsertAuditLog(ctx, &AuditLog{UserID: u.ID, Action: "user.created"})
	})
	if err != nil {
		return nil, err
	}

	return toResponse(u), nil
}

// GetByID returns a user. A read operation — no transaction needed; the
// repository automatically uses the regular connection.
func (s *service) GetByID(ctx context.Context, id int64) (*UserResponse, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return toResponse(u), nil
}

// toResponse maps a User entity → UserResponse DTO
// (hides PasswordHash, formats the timestamp).
func toResponse(u *User) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}
