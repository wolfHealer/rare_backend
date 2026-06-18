package service

import (
	"database/sql"
	"strings"

	"rare_backend/internal/module/community/domain"
	"rare_backend/internal/module/community/repo"
)

type ReportService struct {
	postRepo   *repo.PostRepo
	reportRepo *repo.ReportRepo
}

func NewReportService(postRepo *repo.PostRepo, reportRepo *repo.ReportRepo) *ReportService {
	return &ReportService{postRepo: postRepo, reportRepo: reportRepo}
}

func (s *ReportService) ReportPost(in domain.ReportPostInput) error {
	reasonType := strings.TrimSpace(in.ReasonType)
	if _, ok := domain.ValidReportReasons[reasonType]; !ok {
		return domain.ErrInvalidReportReason
	}

	ownerID, ok, err := s.postRepo.ExistsReportable(in.PostID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrPostNotFound
	}
	if ownerID == in.ReporterID {
		return domain.ErrCannotReportSelf
	}

	description := strings.TrimSpace(in.Description)
	if len([]rune(description)) > 500 {
		description = string([]rune(description)[:500])
	}

	_, err = s.reportRepo.Create(domain.ReportPostInput{
		PostID:      in.PostID,
		ReporterID:  in.ReporterID,
		ReasonType:  reasonType,
		Description: description,
	})
	return err
}

func (s *ReportService) normalizeListFilter(filter domain.ReportListFilter) domain.ReportListFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 10
	}
	filter.ReasonType = strings.TrimSpace(filter.ReasonType)
	if filter.ReasonType != "" {
		if _, ok := domain.ValidReportReasons[filter.ReasonType]; !ok {
			filter.ReasonType = "__invalid__"
		}
	}
	filter.Keyword = strings.TrimSpace(filter.Keyword)
	return filter
}

func (s *ReportService) ListReports(filter domain.ReportListFilter) ([]domain.ReportItem, int64, error) {
	filter = s.normalizeListFilter(filter)
	return s.reportRepo.List(filter)
}

func (s *ReportService) GetReport(id int64) (*domain.ReportItem, error) {
	item, err := s.reportRepo.GetDetailByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrReportNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *ReportService) HandleReport(in domain.HandleReportInput) error {
	if in.Status != domain.ReportStatusHandled && in.Status != domain.ReportStatusRejected {
		return domain.ErrInvalidReportReason
	}
	note := strings.TrimSpace(in.HandleNote)
	if len([]rune(note)) > 200 {
		note = string([]rune(note)[:200])
	}
	in.HandleNote = note

	status, postID, err := s.reportRepo.GetByID(in.ReportID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrReportNotFound
		}
		return err
	}
	if status != domain.ReportStatusPending {
		return domain.ErrReportAlreadyDone
	}

	if err := s.reportRepo.Handle(in); err != nil {
		return err
	}
	if in.RemovePost && in.Status == domain.ReportStatusHandled {
		return s.postRepo.SoftDelete(postID)
	}
	return nil
}
