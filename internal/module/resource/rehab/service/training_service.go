package service

import (
	"strconv"

	"rare_backend/internal/module/resource/rehab/domain"
	"rare_backend/internal/module/resource/rehab/repo"
)

type TrainingService struct {
	repo *repo.TrainingRepo
}

func NewTrainingService(r *repo.TrainingRepo) *TrainingService {
	return &TrainingService{repo: r}
}

func BuildTrainingListFilter(diseaseStr, stage, keyword, auditStatusStr, pageStr, pageSizeStr string) domain.TrainingListFilter {
	page, pageSize := parsePageParams(pageStr, pageSizeStr, 10)
	return domain.TrainingListFilter{
		DiseaseID:   resolveDiseaseID(diseaseStr),
		RehabStage:  stage,
		Keyword:     keyword,
		AuditStatus: resolveAuditStatus(auditStatusStr),
		Page:        page,
		PageSize:    pageSize,
	}
}

type TrainingListItemResult struct {
	ID              uint
	RehabStage      string
	Title           string
	TrainPurpose    string
	TrainContent    string
	ForbiddenAction string
	PicUrls         []string
	GuidePdf        string
	GuideWord       string
	AuditStatus     int8
	RejectReason    string
	Sort            int
	DiseaseIds      []uint64
	Diseases        []domain.DiseaseItem
	CreatedAt       string
	UpdatedAt       string
}

type TrainingListResult struct {
	List  []TrainingListItemResult
	Total int64
}

func (s *TrainingService) List(filter domain.TrainingListFilter) (*TrainingListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, &domain.CountListError{Cause: err}
	}

	rows, err := s.repo.List(filter)
	if err != nil {
		return nil, &domain.QueryListError{Cause: err}
	}

	var guideIDs []uint64
	for _, t := range rows {
		guideIDs = append(guideIDs, uint64(t.ID))
	}

	diseaseIdsMap, diseaseDetailsMap, _ := s.repo.BatchGetDiseaseRels(guideIDs)

	var listItems []TrainingListItemResult
	for _, t := range rows {
		picUrls := parsePicUrls("")
		if t.PicUrls.Valid && t.PicUrls.String != "" {
			picUrls = parsePicUrls(t.PicUrls.String)
		}

		dids := diseaseIdsMap[uint64(t.ID)]
		if dids == nil {
			dids = []uint64{}
		}
		dDetails := diseaseDetailsMap[uint64(t.ID)]
		if dDetails == nil {
			dDetails = []domain.DiseaseItem{}
		}

		rejectReason := ""
		if t.RejectReason.Valid {
			rejectReason = t.RejectReason.String
		}
		forbiddenAction := ""
		if t.ForbiddenAction.Valid {
			forbiddenAction = t.ForbiddenAction.String
		}
		guidePdf := ""
		if t.GuidePdf.Valid {
			guidePdf = t.GuidePdf.String
		}
		guideWord := ""
		if t.GuideWord.Valid {
			guideWord = t.GuideWord.String
		}

		listItems = append(listItems, TrainingListItemResult{
			ID:              t.ID,
			RehabStage:      t.RehabStage,
			Title:           t.Title,
			TrainPurpose:    t.TrainPurpose,
			TrainContent:    t.TrainContent,
			ForbiddenAction: forbiddenAction,
			PicUrls:         picUrls,
			GuidePdf:        guidePdf,
			GuideWord:       guideWord,
			AuditStatus:     t.AuditStatus,
			RejectReason:    rejectReason,
			Sort:            t.Sort,
			DiseaseIds:      dids,
			Diseases:        dDetails,
			CreatedAt:       t.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:       t.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	if listItems == nil {
		listItems = []TrainingListItemResult{}
	}

	return &TrainingListResult{List: listItems, Total: total}, nil
}

type TrainingDetailResult struct {
	ID              uint
	RehabStage      string
	Title           string
	TrainPurpose    string
	TrainContent    string
	ForbiddenAction string
	PicUrls         []string
	GuidePdf        string
	GuideWord       string
	AuditStatus     int8
	RejectReason    *string
	Sort            int
	DiseaseIds      []uint64
	Diseases        []domain.DiseaseItem
	CreatedAt       string
	UpdatedAt       string
}

func (s *TrainingService) GetByID(id uint64) (*TrainingDetailResult, error) {
	t, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrTrainingNotFound
		}
		return nil, err
	}

	diseaseIds, diseases, _ := s.repo.GetDiseasesByGuideID(id)
	if diseaseIds == nil {
		diseaseIds = []uint64{}
	}
	if diseases == nil {
		diseases = []domain.DiseaseItem{}
	}

	picUrls := parsePicUrls("")
	if t.PicUrls.Valid && t.PicUrls.String != "" {
		picUrls = parsePicUrls(t.PicUrls.String)
	}

	var rejectReasonPtr *string
	if t.RejectReason.Valid {
		rejectReasonPtr = &t.RejectReason.String
	}

	forbiddenAction := ""
	if t.ForbiddenAction.Valid {
		forbiddenAction = t.ForbiddenAction.String
	}
	guidePdf := ""
	if t.GuidePdf.Valid {
		guidePdf = t.GuidePdf.String
	}
	guideWord := ""
	if t.GuideWord.Valid {
		guideWord = t.GuideWord.String
	}

	return &TrainingDetailResult{
		ID:              t.ID,
		RehabStage:      t.RehabStage,
		Title:           t.Title,
		TrainPurpose:    t.TrainPurpose,
		TrainContent:    t.TrainContent,
		ForbiddenAction: forbiddenAction,
		PicUrls:         picUrls,
		GuidePdf:        guidePdf,
		GuideWord:       guideWord,
		AuditStatus:     t.AuditStatus,
		RejectReason:    rejectReasonPtr,
		Sort:            t.Sort,
		DiseaseIds:      diseaseIds,
		Diseases:        diseases,
		CreatedAt:       t.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       t.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

type TrainingResourceResult struct {
	DownloadUrl string
	PreviewUrl  string
	FileName    string
	FileSize    string
}

func (s *TrainingService) GetResource(id uint64, resourceType string) (*TrainingResourceResult, error) {
	training, err := s.repo.GetResource(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrTrainingNotFound
		}
		return nil, err
	}

	var downloadUrl string
	var fileName string

	if resourceType == "pdf" {
		downloadUrl = training.GuidePDF.String
		fileName = training.Title + ".pdf"
	} else {
		downloadUrl = training.GuideWord.String
		fileName = training.Title + ".docx"
	}

	if downloadUrl == "" {
		return nil, domain.ErrResourceNotFound
	}

	return &TrainingResourceResult{
		DownloadUrl: downloadUrl,
		PreviewUrl:  "https://example.com/preview/" + strconv.FormatUint(id, 10),
		FileName:    fileName,
		FileSize:    "2.5MB",
	}, nil
}

func (s *TrainingService) Create(in domain.CreateTrainingInput) (int64, error) {
	return s.repo.Create(in)
}

func (s *TrainingService) Update(id uint64, in domain.UpdateTrainingInput) error {
	return s.repo.Update(id, in)
}

func (s *TrainingService) Delete(id uint64) error {
	rowsAffected, err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrTrainingNotFound
	}
	return nil
}
