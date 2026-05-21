package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPhoneExists        = errors.New("phone already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserDisabled       = errors.New("user disabled or not found")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrNoUpdateFields     = errors.New("no update fields")
)
