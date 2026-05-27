package domain

import "time"

// PostListItem repo 层帖子列表行
type PostListItem struct {
	ID            int64
	UserID        int64
	DisplayName   string
	DiseaseID     *int64
	CategoryID    *int64
	Type          string
	Title         string
	Content       string
	Images        []string
	ViewCount     int
	LikeCount     int
	CommentCount  int
	FavoriteCount int
	IsTop         bool
	IsRecommend   bool
	Status        int
	RejectReason  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
