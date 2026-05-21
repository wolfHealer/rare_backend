package service

import (
	"strconv"
	"strings"

	"rare_backend/internal/module/resource/drug/domain"
	"rare_backend/internal/module/resource/drug/repo"
)

type ChannelService struct {
	repo *repo.ChannelRepo
}

func NewChannelService(r *repo.ChannelRepo) *ChannelService {
	return &ChannelService{repo: r}
}

func BuildChannelListFilter(keyword, provinceCode, cityCode, districtCode, channelType, delivery, auditStatusStr, isInsuranceSettleStr, pageStr, pageSizeStr string) domain.ChannelListFilter {
	filter := domain.ChannelListFilter{
		Keyword:           keyword,
		ProvinceCode:      provinceCode,
		CityCode:          cityCode,
		DistrictCode:      districtCode,
		ChannelType:       channelType,
		DeliveryScope:     delivery,
		IsInsuranceSettle: isInsuranceSettleStr,
		Page:              1,
		PageSize:          10,
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

func (s *ChannelService) List(filter domain.ChannelListFilter) (*domain.ChannelListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, domain.ErrCountList
	}

	list, err := s.repo.List(filter)
	if err != nil {
		return nil, domain.ErrQueryList
	}
	if list == nil {
		list = []map[string]interface{}{}
	}

	return &domain.ChannelListResult{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *ChannelService) GetByID(id uint) (*domain.ChannelDetailResponse, error) {
	row, err := s.repo.GetApprovedByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrChannelNotFound
		}
		return nil, err
	}

	drugRows, err := s.repo.ListDrugsByChannelID(id)
	if err != nil {
		drugRows = nil
	}

	drugs := make([]domain.ChannelDrugItem, 0, len(drugRows))
	for _, d := range drugRows {
		drugs = append(drugs, domain.ChannelDrugItem{
			ID:          d.ID,
			GenericName: d.GenericName,
			BrandName:   nullString(d.BrandName),
		})
	}
	if drugs == nil {
		drugs = []domain.ChannelDrugItem{}
	}

	descParts := []string{}
	if row.DeliveryScope != "" {
		descParts = append(descParts, row.DeliveryScope)
	}
	if row.IsInsurance == 1 {
		descParts = append(descParts, "支持医保")
	}
	if row.DeliveryCycle != "" {
		descParts = append(descParts, row.DeliveryCycle)
	}
	desc := ""
	if len(descParts) > 0 {
		desc = strings.Join(descParts, "，")
	}

	return &domain.ChannelDetailResponse{
		ID:            row.ID,
		Name:          row.Name,
		ChannelType:   row.ChannelType,
		ProvinceCode:  row.ProvinceCode,
		CityCode:      row.CityCode,
		DistrictCode:  nullString(row.DistrictCode),
		Address:       nullString(row.Address),
		Desc:          desc,
		ContactPhone:  nullString(row.ContactPhone),
		ContactURL:    nullString(row.ContactURL),
		Qualification: nullString(row.Qualification),
		DeliveryScope: row.DeliveryScope,
		AuditStatus:   row.AuditStatus,
		ProvinceName:  nullString(row.ProvinceName),
		CityName:      nullString(row.CityName),
		DistrictName:  nullString(row.DistrictName),
		Drugs:         drugs,
	}, nil
}

func (s *ChannelService) Create(input domain.CreateChannelInput) (int64, error) {
	return s.repo.Create(input)
}

func (s *ChannelService) Update(id uint, input domain.UpdateChannelInput) error {
	exists, err := s.repo.Exists(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return domain.ErrChannelNotFound
		}
		return err
	}
	if !exists {
		return domain.ErrChannelNotFound
	}

	fields := []string{}
	args := []interface{}{}

	if input.Name != nil {
		fields = append(fields, "name = ?")
		args = append(args, *input.Name)
	}
	if input.ChannelType != nil {
		fields = append(fields, "channel_type = ?")
		args = append(args, *input.ChannelType)
	}
	if input.ProvinceCode != nil {
		fields = append(fields, "province_code = ?")
		args = append(args, *input.ProvinceCode)
	}
	if input.CityCode != nil {
		fields = append(fields, "city_code = ?")
		args = append(args, *input.CityCode)
	}
	if input.DistrictCode != nil {
		fields = append(fields, "district_code = ?")
		args = append(args, *input.DistrictCode)
	}
	if input.Address != nil {
		fields = append(fields, "address = ?")
		args = append(args, *input.Address)
	}
	if input.ContactPhone != nil {
		fields = append(fields, "contact_phone = ?")
		args = append(args, *input.ContactPhone)
	}
	if input.ContactURL != nil {
		fields = append(fields, "contact_url = ?")
		args = append(args, *input.ContactURL)
	}
	if input.DeliveryScope != nil {
		fields = append(fields, "delivery_scope = ?")
		args = append(args, *input.DeliveryScope)
	}
	if input.DeliveryCycle != nil {
		fields = append(fields, "delivery_cycle = ?")
		args = append(args, *input.DeliveryCycle)
	}
	if input.IsInsuranceSettle != nil {
		fields = append(fields, "is_insurance_settle = ?")
		args = append(args, boolToInt(*input.IsInsuranceSettle))
	}
	if input.Qualification != nil {
		fields = append(fields, "qualification = ?")
		args = append(args, *input.Qualification)
	}
	if input.DrugID != nil {
		fields = append(fields, "drug_id = ?")
		args = append(args, *input.DrugID)
	}

	if len(fields) == 0 {
		return domain.ErrNoUpdateFields
	}
	return s.repo.Update(id, fields, args)
}

func (s *ChannelService) Delete(id uint) error {
	exists, err := s.repo.ApprovedExists(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return domain.ErrChannelNotFound
		}
		return err
	}
	if !exists {
		return domain.ErrChannelNotFound
	}
	return s.repo.SoftDelete(id)
}

func (s *ChannelService) Contact(id uint, input domain.ChannelContactInput) (*domain.ChannelContactResponse, error) {
	phone, url, _, err := s.repo.GetContact(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrChannelNotFound
		}
		return nil, err
	}

	resp := &domain.ChannelContactResponse{
		Phone:  phone,
		Wechat: "",
		Email:  "",
	}
	if input.ContactType == "url" {
		resp.Wechat = url
	}
	return resp, nil
}
