package domain

import "errors"

// Sentinel errors for the domain layer.
// Repository implementations MUST translate ORM/driver-specific errors
// (e.g. gorm.ErrRecordNotFound) into these domain errors so that the
// service layer stays completely decoupled from persistence internals.
var (
	ErrNotFound          = errors.New("record not found")
	ErrConflict          = errors.New("record already exists")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
)
