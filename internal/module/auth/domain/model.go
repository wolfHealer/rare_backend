package domain

import "time"

type LoginInput struct {
	Phone    string
	Password string
}

type LoginResult struct {
	UserID    int64
	Role      int
	Nickname  string
	Phone     string
	Avatar    string
	Token     string
	ExpiresIn int
}

type RegisterInput struct {
	Phone    string
	Password string
}

type User struct {
	ID           int64
	Phone        string
	PasswordHash string
	DisplayName  string
	Avatar       string
	Role         int
	Status       int
	LoginCount   int
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserItem struct {
	ID          int64  `json:"id"`
	Phone       string `json:"phone"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Role        int    `json:"role"`
	Status      int    `json:"status"`
	LastLoginAt string `json:"lastLoginAt"`
	LoginCount  int    `json:"loginCount"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type UserListFilter struct {
	Keyword  string
	Page     int
	PageSize int
	Status   *int
	Role     *int
}

type UserListResult struct {
	List     []UserItem
	Total    int64
	Page     int
	PageSize int
}

type UpdateUserInput struct {
	DisplayName *string
	Avatar      *string
	Role        *int
	Status      *int
}

type CreateUserInput struct {
	Phone       string
	Password    string
	DisplayName string
	Avatar      string
	Role        int
	Status      int
}

// SMSCode 验证码记录
type SMSCode struct {
	ID        int64     `json:"id"`
	Phone     string    `json:"phone"`
	Code      string    `json:"code"`
	Scene     string    `json:"scene"` // register, login
	CreatedAt time.Time `json:"createdAt"`
	ExpiredAt time.Time `json:"expiredAt"`
	Used      int       `json:"used"` // 0: 未使用, 1: 已使用
}
