package community

// CreatePostRequest HTTP 绑定
type CreatePostRequest struct {
	DiseaseID  *int64   `json:"disease_id"`
	CategoryID *int64   `json:"category_id"`
	Type       string   `json:"type" binding:"required,oneof=help experience emotion info"`
	Title      *string  `json:"title"`
	Content    string   `json:"content" binding:"required"`
	Images     []string `json:"images"`
}

// UpdatePostRequest HTTP 绑定
type UpdatePostRequest struct {
	Title        *string   `json:"title"`
	Content      *string   `json:"content"`
	Images       *[]string `json:"images"`
	DiseaseID    *int64    `json:"disease_id"`
	CategoryID   *int64    `json:"category_id"`
	Type         *string   `json:"type" binding:"omitempty,oneof=help experience emotion info"`
	IsTop        *int8     `json:"is_top"`
	IsRecommend  *int8     `json:"is_recommend"`
	Status       *int8     `json:"status"`
	RejectReason *string   `json:"reject_reason"`
}

// UpdateCommentRequest HTTP 绑定
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}
