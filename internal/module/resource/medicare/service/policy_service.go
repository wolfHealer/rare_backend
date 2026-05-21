package service

import (
	"time"

	"rare_backend/internal/module/resource/medicare/domain"
	"rare_backend/internal/module/resource/medicare/repo"
)

type PolicyService struct {
	repo *repo.PolicyRepo
}

func NewPolicyService(r *repo.PolicyRepo) *PolicyService {
	return &PolicyService{repo: r}
}

func (s *PolicyService) List(filter domain.PolicyListFilter) (*domain.PolicyListResult, error) {
	mode, selected := determineQueryMode(filter)

	total, err := s.repo.CountList(filter, mode)
	if err != nil {
		return nil, domain.ErrCountList
	}

	rows, err := s.repo.List(filter, mode)
	if err != nil {
		return nil, domain.ErrQueryList
	}

	var nationalList, provinceList, cityList []domain.PolicyItem
	for _, row := range rows {
		item := convertToPolicyItem(row)
		switch row.ScopeLevel {
		case 1:
			nationalList = append(nationalList, item)
		case 2:
			provinceList = append(provinceList, item)
		case 3:
			cityList = append(cityList, item)
		default:
			cityList = append(cityList, item)
		}
	}

	if nationalList == nil {
		nationalList = []domain.PolicyItem{}
	}
	if provinceList == nil {
		provinceList = []domain.PolicyItem{}
	}
	if cityList == nil {
		cityList = []domain.PolicyItem{}
	}

	return &domain.PolicyListResult{
		QueryMode:       mode,
		SelectedFilters: selected,
		Tips:            generateTips(mode, selected),
		Policies: domain.GroupedPolicies{
			National: nationalList,
			Province: provinceList,
			City:     cityList,
		},
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *PolicyService) GetByID(id uint64) (*domain.PolicyDetailResult, error) {
	row, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrPolicyNotFound
		}
		return nil, domain.ErrQueryDetail
	}

	related, _ := s.repo.ListRelated(uint64(row.ID), row.ProvinceCode, row.DiseaseID)
	if related == nil {
		related = []domain.RelatedPolicy{}
	}

	return &domain.PolicyDetailResult{
		ID:                  row.ID,
		Title:               row.PolicyTitle,
		Region:              formatRegionDisplay(row.ScopeLevel, row.ProvinceCode, row.ProvinceName, row.CityCode, row.CityName),
		RegionCode:          row.ProvinceCode,
		PublishDate:         formatOptionalTime(row.PublishDate),
		EffectiveDate:       formatOptionalTime(row.EffectiveDate),
		Content:             row.PopularInterpret,
		FileUrl:             row.PolicyOriginal,
		ReimburseRatio:      row.ReimburseRatio,
		ReimburseLimit:      row.ReimburseLimit,
		ReimburseProcess:    row.ReimburseProcess,
		ReimburseMaterial:   row.ReimburseMaterial,
		RemoteApplyTemplate: row.RemoteApplyTemplate,
		RelatedPolicies:     related,
	}, nil
}

func (s *PolicyService) ListMaterials(materialType string) ([]domain.MaterialItem, error) {
	rows, err := s.repo.ListMaterials(materialType)
	if err != nil {
		return nil, domain.ErrQueryMaterials
	}

	urlSet := make(map[string]bool)
	var materials []domain.MaterialItem

	addMaterial := func(name, mType, url string, updateTime time.Time) {
		if url != "" && !urlSet[url] {
			urlSet[url] = true
			materials = append(materials, domain.MaterialItem{
				Name:       name,
				Type:       mType,
				URL:        url,
				Size:       "未知",
				UpdateTime: updateTime.Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	for _, row := range rows {
		if materialType != "" {
			switch materialType {
			case "flowchart":
				addMaterial("医保报销流程图解", "flowchart", row.UrlFlowchart, row.UpdateTime)
			case "checklist":
				addMaterial("报销材料清单说明", "checklist", row.UrlChecklist, row.UpdateTime)
			case "template":
				addMaterial("异地就医备案模板", "template", row.UrlTemplate, row.UpdateTime)
			}
		} else {
			addMaterial("医保报销流程图解", "flowchart", row.UrlFlowchart, row.UpdateTime)
			addMaterial("报销材料清单说明", "checklist", row.UrlChecklist, row.UpdateTime)
			addMaterial("异地就医备案模板", "template", row.UrlTemplate, row.UpdateTime)
		}
	}

	if materials == nil {
		materials = []domain.MaterialItem{}
	}
	return materials, nil
}
