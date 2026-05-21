package service

import (
	"strconv"
	"time"

	"rare_backend/internal/module/resource/medical/domain"
	"rare_backend/internal/module/resource/medical/repo"
)

type HospitalService struct {
	repo    *repo.HospitalRepo
	relRepo *repo.DiseaseRelRepo
}

func NewHospitalService(r *repo.HospitalRepo, rel *repo.DiseaseRelRepo) *HospitalService {
	return &HospitalService{repo: r, relRepo: rel}
}

func BuildHospitalListFilter(keyword, provinceCode, cityCode, districtCode, level, isRareNetwork, auditStatus, pageStr, pageSizeStr string) domain.HospitalListFilter {
	filter := domain.HospitalListFilter{
		Keyword:       keyword,
		ProvinceCode:  provinceCode,
		CityCode:      cityCode,
		DistrictCode:  districtCode,
		Level:         level,
		IsRareNetwork: isRareNetwork,
		AuditStatus:   auditStatus,
		Page:          1,
		PageSize:      20,
	}
	if page, err := strconv.Atoi(pageStr); err == nil && page >= 1 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
		if pageSize >= 1 && pageSize <= 100 {
			filter.PageSize = pageSize
		}
	}
	return filter
}

func (s *HospitalService) List(filter domain.HospitalListFilter) (*domain.HospitalListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, domain.ErrCountList
	}

	rows, err := s.repo.List(filter)
	if err != nil {
		return nil, domain.ErrQueryList
	}

	var list []map[string]interface{}
	for _, h := range rows {
		list = append(list, map[string]interface{}{
			"id":            h.ID,
			"name":          h.Name,
			"provinceCode":  h.ProvinceCode,
			"cityCode":      h.CityCode,
			"districtCode":  h.DistrictCode,
			"provinceName":  h.ProvinceName,
			"cityName":      h.CityName,
			"districtName":  h.DistrictName,
			"level":         h.Level,
			"address":       h.Address,
			"phone":         h.Phone,
			"hospitalUrl":   h.HospitalURL.String,
			"treatScope":    h.TreatScope.String,
			"isRareNetwork": h.IsRareNetwork == 1,
			"auditStatus":   h.AuditStatus,
			"createdAt":     h.CreatedAt.Format(time.RFC3339),
		})
	}

	return &domain.HospitalListResult{Total: total, List: list}, nil
}

func (s *HospitalService) GetByID(id uint64) (*domain.HospitalDetailResponse, error) {
	row, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrHospitalNotFound
		}
		return nil, err
	}

	diseases, _ := s.relRepo.GetDiseasesByHospital(id)

	return &domain.HospitalDetailResponse{
		ID:            row.ID,
		Name:          row.Name,
		ProvinceCode:  row.ProvinceCode,
		CityCode:      row.CityCode,
		DistrictCode:  row.DistrictCode,
		ProvinceName:  row.ProvinceName,
		CityName:      row.CityName,
		DistrictName:  row.DistrictName,
		Level:         row.Level,
		Address:       row.Address,
		Phone:         row.Phone,
		HospitalURL:   row.HospitalURL,
		TreatScope:    row.TreatScope,
		IsRareNetwork: row.IsRareNetwork == 1,
		AuditStatus:   row.AuditStatus,
		RejectReason:  row.RejectReason,
		Diseases:      diseases,
		CreatedAt:     row.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *HospitalService) Create(in domain.CreateHospitalInput) (int64, error) {
	return s.repo.Create(in)
}

func (s *HospitalService) Update(id uint64, in domain.UpdateHospitalInput) error {
	exists, err := s.repo.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrHospitalNotFound
	}
	return s.repo.Update(id, in)
}

func (s *HospitalService) Delete(id uint64) error {
	err := s.repo.Delete(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return domain.ErrHospitalNotFound
		}
		return err
	}
	return nil
}

func (s *HospitalService) ListOptions(keyword string) ([]domain.HospitalOptionItem, error) {
	return s.repo.ListOptions(keyword)
}
