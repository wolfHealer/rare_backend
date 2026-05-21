package domain

import "time"

// DiseaseSimpleInfo 用于列表展示的疾病简要信息
type DiseaseSimpleInfo struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias,omitempty"`
}

type HospitalOptionItem struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// --- Hospital ---

type HospitalListFilter struct {
	Keyword        string
	ProvinceCode   string
	CityCode       string
	DistrictCode   string
	Level          string
	IsRareNetwork  string
	AuditStatus    string
	Page           int
	PageSize       int
}

type HospitalListResult struct {
	Total int64                    `json:"total"`
	List  []map[string]interface{} `json:"list"`
}

type CreateHospitalInput struct {
	Name          string
	ProvinceCode  string
	CityCode      string
	DistrictCode  string
	ProvinceName  string
	CityName      string
	DistrictName  string
	Level         string
	IsRareNetwork int8
	TreatScope    string
	Address       string
	Phone         string
	HospitalURL   string
	DiseaseIDs    []uint64
}

type UpdateHospitalInput struct {
	Name          string
	ProvinceCode  *string
	CityCode      *string
	DistrictCode  *string
	ProvinceName  *string
	CityName      *string
	DistrictName  *string
	Level         string
	IsRareNetwork *int8
	TreatScope    *string
	Address       string
	Phone         string
	HospitalURL   *string
	AuditStatus   *int8
	RejectReason  *string
	DiseaseIDs    []uint64
	HasDiseaseIDs bool
}

type HospitalDetailResponse struct {
	ID            uint64              `json:"id"`
	Name          string              `json:"name"`
	ProvinceCode  string              `json:"provinceCode"`
	CityCode      string              `json:"cityCode"`
	DistrictCode  string              `json:"districtCode"`
	ProvinceName  string              `json:"provinceName"`
	CityName      string              `json:"cityName"`
	DistrictName  string              `json:"districtName"`
	Level         string              `json:"level"`
	Address       string              `json:"address"`
	Phone         string              `json:"phone"`
	HospitalURL   string              `json:"hospitalUrl"`
	TreatScope    string              `json:"treatScope"`
	IsRareNetwork bool                `json:"isRareNetwork"`
	AuditStatus   int8                `json:"auditStatus"`
	RejectReason  string              `json:"rejectReason"`
	Diseases      []DiseaseSimpleInfo `json:"diseases"`
	CreatedAt     string              `json:"createdAt"`
}

// --- Doctor ---

type DoctorListFilter struct {
	Keyword     string
	DiseaseID   uint64
	Title       string
	HospitalID  uint64
	Level       string
	AuditStatus string
	Page        int
	PageSize    int
}

type DoctorListResult struct {
	Total int64                    `json:"total"`
	List  []map[string]interface{} `json:"list"`
}

type CreateDoctorInput struct {
	Name       string
	Title      string
	Department string
	GoodAt     string
	ClinicTime string
	Contact    string
	HospitalID uint64
	DiseaseIDs []uint64
	AuditStatus *int8
	Score       *float64
	CommentNum  *int
}

type UpdateDoctorInput struct {
	Name         string
	Title        string
	Department   string
	GoodAt       *string
	ClinicTime   *string
	Contact      *string
	HospitalID   *uint64
	AuditStatus  *int8
	RejectReason *string
	DiseaseIDs   []uint64
	HasDiseaseIDs bool
}

type DoctorDetailResponse struct {
	ID            uint64              `json:"id"`
	Name          string              `json:"name"`
	Title         string              `json:"title"`
	Department    string              `json:"department"`
	GoodAt        string              `json:"goodAt"`
	ClinicTime    string              `json:"clinicTime"`
	Contact       string              `json:"contact"`
	Score         float64             `json:"score"`
	CommentNum    int                 `json:"commentNum"`
	AuditStatus   int8                `json:"auditStatus"`
	RejectReason  string              `json:"rejectReason"`
	HospitalID    uint64              `json:"hospitalId"`
	HospitalName  string              `json:"hospitalName"`
	Diseases      []DiseaseSimpleInfo `json:"diseases"`
	DiseaseIDs    []uint64            `json:"diseaseIds"`
	IsRareNetwork bool                `json:"isRareNetwork"`
	Address       string              `json:"address"`
	CityCode      string              `json:"cityCode"`
	CityName      string              `json:"cityName"`
	DistrictCode  string              `json:"districtCode"`
	DistrictName  string              `json:"districtName"`
	Level         string              `json:"level"`
	Phone         string              `json:"phone"`
	ProvinceCode  string              `json:"provinceCode"`
	ProvinceName  string              `json:"provinceName"`
}

// --- Examination ---

type ExaminationListFilter struct {
	Keyword     string
	ExamType    string
	AuditStatus string
	DiseaseID   uint64
	Page        int
	PageSize    int
}

type ExaminationListResult struct {
	Total int64                    `json:"total"`
	List  []map[string]interface{} `json:"list"`
}

type ExaminationTemplates struct {
	Excel   string
	Word    string
	Compare string
}

type CreateExaminationInput struct {
	ExamName          string
	ExamType          string
	ExamPurpose       string
	ReferenceValue    string
	AbnormalInterpret string
	SampleNotes       string
	Institution       string
	Templates         *ExaminationTemplates
	Sort              int
	DiseaseIDs        []uint64
	AuditStatus       *int8
}

type UpdateExaminationInput struct {
	ExamName          *string
	ExamType          *string
	ExamPurpose       *string
	ReferenceValue    *string
	AbnormalInterpret *string
	SampleNotes       *string
	Institution       *string
	Templates         *ExaminationTemplates
	Sort              *int
	AuditStatus       *int8
	RejectReason      *string
	DiseaseIDs        []uint64
	HasDiseaseIDs     bool
}

type ExaminationDetailResponse struct {
	ID                uint64 `json:"id"`
	ExamName          string `json:"examName"`
	ExamType          string `json:"examType"`
	ExamPurpose       string `json:"examPurpose"`
	ReferenceValue    string `json:"referenceValue"`
	AbnormalInterpret string `json:"abnormalInterpret"`
	SampleNotes       string `json:"sampleNotes"`
	Institution       string `json:"institution"`
	Templates         struct {
		Excel   string `json:"excel"`
		Word    string `json:"word"`
		Compare string `json:"compare"`
	} `json:"templates"`
	AuditStatus  int8     `json:"auditStatus"`
	RejectReason string   `json:"rejectReason"`
	Sort         int      `json:"sort"`
	DiseaseIDs   []uint64 `json:"diseaseIds"`
	CreatedAt    string   `json:"createdAt"`
}

// HospitalRow 内部 repo 扫描用
type HospitalRow struct {
	ID            uint64
	Name          string
	ProvinceCode  string
	CityCode      string
	DistrictCode  string
	ProvinceName  string
	CityName      string
	DistrictName  string
	Level         string
	IsRareNetwork int8
	TreatScope    string
	Address       string
	Phone         string
	HospitalURL   string
	AuditStatus   int8
	RejectReason  string
	CreatedAt     time.Time
}
