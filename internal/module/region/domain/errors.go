package domain

import "errors"

var (
	ErrInvalidID      = errors.New("invalid id")
	ErrNotFound       = errors.New("not found")
	ErrCodeExists     = errors.New("code already exists")
	ErrParentNotFound = errors.New("parent not found")
	ErrHasChildren    = errors.New("has children")
	ErrInvalidLevel   = errors.New("invalid level")
	ErrNoUpdateFields = errors.New("no update fields")
)
