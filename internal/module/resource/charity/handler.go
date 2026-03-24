package charity

import (
	"database/sql"
	"net/http"
	"rare_backend/internal/pkg/db"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ProjectItem 救助项目项响应结构
type ProjectItem struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	Org        string `json:"org"`
	Desc       string `json:"desc"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	Disease    int    `json:"disease"`
	Amount     string `json:"amount"`
	Difficulty string `json:"difficulty"`
}

// ProjectListResponse 列表响应结构
type ProjectListResponse struct {
	List     []ProjectItem `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

// ProjectDetailResponse 详情响应结构
type ProjectDetailResponse struct {
	ID           uint       `json:"id"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	Requirements string     `json:"requirements"`
	Deadline     *time.Time `json:"deadline"`
	Contact      string     `json:"contact"`
	ApplyForm    string     `json:"applyForm"`
	ApplyGuide   string     `json:"applyGuide"`
	MaterialList string     `json:"materialList"`
}

// ProjectOptionsResponse 筛选选项响应
type ProjectOptionsResponse struct {
	Types        []OptionItem `json:"types"`
	Diseases     []OptionItem `json:"diseases"`
	Amounts      []OptionItem `json:"amounts"`
	Difficulties []OptionItem `json:"difficulties"`
}

// OptionItem 选项项
type OptionItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// ListProjects 获取救助项目列表
func ListProjects(c *gin.Context) {
	// 获取请求参数
	typeFilter := c.DefaultQuery("type", "")
	diseaseStr := c.DefaultQuery("disease", "")
	difficulty := c.DefaultQuery("difficulty", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	disease := 0
	if diseaseStr != "" {
		disease, _ = strconv.Atoi(diseaseStr)
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

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}

	if typeFilter != "" {
		whereClause += " AND relief_type = ?"
		args = append(args, typeFilter)
	}
	if disease != 0 {
		whereClause += " AND disease_value = ?"
		args = append(args, disease)
	}
	if difficulty != "" {
		whereClause += " AND apply_difficulty = ?"
		args = append(args, difficulty)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM relief_projects " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 查询列表
	listQuery := `
		SELECT id, name, organizer, apply_condition, is_audit,
		       relief_type, disease_value, relief_standard, apply_difficulty
		FROM relief_projects
		` + whereClause + `
		ORDER BY sort DESC, id DESC
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

	var list []ProjectItem
	for rows.Next() {
		var project struct {
			ID              uint   `db:"id"`
			Name            string `db:"name"`
			Organizer       string `db:"organizer"`
			ApplyCondition  string `db:"apply_condition"`
			IsAudit         int    `db:"is_audit"`
			ReliefType      string `db:"relief_type"`
			DiseaseValue    int    `db:"disease_value"`
			ReliefStandard  string `db:"relief_standard"`
			ApplyDifficulty string `db:"apply_difficulty"`
		}
		if err := rows.Scan(
			&project.ID, &project.Name, &project.Organizer, &project.ApplyCondition,
			&project.IsAudit, &project.ReliefType, &project.DiseaseValue,
			&project.ReliefStandard, &project.ApplyDifficulty,
		); err != nil {
			continue
		}

		// 转换状态
		status := "closed"
		if project.IsAudit == 1 {
			status = "open"
		}

		list = append(list, ProjectItem{
			ID:         project.ID,
			Title:      project.Name,
			Org:        project.Organizer,
			Desc:       project.ApplyCondition,
			Status:     status,
			Type:       project.ReliefType,
			Disease:    project.DiseaseValue,
			Amount:     project.ReliefStandard,
			Difficulty: project.ApplyDifficulty,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": ProjectListResponse{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// GetProjectDetail 获取项目详情
func GetProjectDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目 ID",
		})
		return
	}

	query := `
		SELECT id, name, apply_process, apply_condition, apply_deadline,
		       contact, apply_form, apply_guide, material_list
		FROM relief_projects
		WHERE id = ? AND is_audit = 1
	`

	var project struct {
		ID             uint       `db:"id"`
		Name           string     `db:"name"`
		ApplyProcess   string     `db:"apply_process"`
		ApplyCondition string     `db:"apply_condition"`
		ApplyDeadline  *time.Time `db:"apply_deadline"`
		Contact        string     `db:"contact"`
		ApplyForm      string     `db:"apply_form"`
		ApplyGuide     string     `db:"apply_guide"`
		MaterialList   string     `db:"material_list"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&project.ID, &project.Name, &project.ApplyProcess, &project.ApplyCondition,
		&project.ApplyDeadline, &project.Contact, &project.ApplyForm,
		&project.ApplyGuide, &project.MaterialList,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "项目不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询项目详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": ProjectDetailResponse{
			ID:           project.ID,
			Title:        project.Name,
			Content:      project.ApplyProcess,
			Requirements: project.ApplyCondition,
			Deadline:     project.ApplyDeadline,
			Contact:      project.Contact,
			ApplyForm:    project.ApplyForm,
			ApplyGuide:   project.ApplyGuide,
			MaterialList: project.MaterialList,
		},
	})
}

// GetFilters 获取筛选选项
func GetFilters(c *gin.Context) {
	data := ProjectOptionsResponse{
		Types: []OptionItem{
			{Text: "医疗费用救助", Value: "medical"},
			{Text: "生活补助", Value: "living"},
			{Text: "康复补贴", Value: "rehab"},
			{Text: "药品救助", Value: "drug"},
		},
		Diseases: []OptionItem{
			{Text: "渐冻症", Value: "als"},
			{Text: "血友病", Value: "hemophilia"},
			{Text: "罕见病", Value: "rare"},
		},
		Amounts: []OptionItem{
			{Text: "1-5 万", Value: "1w-5w"},
			{Text: "5-10 万", Value: "5w-10w"},
			{Text: "10 万以上", Value: "10w+"},
		},
		Difficulties: []OptionItem{
			{Text: "简单", Value: "easy"},
			{Text: "中等", Value: "medium"},
			{Text: "复杂", Value: "hard"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}

// PolicyItem 政策项响应结构
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

// PolicyListResponse 政策列表响应结构
type PolicyListResponse struct {
	List     []PolicyItem `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
}

// RelatedPolicy 相关政策项
type RelatedPolicy struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

// PolicyDetailResponse 政策详情响应结构
type PolicyDetailResponse struct {
	ID              uint            `json:"id"`
	Title           string          `json:"title"`
	Region          string          `json:"region"`
	RegionCode      string          `json:"regionCode"`
	PublishDate     string          `json:"publishDate"`
	Category        string          `json:"category"`
	Content         string          `json:"content"`
	FileUrl         string          `json:"fileUrl"`
	RelatedPolicies []RelatedPolicy `json:"relatedPolicies"`
}

// GetPolicyList 获取医保政策列表
func GetPolicyList(c *gin.Context) {
	// 获取请求参数
	region := c.DefaultQuery("region", "")
	diseaseStr := c.DefaultQuery("disease", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	disease := 0
	if diseaseStr != "" {
		disease, _ = strconv.Atoi(diseaseStr)
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

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}

	if region != "" {
		whereClause += " AND region = ?"
		args = append(args, region)
	}
	if disease != 0 {
		whereClause += " AND disease_value = ?"
		args = append(args, disease)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM medical_insurance_policies " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 查询列表
	listQuery := `
		SELECT id, disease_value, region, policy_title, policy_original,
		       popular_interpret, is_update, created_at
		FROM medical_insurance_policies
		` + whereClause + `
		ORDER BY is_update DESC, created_at DESC
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

	var list []PolicyItem
	for rows.Next() {
		var policy struct {
			ID               uint      `db:"id"`
			DiseaseValue     int       `db:"disease_value"`
			Region           string    `db:"region"`
			PolicyTitle      string    `db:"policy_title"`
			PolicyOriginal   string    `db:"policy_original"`
			PopularInterpret string    `db:"popular_interpret"`
			IsUpdate         int       `db:"is_update"`
			CreatedAt        time.Time `db:"created_at"`
		}
		if err := rows.Scan(
			&policy.ID, &policy.DiseaseValue, &policy.Region, &policy.PolicyTitle,
			&policy.PolicyOriginal, &policy.PopularInterpret, &policy.IsUpdate, &policy.CreatedAt,
		); err != nil {
			continue
		}

		// 数据转换
		list = append(list, PolicyItem{
			ID:          policy.ID,
			Title:       policy.PolicyTitle,
			Region:      policy.Region,
			RegionCode:  convertRegionToCode(policy.Region),
			Date:        policy.CreatedAt.Format("2006-01"),
			PublishDate: policy.CreatedAt.Format("2006-01-02T15:04:05Z"),
			Summary:     truncateSummary(policy.PopularInterpret, 50),
			Category:    convertDiseaseToCategory(policy.DiseaseValue),
			FileUrl:     policy.PolicyOriginal,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": PolicyListResponse{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// convertDiseaseToCategory 疾病值转分类
func convertDiseaseToCategory(diseaseValue int) string {
	// 根据实际 disease_options 表数据调整映射关系
	categoryMap := map[int]string{
		1: "医保报销",
		2: "门诊特殊病",
		3: "大病保险",
		4: "医疗救助",
		5: "药品保障",
	}
	if category, ok := categoryMap[diseaseValue]; ok {
		return category
	}
	return "其他政策"
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

// GetPolicyDetail 获取政策详情
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
		SELECT id, disease_value, region, policy_title, policy_original,
		       popular_interpret, is_update, created_at, updated_at
		FROM medical_insurance_policies
		WHERE id = ? AND is_audit = 1
	`

	var policy struct {
		ID               uint      `db:"id"`
		DiseaseValue     int       `db:"disease_value"`
		Region           string    `db:"region"`
		PolicyTitle      string    `db:"policy_title"`
		PolicyOriginal   string    `db:"policy_original"`
		PopularInterpret string    `db:"popular_interpret"`
		IsUpdate         int       `db:"is_update"`
		CreatedAt        time.Time `db:"created_at"`
		UpdatedAt        time.Time `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&policy.ID, &policy.DiseaseValue, &policy.Region, &policy.PolicyTitle,
		&policy.PolicyOriginal, &policy.PopularInterpret, &policy.IsUpdate,
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

	// 查询相关政策（同地区或同疾病的其他政策）
	relatedQuery := `
		SELECT id, policy_title 
		FROM medical_insurance_policies 
		WHERE is_audit = 1 
		  AND id != ? 
		  AND (region = ? OR disease_value = ?)
		ORDER BY is_update DESC, created_at DESC 
		LIMIT 5
	`
	relatedRows, err := db.MySQL.Query(relatedQuery, policy.ID, policy.Region, policy.DiseaseValue)
	if err != nil {
		// 相关政策查询失败不影响主流程，记录日志即可
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
			ID:              policy.ID,
			Title:           policy.PolicyTitle,
			Region:          policy.Region,
			RegionCode:      convertRegionToCode(policy.Region),
			PublishDate:     policy.CreatedAt.Format("2006-01-02T15:04:05Z"),
			Category:        convertDiseaseToCategory(policy.DiseaseValue),
			Content:         policy.PopularInterpret,
			FileUrl:         policy.PolicyOriginal,
			RelatedPolicies: relatedPolicies,
		},
	})
}

// RegionItem 地区选项项
type RegionItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// RegionResponse 地区选项响应结构
type RegionResponse struct {
	Regions []RegionItem `json:"regions"`
}

// GetRegions 获取地区筛选选项
func GetRegions(c *gin.Context) {
	// 方案一：从数据库动态查询地区（推荐）
	query := `
		SELECT DISTINCT region 
		FROM medical_insurance_policies 
		WHERE is_audit = 1 
		ORDER BY region ASC
	`

	rows, err := db.MySQL.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询地区选项失败",
		})
		return
	}
	defer rows.Close()

	var regions []RegionItem
	for rows.Next() {
		var region string
		if err := rows.Scan(&region); err != nil {
			continue
		}
		regions = append(regions, RegionItem{
			Text:  region,
			Value: convertRegionToCode(region),
		})
	}

	// 确保全国选项在前
	regions = sortRegions(regions)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": RegionResponse{
			Regions: regions,
		},
	})
}

// convertRegionToCode 地区名称转地区代码
func convertRegionToCode(region string) string {
	regionMap := map[string]string{
		"全国": "all",
		"北京": "bj",
		"上海": "sh",
		"广东": "gd",
		"广州": "gz",
		"深圳": "sz",
		"浙江": "zj",
		"江苏": "js",
		"四川": "sc",
		"湖北": "hb",
		"山东": "sd",
		"河南": "hn",
		"福建": "fj",
		"湖南": "hun",
		"安徽": "ah",
		"辽宁": "ln",
		"陕西": "sx",
		"重庆": "cq",
		"天津": "tj",
	}
	if code, ok := regionMap[region]; ok {
		return code
	}
	// 默认返回拼音首字母或其他标识
	return "other"
}

// sortRegions 排序地区，确保全国在前
func sortRegions(regions []RegionItem) []RegionItem {
	if len(regions) == 0 {
		return regions
	}

	// 查找全国选项
	var nationalIndex = -1
	for i, r := range regions {
		if r.Text == "全国" || r.Value == "all" {
			nationalIndex = i
			break
		}
	}

	// 如果全国不在第一位，将其移动到第一位
	if nationalIndex > 0 {
		national := regions[nationalIndex]
		regions = append(regions[:nationalIndex], regions[nationalIndex+1:]...)
		regions = append([]RegionItem{national}, regions...)
	}

	return regions
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
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}

	if materialType != "" {
		// 根据类型映射到对应字段
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
		SELECT DISTINCT 
			CASE 
				WHEN reimburse_process != '' THEN 'flowchart'
				WHEN remote_apply_template != '' THEN 'template'
				WHEN reimburse_material != '' THEN 'checklist'
			END AS type,
			CASE 
				WHEN reimburse_process != '' THEN '医保报销流程图解'
				WHEN remote_apply_template != '' THEN '异地就医备案模板'
				WHEN reimburse_material != '' THEN '报销材料清单说明'
			END AS name,
			CASE 
				WHEN reimburse_process != '' THEN reimburse_process
				WHEN remote_apply_template != '' THEN remote_apply_template
				WHEN reimburse_material != '' THEN reimburse_material
			END AS url,
			updated_at
		FROM medical_insurance_policies 
		` + whereClause + `
		ORDER BY updated_at DESC
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
	for rows.Next() {
		var material struct {
			Type       string    `db:"type"`
			Name       string    `db:"name"`
			URL        string    `db:"url"`
			UpdateTime time.Time `db:"updated_at"`
		}
		if err := rows.Scan(&material.Type, &material.Name, &material.URL, &material.UpdateTime); err != nil {
			continue
		}
		materials = append(materials, MaterialItem{
			Name:       material.Name,
			Type:       material.Type,
			URL:        material.URL,
			Size:       "未知", // 数据库未存储文件大小，可后续扩展
			UpdateTime: material.UpdateTime.Format("2006-01-02T15:04:05Z"),
		})
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

// ChannelItem 求助渠道项
type ChannelItem struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Desc         string `json:"desc"`
	Type         string `json:"type"`
	ContactValue string `json:"contactValue"`
	ServiceTime  string `json:"serviceTime"`
	Available    bool   `json:"available"`
}

// ChannelResponse 渠道响应结构
type ChannelResponse struct {
	Channels []ChannelItem `json:"channels"`
}

// TemplateItem 模板项
type TemplateItem struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	Size       string `json:"size"`
	UpdateTime string `json:"updateTime"`
}

// TemplateResponse 模板响应结构
type TemplateResponse struct {
	Templates []TemplateItem `json:"templates"`
}

// SubmitHelpRequest 提交求助请求体
type SubmitHelpRequest struct {
	ChannelID   uint   `json:"channelId"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// SubmitHelpResponse 提交求助响应结构
type SubmitHelpResponse struct {
	RequestID            string `json:"requestId"`
	Status               string `json:"status"`
	ExpectedResponseTime string `json:"expectedResponseTime"`
}

// GetChannels 获取求助渠道列表
func GetChannels(c *gin.Context) {
	// 获取请求参数
	channelType := c.DefaultQuery("type", "")

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}

	if channelType != "" {
		whereClause += " AND channel_type = ?"
		args = append(args, channelType)
	}

	// 查询渠道列表
	query := `
		SELECT id, channel_type, name, apply_condition, response_time, contact
		FROM help_channels 
		` + whereClause + `
		ORDER BY sort DESC, id ASC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询渠道列表失败",
		})
		return
	}
	defer rows.Close()

	var channels []ChannelItem
	for rows.Next() {
		var channel struct {
			ID             uint   `db:"id"`
			ChannelType    string `db:"channel_type"`
			Name           string `db:"name"`
			ApplyCondition string `db:"apply_condition"`
			ResponseTime   string `db:"response_time"`
			Contact        string `db:"contact"`
		}
		if err := rows.Scan(
			&channel.ID, &channel.ChannelType, &channel.Name,
			&channel.ApplyCondition, &channel.ResponseTime, &channel.Contact,
		); err != nil {
			continue
		}

		// 转换渠道类型
		channelType := convertChannelType(channel.ChannelType)

		// 提取联系方式
		contactValue := extractContactValue(channel.Contact, channelType)

		channels = append(channels, ChannelItem{
			ID:           channel.ID,
			Name:         channel.Name,
			Desc:         "响应时间：" + channel.ResponseTime,
			Type:         channelType,
			ContactValue: contactValue,
			ServiceTime:  channel.ResponseTime,
			Available:    true,
		})
	}

	// 确保数组不为 null
	if channels == nil {
		channels = []ChannelItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": ChannelResponse{
			Channels: channels,
		},
	})
}

// GetTemplates 获取求助模板列表
func GetTemplates(c *gin.Context) {
	// 获取请求参数
	templateType := c.DefaultQuery("type", "")

	// 定义模板配置
	allTemplates := []TemplateItem{
		{
			Name:       "通用求助信模板",
			Type:       "letter",
			URL:        "https://example.com/templates/help_letter.docx",
			Size:       "50KB",
			UpdateTime: "2024-01-10T00:00:00Z",
		},
		{
			Name:       "医疗救助申请表",
			Type:       "application",
			URL:        "https://example.com/templates/medical_application.docx",
			Size:       "80KB",
			UpdateTime: "2024-01-10T00:00:00Z",
		},
		{
			Name:       "紧急求助信模板",
			Type:       "urgent_letter",
			URL:        "https://example.com/templates/urgent_letter.docx",
			Size:       "45KB",
			UpdateTime: "2024-01-08T00:00:00Z",
		},
		{
			Name:       "众筹求助模板",
			Type:       "crowdfunding",
			URL:        "https://example.com/templates/crowdfunding.docx",
			Size:       "60KB",
			UpdateTime: "2024-01-05T00:00:00Z",
		},
	}

	// 根据类型筛选
	var templates []TemplateItem
	if templateType != "" {
		for _, t := range allTemplates {
			if t.Type == templateType {
				templates = append(templates, t)
			}
		}
	} else {
		templates = allTemplates
	}

	// 确保数组不为 null
	if templates == nil {
		templates = []TemplateItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": TemplateResponse{
			Templates: templates,
		},
	})
}

// SubmitHelp 提交求助请求
func SubmitHelp(c *gin.Context) {
	var req SubmitHelpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 参数校验
	if req.ChannelID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "渠道 ID 不能为空",
		})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "姓名不能为空",
		})
		return
	}
	if req.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "手机号不能为空",
		})
		return
	}
	if req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "求助描述不能为空",
		})
		return
	}

	// 生成请求 ID
	requestID := generateRequestID()

	// 插入求助请求记录（需要创建 help_requests 表）
	insertQuery := `
		INSERT INTO help_requests 
		(channel_id, name, phone, help_type, description, request_id, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, 'submitted', NOW())
	`
	_, err := db.MySQL.Exec(insertQuery,
		req.ChannelID, req.Name, req.Phone, req.Type, req.Description, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交求助请求失败",
		})
		return
	}

	// 计算预期响应时间（根据渠道类型）
	expectedTime := calculateExpectedResponseTime(req.ChannelID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "求助请求已提交",
		"data": SubmitHelpResponse{
			RequestID:            requestID,
			Status:               "submitted",
			ExpectedResponseTime: expectedTime,
		},
	})
}

// convertChannelType 转换渠道类型
func convertChannelType(channelType string) string {
	typeMap := map[string]string{
		"紧急求助": "urgent",
		"众筹求助": "crowdfunding",
		"医疗救助": "medical",
		"生活补助": "living",
	}
	if t, ok := typeMap[channelType]; ok {
		return t
	}
	return "other"
}

// extractContactValue 提取联系方式
func extractContactValue(contact string, channelType string) string {
	// 根据渠道类型提取对应的联系方式
	// 这里简化处理，实际可根据 contact 字段格式解析
	return contact
}

// generateRequestID 生成请求 ID
func generateRequestID() string {
	timestamp := time.Now().Format("20060102150405")
	// 可添加随机数或 UUID 确保唯一性
	return "REQ" + timestamp
}

// calculateExpectedResponseTime 计算预期响应时间
func calculateExpectedResponseTime(channelID uint) string {
	// 根据渠道 ID 查询响应时间，这里简化处理
	// 实际可查询 help_channels 表的 response_time 字段
	return time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04:05Z")
}

// ChannelTemplate 渠道模板项
type ChannelTemplate struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ChannelDetailResponse 渠道详情响应结构
type ChannelDetailResponse struct {
	ID             uint              `json:"id"`
	Name           string            `json:"name"`
	Type           string            `json:"type"`
	TypeName       string            `json:"typeName"`
	Desc           string            `json:"desc"`
	ApplyCondition string            `json:"applyCondition"`
	ResponseTime   string            `json:"responseTime"`
	ContactValue   string            `json:"contactValue"`
	ServiceTime    string            `json:"serviceTime"`
	Templates      []ChannelTemplate `json:"templates"`
	Available      bool              `json:"available"`
	PublishDate    string            `json:"publishDate"`
	UpdateTime     string            `json:"updateTime"`
}

// GetChannelDetail 获取求助渠道详情
func GetChannelDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道 ID",
		})
		return
	}

	// 查询渠道详情
	query := `
		SELECT id, channel_type, name, apply_condition, response_time, 
		       contact, help_letter_template, crowdfunding_template, 
		       sort, created_at, updated_at
		FROM help_channels
		WHERE id = ? AND is_audit = 1
	`

	var channel struct {
		ID                   uint      `db:"id"`
		ChannelType          string    `db:"channel_type"`
		Name                 string    `db:"name"`
		ApplyCondition       string    `db:"apply_condition"`
		ResponseTime         string    `db:"response_time"`
		Contact              string    `db:"contact"`
		HelpLetterTemplate   string    `db:"help_letter_template"`
		CrowdfundingTemplate string    `db:"crowdfunding_template"`
		Sort                 int       `db:"sort"`
		CreatedAt            time.Time `db:"created_at"`
		UpdatedAt            time.Time `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&channel.ID, &channel.ChannelType, &channel.Name,
		&channel.ApplyCondition, &channel.ResponseTime, &channel.Contact,
		&channel.HelpLetterTemplate, &channel.CrowdfundingTemplate,
		&channel.Sort, &channel.CreatedAt, &channel.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "渠道不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询渠道详情失败",
		})
		return
	}

	// 构建模板列表
	var templates []ChannelTemplate
	if channel.HelpLetterTemplate != "" {
		templates = append(templates, ChannelTemplate{
			Name: "紧急求助信模板",
			URL:  channel.HelpLetterTemplate,
		})
	}
	if channel.CrowdfundingTemplate != "" {
		templates = append(templates, ChannelTemplate{
			Name: "众筹求助模板",
			URL:  channel.CrowdfundingTemplate,
		})
	}

	// 确保数组不为 null
	if templates == nil {
		templates = []ChannelTemplate{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": ChannelDetailResponse{
			ID:             channel.ID,
			Name:           channel.Name,
			Type:           convertChannelType(channel.ChannelType),
			TypeName:       channel.ChannelType,
			Desc:           "响应时间：" + channel.ResponseTime,
			ApplyCondition: channel.ApplyCondition,
			ResponseTime:   channel.ResponseTime,
			ContactValue:   extractContactValue(channel.Contact, convertChannelType(channel.ChannelType)),
			ServiceTime:    channel.ResponseTime,
			Templates:      templates,
			Available:      true,
			PublishDate:    channel.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdateTime:     channel.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// CaseItem 案例项响应结构
type CaseItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Disease     string `json:"disease"`
	DiseaseCode string `json:"diseaseCode"`
	Amount      string `json:"amount"`
	AmountValue int    `json:"amountValue"`
	Summary     string `json:"summary"`
	PublishDate string `json:"publishDate"`
	CoverUrl    string `json:"coverUrl"`
	ViewCount   int    `json:"viewCount"`
	Status      string `json:"status"`
}

// CaseListResponse 案例列表响应结构
type CaseListResponse struct {
	List     []CaseItem `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

// CaseDetailResponse 案例详情响应结构
type CaseDetailResponse struct {
	ID           uint     `json:"id"`
	Title        string   `json:"title"`
	Disease      string   `json:"disease"`
	DiseaseCode  string   `json:"diseaseCode"`
	Amount       string   `json:"amount"`
	AmountValue  int      `json:"amountValue"`
	Summary      string   `json:"summary"`
	Content      string   `json:"content"`
	PublishDate  string   `json:"publishDate"`
	CoverUrl     string   `json:"coverUrl"`
	Images       []string `json:"images"`
	VideoUrl     string   `json:"videoUrl"`
	ProjectName  string   `json:"projectName"`
	ProjectID    uint     `json:"projectId"`
	ApplyProcess string   `json:"applyProcess"`
	Tips         string   `json:"tips"`
	ViewCount    int      `json:"viewCount"`
	Status       string   `json:"status"`
	PdfUrl       string   `json:"pdfUrl"`
}

// GetCases 获取救助案例列表
func GetCases(c *gin.Context) {
	// 获取请求参数
	diseaseStr := c.DefaultQuery("disease", "")
	keyword := c.DefaultQuery("keyword", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}

	if diseaseStr != "" {
		whereClause += " AND disease_value = ?"
		args = append(args, diseaseStr)
	}
	if keyword != "" {
		whereClause += " AND (case_title LIKE ? OR patient_desc LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM relief_cases " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 查询列表
	listQuery := `
		SELECT id, disease_value, case_title, actual_relief, 
		       experience, created_at
		FROM relief_cases
		` + whereClause + `
		ORDER BY created_at DESC
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

	var list []CaseItem
	for rows.Next() {
		var caseItem struct {
			ID           uint      `db:"id"`
			DiseaseValue string    `db:"disease_value"`
			CaseTitle    string    `db:"case_title"`
			ActualRelief string    `db:"actual_relief"`
			Experience   string    `db:"experience"`
			CreatedAt    time.Time `db:"created_at"`
		}
		if err := rows.Scan(
			&caseItem.ID, &caseItem.DiseaseValue, &caseItem.CaseTitle,
			&caseItem.ActualRelief, &caseItem.Experience, &caseItem.CreatedAt,
		); err != nil {
			continue
		}

		// 转换疾病信息
		diseaseName, diseaseCode := getDiseaseInfo(caseItem.DiseaseValue)

		// 提取金额数值
		amountValue := extractAmountValue(caseItem.ActualRelief)

		list = append(list, CaseItem{
			ID:          caseItem.ID,
			Title:       caseItem.CaseTitle,
			Disease:     diseaseName,
			DiseaseCode: diseaseCode,
			Amount:      caseItem.ActualRelief,
			AmountValue: amountValue,
			Summary:     truncateSummary(caseItem.Experience, 50),
			PublishDate: caseItem.CreatedAt.Format("2006-01-02T15:04:05Z"),
			CoverUrl:    "https://example.com/cases/cover/" + strconv.Itoa(int(caseItem.ID)) + ".jpg",
			ViewCount:   0, // 可后续增加浏览量统计字段
			Status:      "published",
		})
	}

	// 确保数组不为 null
	if list == nil {
		list = []CaseItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": CaseListResponse{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// GetCaseDetail 获取案例详情
func GetCaseDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的案例 ID",
		})
		return
	}

	// 查询案例详情（关联项目名称）
	query := `
		SELECT rc.id, rc.disease_value, rc.project_id, rc.case_title, 
		       rc.patient_desc, rc.actual_relief, rc.experience, 
		       rc.pitfall_guide, rc.case_pdf, rc.material_template, 
		       rc.created_at, rp.name as project_name
		FROM relief_cases rc
		LEFT JOIN relief_projects rp ON rc.project_id = rp.id
		WHERE rc.id = ? AND rc.is_audit = 1
	`

	var caseItem struct {
		ID               uint      `db:"id"`
		DiseaseValue     string    `db:"disease_value"`
		ProjectID        uint      `db:"project_id"`
		CaseTitle        string    `db:"case_title"`
		PatientDesc      string    `db:"patient_desc"`
		ActualRelief     string    `db:"actual_relief"`
		Experience       string    `db:"experience"`
		PitfallGuide     string    `db:"pitfall_guide"`
		CasePdf          string    `db:"case_pdf"`
		MaterialTemplate string    `db:"material_template"`
		CreatedAt        time.Time `db:"created_at"`
		ProjectName      string    `db:"project_name"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&caseItem.ID, &caseItem.DiseaseValue, &caseItem.ProjectID,
		&caseItem.CaseTitle, &caseItem.PatientDesc, &caseItem.ActualRelief,
		&caseItem.Experience, &caseItem.PitfallGuide, &caseItem.CasePdf,
		&caseItem.MaterialTemplate, &caseItem.CreatedAt, &caseItem.ProjectName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "案例不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询案例详情失败",
		})
		return
	}

	// 转换疾病信息
	diseaseName, diseaseCode := getDiseaseInfo(caseItem.DiseaseValue)

	// 提取金额数值
	amountValue := extractAmountValue(caseItem.ActualRelief)

	// 构建图片列表（示例，实际可从数据库或配置获取）
	images := []string{
		"https://example.com/cases/images/" + strconv.Itoa(int(caseItem.ID)) + "_1.jpg",
		"https://example.com/cases/images/" + strconv.Itoa(int(caseItem.ID)) + "_2.jpg",
		"https://example.com/cases/images/" + strconv.Itoa(int(caseItem.ID)) + "_3.jpg",
	}

	// 确保数组不为 null
	if images == nil {
		images = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": CaseDetailResponse{
			ID:           caseItem.ID,
			Title:        caseItem.CaseTitle,
			Disease:      diseaseName,
			DiseaseCode:  diseaseCode,
			Amount:       caseItem.ActualRelief,
			AmountValue:  amountValue,
			Summary:      truncateSummary(caseItem.Experience, 50),
			Content:      caseItem.PatientDesc + "\n\n" + caseItem.Experience,
			PublishDate:  caseItem.CreatedAt.Format("2006-01-02T15:04:05Z"),
			CoverUrl:     "https://example.com/cases/cover/" + strconv.Itoa(int(caseItem.ID)) + ".jpg",
			Images:       images,
			VideoUrl:     "", // 可后续扩展视频字段
			ProjectName:  caseItem.ProjectName,
			ProjectID:    caseItem.ProjectID,
			ApplyProcess: "1.提交申请材料；2.基金会初审；3.专家评估；4.公示；5.发放救助金",
			Tips:         caseItem.PitfallGuide,
			ViewCount:    0, // 可后续增加浏览量统计字段
			Status:       "published",
			PdfUrl:       caseItem.CasePdf,
		},
	})
}

// getDiseaseInfo 获取疾病信息（名称和代码）
func getDiseaseInfo(diseaseValue string) (string, string) {
	// 根据实际 disease_options 表数据调整映射关系
	diseaseMap := map[string]struct {
		Name string
		Code string
	}{
		"1": {Name: "渐冻症", Code: "als"},
		"2": {Name: "血友病", Code: "hemophilia"},
		"3": {Name: "罕见病", Code: "rare"},
		"4": {Name: "戈谢病", Code: "gaucher"},
		"5": {Name: "庞贝症", Code: "pompe"},
	}
	if info, ok := diseaseMap[diseaseValue]; ok {
		return info.Name, info.Code
	}
	return "其他", "other"
}

// extractAmountValue 提取金额数值
func extractAmountValue(amountStr string) int {
	// 简化处理，从字符串中提取数字
	// 实际可根据具体格式解析
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(amountStr, -1)
	if len(matches) > 0 {
		value, _ := strconv.Atoi(matches[0])
		// 判断是否有"万"单位
		if strings.Contains(amountStr, "万") {
			value *= 10000
		}
		return value
	}
	return 0
}

// CasePDFResponse 案例 PDF 响应结构
type CasePDFResponse struct {
	DownloadURL string `json:"downloadUrl"`
	FileName    string `json:"fileName"`
	FileSize    string `json:"fileSize"`
	ExpireTime  int    `json:"expireTime"`
}

// DiseaseItem 疾病选项项
type DiseaseItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// DiseaseResponse 疾病选项响应结构
type DiseaseResponse struct {
	Diseases []DiseaseItem `json:"diseases"`
}

// GetCasePDF 获取案例 PDF 下载链接
func GetCasePDF(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的案例 ID",
		})
		return
	}

	// 查询案例 PDF 地址
	query := `
		SELECT case_pdf, case_title 
		FROM relief_cases 
		WHERE id = ? AND is_audit = 1
	`

	var casePDF struct {
		CasePDF   string `db:"case_pdf"`
		CaseTitle string `db:"case_title"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(&casePDF.CasePDF, &casePDF.CaseTitle)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "案例不存在或未审核",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询案例失败",
		})
		return
	}

	// 检查 PDF 地址是否为空
	if casePDF.CasePDF == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "该案例暂无 PDF 版本",
		})
		return
	}

	// 生成文件名
	fileName := casePDF.CaseTitle + ".pdf"

	// 获取文件大小（简化处理，实际可调用文件服务获取）
	fileSize := getFileSize(casePDF.CasePDF)

	// 设置链接过期时间（秒）
	expireTime := 3600

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": CasePDFResponse{
			DownloadURL: casePDF.CasePDF,
			FileName:    fileName,
			FileSize:    fileSize,
			ExpireTime:  expireTime,
		},
	})
}

// GetCaseDiseases 获取疾病筛选选项
func GetCaseDiseases(c *gin.Context) {
	// 方案一：从数据库动态查询（推荐）
	query := `
		SELECT DISTINCT rc.disease_value, do.name, do.code
		FROM relief_cases rc
		INNER JOIN disease_options do ON rc.disease_value = do.value
		WHERE rc.is_audit = 1
		ORDER BY do.sort ASC
	`

	rows, err := db.MySQL.Query(query)
	if err != nil {
		// 查询失败则返回硬编码数据
		diseases := getHardcodedDiseases()
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"data": DiseaseResponse{
				Diseases: diseases,
			},
		})
		return
	}
	defer rows.Close()

	var diseases []DiseaseItem
	// 添加"全部疾病"选项
	diseases = append(diseases, DiseaseItem{
		Text:  "全部疾病",
		Value: "all",
	})

	for rows.Next() {
		var disease struct {
			DiseaseValue string `db:"disease_value"`
			Name         string `db:"name"`
			Code         string `db:"code"`
		}
		if err := rows.Scan(&disease.DiseaseValue, &disease.Name, &disease.Code); err != nil {
			continue
		}
		diseases = append(diseases, DiseaseItem{
			Text:  disease.Name,
			Value: disease.Code,
		})
	}

	// 确保数组不为 null
	if diseases == nil {
		diseases = getHardcodedDiseases()
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": DiseaseResponse{
			Diseases: diseases,
		},
	})
}

// getFileSize 获取文件大小（简化实现）
func getFileSize(url string) string {
	// 实际可调用文件服务获取真实大小
	// 这里根据 URL 返回示例值
	if strings.Contains(url, ".pdf") {
		return "1.2MB"
	}
	return "未知"
}

// getHardcodedDiseases 获取硬编码疾病列表（备用）
func getHardcodedDiseases() []DiseaseItem {
	return []DiseaseItem{
		{Text: "全部疾病", Value: "all"},
		{Text: "血友病", Value: "hemophilia"},
		{Text: "渐冻症", Value: "als"},
		{Text: "罕见病", Value: "rare"},
		{Text: "戈谢病", Value: "gaucher"},
		{Text: "庞贝症", Value: "pompe"},
		{Text: "白血病", Value: "leukemia"},
		{Text: "脑瘫", Value: "cerebral_palsy"},
	}
}

// CreateProjectRequest 创建项目请求结构
type CreateProjectRequest struct {
	Title           string `json:"title" binding:"required"`
	Content         string `json:"content" binding:"required"`
	Requirements    string `json:"requirements"`
	ReliefType      string `json:"reliefType" binding:"required"`
	DiseaseValue    int    `json:"diseaseValue"`
	ReliefStandard  string `json:"reliefStandard"`
	ApplyDifficulty string `json:"applyDifficulty"`
	ApplyDeadline   string `json:"applyDeadline"`
	Contact         string `json:"contact"`
	ApplyForm       string `json:"applyForm"`
	ApplyGuide      string `json:"applyGuide"`
	MaterialList    string `json:"materialList"`
	Organizer       string `json:"organizer"`
	Sort            int    `json:"sort"`
}

// CreateProject 新增救助项目
func CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 参数校验
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "项目名称不能为空",
		})
		return
	}

	// 插入数据库
	insertQuery := `
		INSERT INTO relief_projects 
		(name, apply_process, apply_condition, relief_type, disease_value,
		 relief_standard, apply_difficulty, apply_deadline, contact,
		 apply_form, apply_guide, material_list, organizer, sort,
		 is_audit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW())
	`
	result, err := db.MySQL.Exec(insertQuery,
		req.Title, req.Content, req.Requirements, req.ReliefType, req.DiseaseValue,
		req.ReliefStandard, req.ApplyDifficulty, req.ApplyDeadline, req.Contact,
		req.ApplyForm, req.ApplyGuide, req.MaterialList, req.Organizer, req.Sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建项目失败",
		})
		return
	}

	// 获取新增的 ID
	id, _ := result.LastInsertId()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data": gin.H{
			"id": id,
		},
	})
}

// UpdateProjectRequest 更新项目请求结构
type UpdateProjectRequest struct {
	Title           string `json:"title"`
	Content         string `json:"content"`
	Requirements    string `json:"requirements"`
	ReliefType      string `json:"reliefType"`
	DiseaseValue    int    `json:"diseaseValue"`
	ReliefStandard  string `json:"reliefStandard"`
	ApplyDifficulty string `json:"applyDifficulty"`
	ApplyDeadline   string `json:"applyDeadline"`
	Contact         string `json:"contact"`
	ApplyForm       string `json:"applyForm"`
	ApplyGuide      string `json:"applyGuide"`
	MaterialList    string `json:"materialList"`
	Organizer       string `json:"organizer"`
	Sort            int    `json:"sort"`
	IsAudit         int    `json:"isAudit"`
}

// UpdateProject 更新救助项目
func UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目 ID",
		})
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 构建动态更新语句
	updateFields := []string{}
	args := []interface{}{}

	if req.Title != "" {
		updateFields = append(updateFields, "name = ?")
		args = append(args, req.Title)
	}
	if req.Content != "" {
		updateFields = append(updateFields, "apply_process = ?")
		args = append(args, req.Content)
	}
	if req.Requirements != "" {
		updateFields = append(updateFields, "apply_condition = ?")
		args = append(args, req.Requirements)
	}
	if req.ReliefType != "" {
		updateFields = append(updateFields, "relief_type = ?")
		args = append(args, req.ReliefType)
	}
	if req.DiseaseValue != 0 {
		updateFields = append(updateFields, "disease_value = ?")
		args = append(args, req.DiseaseValue)
	}
	if req.ReliefStandard != "" {
		updateFields = append(updateFields, "relief_standard = ?")
		args = append(args, req.ReliefStandard)
	}
	if req.ApplyDifficulty != "" {
		updateFields = append(updateFields, "apply_difficulty = ?")
		args = append(args, req.ApplyDifficulty)
	}
	if req.ApplyDeadline != "" {
		updateFields = append(updateFields, "apply_deadline = ?")
		args = append(args, req.ApplyDeadline)
	}
	if req.Contact != "" {
		updateFields = append(updateFields, "contact = ?")
		args = append(args, req.Contact)
	}
	if req.ApplyForm != "" {
		updateFields = append(updateFields, "apply_form = ?")
		args = append(args, req.ApplyForm)
	}
	if req.ApplyGuide != "" {
		updateFields = append(updateFields, "apply_guide = ?")
		args = append(args, req.ApplyGuide)
	}
	if req.MaterialList != "" {
		updateFields = append(updateFields, "material_list = ?")
		args = append(args, req.MaterialList)
	}
	if req.Organizer != "" {
		updateFields = append(updateFields, "organizer = ?")
		args = append(args, req.Organizer)
	}
	updateFields = append(updateFields, "sort = ?")
	args = append(args, req.Sort)

	// 审核状态单独处理
	if req.IsAudit != 0 {
		updateFields = append(updateFields, "is_audit = ?")
		args = append(args, req.IsAudit)
	}

	updateFields = append(updateFields, "updated_at = NOW()")
	args = append(args, id)

	updateQuery := `UPDATE relief_projects SET ` + strings.Join(updateFields, ", ") + ` WHERE id = ?`
	_, err = db.MySQL.Exec(updateQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新项目失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    nil,
	})
}

// DeleteProject 删除救助项目（软删除）
func DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目 ID",
		})
		return
	}

	// 软删除：更新 is_audit = 0
	updateQuery := `UPDATE relief_projects SET is_audit = 0, updated_at = NOW() WHERE id = ?`
	_, err = db.MySQL.Exec(updateQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除项目失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    nil,
	})
}

// CreateCaseRequest 创建案例请求结构
type CreateCaseRequest struct {
	Title            string `json:"title" binding:"required"`
	PatientDesc      string `json:"patientDesc" binding:"required"`
	DiseaseValue     string `json:"diseaseValue" binding:"required"`
	ActualRelief     string `json:"actualRelief"`
	Experience       string `json:"experience"`
	PitfallGuide     string `json:"pitfallGuide"`
	CasePdf          string `json:"casePdf"`
	MaterialTemplate string `json:"materialTemplate"`
	ProjectID        uint   `json:"projectId"`
	CoverUrl         string `json:"coverUrl"`
	Sort             int    `json:"sort"`
}

// CreateCase 新增救助案例
func CreateCase(c *gin.Context) {
	var req CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 参数校验
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "案例标题不能为空",
		})
		return
	}
	if req.PatientDesc == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "患者描述不能为空",
		})
		return
	}

	// 插入数据库
	insertQuery := `
		INSERT INTO relief_cases 
		(case_title, patient_desc, disease_value, actual_relief, experience,
		 pitfall_guide, case_pdf, material_template, project_id, cover_url, sort,
		 is_audit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW())
	`
	result, err := db.MySQL.Exec(insertQuery,
		req.Title, req.PatientDesc, req.DiseaseValue, req.ActualRelief, req.Experience,
		req.PitfallGuide, req.CasePdf, req.MaterialTemplate, req.ProjectID, req.CoverUrl, req.Sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建案例失败",
		})
		return
	}

	// 获取新增的 ID
	id, _ := result.LastInsertId()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data": gin.H{
			"id": id,
		},
	})
}

// UpdateCaseRequest 更新案例请求结构
type UpdateCaseRequest struct {
	Title            string `json:"title"`
	PatientDesc      string `json:"patientDesc"`
	DiseaseValue     string `json:"diseaseValue"`
	ActualRelief     string `json:"actualRelief"`
	Experience       string `json:"experience"`
	PitfallGuide     string `json:"pitfallGuide"`
	CasePdf          string `json:"casePdf"`
	MaterialTemplate string `json:"materialTemplate"`
	ProjectID        uint   `json:"projectId"`
	CoverUrl         string `json:"coverUrl"`
	Sort             int    `json:"sort"`
	IsAudit          int    `json:"isAudit"`
}

// UpdateCase 更新救助案例
func UpdateCase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的案例 ID",
		})
		return
	}

	var req UpdateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 构建动态更新语句
	updateFields := []string{}
	args := []interface{}{}

	if req.Title != "" {
		updateFields = append(updateFields, "case_title = ?")
		args = append(args, req.Title)
	}
	if req.PatientDesc != "" {
		updateFields = append(updateFields, "patient_desc = ?")
		args = append(args, req.PatientDesc)
	}
	if req.DiseaseValue != "" {
		updateFields = append(updateFields, "disease_value = ?")
		args = append(args, req.DiseaseValue)
	}
	if req.ActualRelief != "" {
		updateFields = append(updateFields, "actual_relief = ?")
		args = append(args, req.ActualRelief)
	}
	if req.Experience != "" {
		updateFields = append(updateFields, "experience = ?")
		args = append(args, req.Experience)
	}
	if req.PitfallGuide != "" {
		updateFields = append(updateFields, "pitfall_guide = ?")
		args = append(args, req.PitfallGuide)
	}
	if req.CasePdf != "" {
		updateFields = append(updateFields, "case_pdf = ?")
		args = append(args, req.CasePdf)
	}
	if req.MaterialTemplate != "" {
		updateFields = append(updateFields, "material_template = ?")
		args = append(args, req.MaterialTemplate)
	}
	if req.ProjectID != 0 {
		updateFields = append(updateFields, "project_id = ?")
		args = append(args, req.ProjectID)
	}
	if req.CoverUrl != "" {
		updateFields = append(updateFields, "cover_url = ?")
		args = append(args, req.CoverUrl)
	}
	updateFields = append(updateFields, "sort = ?")
	args = append(args, req.Sort)

	// 审核状态单独处理
	if req.IsAudit != 0 {
		updateFields = append(updateFields, "is_audit = ?")
		args = append(args, req.IsAudit)
	}

	updateFields = append(updateFields, "updated_at = NOW()")
	args = append(args, id)

	updateQuery := `UPDATE relief_cases SET ` + strings.Join(updateFields, ", ") + ` WHERE id = ?`
	_, err = db.MySQL.Exec(updateQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新案例失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    nil,
	})
}

// DeleteCase 删除救助案例（软删除）
func DeleteCase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的案例 ID",
		})
		return
	}

	// 软删除：更新 is_audit = 0
	updateQuery := `UPDATE relief_cases SET is_audit = 0, updated_at = NOW() WHERE id = ?`
	_, err = db.MySQL.Exec(updateQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除案例失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    nil,
	})
}
