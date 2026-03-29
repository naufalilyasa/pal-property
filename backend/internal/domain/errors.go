package domain

import "errors"

// Sentinel errors for the domain layer.
// Repository implementations MUST translate ORM/driver-specific errors
// (e.g. gorm.ErrRecordNotFound) into these domain errors so that the
// service layer stays completely decoupled from persistence internals.
var (
	ErrNotFound           = errors.New("record not found")
	ErrConflict           = errors.New("record already exists")
	ErrInvalidCredential  = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidImageFile   = errors.New("invalid image file")
	ErrImageLimitReached  = errors.New("listing image limit reached")
	ErrImageOrderInvalid  = errors.New("invalid listing image order")
	ErrImageStorageUnset  = errors.New("listing image storage is not configured")
	ErrInvalidVideoFile   = errors.New("invalid video file")
	ErrVideoTooLarge      = errors.New("video file is too large")
	ErrVideoTooLong       = errors.New("video duration exceeds limit")
	ErrVideoAlreadyExists = errors.New("listing already has a video")
	ErrVideoStorageUnset  = errors.New("listing video storage is not configured")
)
