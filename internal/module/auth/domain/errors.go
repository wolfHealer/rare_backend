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
	ErrAccountDeactivated = errors.New("account already deactivated")
	ErrCannotDeactivateAdmin = errors.New("admin cannot deactivate self")
)

const AnonymousDisplayName = "已注销用户"

var (
	ErrInvalidScene      = errors.New("无效的场景")
	ErrTooManyRequests   = errors.New("发送过于频繁，请稍后再试")
	ErrInvalidCode       = errors.New("验证码无效或已过期")
	ErrFavoriteNotFound  = errors.New("favorite not found")
	ErrInvalidTargetType = errors.New("invalid target type")
)
