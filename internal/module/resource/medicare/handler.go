package medicare

import (
	"database/sql"
	"fmt"
	"net/http"
	"rare_backend/internal/pkg/db"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PolicyItem 政策项响应结构
type PolicyItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Region      string `json:"region"`
	RegionCode  string `json:"regionCode"`
	Date        string `json:"date"`
	PublishDate string `json:"publishDate"`
	Summary     string `json:"summary"`
	Category    string `json:"category"`
	FileUrl     string `json:"fileUrl"`
}

// RelatedPolicy 相关政策项
type RelatedPolicy struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

// truncateSummary 截断摘要文本
func truncateSummary(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	// 按字节截断，避免中文乱码
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

// PolicyDetailResponse 政策详情响应结构
type PolicyDetailResponse struct {
	ID                  uint            `json:"id"`
	Title               string          `json:"title"`
	Region              string          `json:"region"`
	RegionCode          string          `json:"regionCode"`
	PublishDate         string          `json:"publishDate"`
	EffectiveDate       string          `json:"effectiveDate"` // 新增
	Category            string          `json:"category"`
	Content             string          `json:"content"`
	FileUrl             string          `json:"fileUrl"`
	ReimburseRatio      string          `json:"reimburseRatio"`      // 新增
	ReimburseLimit      string          `json:"reimburseLimit"`      // 新增
	ReimburseProcess    string          `json:"reimburseProcess"`    // 新增
	ReimburseMaterial   string          `json:"reimburseMaterial"`   // 新增
	RemoteApplyTemplate string          `json:"remoteApplyTemplate"` // 新增
	RelatedPolicies     []RelatedPolicy `json:"relatedPolicies"`
}

// GetPolicyDetail 获取政策详情
func GetPolicyDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的政策 ID",
		})
		return
	}

	// 查询政策详情
	query := `
		SELECT id, disease_id, scope_level, province_code, city_code, province_name, city_name,
		       policy_title, policy_original, popular_interpret, reimburse_ratio, reimburse_limit, 
			   reimburse_process, reimburse_material, remote_apply_template,
		       is_latest, publish_date, effective_date, created_at, updated_at
		FROM medical_insurance_policy
		WHERE id = ? AND audit_status = 1
	`

	var policy struct {
		ID                  uint       `db:"id"`
		DiseaseID           int        `db:"disease_id"`
		ScopeLevel          int        `db:"scope_level"`
		ProvinceCode        string     `db:"province_code"`
		CityCode            string     `db:"city_code"`
		ProvinceName        string     `db:"province_name"`
		CityName            string     `db:"city_name"`
		PolicyTitle         string     `db:"policy_title"`
		PolicyOriginal      string     `db:"policy_original"`
		PopularInterpret    string     `db:"popular_interpret"`
		ReimburseRatio      string     `db:"reimburse_ratio"`
		ReimburseLimit      string     `db:"reimburse_limit"`
		ReimburseProcess    string     `db:"reimburse_process"`
		ReimburseMaterial   string     `db:"reimburse_material"`
		RemoteApplyTemplate string     `db:"remote_apply_template"`
		IsLatest            int        `db:"is_latest"`
		PublishDate         *time.Time `db:"publish_date"`
		EffectiveDate       *time.Time `db:"effective_date"`
		CreatedAt           time.Time  `db:"created_at"`
		UpdatedAt           time.Time  `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&policy.ID, &policy.DiseaseID, &policy.ScopeLevel,
		&policy.ProvinceCode, &policy.CityCode, &policy.ProvinceName, &policy.CityName,
		&policy.PolicyTitle, &policy.PolicyOriginal, &policy.PopularInterpret,
		&policy.ReimburseRatio, &policy.ReimburseLimit,
		&policy.ReimburseProcess, &policy.ReimburseMaterial, &policy.RemoteApplyTemplate,
		&policy.IsLatest, &policy.PublishDate, &policy.EffectiveDate,
		&policy.CreatedAt, &policy.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "政策不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询政策详情失败",
		})
		return
	}

	// 确定显示的地区名称
	regionDisplay := ""
	if policy.ScopeLevel == 1 {
		regionDisplay = "全国"
	} else if policy.ScopeLevel == 2 {
		regionDisplay = policy.ProvinceName
		if regionDisplay == "" {
			regionDisplay = policy.ProvinceCode
		}
	} else {
		regionDisplay = policy.ProvinceName
		if policy.CityName != "" {
			regionDisplay += " " + policy.CityName
		} else if policy.CityCode != "" {
			regionDisplay += " " + policy.CityCode
		}
		if regionDisplay == "" {
			regionDisplay = policy.ProvinceCode
		}
	}

	// 格式化日期
	publishDateStr := ""
	if policy.PublishDate != nil {
		publishDateStr = policy.PublishDate.Format("2006-01-02T15:04:05Z")
	}

	effectiveDateStr := ""
	if policy.EffectiveDate != nil {
		effectiveDateStr = policy.EffectiveDate.Format("2006-01-02T15:04:05Z")
	}

	// 查询相关政策（同省份或同疾病的其他政策）
	// 注意：关联条件也需调整为 code
	relatedQuery := `
		SELECT id, policy_title 
		FROM medical_insurance_policy 
		WHERE audit_status = 1 
		  AND id != ? 
		  AND (province_code = ? OR disease_id = ?)
		ORDER BY is_latest DESC, publish_date DESC 
		LIMIT 5
	`
	relatedRows, err := db.MySQL.Query(relatedQuery, policy.ID, policy.ProvinceCode, policy.DiseaseID)
	if err != nil {
		// 相关政策查询失败不影响主流程
		relatedRows = nil
	}

	var relatedPolicies []RelatedPolicy
	if relatedRows != nil {
		defer relatedRows.Close()
		for relatedRows.Next() {
			var related struct {
				ID    uint   `db:"id"`
				Title string `db:"policy_title"`
			}
			if err := relatedRows.Scan(&related.ID, &related.Title); err != nil {
				continue
			}
			relatedPolicies = append(relatedPolicies, RelatedPolicy{
				ID:    related.ID,
				Title: related.Title,
			})
		}
	}

	// 确保数组不为 null
	if relatedPolicies == nil {
		relatedPolicies = []RelatedPolicy{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": PolicyDetailResponse{
			ID:                  policy.ID,
			Title:               policy.PolicyTitle,
			Region:              regionDisplay,
			RegionCode:          policy.ProvinceCode,
			PublishDate:         publishDateStr,
			EffectiveDate:       effectiveDateStr,
			Content:             policy.PopularInterpret,
			FileUrl:             policy.PolicyOriginal,
			ReimburseRatio:      policy.ReimburseRatio,
			ReimburseLimit:      policy.ReimburseLimit,
			ReimburseProcess:    policy.ReimburseProcess,
			ReimburseMaterial:   policy.ReimburseMaterial,
			RemoteApplyTemplate: policy.RemoteApplyTemplate,
			RelatedPolicies:     relatedPolicies,
		},
	})
}

// MaterialItem 资料项
type MaterialItem struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	Size       string `json:"size"`
	UpdateTime string `json:"updateTime"`
}

// MaterialResponse 资料响应结构
type MaterialResponse struct {
	Materials []MaterialItem `json:"materials"`
}

// DownloadMaterials 下载资料（数据库查询版本）
func DownloadMaterials(c *gin.Context) {
	// 获取请求参数
	materialType := c.DefaultQuery("type", "")

	// 构建查询条件
	whereClause := "WHERE audit_status = 1"
	args := []interface{}{}

	if materialType != "" {
		// 根据类型映射到新表对应字段
		typeFieldMap := map[string]string{
			"flowchart": "reimburse_process",
			"guide":     "popular_interpret",
			"template":  "remote_apply_template",
			"checklist": "reimburse_material",
		}
		if field, ok := typeFieldMap[materialType]; ok {
			whereClause += " AND " + field + " != ''"
		}
	}

	// 查询资料
	query := `
		SELECT 
			reimburse_process as url_flowchart,
			reimburse_material as url_checklist,
			remote_apply_template as url_template,
			updated_at
		FROM medical_insurance_policy 
		` + whereClause + `
		ORDER BY is_latest DESC, updated_at DESC
		LIMIT 50
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询资料失败",
		})
		return
	}
	defer rows.Close()

	var materials []MaterialItem

	// 用于去重，避免相同URL重复添加
	urlSet := make(map[string]bool)

	addMaterial := func(name, mType, url string, updateTime time.Time) {
		if url != "" && !urlSet[url] {
			urlSet[url] = true
			materials = append(materials, MaterialItem{
				Name:       name,
				Type:       mType,
				URL:        url,
				Size:       "未知",
				UpdateTime: updateTime.Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	for rows.Next() {
		var row struct {
			UrlFlowchart string    `db:"url_flowchart"`
			UrlChecklist string    `db:"url_checklist"`
			UrlTemplate  string    `db:"url_template"`
			UpdateTime   time.Time `db:"updated_at"`
		}
		if err := rows.Scan(&row.UrlFlowchart, &row.UrlChecklist, &row.UrlTemplate, &row.UpdateTime); err != nil {
			continue
		}

		// 如果指定了类型，只添加该类型
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
			// 未指定类型，返回所有非空资料
			addMaterial("医保报销流程图解", "flowchart", row.UrlFlowchart, row.UpdateTime)
			addMaterial("报销材料清单说明", "checklist", row.UrlChecklist, row.UpdateTime)
			addMaterial("异地就医备案模板", "template", row.UrlTemplate, row.UpdateTime)
		}
	}

	// 确保数组不为 null
	if materials == nil {
		materials = []MaterialItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": MaterialResponse{
			Materials: materials,
		},
	})
}

// QueryMode 查询模式
type QueryMode string

const (
	ModeDefaultRecommend QueryMode = "default_recommend"
	ModeRegionOnly       QueryMode = "region_only"
	ModeDiseaseOnly      QueryMode = "disease_only"
	ModeRegionDisease    QueryMode = "region_disease"
)

// SelectedFilters 选中的筛选条件详情
type SelectedFilters struct {
	DiseaseID    int    `json:"disease_id"`
	DiseaseName  string `json:"disease_name,omitempty"`
	ProvinceCode string `json:"province_code"`
	CityCode     string `json:"city_code"`
	ProvinceName string `json:"province_name,omitempty"`
	CityName     string `json:"city_name,omitempty"`
}

// GroupedPolicies 分层级的政策列表
type GroupedPolicies struct {
	National []PolicyItem `json:"national"`
	Province []PolicyItem `json:"province"`
	City     []PolicyItem `json:"city"`
}

// NewPolicyListResponse 新的政策列表响应结构
type NewPolicyListResponse struct {
	QueryMode       QueryMode       `json:"query_mode"`
	SelectedFilters SelectedFilters `json:"selected_filters"`
	Tips            string          `json:"tips"`
	Policies        GroupedPolicies `json:"policies"`
	Total           int64           `json:"total"`
	Page            int             `json:"page"`
	PageSize        int             `json:"pageSize"`
}

// --- 辅助函数 ---

// getDiseaseNameByID 根据ID获取疾病名称
// TODO: 实际项目中应查询 disease_options 表或缓存
func getDiseaseNameByID(id int) string {
	if id == 1 {
		return "戈谢病"
	}
	// 这里可以接入数据库查询: SELECT name FROM disease_options WHERE id = ?
	return fmt.Sprintf("疾病ID-%d", id)
}

// getRegionNameByCode 根据代码获取地区名称
// TODO: 实际项目中应查询 region_dict 表或缓存
func getRegionNameByCode(code string) string {
	if code == "" {
		return ""
	}
	// 简单模拟，实际应查库
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
	return code // 兜底返回代码
}

// generateTips 生成用户提示语
func generateTips(mode QueryMode, filters SelectedFilters) string {
	switch mode {
	case ModeRegionDisease:
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

	case ModeRegionOnly:
		loc := filters.ProvinceName
		if filters.CityName != "" {
			loc = filters.ProvinceName + " " + filters.CityName
		} else if loc == "" {
			loc = filters.ProvinceCode
		}
		return fmt.Sprintf("已为您展示%s地区的医保政策", loc)

	case ModeDiseaseOnly:
		disease := filters.DiseaseName
		if disease == "" {
			disease = "指定病种"
		}
		return fmt.Sprintf("已为您展示全国及各地关于【%s】的医保政策", disease)

	default:
		return "为您推荐全国最新医保政策"
	}
}

// GetPolicyList 获取医保政策列表（重构版：支持分层级返回与智能推荐）
func GetPolicyList(c *gin.Context) {
	// 1. 获取请求参数
	provinceCode := c.DefaultQuery("province_code", "")
	cityCode := c.DefaultQuery("city_code", "")
	diseaseStr := c.DefaultQuery("disease_id", "") // 注意：参数名统一为 disease_id
	needLatestStr := c.DefaultQuery("need_latest_only", "1")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10") // 注意：参数名统一为 page_size

	// 参数解析
	diseaseID := 0
	if diseaseStr != "" {
		diseaseID, _ = strconv.Atoi(diseaseStr)
	}

	needLatest := true
	if needLatestStr == "0" || needLatestStr == "false" {
		needLatest = false
	}

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 2. 确定查询模式 (Query Mode)
	hasDisease := diseaseID != 0
	hasRegion := provinceCode != ""

	var mode QueryMode
	var selectedFilters SelectedFilters

	// 初始化筛选条件
	selectedFilters.DiseaseID = diseaseID
	selectedFilters.ProvinceCode = provinceCode
	selectedFilters.CityCode = cityCode

	// 获取名称用于展示和Tips
	if diseaseID != 0 {
		selectedFilters.DiseaseName = getDiseaseNameByID(diseaseID)
	}
	if provinceCode != "" {
		selectedFilters.ProvinceName = getRegionNameByCode(provinceCode)
		if cityCode != "" {
			selectedFilters.CityName = getRegionNameByCode(cityCode)
		}
	}

	// 判定模式
	if hasDisease && hasRegion {
		mode = ModeRegionDisease
	} else if hasDisease {
		mode = ModeDiseaseOnly
	} else if hasRegion {
		mode = ModeRegionOnly
	} else {
		mode = ModeDefaultRecommend
	}

	tips := generateTips(mode, selectedFilters)

	// 3. 构建 SQL 查询条件
	whereClause := "WHERE audit_status = 1"
	args := []interface{}{}

	// 基础过滤：是否只看最新
	if needLatest {
		whereClause += " AND is_latest = 1"
	}

	// 根据模式添加特定条件
	switch mode {
	case ModeRegionDisease:
		// 病种 + 地区：全国 OR (省匹配) OR (市匹配)
		whereClause += " AND disease_id = ?"
		args = append(args, diseaseID)
		whereClause += " AND (scope_level = 1 OR (scope_level = 2 AND province_code = ?) OR (scope_level = 3 AND city_code = ?))"
		args = append(args, provinceCode, cityCode)

	case ModeRegionOnly:
		// 仅地区：全国 OR (省匹配) OR (市匹配)
		whereClause += " AND (scope_level = 1 OR (scope_level = 2 AND province_code = ?) OR (scope_level = 3 AND city_code = ?))"
		args = append(args, provinceCode, cityCode)

	case ModeDiseaseOnly:
		// 仅病种：所有层级
		whereClause += " AND disease_id = ?"
		args = append(args, diseaseID)

	case ModeDefaultRecommend:
		// 默认推荐：仅全国最新
		whereClause += " AND scope_level = 1"
	}

	// 排序：优先按层级（全国1 > 省2 > 市3），再按发布时间倒序
	// 注意：scope_level ASC 会让 1 排在前面
	orderClause := "ORDER BY scope_level ASC, publish_date DESC, id DESC"

	// 4. 查询总数
	countQuery := "SELECT COUNT(*) FROM medical_insurance_policy " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 5. 查询列表数据
	listQuery := `
		SELECT id, disease_id, scope_level, province_code, city_code, province_name, city_name, 
		       policy_title, policy_original, popular_interpret, is_latest, publish_date, created_at
		FROM medical_insurance_policy
		` + whereClause + `
		` + orderClause + `
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询列表失败",
		})
		return
	}
	defer rows.Close()

	// 6. 数据处理与分组
	var nationalList []PolicyItem
	var provinceList []PolicyItem
	var cityList []PolicyItem

	for rows.Next() {
		var policy struct {
			ID               uint       `db:"id"`
			DiseaseID        int        `db:"disease_id"`
			ScopeLevel       int        `db:"scope_level"`
			ProvinceCode     string     `db:"province_code"`
			CityCode         string     `db:"city_code"`
			ProvinceName     string     `db:"province_name"`
			CityName         string     `db:"city_name"`
			PolicyTitle      string     `db:"policy_title"`
			PolicyOriginal   string     `db:"policy_original"`
			PopularInterpret string     `db:"popular_interpret"`
			IsLatest         int        `db:"is_latest"`
			PublishDate      *time.Time `db:"publish_date"`
			CreatedAt        time.Time  `db:"created_at"`
		}
		if err := rows.Scan(
			&policy.ID, &policy.DiseaseID, &policy.ScopeLevel,
			&policy.ProvinceCode, &policy.CityCode, &policy.ProvinceName, &policy.CityName,
			&policy.PolicyTitle, &policy.PolicyOriginal, &policy.PopularInterpret,
			&policy.IsLatest, &policy.PublishDate, &policy.CreatedAt,
		); err != nil {
			continue
		}

		// 转换为通用 Item 结构
		item := convertToPolicyItem(policy)

		// 按层级分组
		switch policy.ScopeLevel {
		case 1:
			nationalList = append(nationalList, item)
		case 2:
			provinceList = append(provinceList, item)
		case 3:
			cityList = append(cityList, item)
		default:
			// 其他层级（如区县）可根据需求归类，这里暂归入城市或忽略
			cityList = append(cityList, item)
		}
	}

	// 确保切片不为 nil，避免前端解析为 null
	if nationalList == nil {
		nationalList = []PolicyItem{}
	}
	if provinceList == nil {
		provinceList = []PolicyItem{}
	}
	if cityList == nil {
		cityList = []PolicyItem{}
	}

	// 7. 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": NewPolicyListResponse{
			QueryMode:       mode,
			SelectedFilters: selectedFilters,
			Tips:            tips,
			Policies: GroupedPolicies{
				National: nationalList,
				Province: provinceList,
				City:     cityList,
			},
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// convertToPolicyItem 将数据库行转换为 API 响应项 (提取公共逻辑)
func convertToPolicyItem(p struct {
	ID               uint       `db:"id"`
	DiseaseID        int        `db:"disease_id"`
	ScopeLevel       int        `db:"scope_level"`
	ProvinceCode     string     `db:"province_code"`
	CityCode         string     `db:"city_code"`
	ProvinceName     string     `db:"province_name"`
	CityName         string     `db:"city_name"`
	PolicyTitle      string     `db:"policy_title"`
	PolicyOriginal   string     `db:"policy_original"`
	PopularInterpret string     `db:"popular_interpret"`
	IsLatest         int        `db:"is_latest"`
	PublishDate      *time.Time `db:"publish_date"`
	CreatedAt        time.Time  `db:"created_at"`
}) PolicyItem {

	// 确定显示的地区名称
	regionDisplay := ""
	if p.ScopeLevel == 1 {
		regionDisplay = "全国"
	} else if p.ScopeLevel == 2 {
		regionDisplay = p.ProvinceName
		if regionDisplay == "" {
			regionDisplay = p.ProvinceCode
		}
	} else {
		regionDisplay = p.ProvinceName
		if p.CityName != "" {
			regionDisplay += " " + p.CityName
		} else if p.CityCode != "" {
			regionDisplay += " " + p.CityCode
		}
		if regionDisplay == "" {
			regionDisplay = p.ProvinceCode
		}
	}

	// 确定日期
	dateStr := p.CreatedAt.Format("2006-01")
	if p.PublishDate != nil {
		dateStr = p.PublishDate.Format("2006-01")
	}

	publishDateStr := p.CreatedAt.Format("2006-01-02T15:04:05Z")
	if p.PublishDate != nil {
		publishDateStr = p.PublishDate.Format("2006-01-02T15:04:05Z")
	}

	return PolicyItem{
		ID:          p.ID,
		Title:       p.PolicyTitle,
		Region:      regionDisplay,
		RegionCode:  p.ProvinceCode,
		Date:        dateStr,
		PublishDate: publishDateStr,
		Summary:     truncateSummary(p.PopularInterpret, 50),
		FileUrl:     p.PolicyOriginal,
	}
}
