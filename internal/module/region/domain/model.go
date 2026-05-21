package domain

import "time"

type RegionItem struct {
	ID         uint   `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	FullName   string `json:"fullName"`
	ParentCode string `json:"parentCode"`
	Level      int    `json:"level"`
	Sort       int    `json:"sort"`
	IsEnabled  int    `json:"isEnabled"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type CreateRegionInput struct {
	Code       string
	Name       string
	FullName   string
	ParentCode string
	Level      int
	Sort       int
	IsEnabled  int
}

type UpdateRegionInput struct {
	Name       *string
	FullName   *string
	ParentCode *string
	Level      *int
	Sort       *int
	IsEnabled  *int
}

type RegionListFilter struct {
	ParentCode string
	Level      int
	IsEnabled  int
	Keyword    string
	Page       int
	PageSize   int
}

type RegionListResult struct {
	List     []RegionItem
	Total    int64
	Page     int
	PageSize int
}

type FlatRegion struct {
	Code       string
	Name       string
	Level      int
	ParentCode string
}

type RegionTreeNode struct {
	Code     string            `json:"code"`
	Name     string            `json:"name"`
	Level    int               `json:"level"`
	Children []*RegionTreeNode `json:"children,omitempty"`
}

type RegionRow struct {
	ID         uint
	Code       string
	Name       string
	FullName   string
	ParentCode string
	Level      int
	Sort       int
	IsEnabled  int8
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func ToRegionItem(r RegionRow) RegionItem {
	return RegionItem{
		ID:         r.ID,
		Code:       r.Code,
		Name:       r.Name,
		FullName:   r.FullName,
		ParentCode: r.ParentCode,
		Level:      r.Level,
		Sort:       r.Sort,
		IsEnabled:  int(r.IsEnabled),
		CreatedAt:  r.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  r.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
