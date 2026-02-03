package post

type CreatePostReq struct {
	Content    string   `json:"content" binding:"required"`
	Images     []string `json:"images"`
	DiseaseTag string   `json:"disease_tag"`
}

type PostListResp struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}
