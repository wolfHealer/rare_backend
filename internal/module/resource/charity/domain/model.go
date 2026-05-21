package domain

import "time"

// --- Project ---

type ProjectItem struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Organizer       string `json:"organizer"`
	ApplyCondition  string `json:"applyCondition"`
	Status          string `json:"status"`
	ReliefType      string `json:"reliefType"`
	DiseaseIds      []int  `json:"diseaseIds"`
	ReliefStandard  string `json:"reliefStandard"`
	ApplyDifficulty string `json:"applyDifficulty"`
	AuditStatus     int    `json:"auditStatus"`
	RejectReason    string `json:"rejectReason"`
	Sort            int    `json:"sort"`
	UpdatedAt       string `json:"updatedAt"`
}

type ProjectListResult struct {
	List     []ProjectItem `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

type ProjectDiseaseDetail struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type ProjectDetailResponse struct {
	ID              uint                   `json:"id"`
	Name            string                 `json:"name"`
	ApplyProcess    string                 `json:"applyProcess"`
	ApplyCondition  string                 `json:"applyCondition"`
	ApplyDeadline   *time.Time             `json:"applyDeadline"`
	ContactPhone    string                 `json:"contactPhone"`
	ContactUrl      string                 `json:"contactUrl"`
	ApplyForm       string                 `json:"applyForm"`
	ApplyGuide      string                 `json:"applyGuide"`
	MaterialList    string                 `json:"materialList"`
	ReliefType      string                 `json:"reliefType"`
	ReliefStandard  string                 `json:"reliefStandard"`
	ApplyDifficulty string                 `json:"applyDifficulty"`
	Organizer       string                 `json:"organizer"`
	DiseaseIds      []int                  `json:"diseaseIds"`
	Diseases        []ProjectDiseaseDetail `json:"diseases"`
	CreatedAt       string                 `json:"createdAt"`
	UpdatedAt       string                 `json:"updatedAt"`
	Sort            int                    `json:"sort"`
	AuditStatus     int                    `json:"auditStatus"`
	RejectReason    string                 `json:"rejectReason"`
}

type ProjectOptionItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ProjectOptionsResult struct {
	List []ProjectOptionItem `json:"list"`
}

type CreateProjectInput struct {
	Name            string
	ApplyProcess    string
	ApplyCondition  string
	ReliefType      string
	DiseaseIDs      []int
	ReliefStandard  string
	ApplyDifficulty string
	ApplyDeadline   *string
	ContactPhone    string
	ContactURL      string
	ApplyForm       string
	ApplyGuide      string
	MaterialList    string
	Organizer       string
	Sort            int
}

type UpdateProjectInput struct {
	Name            string
	ApplyProcess    string
	ApplyCondition  string
	ReliefType      string
	Type            string
	DiseaseIDs      []int
	ReliefStandard  string
	ApplyDifficulty string
	ApplyDeadline   *string
	ContactPhone    string
	ContactURL      string
	ApplyForm       string
	ApplyGuide      string
	MaterialList    string
	Organizer       string
	Sort            *int
	AuditStatus     *int
	RejectReason    *string
	HasDiseaseIDs   bool
}

type ProjectListFilter struct {
	ReliefType      string
	DiseaseID       int
	ApplyDifficulty string
	AuditStatus     int // -1 = default to approved; -2 = no filter; 0/1/2 = explicit
	Keyword         string
	Page            int
	PageSize        int
}

type ProjectOptionsFilter struct {
	Keyword     string
	AuditStatus int // -1 = no filter; 0/1/2 = explicit
	Page        int
	PageSize    int
}

type CreateIDResult struct {
	ID int64 `json:"id"`
}

// --- Channel ---

type ChannelItem struct {
	ID                   uint    `json:"id"`
	ChannelType          string  `json:"channelType"`
	Name                 string  `json:"name"`
	ApplyCondition       string  `json:"applyCondition"`
	ResponseTime         string  `json:"responseTime"`
	ContactPhone         string  `json:"contactPhone"`
	ContactUrl           string  `json:"contactUrl"`
	HelpLetterTemplate   string  `json:"helpLetterTemplate"`
	CrowdfundingTemplate string  `json:"crowdfundingTemplate"`
	AuditStatus          int     `json:"auditStatus"`
	RejectReason         *string `json:"rejectReason"`
	Sort                 int     `json:"sort"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
}

type ChannelListResult struct {
	List  []ChannelItem `json:"list"`
	Total int64         `json:"total"`
}

type CreateChannelInput struct {
	ChannelType          string
	Name                 string
	ApplyCondition       string
	ResponseTime         string
	ContactPhone         string
	ContactUrl           string
	HelpLetterTemplate   string
	CrowdfundingTemplate string
	Sort                 int
	AuditStatus          *int
}

type UpdateChannelInput struct {
	ChannelType          string
	Name                 string
	ApplyCondition       string
	ResponseTime         string
	ContactPhone         string
	ContactUrl           string
	HelpLetterTemplate   string
	CrowdfundingTemplate string
	Sort                 *int
	AuditStatus          *int
	RejectReason         *string
}

type ChannelListFilter struct {
	ChannelType string
	Keyword     string
	AuditStatus int // -1 = default approved; -2 = no filter; 0/1/2 = explicit
	Page        int
	PageSize    int
}

// --- Case ---

type CaseItem struct {
	ID               uint    `json:"id"`
	DiseaseID        int64   `json:"diseaseId"`
	DiseaseName      string  `json:"diseaseName"`
	ProjectID        uint    `json:"projectId"`
	ProjectName      string  `json:"projectName"`
	CaseTitle        string  `json:"caseTitle"`
	PatientDesc      string  `json:"patientDesc"`
	ApplyCycle       string  `json:"applyCycle"`
	ActualRelief     string  `json:"actualRelief"`
	Experience       string  `json:"experience"`
	PitfallGuide     string  `json:"pitfallGuide"`
	CasePdf          *string `json:"casePdf"`
	MaterialTemplate *string `json:"materialTemplate"`
	AuditStatus      int     `json:"auditStatus"`
	RejectReason     *string `json:"rejectReason"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}

type CaseListResult struct {
	Total int64      `json:"total"`
	List  []CaseItem `json:"list"`
}

type CaseDetailResponse struct {
	ID               uint    `json:"id"`
	DiseaseID        int64   `json:"diseaseId"`
	DiseaseName      string  `json:"diseaseName"`
	ProjectID        uint    `json:"projectId"`
	ProjectName      string  `json:"projectName"`
	CaseTitle        string  `json:"caseTitle"`
	PatientDesc      string  `json:"patientDesc"`
	ApplyCycle       string  `json:"applyCycle"`
	ActualRelief     string  `json:"actualRelief"`
	Experience       string  `json:"experience"`
	PitfallGuide     string  `json:"pitfallGuide"`
	CasePdf          *string `json:"casePdf"`
	MaterialTemplate *string `json:"materialTemplate"`
	AuditStatus      int     `json:"auditStatus"`
	RejectReason     *string `json:"rejectReason"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}

type CasePDFResponse struct {
	DownloadURL string `json:"downloadUrl"`
	FileName    string `json:"fileName"`
	FileSize    string `json:"fileSize"`
	ExpireTime  int    `json:"expireTime"`
}

type DiseaseOptionItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

type DiseaseOptionsResult struct {
	Diseases []DiseaseOptionItem `json:"diseases"`
}

type CreateCaseInput struct {
	DiseaseID        int64
	ProjectID        uint
	CaseTitle        string
	PatientDesc      string
	ApplyCycle       string
	ActualRelief     string
	Experience       string
	PitfallGuide     string
	CasePdf          string
	MaterialTemplate string
	AuditStatus      *int
	RejectReason     string
}

type UpdateCaseInput struct {
	Title            string
	PatientDesc      string
	DiseaseID        *int64
	ActualRelief     string
	Experience       string
	PitfallGuide     string
	ApplyCycle       string
	CasePdf          string
	MaterialTemplate string
	ProjectID        *uint
	AuditStatus      *int
}

type CaseListFilter struct {
	DiseaseID   string
	Keyword     string
	AuditStatus int // -1 = no filter; 0/1/2 = explicit
	Page        int
	PageSize    int
}
