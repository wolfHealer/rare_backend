package service

import (
	"strconv"

	"rare_backend/internal/module/resource/charity/domain"
	"rare_backend/internal/module/resource/charity/repo"
)

type ChannelService struct {
	repo *repo.ChannelRepo
}

func NewChannelService(r *repo.ChannelRepo) *ChannelService {
	return &ChannelService{repo: r}
}

func (s *ChannelService) List(filter domain.ChannelListFilter) (*domain.ChannelListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, domain.ErrCountList
	}

	rows, err := s.repo.List(filter)
	if err != nil {
		return nil, domain.ErrQueryList
	}

	channels := make([]domain.ChannelItem, 0, len(rows))
	for _, row := range rows {
		channels = append(channels, mapChannelRow(row))
	}
	if channels == nil {
		channels = []domain.ChannelItem{}
	}

	return &domain.ChannelListResult{
		List:  channels,
		Total: total,
	}, nil
}

func (s *ChannelService) GetByID(id uint) (*domain.ChannelItem, error) {
	row, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrChannelNotFound
		}
		return nil, err
	}
	item := mapChannelRow(*row)
	return &item, nil
}

func (s *ChannelService) Create(in domain.CreateChannelInput) (int64, error) {
	auditStatus := 0
	if in.AuditStatus != nil {
		if *in.AuditStatus >= 0 && *in.AuditStatus <= 2 {
			auditStatus = *in.AuditStatus
		}
	}
	return s.repo.Create(in, auditStatus)
}

func (s *ChannelService) Update(id uint, in domain.UpdateChannelInput) error {
	updateFields := []string{}
	args := []interface{}{}

	if in.ChannelType != "" {
		updateFields = append(updateFields, "channel_type = ?")
		args = append(args, in.ChannelType)
	}
	if in.Name != "" {
		updateFields = append(updateFields, "name = ?")
		args = append(args, in.Name)
	}
	if in.ApplyCondition != "" {
		updateFields = append(updateFields, "apply_condition = ?")
		args = append(args, in.ApplyCondition)
	}
	if in.ResponseTime != "" {
		updateFields = append(updateFields, "response_time = ?")
		args = append(args, in.ResponseTime)
	}
	if in.ContactPhone != "" {
		updateFields = append(updateFields, "contact_phone = ?")
		args = append(args, in.ContactPhone)
	}
	if in.ContactUrl != "" {
		updateFields = append(updateFields, "contact_url = ?")
		args = append(args, in.ContactUrl)
	}
	if in.HelpLetterTemplate != "" {
		updateFields = append(updateFields, "help_letter_template = ?")
		args = append(args, in.HelpLetterTemplate)
	}
	if in.CrowdfundingTemplate != "" {
		updateFields = append(updateFields, "crowdfunding_template = ?")
		args = append(args, in.CrowdfundingTemplate)
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

	if len(updateFields) == 0 {
		return domain.ErrNoUpdateFields
	}

	return s.repo.Update(id, updateFields, args)
}

func (s *ChannelService) Delete(id uint) error {
	rowsAffected, err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrChannelNotFound
	}
	return nil
}

func mapChannelRow(row repo.ChannelRow) domain.ChannelItem {
	return domain.ChannelItem{
		ID:                   row.ID,
		ChannelType:          row.ChannelType,
		Name:                 row.Name,
		ApplyCondition:       row.ApplyCondition,
		ResponseTime:         row.ResponseTime,
		ContactPhone:         row.ContactPhone,
		ContactUrl:           row.ContactUrl,
		HelpLetterTemplate:   row.HelpLetterTemplate,
		CrowdfundingTemplate: row.CrowdfundingTemplate,
		AuditStatus:          row.AuditStatus,
		RejectReason:         nullStringPtr(row.RejectReason),
		Sort:                 row.Sort,
		CreatedAt:            row.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:            row.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func BuildChannelListFilter(channelType, keyword, auditStatusStr, pageStr, pageSizeStr string) domain.ChannelListFilter {
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	page, pageSize = normalizePage(page, pageSize, 10)

	return domain.ChannelListFilter{
		ChannelType: channelType,
		Keyword:     keyword,
		AuditStatus: parseAuditStatusParam(auditStatusStr, -1),
		Page:        page,
		PageSize:    pageSize,
	}
}
