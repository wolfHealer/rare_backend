package domain

import (
	"database/sql"
	"time"
)

type DiseaseItem struct {
	ID    int64
	Name  string
	Alias string
}

// --- Institution ---

type InstitutionListFilter struct {
	ProvinceCode string
	CityCode     string
	DistrictCode string
	DiseaseID    string
	Keyword      string
	AuditStatus  string
	Page         int
	PageSize     int
}

type InstitutionRow struct {
	ID            uint
	Name          string
	ProvinceCode  string
	CityCode      string
	DistrictCode  string
	ProvinceName  string
	CityName      string
	DistrictName  string
	Qualification string
	RehabProjects string
	FeeStandard   string
	ContactPhone  string
	ContactUrl    string
	Address       string
	AuditStatus   int8
	RejectReason  sql.NullString
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateInstitutionInput struct {
	Name          string
	ProvinceCode  string
	CityCode      string
	DistrictCode  string
	ProvinceName  string
	CityName      string
	DistrictName  string
	Qualification string
	RehabProjects string
	FeeStandard   string
	ContactPhone  string
	ContactUrl    string
	Address       string
	DiseaseIDs    []int
	AuditStatus   int
	RejectReason  string
}

type UpdateInstitutionInput struct {
	Name          string
	ProvinceCode  *string
	CityCode      *string
	DistrictCode  *string
	ProvinceName  *string
	CityName      *string
	DistrictName  *string
	Qualification string
	RehabProjects string
	FeeStandard   string
	ContactPhone  string
	ContactUrl    string
	Address       string
	DiseaseIDs    []int
	Sort          int
	AuditStatus   *int
	RejectReason  *string
	HasDiseaseIDs bool
}

// --- Psychological Org ---

type PsychOrgListFilter struct {
	ProvinceCode string
	CityCode     string
	DistrictCode string
	ConsultWay   string
	DiseaseID    string
	IsFree       string
	Keyword      string
	AuditStatus  string
	Page         int
	PageSize     int
}

type PsychOrgRow struct {
	ID           uint
	Name         string
	ProvinceCode string
	CityCode     string
	DistrictCode sql.NullString
	ProvinceName sql.NullString
	CityName     sql.NullString
	DistrictName sql.NullString
	Address      string
	ContactPhone string
	ContactUrl   string
	IsFree       sql.NullInt32
	ConsultWay   string
	ContentIntro string
	AuditStatus  int8
	RejectReason sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreatePsychOrgInput struct {
	Name         string
	ProvinceCode string
	CityCode     string
	DistrictCode string
	Address      string
	ContactPhone string
	ContactUrl   string
	IsFree       bool
	ConsultWay   string
	ContentIntro string
	DiseaseIDs   []int
}

type UpdatePsychOrgInput struct {
	Name         string
	ProvinceCode *string
	CityCode     *string
	DistrictCode *string
	Address      string
	ContactPhone string
	ContactUrl   string
	IsFree       *bool
	ConsultWay   string
	ContentIntro string
	DiseaseIDs   []int
	AuditStatus  int
	HasDiseaseIDs bool
}

// --- Training ---

type TrainingListFilter struct {
	DiseaseID   string
	RehabStage  string
	Keyword     string
	AuditStatus string
	Page        int
	PageSize    int
}

type TrainingRow struct {
	ID              uint
	RehabStage      string
	Title           string
	TrainPurpose    string
	TrainContent    string
	ForbiddenAction sql.NullString
	PicUrls         sql.NullString
	GuidePdf        sql.NullString
	GuideWord       sql.NullString
	AuditStatus     int8
	RejectReason    sql.NullString
	Sort            int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreateTrainingInput struct {
	Title           string
	TrainContent    string
	RehabStage      string
	TrainPurpose    string
	ForbiddenAction string
	PicUrls         []string
	GuidePDF        string
	GuideWord       string
	Sort            int
	AuditStatus     *int
	RejectReason    *string
	DiseaseIDs      []int
}

type UpdateTrainingInput struct {
	Title           string
	TrainContent    string
	RehabStage      string
	TrainPurpose    string
	ForbiddenAction string
	PicUrls         []string
	HasPicUrls      bool
	GuidePDF        string
	GuideWord       string
	Sort            int
	AuditStatus     *int
	RejectReason    *string
	DiseaseIDs      []int
	HasDiseaseIDs   bool
}

type TrainingResourceRow struct {
	Title     string
	GuidePDF  sql.NullString
	GuideWord sql.NullString
}

// --- Options ---

type OptionItem struct {
	Text  string
	Value string
}

type RegionItem struct {
	Text  string
	Value string
}

type OrgTypeItem struct {
	Text  string
	Value string
}

type InstitutionOptionsResult struct {
	Regions  []RegionItem
	Types    []OrgTypeItem
	Diseases []OptionItem
}

type TargetItem struct {
	Text  string
	Value string
}

type TargetResponse struct {
	Targets []TargetItem
}

type OrgTypeResponse struct {
	Types []OrgTypeItem
}

type RegionResponse struct {
	Regions []RegionItem
}

type DiseaseRelBatch struct {
	EntityID   uint64
	DiseaseIDs []uint64
	Diseases   []DiseaseItem
}
