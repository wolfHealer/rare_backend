package service

import (
	"strconv"

	"rare_backend/internal/module/resource/charity/domain"
	"rare_backend/internal/module/resource/charity/repo"
)

type CaseService struct {
	repo *repo.CaseRepo
}

func NewCaseService(r *repo.CaseRepo) *CaseService {
	return &CaseService{repo: r}
}

func (s *CaseService) List(filter domain.CaseListFilter) (*domain.CaseListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, domain.ErrCountList
	}

	rows, err := s.repo.List(filter)
	if err != nil {
		return nil, domain.ErrQueryList
	}

	list := make([]domain.CaseItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, mapCaseRow(row))
	}
	if list == nil {
		list = []domain.CaseItem{}
	}

	return &domain.CaseListResult{
		Total: total,
		List:  list,
	}, nil
}

func (s *CaseService) GetByID(id uint) (*domain.CaseDetailResponse, error) {
	row, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrCaseNotFound
		}
		return nil, err
	}
	item := mapCaseRow(*row)
	return &domain.CaseDetailResponse{
		ID:               item.ID,
		DiseaseID:        item.DiseaseID,
		DiseaseName:      item.DiseaseName,
		ProjectID:        item.ProjectID,
		ProjectName:      item.ProjectName,
		CaseTitle:        item.CaseTitle,
		PatientDesc:      item.PatientDesc,
		ApplyCycle:       item.ApplyCycle,
		ActualRelief:     item.ActualRelief,
		Experience:       item.Experience,
		PitfallGuide:     item.PitfallGuide,
		CasePdf:          item.CasePdf,
		MaterialTemplate: item.MaterialTemplate,
		AuditStatus:      item.AuditStatus,
		RejectReason:     item.RejectReason,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}, nil
}

func (s *CaseService) GetPDF(id uint) (*domain.CasePDFResponse, error) {
	row, err := s.repo.GetApprovedPDF(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrCaseNotAudited
		}
		return nil, err
	}
	if row.CasePDF == "" {
		return nil, domain.ErrCaseNoPDF
	}

	return &domain.CasePDFResponse{
		DownloadURL: row.CasePDF,
		FileName:    row.CaseTitle + ".pdf",
		FileSize:    getFileSize(row.CasePDF),
		ExpireTime:  3600,
	}, nil
}

func (s *CaseService) ListDiseaseOptions() (*domain.DiseaseOptionsResult, error) {
	rows, err := s.repo.ListDiseaseOptions()
	if err != nil {
		return nil, domain.ErrDiseaseOptions
	}

	diseases := []domain.DiseaseOptionItem{
		{Text: "全部疾病", Value: "all"},
	}
	for _, row := range rows {
		diseases = append(diseases, domain.DiseaseOptionItem{
			Text:  row.Name,
			Value: strconv.FormatInt(row.DiseaseID, 10),
		})
	}

	return &domain.DiseaseOptionsResult{Diseases: diseases}, nil
}

func (s *CaseService) Create(in domain.CreateCaseInput) (int64, error) {
	if in.DiseaseID == 0 {
		return 0, domain.ErrDiseaseIDRequired
	}
	if in.ProjectID == 0 {
		return 0, domain.ErrProjectIDRequired
	}
	if in.CaseTitle == "" {
		return 0, domain.ErrCaseTitleRequired
	}
	if in.PatientDesc == "" {
		return 0, domain.ErrPatientDescRequired
	}
	if in.ApplyCycle == "" {
		return 0, domain.ErrApplyCycleRequired
	}
	if in.ActualRelief == "" {
		return 0, domain.ErrActualReliefRequired
	}
	if in.Experience == "" {
		return 0, domain.ErrExperienceRequired
	}
	if in.PitfallGuide == "" {
		return 0, domain.ErrPitfallGuideRequired
	}

	auditStatus := 0
	if in.AuditStatus != nil {
		if *in.AuditStatus >= 0 && *in.AuditStatus <= 2 {
			auditStatus = *in.AuditStatus
		}
	}
	return s.repo.Create(in, auditStatus)
}

func (s *CaseService) Update(id uint, in domain.UpdateCaseInput) error {
	updateFields := []string{}
	args := []interface{}{}

	if in.Title != "" {
		updateFields = append(updateFields, "case_title = ?")
		args = append(args, in.Title)
	}
	if in.PatientDesc != "" {
		updateFields = append(updateFields, "patient_desc = ?")
		args = append(args, in.PatientDesc)
	}
	if in.DiseaseID != nil {
		updateFields = append(updateFields, "disease_id = ?")
		args = append(args, *in.DiseaseID)
	}
	if in.ActualRelief != "" {
		updateFields = append(updateFields, "actual_relief = ?")
		args = append(args, in.ActualRelief)
	}
	if in.Experience != "" {
		updateFields = append(updateFields, "experience = ?")
		args = append(args, in.Experience)
	}
	if in.PitfallGuide != "" {
		updateFields = append(updateFields, "pitfall_guide = ?")
		args = append(args, in.PitfallGuide)
	}
	if in.ApplyCycle != "" {
		updateFields = append(updateFields, "apply_cycle = ?")
		args = append(args, in.ApplyCycle)
	}
	if in.CasePdf != "" {
		updateFields = append(updateFields, "case_pdf = ?")
		args = append(args, in.CasePdf)
	}
	if in.MaterialTemplate != "" {
		updateFields = append(updateFields, "material_template = ?")
		args = append(args, in.MaterialTemplate)
	}
	if in.ProjectID != nil {
		updateFields = append(updateFields, "project_id = ?")
		args = append(args, *in.ProjectID)
	}
	if in.AuditStatus != nil {
		updateFields = append(updateFields, "audit_status = ?")
		args = append(args, *in.AuditStatus)
	}

	if len(updateFields) == 0 {
		return domain.ErrNoUpdateFields
	}

	return s.repo.Update(id, updateFields, args)
}

func (s *CaseService) Delete(id uint) error {
	rowsAffected, err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrCaseNotFound
	}
	return nil
}

func mapCaseRow(row repo.CaseRow) domain.CaseItem {
	return domain.CaseItem{
		ID:               row.ID,
		DiseaseID:        row.DiseaseID,
		DiseaseName:      row.DiseaseName,
		ProjectID:        row.ProjectID,
		ProjectName:      row.ProjectName,
		CaseTitle:        row.CaseTitle,
		PatientDesc:      row.PatientDesc,
		ApplyCycle:       row.ApplyCycle,
		ActualRelief:     row.ActualRelief,
		Experience:       row.Experience,
		PitfallGuide:     row.PitfallGuide,
		CasePdf:          nullStringPtr(row.CasePdf),
		MaterialTemplate: nullStringPtr(row.MaterialTemplate),
		AuditStatus:      row.AuditStatus,
		RejectReason:     nullStringPtr(row.RejectReason),
		CreatedAt:        row.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        row.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func BuildCaseListFilter(diseaseID, keyword, auditStatusStr, pageStr, pageSizeStr string) domain.CaseListFilter {
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	page, pageSize = normalizePage(page, pageSize, 10)

	return domain.CaseListFilter{
		DiseaseID:   diseaseID,
		Keyword:     keyword,
		AuditStatus: parseAuditStatusParam(auditStatusStr, -1),
		Page:        page,
		PageSize:    pageSize,
	}
}
