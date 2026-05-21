package domain

import "errors"

var (
	ErrInvalidID       = errors.New("invalid id")
	ErrNotFound        = errors.New("not found")
	ErrCodeExists      = errors.New("code already exists")
	ErrNoUpdateFields  = errors.New("no update fields")
	ErrNameRequired    = errors.New("name required")
	ErrCodeRequired    = errors.New("code required")
	ErrInvalidLevel    = errors.New("invalid level")
	ErrKeywordRequired = errors.New("keyword required")
)
