package domain

import "time"

const (
	TargetTypePost    = "post"
	TargetTypeArticle = "article"
	TargetTypeResource = "resource"
)

// FavoriteItem 收藏列表项（聚合展示）
type FavoriteItem struct {
	ID         int64     `json:"id"`
	TargetType string    `json:"target_type"`
	TargetID   int64     `json:"target_id"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary,omitempty"`
	Cover      string    `json:"cover,omitempty"`
	Extra      any       `json:"extra,omitempty"`
	CreatedAt  time.Time `json:"-"`
	CreatedAtS string    `json:"created_at"`
}

type FavoriteListFilter struct {
	UserID     int64
	TargetType string
	Keyword    string
	Page       int
	PageSize   int
}

type RemoveFavoriteByTargetInput struct {
	UserID     int64
	TargetType string
	TargetID   int64
}
