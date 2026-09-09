// Package user contains the "user" business module: domain entities, DTOs,
// ports (Repository & Service), service logic, and the HTTP handler.
package user

import "time"

// User is the user domain entity — a representation of a row in `users`.
type User struct {
	ID           int64
	Name         string
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AuditLog is an entity for recording user activity (table `user_audit_logs`).
// It serves as the second write operation inside one transaction (proving that
// Create writes 2 rows atomically).
type AuditLog struct {
	ID        int64     // Primary key
	UserID    int64     // FK to users.id
	Action    string    // Action name, e.g. "user.created"
	CreatedAt time.Time // Log creation time
}
