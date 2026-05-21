package domain

import "errors"

var (
	ErrInvalidPolicyID  = errors.New("invalid policy id")
	ErrPolicyNotFound   = errors.New("policy not found")
	ErrCountList        = errors.New("count list failed")
	ErrQueryList        = errors.New("query list failed")
	ErrQueryDetail      = errors.New("query detail failed")
	ErrQueryMaterials   = errors.New("query materials failed")
)
