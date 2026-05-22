package service

import (
	"strconv"
	"time"

	"rare_backend/internal/module/resource/medical/domain"
	"rare_backend/internal/module/resource/medical/repo"
)

type ExaminationService struct {
	repo    *repo.ExaminationRepo
	relRepo *repo.DiseaseRelRepo
}

func NewExaminationService(r *repo.ExaminationRepo, rel *repo.DiseaseRelRepo) *ExaminationService {
	return &ExaminationService{repo: r, relRepo: rel}
}

func BuildExaminationListFilter(keyword, examType, auditStatusStr, diseaseStr, pageStr, pageSizeStr string) domain.ExaminationListFilter {
	diseaseID, _ := strconv.ParseUint(diseaseStr, 10, 64)

	filter := domain.ExaminationListFilter{
		Keyword:     keyword,
		ExamType:    examType,
		AuditStatus: auditStatusStr,
		DiseaseID:   diseaseID,
		Page:        1,
		PageSize:    10,
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

func (s *ExaminationService) List(filter domain.ExaminationListFilter) (*domain.ExaminationListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, domain.ErrCountList
	}

	rows, err := s.repo.List(filter)
	if err != nil {
		return nil, domain.ErrQueryList
	}

	manualIDs := make([]uint64, 0, len(rows))
	for _, item := range rows {
		manualIDs = append(manualIDs, item.ID)
	}
	diseaseMap, _ := s.relRepo.ListDiseaseIDsByExamManualIDs(manualIDs)

	var list []map[string]interface{}
	for _, item := range rows {
		price, duration := mapPriceAndDuration(item.ExamType)
		diseaseIDs := diseaseMap[item.ID]
		if diseaseIDs == nil {
			diseaseIDs = []uint64{}
		}

		list = append(list, map[string]interface{}{
			"id":          item.ID,
			"examName":    item.ExamName,
			"examType":    item.ExamType,
			"examPurpose": item.ExamPurpose,
			"sampleNotes": item.SampleNotes,
			"institution": item.Institution,
			"sort":        item.Sort,
			"price":       price,
			"duration":    duration,
			"diseaseIds":  diseaseIDs,
			"createdAt":   item.CreatedAt.Format(time.RFC3339),
			"updatedAt":   item.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &domain.ExaminationListResult{Total: total, List: list}, nil
}

func (s *ExaminationService) GetByID(id uint64) (*domain.ExaminationDetailResponse, error) {
	row, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrExaminationNotFound
		}
		return nil, err
	}

	diseaseIDs, _ := s.relRepo.GetDiseaseIDsByExamManual(id)

	resp := &domain.ExaminationDetailResponse{
		ID:                row.ID,
		ExamName:          row.ExamName,
		ExamType:          row.ExamType,
		ExamPurpose:       row.ExamPurpose,
		ReferenceValue:    row.ReferenceValue.String,
		AbnormalInterpret: row.AbnormalInterpret.String,
		SampleNotes:       row.SampleNotes.String,
		Institution:       row.Institution.String,
		AuditStatus:       row.AuditStatus,
		RejectReason:      row.RejectReason.String,
		Sort:              row.Sort,
		DiseaseIDs:        diseaseIDs,
		CreatedAt:         row.CreatedAt.Format(time.RFC3339),
	}
	resp.Templates.Excel = row.TemplateExcel.String
	resp.Templates.Word = row.TemplateWord.String
	resp.Templates.Compare = row.CompareTemplate.String

	return resp, nil
}

func (s *ExaminationService) Create(in domain.CreateExaminationInput) (int64, error) {
	return s.repo.Create(in)
}

func (s *ExaminationService) Update(id uint64, in domain.UpdateExaminationInput) error {
	return s.repo.Update(id, in)
}

func (s *ExaminationService) Delete(id uint64) error {
	err := s.repo.Delete(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return domain.ErrExaminationNotFound
		}
		return err
	}
	return nil
}
