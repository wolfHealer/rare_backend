package domain

// --- Category ---

type CategoryItem struct {
	ID          uint   `json:"id"`
	ParentID    uint   `json:"parentId"`
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	SortOrder   int    `json:"sortOrder"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type CategoryTreeNode struct {
	ID          uint               `json:"id"`
	ParentID    uint               `json:"parentId"`
	Level       int                `json:"level"`
	Name        string             `json:"name"`
	Code        string             `json:"code"`
	Description string             `json:"description"`
	IconURL     string             `json:"iconUrl"`
	SortOrder   int                `json:"sortOrder"`
	Status      int                `json:"status"`
	CreatedAt   string             `json:"createdAt"`
	UpdatedAt   string             `json:"updatedAt"`
	Children    []CategoryTreeNode `json:"children"`
}

type CreateCategoryInput struct {
	ParentID    uint
	Level       int
	Name        string
	Code        string
	Description string
	IconURL     string
	SortOrder   int
}

type UpdateCategoryInput struct {
	ParentID    *uint
	Level       *int
	Name        *string
	Code        *string
	Description *string
	IconURL     *string
	SortOrder   *int
	Status      *int
}

type CategoryListFilter struct {
	ParentID *int
	Level    *int
}

type CategoryTreeFilter struct {
	Keyword string
	Status  *int
}

type CreateIDResult struct {
	ID int64 `json:"id"`
}

// --- Tag ---

type TagItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	SortOrder int    `json:"sortOrder"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type CreateTagInput struct {
	Name      string
	Code      string
	SortOrder int
}

type UpdateTagInput struct {
	Name      *string
	Code      *string
	SortOrder *int
	Status    *int
}

// --- Disease ---

type SimpleCategoryItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type SimpleTagItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type DiseaseItem struct {
	ID           uint        `json:"id"`
	Name         string      `json:"name"`
	Alias        string      `json:"alias"`
	Introduction string      `json:"introduction"`
	Symptoms     string      `json:"symptoms"`
	Images       interface{} `json:"images"`

	PrimaryCategoryID   *uint   `json:"primaryCategoryId"`
	PrimaryCategoryName *string `json:"primaryCategoryName"`

	CategoryIDs []uint               `json:"categoryIds"`
	Categories  []SimpleCategoryItem `json:"categories"`

	TagIDs []uint          `json:"tagIds"`
	Tags   []SimpleTagItem `json:"tags"`

	Articles []ArticleItem `json:"articles"`

	Status    int    `json:"status"`
	CreatorID uint   `json:"creatorId"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type CreateDiseaseInput struct {
	Name              string
	Alias             string
	Introduction      string
	Symptoms          string
	Images            []string
	Status            int
	CreatorID         uint
	PrimaryCategoryID uint
	CategoryIDs       []uint
	TagIDs            []uint
}

type UpdateDiseaseInput struct {
	Name              *string
	Alias             *string
	Introduction      *string
	Symptoms          *string
	Images            *[]string
	Status            *int
	PrimaryCategoryID *uint
	CategoryIDs       *[]uint
	TagIDs            *[]uint
}

type DiseaseListFilter struct {
	Keyword    string
	Status     *int
	CategoryID *uint
	TagID      *uint
	Page       int
	PageSize   int
}

type DiseaseListResult struct {
	List     []DiseaseItem `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

type DiseasesByCategoryFilter struct {
	CategoryID uint
	Page       int
	PageSize   int
}

type SimpleDiseaseItem struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type SearchDiseasesResult struct {
	List []SimpleDiseaseItem `json:"list"`
}

type DiseaseOptionItem struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

// --- Article ---

type ArticleItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	CoverImage  string `json:"coverImage"`
	SourceName  string `json:"sourceName"`
	Status      int    `json:"status"`
	PublishTime string `json:"publishTime"`
	ViewCount   int64  `json:"viewCount"`
	IsTop       int    `json:"isTop"`
	IsRecommend int    `json:"isRecommend"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type AdminArticleItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	CoverImage  string `json:"coverImage"`
	SourceName  string `json:"sourceName"`
	Status      int    `json:"status"`
	PublishTime string `json:"publishTime"`
	ViewCount   int64  `json:"viewCount"`
	IsTop       int    `json:"isTop"`
	IsRecommend int    `json:"isRecommend"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type ArticleBlockItem struct {
	ID        uint        `json:"id"`
	BlockType string      `json:"blockType"`
	SortNo    int         `json:"sortNo"`
	Title     string      `json:"title"`
	Content   string      `json:"content"`
	Extra     interface{} `json:"extra"`
}

type ArticleTagItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type ArticleDetailResponse struct {
	ID             uint   `json:"id"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	CoverImage     string `json:"coverImage"`
	ContentType    string `json:"contentType"`
	AuthorID       *uint  `json:"authorId"`
	SourceName     string `json:"sourceName"`
	SourceURL      string `json:"sourceUrl"`
	Status         int    `json:"status"`
	PublishTime    string `json:"publishTime"`
	ViewCount      int64  `json:"viewCount"`
	LikeCount      int64  `json:"likeCount"`
	FavoriteCount  int64  `json:"favoriteCount"`
	IsTop          int    `json:"isTop"`
	IsRecommend    int    `json:"isRecommend"`
	SeoTitle       string `json:"seoTitle"`
	SeoKeywords    string `json:"seoKeywords"`
	SeoDescription string `json:"seoDescription"`

	Blocks []ArticleBlockItem `json:"blocks"`
	Tags   []ArticleTagItem   `json:"tags"`

	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type CreateArticleBlockInput struct {
	BlockType string
	SortNo    int
	Title     string
	Content   string
	Extra     interface{}
}

type CreateArticleInput struct {
	Title          string
	Summary        string
	CoverImage     string
	ContentType    string
	AuthorID       *uint
	SourceName     string
	SourceURL      string
	Status         int
	PublishTime    *string
	IsTop          int
	IsRecommend    int
	SeoTitle       string
	SeoKeywords    string
	SeoDescription string
	Blocks         []CreateArticleBlockInput
	TagIDs         []uint
	DiseaseIDs     []uint
}

type UpdateArticleInput struct {
	Title          *string
	Summary        *string
	CoverImage     *string
	ContentType    *string
	AuthorID       *uint
	SourceName     *string
	SourceURL      *string
	Status         *int
	PublishTime    *string
	IsTop          *int
	IsRecommend    *int
	SeoTitle       *string
	SeoKeywords    *string
	SeoDescription *string
	Blocks         []CreateArticleBlockInput
	TagIDs         *[]uint
	DiseaseIDs     *[]uint
}

type ArticleListFilter struct {
	Keyword   string
	Status    string
	DiseaseID string
	Page      int
	PageSize  int
}

type ArticleListResult struct {
	List     []AdminArticleItem `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}
