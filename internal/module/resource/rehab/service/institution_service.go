package service

import (
	"strconv"

	"rare_backend/internal/module/resource/rehab/domain"
	"rare_backend/internal/module/resource/rehab/repo"
)

type InstitutionService struct {
	repo        *repo.InstitutionRepo
	optionsRepo *repo.OptionsRepo
}

func NewInstitutionService(r *repo.InstitutionRepo, o *repo.OptionsRepo) *InstitutionService {
	return &InstitutionService{repo: r, optionsRepo: o}
}

func BuildInstitutionListFilter(provinceCode, cityCode, districtCode, diseaseStr, keyword, auditStatusStr, pageStr, pageSizeStr string) domain.InstitutionListFilter {
	page, pageSize := parsePageParams(pageStr, pageSizeStr, 10)
	return domain.InstitutionListFilter{
		ProvinceCode: provinceCode,
		CityCode:     cityCode,
		DistrictCode: districtCode,
		DiseaseID:    resolveDiseaseID(diseaseStr),
		Keyword:      keyword,
		AuditStatus:  resolveAuditStatus(auditStatusStr),
		Page:         page,
		PageSize:     pageSize,
	}
}

type InstitutionListItemResult struct {
	ID            uint
	Name          string
	ProvinceCode  string
	CityCode      string
	DistrictCode  string
	ProvinceName  string
	CityName      string
	DistrictName  string
	Address       string
	ContactPhone  string
	ContactUrl    string
	Qualification string
	RehabProjects string
	FeeStandard   string
	DiseaseIds    []uint64
	AuditStatus   int8
	Rating        float64
	Status        string
	UpdateAt      string
}

type InstitutionListResult struct {
	List     []InstitutionListItemResult
	Total    int64
	Page     int
	PageSize int
}

func (s *InstitutionService) List(filter domain.InstitutionListFilter) (*InstitutionListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, &domain.CountListError{Cause: err}
	}

	rows, err := s.repo.List(filter)
	if err != nil {
		return nil, &domain.QueryListError{Cause: err}
	}

	var institutionIDs []uint64
	for _, t := range rows {
		institutionIDs = append(institutionIDs, uint64(t.ID))
	}

	diseaseMap, _ := s.repo.BatchGetDiseaseIDs(institutionIDs)

	var list []InstitutionListItemResult
	for _, t := range rows {
		dids := diseaseMap[uint64(t.ID)]
		if dids == nil {
			dids = []uint64{}
		}
		list = append(list, InstitutionListItemResult{
			ID:            t.ID,
			Name:          t.Name,
			ProvinceCode:  t.ProvinceCode,
			CityCode:      t.CityCode,
			DistrictCode:  t.DistrictCode,
			ProvinceName:  t.ProvinceName,
			CityName:      t.CityName,
			DistrictName:  t.DistrictName,
			Address:       t.Address,
			ContactPhone:  t.ContactPhone,
			ContactUrl:    t.ContactUrl,
			Qualification: t.Qualification,
			RehabProjects: t.RehabProjects,
			FeeStandard:   t.FeeStandard,
			DiseaseIds:    dids,
			AuditStatus:   t.AuditStatus,
			UpdateAt:      t.UpdatedAt.Format("2006-01-02 15:04:05"),
			Rating:        calcRating(t.ID),
			Status:        "active",
		})
	}
	if list == nil {
		list = []InstitutionListItemResult{}
	}

	return &InstitutionListResult{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

type InstitutionDetailResult struct {
	ID            uint
	Name          string
	Type          string
	TypeName      string
	ProvinceCode  string
	CityCode      string
	DistrictCode  string
	ProvinceName  string
	CityName      string
	DistrictName  string
	Address       string
	ContactPhone  string
	ContactUrl    string
	Qualification string
	RehabProjects string
	FeeStandard   string
	Services      []string
	DiseaseIds    []uint64
	Diseases      []domain.DiseaseItem
	AuditStatus   int8
	RejectReason  string
	CreatedAt     string
	UpdatedAt     string
	Rating        float64
	IsInsurance   bool
	CoverUrl      string
	Images        []string
	BusinessHours string
	Facilities    []string
	Doctors       []DoctorItem
	Status        string
}

func (s *InstitutionService) GetByID(id uint64) (*InstitutionDetailResult, error) {
	inst, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrInstitutionNotFound
		}
		return nil, err
	}

	diseaseIds, diseases, _ := s.repo.GetDiseasesByInstitutionID(id)
	if diseaseIds == nil {
		diseaseIds = []uint64{}
	}
	if diseases == nil {
		diseases = []domain.DiseaseItem{}
	}

	services := parseServices(inst.RehabProjects)
	instType, instTypeName := convertInstitutionType(inst.Name)
	images := []string{
		"https://example.com/institutions/" + strconv.FormatUint(uint64(inst.ID), 10) + "_1.jpg",
	}
	facilities := []string{"无障碍通道", "停车场"}
	doctors := getInstitutionDoctors(inst.ID)

	return &InstitutionDetailResult{
		ID:            inst.ID,
		Name:          inst.Name,
		Type:          instType,
		TypeName:      instTypeName,
		ProvinceCode:  inst.ProvinceCode,
		CityCode:      inst.CityCode,
		DistrictCode:  inst.DistrictCode,
		ProvinceName:  inst.ProvinceName,
		CityName:      inst.CityName,
		DistrictName:  inst.DistrictName,
		Address:       inst.Address,
		ContactPhone:  inst.ContactPhone,
		ContactUrl:    inst.ContactUrl,
		Qualification: inst.Qualification,
		RehabProjects: inst.RehabProjects,
		FeeStandard:   inst.FeeStandard,
		Services:      services,
		DiseaseIds:    diseaseIds,
		Diseases:      diseases,
		AuditStatus:   inst.AuditStatus,
		RejectReason:  inst.RejectReason.String,
		CreatedAt:     inst.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     inst.UpdatedAt.Format("2006-01-02 15:04:05"),
		Rating:        calcRating(inst.ID),
		IsInsurance:   true,
		CoverUrl:      "https://example.com/institutions/" + strconv.FormatUint(uint64(inst.ID), 10) + ".jpg",
		Images:        images,
		BusinessHours: "周一至周日 08:00-17:00",
		Facilities:    facilities,
		Doctors:       doctors,
		Status:        "active",
	}, nil
}

func (s *InstitutionService) GetOptions() (*domain.InstitutionOptionsResult, error) {
	provinces, err := s.optionsRepo.GetDistinctProvinces()
	if err != nil {
		return nil, domain.ErrProvinceOptions
	}

	regions := []domain.RegionItem{{Text: "全部地区", Value: "all"}}
	for _, province := range provinces {
		regions = append(regions, domain.RegionItem{
			Text:  province,
			Value: convertRegionToCode(province),
		})
	}

	types := []domain.OrgTypeItem{
		{Text: "全部类型", Value: "all"},
		{Text: "康复医院", Value: "hospital"},
		{Text: "康复中心", Value: "center"},
		{Text: "康复诊所", Value: "clinic"},
		{Text: "社区康复站", Value: "community"},
	}

	rows, err := s.optionsRepo.GetDiseaseOptionRows()
	if err != nil {
		return nil, domain.ErrDiseaseOptions
	}

	var diseases []domain.OptionItem
	diseases = append(diseases, domain.OptionItem{Text: "全部疾病", Value: "all"})
	for _, row := range rows {
		diseases = append(diseases, domain.OptionItem{
			Text:  row.Name,
			Value: strconv.FormatInt(row.ID, 10),
		})
	}
	if diseases == nil {
		diseases = []domain.OptionItem{}
	}

	return &domain.InstitutionOptionsResult{
		Regions:  regions,
		Types:    types,
		Diseases: diseases,
	}, nil
}

func (s *InstitutionService) GetRegions() domain.RegionResponse {
	return domain.RegionResponse{
		Regions: []domain.RegionItem{
			{Text: "全部地区", Value: "all"},
			{Text: "北京", Value: "bj"},
			{Text: "上海", Value: "sh"},
			{Text: "广州", Value: "gz"},
			{Text: "深圳", Value: "sz"},
			{Text: "浙江", Value: "zj"},
			{Text: "江苏", Value: "js"},
			{Text: "四川", Value: "sc"},
			{Text: "湖北", Value: "hb"},
			{Text: "山东", Value: "sd"},
		},
	}
}

func (s *InstitutionService) Create(in domain.CreateInstitutionInput) (int64, error) {
	return s.repo.Create(in)
}

func (s *InstitutionService) Update(id uint64, in domain.UpdateInstitutionInput) error {
	return s.repo.Update(id, in)
}

func (s *InstitutionService) Delete(id uint64) error {
	return s.repo.Delete(id)
}
