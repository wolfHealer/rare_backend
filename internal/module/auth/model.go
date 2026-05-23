package auth

// LoginRequest HTTP 绑定
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest HTTP 绑定
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateUserRequest HTTP 绑定（管理员）
type UpdateUserRequest struct {
	DisplayName *string `json:"displayName"`
	Avatar      *string `json:"avatar"`
	Role        *int    `json:"role"`
	Status      *int    `json:"status"`
}

// UpdateProfileRequest HTTP 绑定（用户自己更新）
type UpdateProfileRequest struct {
	DisplayName *string `json:"displayName"`
	Avatar      *string `json:"avatar"`
}

// UpdateUserRoleRequest HTTP 绑定
type UpdateUserRoleRequest struct {
	Role int `json:"role" binding:"required,min=1,max=9"`
}

// ResetPasswordRequest HTTP 绑定
type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// CreateUserRequest HTTP 绑定
type CreateUserRequest struct {
	Phone       string `json:"phone" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Role        int    `json:"role" binding:"required,oneof=1 2 9"`
	Status      int    `json:"status" binding:"required,oneof=0 1"`
}
