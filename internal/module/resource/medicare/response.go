package medicare

import (
	"rare_backend/internal/module/resource/medicare/domain"
	"rare_backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type PolicyItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Region      string `json:"region"`
	RegionCode  string `json:"regionCode"`
	Date        string `json:"date"`
	PublishDate string `json:"publishDate"`
	Summary     string `json:"summary"`
	Category    string `json:"category"`
	FileUrl     string `json:"fileUrl"`
}

type RelatedPolicy struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

type PolicyDetailResponse struct {
	ID                  uint            `json:"id"`
	Title               string          `json:"title"`
	Region              string          `json:"region"`
	RegionCode          string          `json:"regionCode"`
	PublishDate         string          `json:"publishDate"`
	EffectiveDate       string          `json:"effectiveDate"`
	Category            string          `json:"category"`
	Content             string          `json:"content"`
	FileUrl             string          `json:"fileUrl"`
	ReimburseRatio      string          `json:"reimburseRatio"`
	ReimburseLimit      string          `json:"reimburseLimit"`
	ReimburseProcess    string          `json:"reimburseProcess"`
	ReimburseMaterial   string          `json:"reimburseMaterial"`
	RemoteApplyTemplate string          `json:"remoteApplyTemplate"`
	RelatedPolicies     []RelatedPolicy `json:"relatedPolicies"`
}

type QueryMode string

const (
	ModeDefaultRecommend QueryMode = "default_recommend"
	ModeRegionOnly       QueryMode = "region_only"
	ModeDiseaseOnly      QueryMode = "disease_only"
	ModeRegionDisease    QueryMode = "region_disease"
)

type SelectedFilters struct {
	DiseaseID    int    `json:"disease_id"`
	DiseaseName  string `json:"disease_name,omitempty"`
	ProvinceCode string `json:"province_code"`
	CityCode     string `json:"city_code"`
	ProvinceName string `json:"province_name,omitempty"`
	CityName     string `json:"city_name,omitempty"`
}

type GroupedPolicies struct {
	National []PolicyItem `json:"national"`
	Province []PolicyItem `json:"province"`
	City     []PolicyItem `json:"city"`
}

type NewPolicyListResponse struct {
	QueryMode       QueryMode       `json:"query_mode"`
	SelectedFilters SelectedFilters `json:"selected_filters"`
	Tips            string          `json:"tips"`
	Policies        GroupedPolicies `json:"policies"`
	Total           int64           `json:"total"`
	Page            int             `json:"page"`
	PageSize        int             `json:"pageSize"`
}

type MaterialItem struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	Size       string `json:"size"`
	UpdateTime string `json:"updateTime"`
}

type MaterialResponse struct {
	Materials []MaterialItem `json:"materials"`
}

func respondOK(c *gin.Context, data any) {
	response.OK(c, data)
}

func respondNotFound(c *gin.Context, message string) {
	response.NotFound(c, message)
}

func respondInvalidPolicyID(c *gin.Context) {
	response.BadRequest(c, "无效的政策 ID")
}

func respondPolicyListCountError(c *gin.Context) {
	response.InternalError(c, "查询总数失败")
}

func respondPolicyListError(c *gin.Context) {
	response.InternalError(c, "查询列表失败")
}

func respondPolicyDetailError(c *gin.Context) {
	response.InternalError(c, "查询政策详情失败")
}

func respondMaterialsError(c *gin.Context) {
	response.InternalError(c, "查询资料失败")
}

func toPolicyItem(item domain.PolicyItem) PolicyItem {
	return PolicyItem{
		ID:          item.ID,
		Title:       item.Title,
		Region:      item.Region,
		RegionCode:  item.RegionCode,
		Date:        item.Date,
		PublishDate: item.PublishDate,
		Summary:     item.Summary,
		Category:    item.Category,
		FileUrl:     item.FileUrl,
	}
}

func toPolicyItems(items []domain.PolicyItem) []PolicyItem {
	result := make([]PolicyItem, 0, len(items))
	for _, item := range items {
		result = append(result, toPolicyItem(item))
	}
	return result
}

func toRelatedPolicies(items []domain.RelatedPolicy) []RelatedPolicy {
	result := make([]RelatedPolicy, 0, len(items))
	for _, item := range items {
		result = append(result, RelatedPolicy{ID: item.ID, Title: item.Title})
	}
	return result
}

func toMaterialItems(items []domain.MaterialItem) []MaterialItem {
	result := make([]MaterialItem, 0, len(items))
	for _, item := range items {
		result = append(result, MaterialItem{
			Name:       item.Name,
			Type:       item.Type,
			URL:        item.URL,
			Size:       item.Size,
			UpdateTime: item.UpdateTime,
		})
	}
	return result
}

func toSelectedFilters(f domain.SelectedFilters) SelectedFilters {
	return SelectedFilters{
		DiseaseID:    f.DiseaseID,
		DiseaseName:  f.DiseaseName,
		ProvinceCode: f.ProvinceCode,
		CityCode:     f.CityCode,
		ProvinceName: f.ProvinceName,
		CityName:     f.CityName,
	}
}

func toPolicyListResponse(result *domain.PolicyListResult) NewPolicyListResponse {
	return NewPolicyListResponse{
		QueryMode:       QueryMode(result.QueryMode),
		SelectedFilters: toSelectedFilters(result.SelectedFilters),
		Tips:            result.Tips,
		Policies: GroupedPolicies{
			National: toPolicyItems(result.Policies.National),
			Province: toPolicyItems(result.Policies.Province),
			City:     toPolicyItems(result.Policies.City),
		},
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}
}

func toPolicyDetailResponse(result *domain.PolicyDetailResult) PolicyDetailResponse {
	return PolicyDetailResponse{
		ID:                  result.ID,
		Title:               result.Title,
		Region:              result.Region,
		RegionCode:          result.RegionCode,
		PublishDate:         result.PublishDate,
		EffectiveDate:       result.EffectiveDate,
		Category:            result.Category,
		Content:             result.Content,
		FileUrl:             result.FileUrl,
		ReimburseRatio:      result.ReimburseRatio,
		ReimburseLimit:      result.ReimburseLimit,
		ReimburseProcess:    result.ReimburseProcess,
		ReimburseMaterial:   result.ReimburseMaterial,
		RemoteApplyTemplate: result.RemoteApplyTemplate,
		RelatedPolicies:     toRelatedPolicies(result.RelatedPolicies),
	}
}
