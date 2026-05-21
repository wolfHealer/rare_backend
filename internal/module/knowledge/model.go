package knowledge

// CreateCategoryRequest 新增分类请求
type CreateCategoryRequest struct {
	ParentID    uint   `json:"parentId"`
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	SortOrder   int    `json:"sortOrder"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	ParentID    *uint   `json:"parentId"`
	Level       *int    `json:"level"`
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
	IconURL     *string `json:"iconUrl"`
	SortOrder   *int    `json:"sortOrder"`
	Status      *int    `json:"status"`
}

// CreateTagRequest 新增标签请求
type CreateTagRequest struct {
	Name      string `json:"name"`
	Code      string `json:"code"`
	SortOrder int    `json:"sortOrder"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	Name      *string `json:"name"`
	Code      *string `json:"code"`
	SortOrder *int    `json:"sortOrder"`
	Status    *int    `json:"status"`
}

// CreateDiseaseRequest 新增疾病请求
type CreateDiseaseRequest struct {
	Name         string   `json:"name"`
	Alias        string   `json:"alias"`
	Introduction string   `json:"introduction"`
	Symptoms     string   `json:"symptoms"`
	Images       []string `json:"images"`
	Status       int      `json:"status"`
	CreatorID    uint     `json:"creatorId"`

	PrimaryCategoryID uint   `json:"primary_category_id"`
	CategoryIDs       []uint `json:"category_ids"`
	TagIDs            []uint `json:"tag_ids"`
}

// UpdateDiseaseRequest 更新疾病请求
type UpdateDiseaseRequest struct {
	Name         *string   `json:"name"`
	Alias        *string   `json:"alias"`
	Introduction *string   `json:"introduction"`
	Symptoms     *string   `json:"symptoms"`
	Images       *[]string `json:"images"`
	Status       *int      `json:"status"`

	PrimaryCategoryID *uint   `json:"primary_category_id"`
	CategoryIDs       *[]uint `json:"category_ids"`
	TagIDs            *[]uint `json:"tag_ids"`
}

// CreateArticleRequest 创建文章请求
type CreateArticleRequest struct {
	Title          string  `json:"title" binding:"required"`
	Summary        string  `json:"summary"`
	CoverImage     string  `json:"coverImage"`
	ContentType    string  `json:"contentType"`
	AuthorID       *uint   `json:"authorId"`
	SourceName     string  `json:"sourceName"`
	SourceURL      string  `json:"sourceUrl"`
	Status         int     `json:"status"`
	PublishTime    *string `json:"publishTime"`
	IsTop          int     `json:"isTop"`
	IsRecommend    int     `json:"isRecommend"`
	SeoTitle       string  `json:"seoTitle"`
	SeoKeywords    string  `json:"seoKeywords"`
	SeoDescription string  `json:"seoDescription"`

	Blocks     []CreateArticleBlockRequest `json:"blocks"`
	TagIDs     []uint                      `json:"tagIds"`
	DiseaseIDs []uint                      `json:"diseaseIds"`
}

// CreateArticleBlockRequest 创建文章块请求
type CreateArticleBlockRequest struct {
	BlockType string      `json:"blockType" binding:"required"`
	SortNo    int         `json:"sortNo"`
	Title     string      `json:"title"`
	Content   string      `json:"content"`
	Extra     interface{} `json:"extra"`
}

// UpdateArticleRequest 更新文章请求
type UpdateArticleRequest struct {
	Title          *string `json:"title"`
	Summary        *string `json:"summary"`
	CoverImage     *string `json:"coverImage"`
	ContentType    *string `json:"contentType"`
	AuthorID       *uint   `json:"authorId"`
	SourceName     *string `json:"sourceName"`
	SourceURL      *string `json:"sourceUrl"`
	Status         *int    `json:"status"`
	PublishTime    *string `json:"publishTime"`
	IsTop          *int    `json:"isTop"`
	IsRecommend    *int    `json:"isRecommend"`
	SeoTitle       *string `json:"seoTitle"`
	SeoKeywords    *string `json:"seoKeywords"`
	SeoDescription *string `json:"seoDescription"`

	Blocks     []CreateArticleBlockRequest `json:"blocks"`
	TagIDs     *[]uint                     `json:"tagIds"`
	DiseaseIDs *[]uint                     `json:"diseaseIds"`
}
