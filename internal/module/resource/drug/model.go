package drug

type CreateDrugRequest struct {
	GenericName      string   `json:"genericName"`
	BrandName        string   `json:"brandName"`
	Indication       string   `json:"indication"`
	DrugType         string   `json:"drugType"`
	IsInsurance      bool     `json:"isInsurance"`
	DosageForm       string   `json:"dosageForm"`
	Spec             string   `json:"spec"`
	RefPrice         float64  `json:"refPrice"`
	HasRelief        bool     `json:"hasRelief"`
	IsLaunched       bool     `json:"isLaunched"`
	NeedPrescription bool     `json:"needPrescription"`
	ManualOriginal   string   `json:"manualOriginal"`
	ManualPopular    string   `json:"manualPopular"`
	AuditStatus      int8     `json:"auditStatus"`
	RejectReason     string   `json:"rejectReason"`
	DiseaseIds       []uint64 `json:"diseaseIds"`
}

type UpdateDrugRequest struct {
	GenericName      *string  `json:"genericName"`
	BrandName        *string  `json:"brandName"`
	Indication       *string  `json:"indication"`
	DrugType         *string  `json:"drugType"`
	IsInsurance      *bool    `json:"isInsurance"`
	DosageForm       *string  `json:"dosageForm"`
	Spec             *string  `json:"spec"`
	RefPrice         *float64 `json:"refPrice"`
	HasRelief        *bool    `json:"hasRelief"`
	IsLaunched       *bool    `json:"isLaunched"`
	NeedPrescription *bool    `json:"needPrescription"`
	ManualOriginal   *string  `json:"manualOriginal"`
	ManualPopular    *string  `json:"manualPopular"`
	AuditStatus      *int8    `json:"auditStatus"`
	RejectReason     *string  `json:"rejectReason"`
	DiseaseIds       []uint64 `json:"diseaseIds"`
}

type CreateChannelRequest struct {
	DrugID            uint   `json:"drugId"`
	Name              string `json:"name"`
	ChannelType       string `json:"channelType"`
	ProvinceCode      string `json:"provinceCode"`
	CityCode          string `json:"cityCode"`
	DistrictCode      string `json:"districtCode"`
	Address           string `json:"address"`
	ContactPhone      string `json:"contactPhone"`
	ContactURL        string `json:"contactUrl"`
	DeliveryScope     string `json:"deliveryScope"`
	DeliveryCycle     string `json:"deliveryCycle"`
	IsInsuranceSettle bool   `json:"isInsuranceSettle"`
	Qualification     string `json:"qualification"`
}

type UpdateChannelRequest struct {
	Name              *string `json:"name"`
	ChannelType       *string `json:"channelType"`
	ProvinceCode      *string `json:"provinceCode"`
	CityCode          *string `json:"cityCode"`
	DistrictCode      *string `json:"districtCode"`
	Address           *string `json:"address"`
	ContactPhone      *string `json:"contactPhone"`
	ContactURL        *string `json:"contactUrl"`
	DeliveryScope     *string `json:"deliveryScope"`
	DeliveryCycle     *string `json:"deliveryCycle"`
	IsInsuranceSettle *bool   `json:"isInsuranceSettle"`
	Qualification     *string `json:"qualification"`
	DrugID            *uint   `json:"drugId"`
}

type CreateDonationRequest struct {
	DrugID         uint   `json:"drugId"`
	DiseaseValue   int    `json:"diseaseValue"`
	Name           string `json:"name"`
	Organizer      string `json:"organizer"`
	ApplyCondition string `json:"applyCondition"`
	ReliefCycle    string `json:"reliefCycle"`
	DrugDosage     string `json:"drugDosage"`
	ApplyForm      string `json:"applyForm"`
	ApplyGuide     string `json:"applyGuide"`
	MaterialList   string `json:"materialList"`
	ProgressQuery  string `json:"progressQuery"`
}

type UpdateDonationRequest struct {
	DrugID         *uint   `json:"drugId"`
	DiseaseValue   *int    `json:"diseaseValue"`
	Name           *string `json:"name"`
	Organizer      *string `json:"organizer"`
	ApplyCondition *string `json:"applyCondition"`
	ReliefCycle    *string `json:"reliefCycle"`
	DrugDosage     *string `json:"drugDosage"`
	ApplyForm      *string `json:"applyForm"`
	ApplyGuide     *string `json:"applyGuide"`
	MaterialList   *string `json:"materialList"`
	ProgressQuery  *string `json:"progressQuery"`
}

type ContactChannelRequest struct {
	ContactType string `json:"contactType"`
}

type ApplyDonationRequest struct {
	UserID         uint   `json:"userId"`
	PatientName    string `json:"patientName"`
	PatientIdCard  string `json:"patientIdCard"`
	DiagnosisProof string `json:"diagnosisProof"`
	IncomeProof    string `json:"incomeProof"`
	ContactPhone   string `json:"contactPhone"`
}
