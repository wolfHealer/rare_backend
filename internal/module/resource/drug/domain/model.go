package domain

import "time"

// --- Drug ---

type OptionItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type DrugOptionsResult struct {
	Types      []OptionItem `json:"types"`
	Insurances []OptionItem `json:"insurances"`
}

type DrugListItem struct {
	ID               uint      `json:"id"`
	GenericName      string    `json:"generic_name"`
	BrandName        string    `json:"brand_name"`
	Indication       string    `json:"indication"`
	DrugType         string    `json:"drug_type"`
	IsInsurance      bool      `json:"is_insurance"`
	DosageForm       string    `json:"dosage_form"`
	Spec             string    `json:"spec"`
	RefPrice         string    `json:"ref_price"`
	HasRelief        bool      `json:"has_relief"`
	IsLaunched       bool      `json:"is_launched"`
	NeedPrescription bool      `json:"need_prescription"`
	ManualOriginal   string    `json:"manual_original"`
	ManualPopular    string    `json:"manual_popular"`
	AuditStatus      int8      `json:"audit_status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	DiseaseIds       []uint64  `json:"disease_ids"`
}

type DrugListResult struct {
	List     []DrugListItem `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

type DiseaseSimple struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type DrugDetailResponse struct {
	ID               uint            `json:"id"`
	GenericName      string          `json:"genericName"`
	BrandName        string          `json:"brandName"`
	Indication       string          `json:"indication"`
	DrugType         string          `json:"drugType"`
	IsInsurance      bool            `json:"isInsurance"`
	DosageForm       string          `json:"dosageForm"`
	Spec             string          `json:"spec"`
	RefPrice         string          `json:"refPrice"`
	HasRelief        bool            `json:"hasRelief"`
	IsLaunched       bool            `json:"isLaunched"`
	NeedPrescription bool            `json:"needPrescription"`
	ManualOriginal   string          `json:"manualOriginal"`
	ManualPopular    string          `json:"manualPopular"`
	AuditStatus      int8            `json:"auditStatus"`
	CreatedAt        string          `json:"createdAt"`
	UpdatedAt        string          `json:"updatedAt"`
	Diseases         []DiseaseSimple `json:"diseases"`
}

type ManualResponse struct {
	Original string `json:"original"`
	Popular  string `json:"popular"`
	URL      string `json:"url"`
}

type CreateDrugInput struct {
	GenericName      string
	BrandName        string
	Indication       string
	DrugType         string
	IsInsurance      bool
	DosageForm       string
	Spec             string
	RefPrice         float64
	HasRelief        bool
	IsLaunched       bool
	NeedPrescription bool
	ManualOriginal   string
	ManualPopular    string
	AuditStatus      int8
	RejectReason     string
	DiseaseIds       []uint64
}

type UpdateDrugInput struct {
	GenericName      *string
	BrandName        *string
	Indication       *string
	DrugType         *string
	IsInsurance      *bool
	DosageForm       *string
	Spec             *string
	RefPrice         *float64
	HasRelief        *bool
	IsLaunched       *bool
	NeedPrescription *bool
	ManualOriginal   *string
	ManualPopular    *string
	AuditStatus      *int8
	RejectReason     *string
	DiseaseIds       []uint64
	HasDiseaseIds    bool
}

type DrugListFilter struct {
	Keyword             string
	DrugType            string
	IsInsurance         string
	HasRelief           string
	AuditStatus         int
	AuditStatusExplicit bool
	Page                int
	PageSize            int
}

type DrugExportFilter struct {
	DiseaseID   int
	Keyword     string
	TypeFilter  string
	Insurance   string
}

type DrugExportRow struct {
	GenericName string
	BrandName   string
	Indication  string
	DrugType    string
	IsInsurance int8
	DosageForm  string
	Spec        string
	RefPrice    string
	HasRelief   int8
	IsLaunched  int8
}

type CreateIDResult struct {
	ID int64 `json:"id"`
}

// --- Channel ---

type ChannelListResult struct {
	List     []map[string]interface{} `json:"list"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
}

type ChannelDrugItem struct {
	ID          uint   `json:"id"`
	GenericName string `json:"genericName"`
	BrandName   string `json:"brandName"`
}

type ChannelDetailResponse struct {
	ID            uint              `json:"id"`
	Name          string            `json:"name"`
	ChannelType   string            `json:"channelType"`
	ProvinceCode  string            `json:"provinceCode"`
	CityCode      string            `json:"cityCode"`
	DistrictCode  string            `json:"districtCode"`
	Address       string            `json:"address"`
	Desc          string            `json:"desc"`
	ContactPhone  string            `json:"contactPhone"`
	ContactURL    string            `json:"contactUrl"`
	Qualification string            `json:"qualification"`
	DeliveryScope string            `json:"deliveryScope"`
	AuditStatus   int8              `json:"auditStatus"`
	ProvinceName  string            `json:"provinceName"`
	CityName      string            `json:"cityName"`
	DistrictName  string            `json:"districtName"`
	Drugs         []ChannelDrugItem `json:"drugs"`
}

type ChannelContactResponse struct {
	Phone  string `json:"phone"`
	Wechat string `json:"wechat"`
	Email  string `json:"email"`
}

type CreateChannelInput struct {
	DrugID            uint
	Name              string
	ChannelType       string
	ProvinceCode      string
	CityCode          string
	DistrictCode      string
	Address           string
	ContactPhone      string
	ContactURL        string
	DeliveryScope     string
	DeliveryCycle     string
	IsInsuranceSettle bool
	Qualification     string
}

type UpdateChannelInput struct {
	Name              *string
	ChannelType       *string
	ProvinceCode      *string
	CityCode          *string
	DistrictCode      *string
	Address           *string
	ContactPhone      *string
	ContactURL        *string
	DeliveryScope     *string
	DeliveryCycle     *string
	IsInsuranceSettle *bool
	Qualification     *string
	DrugID            *uint
}

type ChannelListFilter struct {
	Keyword                string
	ProvinceCode           string
	CityCode               string
	DistrictCode           string
	ChannelType            string
	DeliveryScope          string
	AuditStatus            int
	AuditStatusExplicit    bool
	IsInsuranceSettle      string
	Page                   int
	PageSize               int
}

type ChannelContactInput struct {
	ContactType string
}

// --- Donation ---

type DonationListResult struct {
	List  []map[string]interface{} `json:"list"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
}

type DonationDetailResponse struct {
	ID            uint   `json:"id"`
	DrugID        uint   `json:"drugId"`
	DiseaseID     int    `json:"diseaseId"`
	Name          string `json:"name"`
	Organizer     string `json:"organizer"`
	Condition     string `json:"condition"`
	Period        string `json:"period"`
	Dosage        string `json:"dosage"`
	ApplyForm     string `json:"applyForm"`
	ApplyGuide    string `json:"applyGuide"`
	MaterialList  string `json:"materialList"`
	ProgressQuery string `json:"progressQuery"`
	DrugName      string `json:"drugName"`
	DiseaseName   string `json:"diseaseName"`
	AuditStatus   int8   `json:"auditStatus"`
	RejectReason  string `json:"rejectReason"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type DonationOptionsResult struct {
	Diseases []OptionItem `json:"diseases"`
	Drugs    []OptionItem `json:"drugs"`
}

type CreateDonationInput struct {
	DrugID         uint
	DiseaseValue   int
	Name           string
	Organizer      string
	ApplyCondition string
	ReliefCycle    string
	DrugDosage     string
	ApplyForm      string
	ApplyGuide     string
	MaterialList   string
	ProgressQuery  string
}

type UpdateDonationInput struct {
	DrugID         *uint
	DiseaseValue   *int
	Name           *string
	Organizer      *string
	ApplyCondition *string
	ReliefCycle    *string
	DrugDosage     *string
	ApplyForm      *string
	ApplyGuide     *string
	MaterialList   *string
	ProgressQuery  *string
}

type DonationListFilter struct {
	Keyword             string
	DiseaseID           int
	DrugID              int
	AuditStatus         int
	AuditStatusExplicit bool
	Organizer           string
	Page                int
	PageSize            int
}

type ApplyDonationInput struct {
	UserID         uint
	PatientName    string
	PatientIdCard  string
	DiagnosisProof string
	IncomeProof    string
	ContactPhone   string
}

type ApplyDonationResult struct {
	ApplicationID string `json:"applicationId"`
	Status        string `json:"status"`
}

type DonationProgressLog struct {
	Time   string `json:"time"`
	Status string `json:"status"`
	Desc   string `json:"desc"`
}

type DonationProgressResponse struct {
	ApplicationID uint64                `json:"applicationId"`
	ProjectID     uint                  `json:"projectId"`
	ProjectName   string                `json:"projectName"`
	Status        string                `json:"status"`
	SubmitTime    string                `json:"submitTime"`
	UpdateTime    string                `json:"updateTime"`
	Logs          []DonationProgressLog `json:"logs"`
}

type DonationGuideResult struct {
	GuideURL string
	Name     string
	IsLocal  bool
	FilePath string
}
