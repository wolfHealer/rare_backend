package service

import (
	"fmt"
	"strconv"
	"time"

	"rare_backend/internal/module/resource/medicare/domain"
)

func truncateSummary(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

func formatRegionDisplay(scopeLevel int, provinceCode, provinceName, cityCode, cityName string) string {
	switch scopeLevel {
	case 1:
		return "全国"
	case 2:
		if provinceName != "" {
			return provinceName
		}
		return provinceCode
	default:
		regionDisplay := provinceName
		if cityName != "" {
			regionDisplay += " " + cityName
		} else if cityCode != "" {
			regionDisplay += " " + cityCode
		}
		if regionDisplay == "" {
			return provinceCode
		}
		return regionDisplay
	}
}

func convertToPolicyItem(row domain.PolicyRow) domain.PolicyItem {
	dateStr := row.CreatedAt.Format("2006-01")
	if row.PublishDate != nil {
		dateStr = row.PublishDate.Format("2006-01")
	}

	publishDateStr := row.CreatedAt.Format("2006-01-02T15:04:05Z")
	if row.PublishDate != nil {
		publishDateStr = row.PublishDate.Format("2006-01-02T15:04:05Z")
	}

	return domain.PolicyItem{
		ID:          row.ID,
		Title:       row.PolicyTitle,
		Region:      formatRegionDisplay(row.ScopeLevel, row.ProvinceCode, row.ProvinceName, row.CityCode, row.CityName),
		RegionCode:  row.ProvinceCode,
		Date:        dateStr,
		PublishDate: publishDateStr,
		Summary:     truncateSummary(row.PopularInterpret, 50),
		FileUrl:     row.PolicyOriginal,
	}
}

func getDiseaseNameByID(id int) string {
	if id == 1 {
		return "戈谢病"
	}
	return fmt.Sprintf("疾病ID-%d", id)
}

func getRegionNameByCode(code string) string {
	if code == "" {
		return ""
	}
	if len(code) == 6 {
		if code[:2] == "32" {
			if code == "320000" {
				return "江苏省"
			}
			if code == "320100" {
				return "南京市"
			}
		}
	}
	return code
}

func generateTips(mode domain.QueryMode, filters domain.SelectedFilters) string {
	switch mode {
	case domain.ModeRegionDisease:
		loc := filters.ProvinceName
		if filters.CityName != "" {
			loc = filters.ProvinceName + " " + filters.CityName
		} else if loc == "" {
			loc = filters.ProvinceCode
		}
		disease := filters.DiseaseName
		if disease == "" {
			disease = "指定病种"
		}
		return fmt.Sprintf("已为您展示%s关于【%s】的医保政策", loc, disease)
	case domain.ModeRegionOnly:
		loc := filters.ProvinceName
		if filters.CityName != "" {
			loc = filters.ProvinceName + " " + filters.CityName
		} else if loc == "" {
			loc = filters.ProvinceCode
		}
		return fmt.Sprintf("已为您展示%s地区的医保政策", loc)
	case domain.ModeDiseaseOnly:
		disease := filters.DiseaseName
		if disease == "" {
			disease = "指定病种"
		}
		return fmt.Sprintf("已为您展示全国及各地关于【%s】的医保政策", disease)
	default:
		return "为您推荐全国最新医保政策"
	}
}

func determineQueryMode(filter domain.PolicyListFilter) (domain.QueryMode, domain.SelectedFilters) {
	hasDisease := filter.DiseaseID != 0
	hasRegion := filter.ProvinceCode != ""

	selected := domain.SelectedFilters{
		DiseaseID:    filter.DiseaseID,
		ProvinceCode: filter.ProvinceCode,
		CityCode:     filter.CityCode,
	}
	if filter.DiseaseID != 0 {
		selected.DiseaseName = getDiseaseNameByID(filter.DiseaseID)
	}
	if filter.ProvinceCode != "" {
		selected.ProvinceName = getRegionNameByCode(filter.ProvinceCode)
		if filter.CityCode != "" {
			selected.CityName = getRegionNameByCode(filter.CityCode)
		}
	}

	var mode domain.QueryMode
	switch {
	case hasDisease && hasRegion:
		mode = domain.ModeRegionDisease
	case hasDisease:
		mode = domain.ModeDiseaseOnly
	case hasRegion:
		mode = domain.ModeRegionOnly
	default:
		mode = domain.ModeDefaultRecommend
	}
	return mode, selected
}

func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z")
}

func BuildPolicyListFilter(provinceCode, cityCode, diseaseStr, needLatestStr, pageStr, pageSizeStr string) domain.PolicyListFilter {
	filter := domain.PolicyListFilter{
		ProvinceCode: provinceCode,
		CityCode:     cityCode,
		NeedLatest:   true,
		Page:         1,
		PageSize:     10,
	}
	if diseaseStr != "" {
		filter.DiseaseID, _ = strconv.Atoi(diseaseStr)
	}
	if needLatestStr == "0" || needLatestStr == "false" {
		filter.NeedLatest = false
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
