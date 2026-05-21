package region

// CreateRegionRequest 新增行政区划请求
type CreateRegionRequest struct {
	Code       string `json:"code" binding:"required"`
	Name       string `json:"name" binding:"required"`
	FullName   string `json:"fullName" binding:"required"`
	ParentCode string `json:"parentCode"`
	Level      int    `json:"level" binding:"required,min=1,max=3"`
	Sort       int    `json:"sort"`
	IsEnabled  int    `json:"isEnabled"`
}

// UpdateRegionRequest 更新行政区划请求
type UpdateRegionRequest struct {
	Name       *string `json:"name"`
	FullName   *string `json:"fullName"`
	ParentCode *string `json:"parentCode"`
	Level      *int    `json:"level"`
	Sort       *int    `json:"sort"`
	IsEnabled  *int    `json:"isEnabled"`
}

// GetRegionListRequest 列表查询参数
type GetRegionListRequest struct {
	ParentCode string `form:"parentCode"`
	Level      int    `form:"level"`
	IsEnabled  int    `form:"isEnabled"`
	Keyword    string `form:"keyword"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=100"`
}
