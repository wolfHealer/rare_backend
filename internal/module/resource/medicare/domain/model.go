package domain

import "time"

type QueryMode string

const (
	ModeDefaultRecommend QueryMode = "default_recommend"
	ModeRegionOnly       QueryMode = "region_only"
	ModeDiseaseOnly      QueryMode = "disease_only"
	ModeRegionDisease    QueryMode = "region_disease"
)

type PolicyListFilter struct {
	ProvinceCode string
	CityCode     string
	DiseaseID    int
	NeedLatest   bool
	Page         int
	PageSize     int
}

type SelectedFilters struct {
	DiseaseID    int
	DiseaseName  string
	ProvinceCode string
	CityCode     string
	ProvinceName string
	CityName     string
}

type PolicyRow struct {
	ID               uint
	DiseaseID        int
	ScopeLevel       int
	ProvinceCode     string
	CityCode         string
	ProvinceName     string
	CityName         string
	PolicyTitle      string
	PolicyOriginal   string
	PopularInterpret string
	IsLatest         int
	PublishDate      *time.Time
	CreatedAt        time.Time
}

type PolicyItem struct {
	ID          uint
	Title       string
	Region      string
	RegionCode  string
	Date        string
	PublishDate string
	Summary     string
	Category    string
	FileUrl     string
}

type GroupedPolicies struct {
	National []PolicyItem
	Province []PolicyItem
	City     []PolicyItem
}

type PolicyListResult struct {
	QueryMode       QueryMode
	SelectedFilters SelectedFilters
	Tips            string
	Policies        GroupedPolicies
	Total           int64
	Page            int
	PageSize        int
}

type PolicyDetailRow struct {
	ID                  uint
	DiseaseID           int
	ScopeLevel          int
	ProvinceCode        string
	CityCode            string
	ProvinceName        string
	CityName            string
	PolicyTitle         string
	PolicyOriginal      string
	PopularInterpret    string
	ReimburseRatio      string
	ReimburseLimit      string
	ReimburseProcess    string
	ReimburseMaterial   string
	RemoteApplyTemplate string
	IsLatest            int
	PublishDate         *time.Time
	EffectiveDate       *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type RelatedPolicy struct {
	ID    uint
	Title string
}

type PolicyDetailResult struct {
	ID                  uint
	Title               string
	Region              string
	RegionCode          string
	PublishDate         string
	EffectiveDate       string
	Category            string
	Content             string
	FileUrl             string
	ReimburseRatio      string
	ReimburseLimit      string
	ReimburseProcess    string
	ReimburseMaterial   string
	RemoteApplyTemplate string
	RelatedPolicies     []RelatedPolicy
}

type MaterialRow struct {
	UrlFlowchart string
	UrlChecklist string
	UrlTemplate  string
	UpdateTime   time.Time
}

type MaterialItem struct {
	Name       string
	Type       string
	URL        string
	Size       string
	UpdateTime string
}
