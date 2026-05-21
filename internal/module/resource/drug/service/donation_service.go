package service

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"rare_backend/internal/module/resource/drug/domain"
	"rare_backend/internal/module/resource/drug/repo"
)

type DonationService struct {
	repo *repo.DonationRepo
}

func NewDonationService(r *repo.DonationRepo) *DonationService {
	return &DonationService{repo: r}
}

func BuildDonationListFilter(keyword, diseaseStr, drugIDStr, auditStatusStr, organizer, pageStr, pageSizeStr string) domain.DonationListFilter {
	diseaseID, _ := strconv.Atoi(diseaseStr)
	drugID, _ := strconv.Atoi(drugIDStr)
	filter := domain.DonationListFilter{
		Keyword:   keyword,
		DiseaseID: diseaseID,
		DrugID:    drugID,
		Organizer: organizer,
		Page:      1,
		PageSize:  10,
	}
	if page, err := strconv.Atoi(pageStr); err == nil && page >= 1 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
		if pageSize >= 1 && pageSize <= 100 {
			filter.PageSize = pageSize
		}
	}
	if auditStatusStr != "" {
		if auditStatus, err := strconv.Atoi(auditStatusStr); err == nil {
			filter.AuditStatus = auditStatus
			filter.AuditStatusExplicit = true
		}
	}
	return filter
}

func buildDrugDisplayName(genericName string, brandName sql.NullString) string {
	drugName := genericName
	if brandName.Valid && brandName.String != "" {
		drugName = brandName.String + " (" + genericName + ")"
	} else if genericName == "" {
		drugName = "未知药品"
	}
	return drugName
}

func (s *DonationService) List(filter domain.DonationListFilter) (*domain.DonationListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, domain.ErrCountList
	}

	rows, err := s.repo.List(filter)
	if err != nil {
		return nil, domain.ErrQueryList
	}

	list := make([]map[string]interface{}, 0, len(rows))
	for _, project := range rows {
		drugName := buildDrugDisplayName(project.GenericName, project.BrandName)
		diseaseName := "未知疾病"
		if project.DiseaseName.Valid {
			diseaseName = project.DiseaseName.String
		}

		list = append(list, map[string]interface{}{
			"id":               project.ID,
			"drugId":           project.DrugID,
			"diseaseId":        project.DiseaseID,
			"name":             project.Name,
			"organizer":        project.Organizer,
			"applyCondition":   project.ApplyCondition,
			"reliefCycle":      project.ReliefCycle,
			"reliefDosageDesc": project.ReliefDosageDesc,
			"applyForm":        project.ApplyForm.String,
			"applyGuide":       project.ApplyGuide.String,
			"materialList":     project.MaterialList.String,
			"progressQuery":    project.ProgressQuery.String,
			"drugName":         drugName,
			"diseaseName":      diseaseName,
			"auditStatus":      project.AuditStatus,
			"updatedAt":        project.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &domain.DonationListResult{
		List:  list,
		Total: total,
		Page:  filter.Page,
	}, nil
}

func (s *DonationService) GetByID(id uint) (*domain.DonationDetailResponse, error) {
	row, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrDonationNotFound
		}
		return nil, err
	}

	drugName := buildDrugDisplayName(row.GenericName, row.BrandName)
	diseaseName := "未知疾病"
	if row.DiseaseName.Valid {
		diseaseName = row.DiseaseName.String
	}
	rejectReason := ""
	if row.RejectReason.Valid {
		rejectReason = row.RejectReason.String
	}

	return &domain.DonationDetailResponse{
		ID:            row.ID,
		DrugID:        row.DrugID,
		DiseaseID:     row.DiseaseID,
		Name:          row.Name,
		Organizer:     row.Organizer,
		Condition:     row.ApplyCondition,
		Period:        row.ReliefCycle,
		Dosage:        row.ReliefDosageDesc,
		ApplyForm:     row.ApplyForm.String,
		ApplyGuide:    row.ApplyGuide.String,
		MaterialList:  row.MaterialList.String,
		ProgressQuery: row.ProgressQuery.String,
		DrugName:      drugName,
		DiseaseName:   diseaseName,
		AuditStatus:   row.AuditStatus,
		RejectReason:  rejectReason,
		CreatedAt:     row.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     row.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *DonationService) Create(input domain.CreateDonationInput) (int64, error) {
	exists, err := s.repo.DrugApprovedExists(input.DrugID)
	if err != nil {
		if repo.IsNoRows(err) {
			return 0, domain.ErrDrugNotApproved
		}
		return 0, err
	}
	if !exists {
		return 0, domain.ErrDrugNotApproved
	}
	return s.repo.Create(input)
}

func (s *DonationService) Update(id uint, input domain.UpdateDonationInput) error {
	exists, err := s.repo.Exists(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return domain.ErrDonationNotFound
		}
		return err
	}
	if !exists {
		return domain.ErrDonationNotFound
	}

	fields := []string{}
	args := []interface{}{}

	if input.DrugID != nil {
		ok, err := s.repo.DrugApprovedExists(*input.DrugID)
		if err != nil || !ok {
			return domain.ErrDrugNotApproved
		}
		fields = append(fields, "drug_id = ?")
		args = append(args, *input.DrugID)
	}
	if input.DiseaseValue != nil {
		fields = append(fields, "disease_id = ?")
		args = append(args, *input.DiseaseValue)
	}
	if input.Name != nil {
		fields = append(fields, "name = ?")
		args = append(args, *input.Name)
	}
	if input.Organizer != nil {
		fields = append(fields, "organizer = ?")
		args = append(args, *input.Organizer)
	}
	if input.ApplyCondition != nil {
		fields = append(fields, "apply_condition = ?")
		args = append(args, *input.ApplyCondition)
	}
	if input.ReliefCycle != nil {
		fields = append(fields, "relief_cycle = ?")
		args = append(args, *input.ReliefCycle)
	}
	if input.DrugDosage != nil {
		fields = append(fields, "relief_dosage_desc = ?")
		args = append(args, *input.DrugDosage)
	}
	if input.ApplyForm != nil {
		fields = append(fields, "apply_form = ?")
		args = append(args, *input.ApplyForm)
	}
	if input.ApplyGuide != nil {
		fields = append(fields, "apply_guide = ?")
		args = append(args, *input.ApplyGuide)
	}
	if input.MaterialList != nil {
		fields = append(fields, "material_list = ?")
		args = append(args, *input.MaterialList)
	}
	if input.ProgressQuery != nil {
		fields = append(fields, "progress_query = ?")
		args = append(args, *input.ProgressQuery)
	}

	if len(fields) == 0 {
		return domain.ErrNoUpdateFields
	}
	return s.repo.Update(id, fields, args)
}

func (s *DonationService) Delete(id uint) error {
	exists, err := s.repo.ApprovedExists(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return domain.ErrDonationNotFound
		}
		return err
	}
	if !exists {
		return domain.ErrDonationNotFound
	}
	return s.repo.SoftDelete(id)
}

func (s *DonationService) Apply(projectID uint, input domain.ApplyDonationInput) (*domain.ApplyDonationResult, error) {
	_, err := s.repo.GetApprovedProject(projectID)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrDonationNotFound
		}
		return nil, err
	}

	idCardEnc := base64.StdEncoding.EncodeToString([]byte(input.PatientIdCard))
	maskedID := input.PatientIdCard[:6] + "********" + input.PatientIdCard[14:]
	applicationID := fmt.Sprintf("RELIEF%s%03d", time.Now().Format("20060102"), projectID)

	appID, err := s.repo.InsertApplication(projectID, input, idCardEnc, maskedID)
	if err != nil {
		return nil, err
	}
	_ = s.repo.InsertApplicationLog(appID)

	return &domain.ApplyDonationResult{
		ApplicationID: applicationID,
		Status:        "pending",
	}, nil
}

func (s *DonationService) GetProgress(idParam string) (*domain.DonationProgressResponse, error) {
	var progress *repo.DonationProgressRow
	var err error

	if appIDUint, parseErr := strconv.ParseUint(idParam, 10, 64); parseErr == nil {
		progress, err = s.getProgressByNumeric(appIDUint)
	} else {
		progress, err = s.getProgressByApplicationID(idParam)
	}
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrDonationNotFound
		}
		return nil, err
	}

	logRows, err := s.repo.ListProgressLogs(progress.ApplicationID)
	if err != nil {
		return nil, err
	}

	logs := make([]domain.DonationProgressLog, 0, len(logRows))
	for _, log := range logRows {
		logs = append(logs, domain.DonationProgressLog{
			Time:   log.CreatedAt.Format("2006-01-02 15:04:05"),
			Status: log.Status,
			Desc:   log.ActionDesc,
		})
	}

	return &domain.DonationProgressResponse{
		ApplicationID: progress.ApplicationID,
		ProjectID:     progress.ProjectID,
		ProjectName:   progress.ProjectName,
		Status:        progress.Status,
		SubmitTime:    progress.SubmitTime.Format("2006-01-02 15:04:05"),
		UpdateTime:    progress.UpdateTime.Format("2006-01-02 15:04:05"),
		Logs:          logs,
	}, nil
}

func (s *DonationService) GetGuide(projectID uint) (*domain.DonationGuideResult, error) {
	guideURL, name, err := s.repo.GetGuide(projectID)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrDonationNotFound
		}
		return nil, err
	}
	if guideURL == "" {
		return nil, domain.ErrGuideNotFound
	}

	result := &domain.DonationGuideResult{
		GuideURL: guideURL,
		Name:     name,
	}
	if strings.HasPrefix(guideURL, "/") || strings.HasPrefix(guideURL, "./") {
		filePath := guideURL
		if !strings.HasPrefix(filePath, "/") {
			filePath = "./" + filePath
		}
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return nil, domain.ErrGuideNotFound
		}
		result.IsLocal = true
		result.FilePath = filePath
	}
	return result, nil
}

func (s *DonationService) Options() (*domain.DonationOptionsResult, error) {
	diseaseIDs, err := s.repo.ListDiseaseOptions()
	if err != nil {
		return nil, err
	}

	diseases := make([]domain.OptionItem, 0, len(diseaseIDs))
	for _, diseaseID := range diseaseIDs {
		diseases = append(diseases, domain.OptionItem{
			Label: fmt.Sprintf("疾病分类%d", diseaseID),
			Value: strconv.Itoa(diseaseID),
		})
	}

	drugRows, err := s.repo.ListDrugOptions()
	if err != nil {
		return nil, err
	}

	drugs := make([]domain.OptionItem, 0, len(drugRows))
	for _, drug := range drugRows {
		name := drug.GenericName
		if drug.BrandName.Valid && drug.BrandName.String != "" {
			name = drug.BrandName.String
		}
		drugs = append(drugs, domain.OptionItem{
			Label: name,
			Value: strconv.FormatUint(uint64(drug.ID), 10),
		})
	}

	return &domain.DonationOptionsResult{
		Diseases: diseases,
		Drugs:    drugs,
	}, nil
}

func (s *DonationService) getProgressByNumeric(appID uint64) (*repo.DonationProgressRow, error) {
	return s.repo.GetProgressByNumericID(appID)
}

func (s *DonationService) getProgressByApplicationID(applicationID string) (*repo.DonationProgressRow, error) {
	return s.repo.GetProgressByApplicationID(applicationID)
}
