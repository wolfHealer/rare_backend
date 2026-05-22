package service

import (
	"strconv"

	"rare_backend/internal/module/resource/medical/domain"
	"rare_backend/internal/module/resource/medical/repo"
)

type DoctorService struct {
	repo    *repo.DoctorRepo
	relRepo *repo.DiseaseRelRepo
}

func NewDoctorService(r *repo.DoctorRepo, rel *repo.DiseaseRelRepo) *DoctorService {
	return &DoctorService{repo: r, relRepo: rel}
}

func BuildDoctorListFilter(keyword, diseaseStr, title, hospitalIdStr, level, auditStatusStr, pageStr, pageSizeStr string) domain.DoctorListFilter {
	diseaseID, _ := strconv.ParseUint(diseaseStr, 10, 64)
	var hospitalID uint64
	if hospitalIdStr != "" {
		hospitalID, _ = strconv.ParseUint(hospitalIdStr, 10, 64)
	}

	filter := domain.DoctorListFilter{
		Keyword:     keyword,
		DiseaseID:   diseaseID,
		Title:       title,
		HospitalID:  hospitalID,
		Level:       level,
		AuditStatus: auditStatusStr,
		Page:        1,
		PageSize:    20,
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

func (s *DoctorService) List(filter domain.DoctorListFilter) (*domain.DoctorListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, domain.ErrCountList
	}

	rows, err := s.repo.List(filter)
	if err != nil {
		return nil, domain.ErrQueryList
	}

	doctorIDs := make([]uint64, 0, len(rows))
	for _, doc := range rows {
		doctorIDs = append(doctorIDs, doc.ID)
	}
	diseaseMap, _ := s.relRepo.ListDiseaseIDsByDoctorIDs(doctorIDs)

	var list []map[string]interface{}
	for _, doc := range rows {
		diseaseIDs := diseaseMap[doc.ID]
		if diseaseIDs == nil {
			diseaseIDs = []uint64{}
		}

		list = append(list, map[string]interface{}{
			"id":            doc.ID,
			"name":          doc.Name,
			"title":         doc.Title,
			"department":    doc.Department,
			"goodAt":        doc.GoodAt.String,
			"clinicTime":    doc.ClinicTime.String,
			"contact":       doc.Contact.String,
			"rating":        doc.Score,
			"reviewCount":   doc.CommentNum,
			"hospital":      doc.HospitalName,
			"hospitalId":    doc.ID,
			"provinceCode":  doc.ProvinceCode,
			"cityCode":      doc.CityCode,
			"districtCode":  doc.DistrictCode,
			"provinceName":  doc.ProvinceName,
			"cityName":      doc.CityName,
			"districtName":  doc.DistrictName,
			"level":         doc.Level,
			"isRareNetwork": doc.IsRareNetwork == 1,
			"auditStatus":   doc.AuditStatus,
			"diseaseIds":    diseaseIDs,
		})
	}

	return &domain.DoctorListResult{Total: total, List: list}, nil
}

func (s *DoctorService) GetByID(id uint64) (*domain.DoctorDetailResponse, error) {
	row, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrDoctorNotFound
		}
		return nil, err
	}

	diseaseDetails, _ := s.relRepo.GetDiseaseDetailsByDoctor(id)
	diseaseIDs, _ := s.relRepo.GetDiseaseIDsByDoctor(id)

	return &domain.DoctorDetailResponse{
		ID:            row.ID,
		Name:          row.Name,
		Title:         row.Title,
		Department:    row.Department,
		GoodAt:        row.GoodAt.String,
		ClinicTime:    row.ClinicTime.String,
		Contact:       row.Contact.String,
		Score:         row.Score,
		CommentNum:    row.CommentNum,
		AuditStatus:   row.AuditStatus,
		RejectReason:  row.RejectReason.String,
		HospitalID:    row.HospitalID,
		HospitalName:  row.HospitalName,
		Diseases:      diseaseDetails,
		DiseaseIDs:    diseaseIDs,
		IsRareNetwork: row.IsRareNetwork == 1,
		Address:       row.Address,
		CityCode:      row.CityCode,
		CityName:      row.CityName,
		DistrictCode:  row.DistrictCode,
		DistrictName:  row.DistrictName,
		Level:         row.Level,
		Phone:         row.Phone,
		ProvinceCode:  row.ProvinceCode,
		ProvinceName:  row.ProvinceName,
	}, nil
}

func (s *DoctorService) Create(in domain.CreateDoctorInput) (int64, error) {
	return s.repo.Create(in)
}

func (s *DoctorService) Update(id uint64, in domain.UpdateDoctorInput) error {
	return s.repo.Update(id, in)
}

func (s *DoctorService) Delete(id uint64) error {
	err := s.repo.Delete(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return domain.ErrDoctorNotFound
		}
		return err
	}
	return nil
}
