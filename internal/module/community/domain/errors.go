package domain

import "errors"

var (
	ErrInvalidID       = errors.New("invalid id")
	ErrPostNotFound    = errors.New("post not found")
	ErrCommentNotFound = errors.New("comment not found")
	ErrParentComment   = errors.New("parent comment not found")
	ErrForbidden       = errors.New("forbidden")
	ErrNoUpdateFields  = errors.New("no update fields")
	ErrInvalidPostStatus = errors.New("invalid post status filter")
)
