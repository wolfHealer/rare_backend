package domain

import "errors"

var (
	ErrInvalidID       = errors.New("invalid id")
	ErrPostNotFound    = errors.New("post not found")
	ErrCommentNotFound = errors.New("comment not found")
	ErrParentComment   = errors.New("parent comment not found")
	ErrForbidden       = errors.New("forbidden")
	ErrNoUpdateFields    = errors.New("no update fields")
	ErrInvalidPostStatus = errors.New("invalid post status filter")
	ErrReportDuplicate   = errors.New("report duplicate")
	ErrCannotReportSelf  = errors.New("cannot report self post")
	ErrInvalidReportReason = errors.New("invalid report reason")
	ErrReportNotFound      = errors.New("report not found")
	ErrReportAlreadyDone   = errors.New("report already handled")
)
