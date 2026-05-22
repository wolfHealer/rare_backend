package service

import (
	"strconv"

	"rare_backend/internal/module/resource/charity/domain"
	"rare_backend/internal/module/resource/charity/repo"
)

type ProjectService struct {
	repo *repo.ProjectRepo
}

func NewProjectService(r *repo.ProjectRepo) *ProjectService {
	return &ProjectService{repo: r}
}

func (s *ProjectService) List(filter domain.ProjectListFilter) (*domain.ProjectListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, domain.ErrCountList
	}

	rows, err := s.repo.List(filter)
	if err != nil {
		return nil, domain.ErrQueryList
	}

	projectIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		projectIDs = append(projectIDs, row.ID)
	}
	diseaseMap, _ := s.repo.ListDiseaseIDsByProjectIDs(projectIDs)

	list := make([]domain.ProjectItem, 0, len(rows))
	for _, row := range rows {
		diseaseIds := diseaseMap[row.ID]
		if diseaseIds == nil {
			diseaseIds = []int{}
		}

		list = append(list, domain.ProjectItem{
			ID:              row.ID,
			Name:            row.Name,
			Organizer:       row.Organizer,
			ApplyCondition:  row.ApplyCondition,
			Status:          auditStatusToStatusText(row.AuditStatus),
			ReliefType:      row.ReliefType,
			DiseaseIds:      diseaseIds,
			ReliefStandard:  row.ReliefStandard,
			ApplyDifficulty: row.ApplyDifficulty,
			AuditStatus:     row.AuditStatus,
			RejectReason:    row.RejectReason,
			Sort:            row.Sort,
			UpdatedAt:       row.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	if list == nil {
		list = []domain.ProjectItem{}
	}

	return &domain.ProjectListResult{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *ProjectService) GetByID(id uint) (*domain.ProjectDetailResponse, error) {
	row, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}

	diseases, err := s.repo.ListDiseaseDetails(id)
	if err != nil {
		diseases = nil
	}

	diseaseIds := make([]int, 0, len(diseases))
	for _, d := range diseases {
		diseaseIds = append(diseaseIds, d.ID)
	}
	if diseaseIds == nil {
		diseaseIds = []int{}
	}
	if diseases == nil {
		diseases = []domain.ProjectDiseaseDetail{}
	}

	return &domain.ProjectDetailResponse{
		ID:              row.ID,
		Name:            row.Name,
		ApplyProcess:    row.ApplyProcess,
		ApplyCondition:  row.ApplyCondition,
		ApplyDeadline:   row.ApplyDeadline,
		ContactPhone:    row.ContactPhone,
		ContactUrl:      row.ContactUrl,
		ApplyForm:       row.ApplyForm,
		ApplyGuide:      row.ApplyGuide,
		MaterialList:    row.MaterialList,
		ReliefType:      row.ReliefType,
		ReliefStandard:  row.ReliefStandard,
		ApplyDifficulty: row.ApplyDifficulty,
		Organizer:       row.Organizer,
		DiseaseIds:      diseaseIds,
		Diseases:        diseases,
		CreatedAt:       row.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       row.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Sort:            row.Sort,
		AuditStatus:     row.AuditStatus,
		RejectReason:    row.RejectReason,
	}, nil
}

func (s *ProjectService) Create(in domain.CreateProjectInput) (int64, error) {
	return s.repo.Create(in)
}

func (s *ProjectService) Update(id uint, in domain.UpdateProjectInput) error {
	if in.ReliefType == "" && in.Type != "" {
		in.ReliefType = in.Type
	}

	updateFields := []string{}
	args := []interface{}{}

	if in.Name != "" {
		updateFields = append(updateFields, "name = ?")
		args = append(args, in.Name)
	}
	if in.ApplyProcess != "" {
		updateFields = append(updateFields, "apply_process = ?")
		args = append(args, in.ApplyProcess)
	}
	if in.ApplyCondition != "" {
		updateFields = append(updateFields, "apply_condition = ?")
		args = append(args, in.ApplyCondition)
	}
	if in.ReliefType != "" {
		updateFields = append(updateFields, "relief_type = ?")
		args = append(args, in.ReliefType)
	}
	if in.ReliefStandard != "" {
		updateFields = append(updateFields, "relief_standard = ?")
		args = append(args, in.ReliefStandard)
	}
	if in.ApplyDifficulty != "" {
		updateFields = append(updateFields, "apply_difficulty = ?")
		args = append(args, in.ApplyDifficulty)
	}
	if in.ApplyDeadline != nil {
		deadlineStr := trimApplyDeadline(*in.ApplyDeadline)
		updateFields = append(updateFields, "apply_deadline = ?")
		args = append(args, deadlineStr)
	}
	if in.ContactPhone != "" {
		updateFields = append(updateFields, "contact_phone = ?")
		args = append(args, in.ContactPhone)
	}
	if in.ContactURL != "" {
		updateFields = append(updateFields, "contact_url = ?")
		args = append(args, in.ContactURL)
	}
	if in.ApplyForm != "" {
		updateFields = append(updateFields, "apply_form = ?")
		args = append(args, in.ApplyForm)
	}
	if in.ApplyGuide != "" {
		updateFields = append(updateFields, "apply_guide = ?")
		args = append(args, in.ApplyGuide)
	}
	if in.MaterialList != "" {
		updateFields = append(updateFields, "material_list = ?")
		args = append(args, in.MaterialList)
	}
	if in.Organizer != "" {
		updateFields = append(updateFields, "organizer = ?")
		args = append(args, in.Organizer)
	}
	if in.Sort != nil {
		updateFields = append(updateFields, "sort = ?")
		args = append(args, *in.Sort)
	}
	if in.AuditStatus != nil {
		updateFields = append(updateFields, "audit_status = ?")
		args = append(args, *in.AuditStatus)
	}
	if in.RejectReason != nil {
		updateFields = append(updateFields, "reject_reason = ?")
		args = append(args, *in.RejectReason)
	}

	if len(updateFields) == 0 && !in.HasDiseaseIDs {
		return nil
	}

	return s.repo.UpdateWithDiseaseRels(id, updateFields, args, in.DiseaseIDs, in.HasDiseaseIDs)
}

func (s *ProjectService) Delete(id uint) error {
	rowsAffected, err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (s *ProjectService) ListOptions(filter domain.ProjectOptionsFilter) (*domain.ProjectOptionsResult, error) {
	list, err := s.repo.ListOptions(filter)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []domain.ProjectOptionItem{}
	}
	return &domain.ProjectOptionsResult{List: list}, nil
}

func BuildProjectListFilter(reliefType, diseaseStr, difficulty, auditStatusStr, keyword, pageStr, pageSizeStr string) domain.ProjectListFilter {
	diseaseID := 0
	if diseaseStr != "" {
		diseaseID, _ = strconv.Atoi(diseaseStr)
	}

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	page, pageSize = normalizePage(page, pageSize, 10)

	return domain.ProjectListFilter{
		ReliefType:      reliefType,
		DiseaseID:       diseaseID,
		ApplyDifficulty: difficulty,
		AuditStatus:     parseAuditStatusParam(auditStatusStr, -1),
		Keyword:         keyword,
		Page:            page,
		PageSize:        pageSize,
	}
}

func BuildProjectOptionsFilter(keyword, auditStatusStr, pageStr, pageSizeStr string) domain.ProjectOptionsFilter {
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	page, pageSize = normalizePage(page, pageSize, 20)

	return domain.ProjectOptionsFilter{
		Keyword:     keyword,
		AuditStatus: parseAuditStatusParam(auditStatusStr, -1),
		Page:        page,
		PageSize:    pageSize,
	}
}
