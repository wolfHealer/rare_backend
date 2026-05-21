package service

import (
	"strconv"

	"rare_backend/internal/module/resource/rehab/domain"
	"rare_backend/internal/module/resource/rehab/repo"
)

type PsychologicalService struct {
	repo *repo.PsychologicalRepo
}

func NewPsychologicalService(r *repo.PsychologicalRepo) *PsychologicalService {
	return &PsychologicalService{repo: r}
}

func BuildPsychOrgListFilter(provinceCode, cityCode, districtCode, consultWay, diseaseStr, isFreeStr, keyword, auditStatusStr, pageStr, pageSizeStr string) domain.PsychOrgListFilter {
	page, pageSize := parsePageParams(pageStr, pageSizeStr, 10)
	return domain.PsychOrgListFilter{
		ProvinceCode: provinceCode,
		CityCode:     cityCode,
		DistrictCode: districtCode,
		ConsultWay:   consultWay,
		DiseaseID:    resolvePsychDiseaseID(diseaseStr),
		IsFree:       isFreeStr,
		Keyword:      keyword,
		AuditStatus:  resolveAuditStatus(auditStatusStr),
		Page:         page,
		PageSize:     pageSize,
	}
}

type PsychOrgListItemResult struct {
	ID           uint
	Name         string
	ProvinceCode string
	CityCode     string
	DistrictCode string
	ProvinceName string
	CityName     string
	DistrictName string
	Address      string
	ContactPhone string
	ContactUrl   string
	IsFree       bool
	ConsultWay   string
	ContentIntro string
	AuditStatus  int8
	RejectReason *string
	Type         string
	TypeName     string
	Region       string
	RegionCode   string
	ServiceTime  string
	Description  string
	Services     []string
	Rating       float64
	CoverUrl     string
	Status       string
	DiseaseIds   []uint64
	Diseases     []domain.DiseaseItem
	CreatedAt    string
	UpdatedAt    string
}

type PsychOrgListResult struct {
	List     []PsychOrgListItemResult
	Total    int64
	Page     int
	PageSize int
}

func (s *PsychologicalService) List(filter domain.PsychOrgListFilter) (*PsychOrgListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, &domain.CountListError{Cause: err}
	}

	rows, err := s.repo.List(filter)
	if err != nil {
		return nil, &domain.QueryListError{Cause: err}
	}

	var orgIDs []uint64
	for _, t := range rows {
		orgIDs = append(orgIDs, uint64(t.ID))
	}

	diseaseIdsMap, diseaseDetailsMap, _ := s.repo.BatchGetDiseaseRels(orgIDs)

	var list []PsychOrgListItemResult
	for _, t := range rows {
		isFree := false
		if t.IsFree.Valid && t.IsFree.Int32 == 1 {
			isFree = true
		}

		provinceName := ""
		if t.ProvinceName.Valid {
			provinceName = t.ProvinceName.String
		}
		cityName := ""
		if t.CityName.Valid {
			cityName = t.CityName.String
		}
		districtName := ""
		if t.DistrictName.Valid {
			districtName = t.DistrictName.String
		}

		var rejectReasonPtr *string
		if t.RejectReason.Valid {
			rejectReasonPtr = &t.RejectReason.String
		}

		regionDisplay := t.ProvinceCode
		if cityName != "" {
			regionDisplay += " " + cityName
		}

		orgType, orgTypeName := convertPsychologicalOrgType(t.Name, t.ConsultWay)
		services := parsePsychologicalServices(t.Name, t.ConsultWay)
		serviceTime := getServiceTime(orgType)

		dids := diseaseIdsMap[uint64(t.ID)]
		if dids == nil {
			dids = []uint64{}
		}
		dDetails := diseaseDetailsMap[uint64(t.ID)]
		if dDetails == nil {
			dDetails = []domain.DiseaseItem{}
		}

		list = append(list, PsychOrgListItemResult{
			ID:           t.ID,
			Name:         t.Name,
			ProvinceCode: t.ProvinceCode,
			CityCode:     t.CityCode,
			DistrictCode: "",
			ProvinceName: provinceName,
			CityName:     cityName,
			DistrictName: districtName,
			Address:      t.Address,
			ContactPhone: t.ContactPhone,
			ContactUrl:   t.ContactUrl,
			IsFree:       isFree,
			ConsultWay:   t.ConsultWay,
			ContentIntro: t.ContentIntro,
			AuditStatus:  t.AuditStatus,
			RejectReason: rejectReasonPtr,
			Type:         orgType,
			TypeName:     orgTypeName,
			Region:       regionDisplay,
			RegionCode:   t.ProvinceCode,
			ServiceTime:  serviceTime,
			Description:  t.ContentIntro,
			Services:     services,
			Rating:       calcRating(t.ID),
			CoverUrl:     "https://example.com/orgs/psychological/" + strconv.FormatUint(uint64(t.ID), 10) + ".jpg",
			Status:       "active",
			DiseaseIds:   dids,
			Diseases:     dDetails,
			CreatedAt:    t.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    t.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	if list == nil {
		list = []PsychOrgListItemResult{}
	}

	return &PsychOrgListResult{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

type PsychOrgDetailResult struct {
	ID           uint
	Name         string
	ProvinceCode string
	CityCode     string
	DistrictCode string
	ProvinceName string
	CityName     string
	DistrictName string
	Address      string
	ContactPhone string
	ContactUrl   string
	IsFree       bool
	ConsultWay   string
	ContentIntro string
	AuditStatus  int8
	RejectReason *string
	Type         string
	TypeName     string
	Region       string
	RegionCode   string
	ServiceTime  string
	Description  string
	Services     []string
	Rating       float64
	CoverUrl     string
	Status       string
	Images       []string
	Counselors   []CounselorItem
	DiseaseIds   []uint64
	Diseases     []domain.DiseaseItem
	DiseaseCount int
	CreatedAt    string
	UpdatedAt    string
}

func (s *PsychologicalService) GetByID(id uint64) (*PsychOrgDetailResult, error) {
	org, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrPsychOrgNotFound
		}
		return nil, err
	}

	isFree := false
	if org.IsFree.Valid && org.IsFree.Int32 == 1 {
		isFree = true
	}

	diseaseIds, diseases, _ := s.repo.GetDiseasesByOrgID(id)
	if diseaseIds == nil {
		diseaseIds = []uint64{}
	}
	if diseases == nil {
		diseases = []domain.DiseaseItem{}
	}

	provinceName := ""
	if org.ProvinceName.Valid {
		provinceName = org.ProvinceName.String
	}
	cityName := ""
	if org.CityName.Valid {
		cityName = org.CityName.String
	}
	districtName := ""
	if org.DistrictName.Valid {
		districtName = org.DistrictName.String
	}

	var rejectReasonPtr *string
	if org.RejectReason.Valid {
		rejectReasonPtr = &org.RejectReason.String
	}

	regionDisplay := org.ProvinceCode
	if cityName != "" {
		regionDisplay += " " + cityName
	}

	orgType, orgTypeName := convertPsychologicalOrgType(org.Name, org.ConsultWay)
	services := parsePsychologicalServices(org.Name, org.ConsultWay)
	serviceTime := getServiceTime(orgType)
	images := []string{
		"https://example.com/orgs/psychological/" + strconv.FormatUint(id, 10) + "_1.jpg",
	}
	counselors := getPsychologicalCounselors(org.ID)

	return &PsychOrgDetailResult{
		ID:           org.ID,
		Name:         org.Name,
		ProvinceCode: org.ProvinceCode,
		CityCode:     org.CityCode,
		DistrictCode: "",
		ProvinceName: provinceName,
		CityName:     cityName,
		DistrictName: districtName,
		Address:      org.Address,
		ContactPhone: org.ContactPhone,
		ContactUrl:   org.ContactUrl,
		IsFree:       isFree,
		ConsultWay:   org.ConsultWay,
		ContentIntro: org.ContentIntro,
		AuditStatus:  org.AuditStatus,
		RejectReason: rejectReasonPtr,
		Type:         orgType,
		TypeName:     orgTypeName,
		Region:       regionDisplay,
		RegionCode:   org.ProvinceCode,
		ServiceTime:  serviceTime,
		Description:  org.ContentIntro,
		Services:     services,
		Rating:       calcRating(org.ID),
		CoverUrl:     "https://example.com/orgs/psychological/" + strconv.FormatUint(id, 10) + ".jpg",
		Status:       "active",
		Images:       images,
		Counselors:   counselors,
		DiseaseIds:   diseaseIds,
		Diseases:     diseases,
		DiseaseCount: len(diseases),
		CreatedAt:    org.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    org.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *PsychologicalService) GetGuideTargets() domain.TargetResponse {
	return domain.TargetResponse{
		Targets: []domain.TargetItem{
			{Text: "全部人群", Value: "all"},
			{Text: "患者", Value: "patient"},
			{Text: "家属", Value: "family"},
			{Text: "儿童", Value: "child"},
			{Text: "青少年", Value: "teenager"},
		},
	}
}

func (s *PsychologicalService) GetOrgTypes() domain.OrgTypeResponse {
	return domain.OrgTypeResponse{
		Types: []domain.OrgTypeItem{
			{Text: "全部类型", Value: "all"},
			{Text: "心理热线", Value: "hotline"},
			{Text: "心理中心", Value: "center"},
			{Text: "心理医院", Value: "hospital"},
			{Text: "咨询机构", Value: "clinic"},
			{Text: "在线咨询", Value: "online"},
		},
	}
}

func (s *PsychologicalService) Create(in domain.CreatePsychOrgInput) (int64, error) {
	return s.repo.Create(in)
}

func (s *PsychologicalService) Update(id uint64, in domain.UpdatePsychOrgInput) error {
	return s.repo.Update(id, in)
}

func (s *PsychologicalService) Delete(id uint64) error {
	return s.repo.Delete(id)
}
