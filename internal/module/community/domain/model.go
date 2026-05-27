package domain

import "time"

type LikeResult struct {
	IsLiked   bool `json:"is_liked"`
	LikeCount int  `json:"like_count"`
}

type FavoriteResult struct {
	IsFavorited   bool `json:"is_favorited"`
	FavoriteCount int  `json:"favorite_count"`
}

type CreatePostInput struct {
	UserID     int64
	DiseaseID  *int64
	CategoryID *int64
	Type       string
	Title      *string
	Content    string
	Images     []string
}

type UpdatePostInput struct {
	Title        *string
	Content      *string
	Images       *[]string
	DiseaseID    *int64
	CategoryID   *int64
	Type         *string
	IsTop        *int8
	IsRecommend  *int8
	Status       *int8
	RejectReason *string
}

type CreateCommentInput struct {
	PostID   int64
	UserID   int64
	Content  string
	ParentID *int64
}

type CreateCommentResult struct {
	ID          int64
	PostID      int64
	UserID      int64
	DisplayName string
	Content     string
	ParentID    int64
	RootID      int64
	CreatedAt   time.Time
}
