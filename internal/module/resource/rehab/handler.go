package rehab

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"rare_backend/internal/pkg/db"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ResourceResponse 资源文件响应结构
type ResourceResponse struct {
	DownloadUrl string `json:"downloadUrl"`
	PreviewUrl  string `json:"previewUrl"`
	FileName    string `json:"fileName"`
	FileSize    string `json:"fileSize"`
}

// OptionItem 选项项
type OptionItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// convertStageToValue 将病情阶段转换为枚举值
func convertStageToValue(stage string) string {
	valueMap := map[string]string{
		"早期": "early",
		"中期": "mid",
		"晚期": "late",
	}
	if val, ok := valueMap[stage]; ok {
		return val
	}
	return stage
}

// convertValueToDisease 将疾病 value 转换为前端使用的标识
func convertValueToDisease(value int) string {
	// 根据实际 disease_options 表数据映射
	diseaseMap := map[int]string{
		1: "als",
		2: "huntington",
		3: "rare",
	}
	if val, ok := diseaseMap[value]; ok {
		return val
	}
	return fmt.Sprintf("disease_%d", value)
}

// convertStageToType 将病情阶段转换为训练类型
func convertStageToType(stage string) string {
	typeMap := map[string]string{
		"early":       "基础训练",
		"middle":      "强化训练", // 注意这里是 middle
		"late":        "维持训练",
		"stable":      "维持训练",
		"progressive": "强化训练",
	}
	if val, ok := typeMap[stage]; ok {
		return val
	}
	return "康复训练"
}

// getManualUpdateTime 获取手册更新时间
func getManualUpdateTime() string {
	// 实际项目中可从数据库查询 home_care_manuals 表的最新 updated_at
	// 这里返回示例时间
	return time.Now().Format("2006-01-02T15:04:05Z")
}

// DoctorItem 医生信息项
type DoctorItem struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Specialty string `json:"specialty"`
}

// InstitutionItem 康复机构项响应结构
type InstitutionItem struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	ProvinceCode  string   `json:"provinceCode"`
	CityCode      string   `json:"cityCode"`
	DistrictCode  string   `json:"districtCode"`
	ProvinceName  string   `json:"provinceName"` // 新增
	CityName      string   `json:"cityName"`     // 新增
	DistrictName  string   `json:"districtName"` // 新增
	Address       string   `json:"address"`
	ContactPhone  string   `json:"contactPhone"`
	ContactUrl    string   `json:"contactUrl"`
	Qualification string   `json:"qualification"`
	RehabProjects string   `json:"rehabProjects"`
	FeeStandard   string   `json:"feeStandard"`
	DiseaseIds    []uint64 `json:"diseaseIds"`  // 新增：关联的疾病ID列表
	AuditStatus   int8     `json:"auditStatus"` // 【新增】审核状态字段，解决编译错误
	Rating        float64  `json:"rating"`      // 保留原有逻辑或从其他表获取
	Status        string   `json:"status"`
	UpdateAt      string   `json:"updatedAt"` // 【新增】更新时间字段，解决编译错误
}

// InstitutionListResponse 机构列表响应结构
type InstitutionListResponse struct {
	List     []InstitutionItem `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// InstitutionDetailResponse 机构详情响应结构
type InstitutionDetailResponse struct {
	ID            uint         `json:"id"`
	Name          string       `json:"name"`
	Type          string       `json:"type"`
	TypeName      string       `json:"typeName"`
	Region        string       `json:"region"`
	RegionCode    string       `json:"regionCode"`
	Address       string       `json:"address"`
	Contact       string       `json:"contact"`
	Phone         string       `json:"phone"`
	Email         string       `json:"email"`
	Website       string       `json:"website"`
	Services      []string     `json:"services"`
	Rating        float64      `json:"rating"`
	IsInsurance   bool         `json:"isInsurance"`
	Description   string       `json:"description"`
	CoverUrl      string       `json:"coverUrl"`
	Images        []string     `json:"images"`
	BusinessHours string       `json:"businessHours"`
	Facilities    []string     `json:"facilities"`
	Doctors       []DoctorItem `json:"doctors"`
	Status        string       `json:"status"`
}

// GetInstitutions 获取康复机构列表
func GetInstitutions(c *gin.Context) {
	// 获取请求参数
	provinceCode := c.DefaultQuery("provinceCode", "")
	cityCode := c.DefaultQuery("cityCode", "")
	districtCode := c.DefaultQuery("districtCode", "")
	diseaseStr := c.DefaultQuery("diseaseId", "")
	keyword := c.DefaultQuery("keyword", "")

	// 【新增】获取 auditStatus 参数
	auditStatusStr := c.DefaultQuery("auditStatus", "")

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

	// 1. 构建基础 SQL 片段
	baseFrom := "FROM rehab_institution i"
	whereConditions := []string{}
	args := []interface{}{}

	// 【修改】处理 auditStatus 筛选逻辑
	if auditStatusStr != "" {
		auditStatus, err := strconv.Atoi(auditStatusStr)
		if err == nil {
			// 如果前端明确传了审核状态，则按该状态筛选
			whereConditions = append(whereConditions, "i.audit_status = ?")
			args = append(args, auditStatus)
		}
	} else {
		// 【重要】如果前端没传审核状态，默认只展示已通过的机构 (保持原有业务逻辑一致性)
		whereConditions = append(whereConditions, "i.audit_status = 1")
	}

	// 2. 动态添加其他筛选条件
	if provinceCode != "" && provinceCode != "all" {
		whereConditions = append(whereConditions, "i.province_code = ?")
		args = append(args, provinceCode)
	}
	if cityCode != "" {
		whereConditions = append(whereConditions, "i.city_code = ?")
		args = append(args, cityCode)
	}
	if districtCode != "" {
		whereConditions = append(whereConditions, "i.district_code = ?")
		args = append(args, districtCode)
	}

	// 疾病筛选：需要通过关联表查询
	if diseaseStr != "" {
		diseaseID, err := strconv.Atoi(diseaseStr)
		if err == nil && diseaseID > 0 {
			whereConditions = append(whereConditions, "EXISTS (SELECT 1 FROM rehab_institution_disease_rel r WHERE r.institution_id = i.id AND r.disease_id = ?)")
			args = append(args, diseaseID)
		}
	}

	if keyword != "" {
		whereConditions = append(whereConditions, "(i.name LIKE ? OR i.address LIKE ?)")
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 3. 拼接 WHERE 子句
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 4. 构建并执行 Count 查询
	countQuery := "SELECT COUNT(*) " + baseFrom + " " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败: " + err.Error(),
		})
		return
	}

	// 5. 构建并执行 List 查询
	listQuery := `
		SELECT i.id, i.name, i.province_code, i.city_code, i.district_code, 
		       i.province_name, i.city_name, i.district_name,
		       i.qualification, i.rehab_projects, i.fee_standard, 
		       i.contact_phone, i.contact_url, i.address, i.audit_status, i.created_at, i.updated_at
		` + baseFrom + " " + whereClause + `
		ORDER BY i.id DESC
		LIMIT ? OFFSET ?
	`
	// 追加分页参数
	listArgs := append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, listArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var list []InstitutionItem
	var institutionIDs []uint64

	// 临时存储扫描结果
	type tempInst struct {
		ID            uint
		Name          string
		ProvinceCode  string
		CityCode      string
		DistrictCode  string
		ProvinceName  string
		CityName      string
		DistrictName  string
		Qualification string
		RehabProjects string
		FeeStandard   string
		ContactPhone  string
		ContactUrl    string
		Address       string
		AuditStatus   int8
		CreatedAt     time.Time
		UpdatedAt     time.Time
	}

	var tempList []tempInst

	for rows.Next() {
		var t tempInst
		if err := rows.Scan(
			&t.ID, &t.Name, &t.ProvinceCode, &t.CityCode, &t.DistrictCode,
			&t.ProvinceName, &t.CityName, &t.DistrictName,
			&t.Qualification, &t.RehabProjects, &t.FeeStandard,
			&t.ContactPhone, &t.ContactUrl, &t.Address, &t.AuditStatus, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			// 建议记录日志
			continue
		}
		tempList = append(tempList, t)
		institutionIDs = append(institutionIDs, uint64(t.ID))
	}

	// 6. 批量查询疾病关联 (解决 N+1 问题)
	diseaseMap := make(map[uint64][]uint64)

	if len(institutionIDs) > 0 {
		// 【核心修复】构建 IN 查询的正确姿势
		// 1. 创建占位符字符串 "?, ?, ?"
		placeholders := make([]string, len(institutionIDs))
		queryArgs := make([]interface{}, len(institutionIDs))
		for i, id := range institutionIDs {
			placeholders[i] = "?"
			queryArgs[i] = id
		}
		placeholderStr := strings.Join(placeholders, ",")

		// 2. 构建 SQL
		relQuery := fmt.Sprintf("SELECT institution_id, disease_id FROM rehab_institution_disease_rel WHERE institution_id IN (%s)", placeholderStr)

		// 3. 执行查询
		relRows, err := db.MySQL.Query(relQuery, queryArgs...)
		if err != nil {
			// 记录错误但不中断主流程，疾病列表将为空
			// log.Printf("Query disease rel error: %v", err)
		} else {
			defer relRows.Close()
			for relRows.Next() {
				var instID uint64
				var disID uint64
				if err := relRows.Scan(&instID, &disID); err == nil {
					diseaseMap[instID] = append(diseaseMap[instID], disID)
				}
			}
		}
	}

	// 7. 组装最终返回数据
	for _, t := range tempList {
		// 从 map 中获取疾病 IDs，如果不存在则初始化为空切片，避免前端收到 null
		dids := diseaseMap[uint64(t.ID)]
		if dids == nil {
			dids = []uint64{}
		}

		item := InstitutionItem{
			ID:            t.ID,
			Name:          t.Name,
			ProvinceCode:  t.ProvinceCode,
			CityCode:      t.CityCode,
			DistrictCode:  t.DistrictCode,
			ProvinceName:  t.ProvinceName,
			CityName:      t.CityName,
			DistrictName:  t.DistrictName,
			Address:       t.Address,
			ContactPhone:  t.ContactPhone,
			ContactUrl:    t.ContactUrl,
			Qualification: t.Qualification,
			RehabProjects: t.RehabProjects,
			FeeStandard:   t.FeeStandard,
			DiseaseIds:    dids,
			AuditStatus:   t.AuditStatus,
			UpdateAt:      t.UpdatedAt.Format("2006-01-02 15:04:05"),
			Rating:        4.5 + float64(t.ID%10)/10, // 示例评分逻辑
			Status:        "active",                  // 示例状态
		}

		list = append(list, item)
	}

	if list == nil {
		list = []InstitutionItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": InstitutionListResponse{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// InstitutionDiseaseItem 机构关联疾病详情项
type InstitutionDiseaseItem struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

// GetInstitutionDetail 获取康复机构详情
func GetInstitutionDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的机构 ID",
		})
		return
	}

	// 1. 查询机构基础信息
	query := `
		SELECT id, name, province_code, city_code, district_code, 
		       province_name, city_name, district_name,
		       qualification, rehab_projects, fee_standard, 
		       contact_phone, contact_url, address, 
		       audit_status, reject_reason, created_at, updated_at
		FROM rehab_institution
		WHERE id = ?
	`

	var inst struct {
		ID            uint
		Name          string
		ProvinceCode  string
		CityCode      string
		DistrictCode  string
		ProvinceName  string
		CityName      string
		DistrictName  string
		Qualification string
		RehabProjects string
		FeeStandard   string
		ContactPhone  string
		ContactUrl    string
		Address       string
		AuditStatus   int8
		RejectReason  sql.NullString
		CreatedAt     time.Time
		UpdatedAt     time.Time
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&inst.ID, &inst.Name, &inst.ProvinceCode, &inst.CityCode, &inst.DistrictCode,
		&inst.ProvinceName, &inst.CityName, &inst.DistrictName,
		&inst.Qualification, &inst.RehabProjects, &inst.FeeStandard,
		&inst.ContactPhone, &inst.ContactUrl, &inst.Address,
		&inst.AuditStatus, &inst.RejectReason, &inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "机构不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询机构详情失败: " + err.Error(),
		})
		return
	}

	// 2. 【核心修改】查询关联的疾病 ID 列表和详细信息
	// 使用 JOIN 一次性获取 id, name, alias
	diseaseQuery := `
		SELECT d.id, d.name, d.alias 
		FROM rehab_institution_disease_rel r
		INNER JOIN disease d ON r.disease_id = d.id
		WHERE r.institution_id = ?
		ORDER BY d.id ASC
	`

	rows, err := db.MySQL.Query(diseaseQuery, id)
	if err != nil {
		// 记录错误但不中断主流程，疾病列表将为空
		// log.Printf("Query institution diseases error: %v", err)
		rows = nil
	}

	var diseaseIds []uint64
	var diseases []InstitutionDiseaseItem

	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var d struct {
				ID    int64
				Name  string
				Alias string
			}
			if err := rows.Scan(&d.ID, &d.Name, &d.Alias); err == nil {
				// 填充 diseaseIds (注意类型转换，如果前端需要 uint64 则转换，否则保持 int64 也可，视前端定义而定)
				// 这里为了匹配之前的 DiseaseIds []uint64，我们做转换
				if d.ID > 0 {
					diseaseIds = append(diseaseIds, uint64(d.ID))
				}

				// 填充 diseases 详情
				diseases = append(diseases, InstitutionDiseaseItem{
					ID:    d.ID,
					Name:  d.Name,
					Alias: d.Alias,
				})
			}
		}
	}

	// 确保切片不为 nil，返回空数组而不是 null
	if diseaseIds == nil {
		diseaseIds = []uint64{}
	}
	if diseases == nil {
		diseases = []InstitutionDiseaseItem{}
	}

	// 3. 解析其他字段
	services := parseServices(inst.RehabProjects)
	instType, instTypeName := convertInstitutionType(inst.Name)

	// 示例数据构造
	images := []string{
		"https://example.com/institutions/" + strconv.FormatUint(uint64(inst.ID), 10) + "_1.jpg",
	}
	facilities := []string{"无障碍通道", "停车场"}
	doctors := getInstitutionDoctors(inst.ID)

	// 4. 构造响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":            inst.ID,
			"name":          inst.Name,
			"type":          instType,
			"typeName":      instTypeName,
			"provinceCode":  inst.ProvinceCode,
			"cityCode":      inst.CityCode,
			"districtCode":  inst.DistrictCode,
			"provinceName":  inst.ProvinceName,
			"cityName":      inst.CityName,
			"districtName":  inst.DistrictName,
			"address":       inst.Address,
			"contactPhone":  inst.ContactPhone,
			"contactUrl":    inst.ContactUrl,
			"qualification": inst.Qualification,
			"rehabProjects": inst.RehabProjects,
			"feeStandard":   inst.FeeStandard,
			"services":      services,

			// 【新增】返回疾病相关字段
			"diseaseIds": diseaseIds,
			"diseases":   diseases,

			"auditStatus":   inst.AuditStatus,
			"rejectReason":  inst.RejectReason.String,
			"createdAt":     inst.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt":     inst.UpdatedAt.Format("2006-01-02 15:04:05"),
			"rating":        4.5 + float64(inst.ID%10)/10,
			"isInsurance":   true,
			"coverUrl":      "https://example.com/institutions/" + strconv.FormatUint(uint64(inst.ID), 10) + ".jpg",
			"images":        images,
			"businessHours": "周一至周日 08:00-17:00",
			"facilities":    facilities,
			"doctors":       doctors,
			"status":        "active",
		},
	})
}

// convertRegionToCode 地区名称转地区代码
func convertRegionToCode(region string) string {
	regionMap := map[string]string{
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
	return "other"
}

// convertInstitutionType 转换机构类型
func convertInstitutionType(name string) (string, string) {
	if strings.Contains(name, "医院") {
		return "hospital", "康复医院"
	}
	if strings.Contains(name, "中心") {
		return "center", "康复中心"
	}
	if strings.Contains(name, "诊所") {
		return "clinic", "康复诊所"
	}
	if strings.Contains(name, "社区") {
		return "community", "社区康复站"
	}
	return "other", "其他机构"
}

// parseServices 解析服务项目
func parseServices(projects string) []string {
	if projects == "" {
		return []string{"康复指导", "康复训练", "护理培训"}
	}
	// 按分隔符分割
	services := strings.Split(projects, "，")
	if len(services) == 0 {
		services = strings.Split(projects, ",")
	}
	if len(services) == 0 {
		return []string{"康复指导", "康复训练", "护理培训"}
	}
	return services
}

// getInstitutionDoctors 获取机构医生列表
func getInstitutionDoctors(instID uint) []DoctorItem {
	// 实际项目中可从医生表查询
	// 这里返回示例数据
	doctorsMap := map[uint][]DoctorItem{
		1: {
			{Name: "张医生", Title: "主任医师", Specialty: "神经康复"},
			{Name: "李医生", Title: "副主任医师", Specialty: "肢体康复"},
		},
		2: {
			{Name: "王医生", Title: "主任医师", Specialty: "儿童康复"},
		},
	}

	if doctors, ok := doctorsMap[instID]; ok {
		return doctors
	}
	return []DoctorItem{
		{Name: "赵医生", Title: "主治医师", Specialty: "康复指导"},
	}
}

// RegionItem 地区选项项
type RegionItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// RegionResponse 地区响应结构
type RegionResponse struct {
	Regions []RegionItem `json:"regions"`
}

// GetRegions 获取地区筛选选项
func GetRegions(c *gin.Context) {
	regions := []RegionItem{
		{Text: "全部地区", Value: "all"},
		{Text: "北京", Value: "bj"},
		{Text: "上海", Value: "sh"},
		{Text: "广州", Value: "gz"},
		{Text: "深圳", Value: "sz"},
		{Text: "浙江", Value: "zj"},
		{Text: "江苏", Value: "js"},
		{Text: "四川", Value: "sc"},
		{Text: "湖北", Value: "hb"},
		{Text: "山东", Value: "sd"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": RegionResponse{
			Regions: regions,
		},
	})
}

// getFileSize 获取文件大小
func getFileSize(url string) string {
	if strings.Contains(url, ".pdf") {
		return "1.5MB"
	}
	if strings.Contains(url, ".doc") {
		return "800KB"
	}
	return "未知"
}

// CounselorItem 心理咨询师信息项
type CounselorItem struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Specialty string `json:"specialty"`
}

// PsychologicalOrgItem 心理咨询机构项响应结构
type PsychologicalOrgItem struct {
	ID           uint                     `json:"id"`
	Name         string                   `json:"name"`
	ProvinceCode string                   `json:"provinceCode"`
	CityCode     string                   `json:"cityCode"`
	DistrictCode string                   `json:"districtCode"`
	ProvinceName string                   `json:"provinceName"` // 【新增】
	CityName     string                   `json:"cityName"`     // 【新增】
	DistrictName string                   `json:"districtName"` // 【新增】
	Address      string                   `json:"address"`
	ContactPhone string                   `json:"contactPhone"`
	ContactUrl   string                   `json:"contactUrl"`
	IsFree       bool                     `json:"isFree"`
	ConsultWay   string                   `json:"consultWay"`
	ContentIntro string                   `json:"contentIntro"`
	AuditStatus  int8                     `json:"auditStatus"`  // 【新增】
	RejectReason *string                  `json:"rejectReason"` // 【新增】
	Type         string                   `json:"type"`         // 保留原有逻辑
	TypeName     string                   `json:"typeName"`     // 保留原有逻辑
	Region       string                   `json:"region"`       // 保留原有逻辑，或可废弃
	RegionCode   string                   `json:"regionCode"`   // 保留原有逻辑
	ServiceTime  string                   `json:"serviceTime"`  // 保留原有逻辑
	Description  string                   `json:"description"`  // 保留原有逻辑
	Services     []string                 `json:"services"`     // 保留原有逻辑
	Rating       float64                  `json:"rating"`       // 保留原有逻辑
	CoverUrl     string                   `json:"coverUrl"`     // 保留原有逻辑
	Status       string                   `json:"status"`       // 保留原有逻辑
	DiseaseIds   []uint64                 `json:"diseaseIds"`   // 【新增】
	Diseases     []InstitutionDiseaseItem `json:"diseases"`     // 【新增】复用已有的 InstitutionDiseaseItem 或定义新的 PsychDiseaseItem
	CreatedAt    string                   `json:"createdAt"`    // 【新增】
	UpdatedAt    string                   `json:"updatedAt"`    // 【新增】
}

// PsychologicalOrgListResponse 机构列表响应结构
type PsychologicalOrgListResponse struct {
	List     []PsychologicalOrgItem `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

// PsychologicalOrgDetailResponse 机构详情响应结构
type PsychologicalOrgDetailResponse struct {
	ID           uint                     `json:"id"`
	Name         string                   `json:"name"`
	ProvinceCode string                   `json:"provinceCode"`
	CityCode     string                   `json:"cityCode"`
	DistrictCode string                   `json:"districtCode"`
	ProvinceName string                   `json:"provinceName"` // 【新增】
	CityName     string                   `json:"cityName"`     // 【新增】
	DistrictName string                   `json:"districtName"` // 【新增】
	Address      string                   `json:"address"`
	ContactPhone string                   `json:"contactPhone"`
	ContactUrl   string                   `json:"contactUrl"`
	IsFree       bool                     `json:"isFree"`
	ConsultWay   string                   `json:"consultWay"`
	ContentIntro string                   `json:"contentIntro"`
	AuditStatus  int8                     `json:"auditStatus"`  // 【新增】
	RejectReason *string                  `json:"rejectReason"` // 【新增】
	Type         string                   `json:"type"`         // 保留原有逻辑
	TypeName     string                   `json:"typeName"`     // 保留原有逻辑
	Region       string                   `json:"region"`       // 保留原有逻辑
	RegionCode   string                   `json:"regionCode"`   // 保留原有逻辑
	ServiceTime  string                   `json:"serviceTime"`  // 保留原有逻辑
	Description  string                   `json:"description"`  // 保留原有逻辑
	Services     []string                 `json:"services"`     // 保留原有逻辑
	Rating       float64                  `json:"rating"`       // 保留原有逻辑
	CoverUrl     string                   `json:"coverUrl"`     // 保留原有逻辑
	Status       string                   `json:"status"`       // 保留原有逻辑
	DiseaseIds   []uint64                 `json:"diseaseIds"`   // 【新增】
	Diseases     []InstitutionDiseaseItem `json:"diseases"`     // 【新增】复用 InstitutionDiseaseItem
	DiseaseCount int                      `json:"diseaseCount"` // 【新增】
	Images       []string                 `json:"images"`       // 保留原有逻辑
	Counselors   []CounselorItem          `json:"counselors"`   // 保留原有逻辑
	CreatedAt    string                   `json:"createdAt"`    // 【新增】
	UpdatedAt    string                   `json:"updatedAt"`    // 【新增】
}

// GetPsychologicalOrgs 获取心理咨询机构列表
func GetPsychologicalOrgs(c *gin.Context) {
	// 获取请求参数
	provinceCode := c.DefaultQuery("provinceCode", "")
	cityCode := c.DefaultQuery("cityCode", "")
	districtCode := c.DefaultQuery("districtCode", "")
	consultWay := c.DefaultQuery("consultWay", "")
	diseaseStr := c.DefaultQuery("diseaseId", "")
	isFreeStr := c.DefaultQuery("isFree", "")
	keyword := c.DefaultQuery("keyword", "")

	// 【新增】支持审核状态筛选
	auditStatusStr := c.DefaultQuery("auditStatus", "")

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
	whereConditions := []string{}
	args := []interface{}{}

	// 【修改】处理 auditStatus 筛选逻辑
	if auditStatusStr != "" {
		auditStatus, err := strconv.Atoi(auditStatusStr)
		if err == nil {
			whereConditions = append(whereConditions, "audit_status = ?")
			args = append(args, auditStatus)
		}
	} else {
		// 默认只展示已通过的机构
		whereConditions = append(whereConditions, "audit_status = 1")
	}

	if provinceCode != "" && provinceCode != "all" {
		whereConditions = append(whereConditions, "province_code = ?")
		args = append(args, provinceCode)
	}
	if cityCode != "" {
		whereConditions = append(whereConditions, "city_code = ?")
		args = append(args, cityCode)
	}
	if districtCode != "" {
		whereConditions = append(whereConditions, "district_code = ?")
		args = append(args, districtCode)
	}
	if consultWay != "" {
		whereConditions = append(whereConditions, "consult_way = ?")
		args = append(args, consultWay)
	}

	// 疾病筛选：通过关联表
	if diseaseStr != "" {
		diseaseID, _ := strconv.Atoi(diseaseStr)
		if diseaseID > 0 {
			whereConditions = append(whereConditions, "id IN (SELECT org_id FROM psych_support_org_disease_rel WHERE disease_id = ?)")
			args = append(args, diseaseID)
		}
	}

	if isFreeStr != "" {
		isFree := 0
		if isFreeStr == "true" {
			isFree = 1
		}
		whereConditions = append(whereConditions, "is_free = ?")
		args = append(args, isFree)
	}

	if keyword != "" {
		whereConditions = append(whereConditions, "(name LIKE ? OR content_intro LIKE ?)")
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 拼接 WHERE 子句
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 1. 查询总数
	countQuery := "SELECT COUNT(*) FROM psych_support_org " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败: " + err.Error(),
		})
		return
	}

	// 2. 查询列表主数据
	// 【修改】SQL 中增加了 province_name, city_name, district_name, audit_status, reject_reason, created_at, updated_at
	listQuery := `
		SELECT id, name, province_code, city_code, district_code, 
		       province_name, city_name, district_name,
		       address, contact_phone, contact_url, is_free, consult_way, content_intro, 
		       audit_status, reject_reason, created_at, updated_at
		FROM psych_support_org
		` + whereClause + `
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var list []PsychologicalOrgItem
	var orgIDs []uint64

	// 临时存储扫描结果
	type tempOrg struct {
		ID           uint
		Name         string
		ProvinceCode string
		CityCode     string
		DistrictCode sql.NullString
		ProvinceName sql.NullString
		CityName     sql.NullString
		DistrictName sql.NullString
		Address      string
		ContactPhone string
		ContactUrl   string
		IsFree       sql.NullInt32
		ConsultWay   string
		ContentIntro string
		AuditStatus  int8
		RejectReason sql.NullString
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}

	var tempList []tempOrg

	for rows.Next() {
		var t tempOrg
		if err := rows.Scan(
			&t.ID, &t.Name, &t.ProvinceCode, &t.CityCode, &t.DistrictCode,
			&t.ProvinceName, &t.CityName, &t.DistrictName,
			&t.Address, &t.ContactPhone, &t.ContactUrl, &t.IsFree, &t.ConsultWay,
			&t.ContentIntro, &t.AuditStatus, &t.RejectReason, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			continue
		}
		tempList = append(tempList, t)
		orgIDs = append(orgIDs, uint64(t.ID))
	}

	// 3. 批量查询疾病关联 (解决 N+1 问题)
	diseaseIdsMap := make(map[uint64][]uint64)
	diseaseDetailsMap := make(map[uint64][]InstitutionDiseaseItem)

	if len(orgIDs) > 0 {
		placeholders := make([]string, len(orgIDs))
		queryArgs := make([]interface{}, len(orgIDs))
		for i, id := range orgIDs {
			placeholders[i] = "?"
			queryArgs[i] = id
		}
		placeholderStr := strings.Join(placeholders, ",")

		// JOIN disease 表获取详细信息
		relQuery := fmt.Sprintf(`
			SELECT r.org_id, d.id, d.name, d.alias 
			FROM psych_support_org_disease_rel r
			INNER JOIN disease d ON r.disease_id = d.id
			WHERE r.org_id IN (%s)
			ORDER BY r.org_id, d.id ASC
		`, placeholderStr)

		relRows, err := db.MySQL.Query(relQuery, queryArgs...)
		if err == nil {
			defer relRows.Close()
			for relRows.Next() {
				var oID uint64
				var dItem InstitutionDiseaseItem
				if err := relRows.Scan(&oID, &dItem.ID, &dItem.Name, &dItem.Alias); err == nil {
					diseaseIdsMap[oID] = append(diseaseIdsMap[oID], uint64(dItem.ID))
					diseaseDetailsMap[oID] = append(diseaseDetailsMap[oID], dItem)
				}
			}
		}
	}

	// 4. 组装最终返回数据
	for _, t := range tempList {
		isFree := false
		if t.IsFree.Valid && t.IsFree.Int32 == 1 {
			isFree = true
		}

		// 处理 Null 字符串字段
		provinceName := ""
		if t.ProvinceName.Valid {
			provinceName = t.ProvinceName.String
		}
		cityName := ""
		if t.CityName.Valid {
			cityName = t.CityName.String
		}
		districtName := ""
		if t.DistrictName.Valid {
			districtName = t.DistrictName.String
		}

		var rejectReasonPtr *string
		if t.RejectReason.Valid {
			rejectReasonPtr = &t.RejectReason.String
		}

		// 提取地区显示 (兼容旧逻辑)
		regionDisplay := t.ProvinceCode
		if cityName != "" {
			regionDisplay += " " + cityName
		}

		// 转换机构类型 (复用现有逻辑)
		orgType, orgTypeName := convertPsychologicalOrgType(t.Name, t.ConsultWay)
		services := parsePsychologicalServices(t.Name, t.ConsultWay)
		serviceTime := getServiceTime(orgType)

		// 获取疾病数据
		dids := diseaseIdsMap[uint64(t.ID)]
		if dids == nil {
			dids = []uint64{}
		}
		dDetails := diseaseDetailsMap[uint64(t.ID)]
		if dDetails == nil {
			dDetails = []InstitutionDiseaseItem{}
		}

		item := PsychologicalOrgItem{
			ID:           t.ID,
			Name:         t.Name,
			ProvinceCode: t.ProvinceCode,
			CityCode:     t.CityCode,
			DistrictCode: "", // 如果 DistrictCode 是 NullString，这里需要处理 .String
			ProvinceName: provinceName,
			CityName:     cityName,
			DistrictName: districtName,
			Address:      t.Address,
			ContactPhone: t.ContactPhone,
			ContactUrl:   t.ContactUrl,
			IsFree:       isFree,
			ConsultWay:   t.ConsultWay,
			ContentIntro: t.ContentIntro,
			AuditStatus:  t.AuditStatus,
			RejectReason: rejectReasonPtr,

			// 保留原有计算字段
			Type:        orgType,
			TypeName:    orgTypeName,
			Region:      regionDisplay,
			RegionCode:  t.ProvinceCode,
			ServiceTime: serviceTime,
			Description: t.ContentIntro,
			Services:    services,
			Rating:      4.5 + float64(t.ID%10)/10,
			CoverUrl:    "https://example.com/orgs/psychological/" + strconv.FormatUint(uint64(t.ID), 10) + ".jpg",
			Status:      "active",

			// 新增关联数据
			DiseaseIds: dids,
			Diseases:   dDetails,
			CreatedAt:  t.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  t.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		list = append(list, item)
	}

	if list == nil {
		list = []PsychologicalOrgItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": PsychologicalOrgListResponse{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// GetPsychologicalOrgDetail 获取心理咨询机构详情
func GetPsychologicalOrgDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的机构 ID",
		})
		return
	}

	// 1. 查询机构基础信息
	// 【修改】SQL 中增加了 province_name, city_name, district_name, audit_status, reject_reason, created_at, updated_at
	query := `
		SELECT id, name, province_code, city_code, district_code, 
		       province_name, city_name, district_name,
		       address, contact_phone, contact_url, is_free, consult_way, content_intro, 
		       audit_status, reject_reason, created_at, updated_at
		FROM psych_support_org
		WHERE id = ?
	`
	// 注意：这里移除了 AND audit_status = 1，通常详情页允许查看待审核或驳回的内容（视业务权限而定）
	// 如果必须只展示已通过的，请加回该条件

	var org struct {
		ID           uint
		Name         string
		ProvinceCode string
		CityCode     string
		DistrictCode sql.NullString
		ProvinceName sql.NullString
		CityName     sql.NullString
		DistrictName sql.NullString
		Address      string
		ContactPhone string
		ContactUrl   string
		IsFree       sql.NullInt32
		ConsultWay   string
		ContentIntro string
		AuditStatus  int8
		RejectReason sql.NullString
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&org.ID, &org.Name, &org.ProvinceCode, &org.CityCode, &org.DistrictCode,
		&org.ProvinceName, &org.CityName, &org.DistrictName,
		&org.Address, &org.ContactPhone, &org.ContactUrl, &org.IsFree, &org.ConsultWay,
		&org.ContentIntro, &org.AuditStatus, &org.RejectReason, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "机构不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询机构详情失败: " + err.Error(),
		})
		return
	}

	isFree := false
	if org.IsFree.Valid && org.IsFree.Int32 == 1 {
		isFree = true
	}

	// 2. 【核心修改】查询关联的疾病 ID 列表和详细信息
	// 使用 JOIN 一次性获取 id, name, alias
	diseaseQuery := `
		SELECT d.id, d.name, d.alias 
		FROM psych_support_org_disease_rel r
		INNER JOIN disease d ON r.disease_id = d.id
		WHERE r.org_id = ?
		ORDER BY d.id ASC
	`

	rows, err := db.MySQL.Query(diseaseQuery, id)
	if err != nil {
		// 记录错误但不中断主流程，疾病列表将为空
		// log.Printf("Query org diseases error: %v", err)
		rows = nil
	}

	var diseaseIds []uint64
	var diseases []InstitutionDiseaseItem

	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var d InstitutionDiseaseItem
			if err := rows.Scan(&d.ID, &d.Name, &d.Alias); err == nil {
				diseaseIds = append(diseaseIds, uint64(d.ID))
				diseases = append(diseases, d)
			}
		}
	}

	// 确保切片不为 nil，返回空数组而不是 null
	if diseaseIds == nil {
		diseaseIds = []uint64{}
	}
	if diseases == nil {
		diseases = []InstitutionDiseaseItem{}
	}

	// 3. 处理 Null 字符串字段
	provinceName := ""
	if org.ProvinceName.Valid {
		provinceName = org.ProvinceName.String
	}
	cityName := ""
	if org.CityName.Valid {
		cityName = org.CityName.String
	}
	districtName := ""
	if org.DistrictName.Valid {
		districtName = org.DistrictName.String
	}

	var rejectReasonPtr *string
	if org.RejectReason.Valid {
		rejectReasonPtr = &org.RejectReason.String
	}

	// 解析联系方式
	phone := org.ContactPhone
	website := org.ContactUrl

	// 提取地区显示 (兼容旧逻辑)
	regionDisplay := org.ProvinceCode
	if cityName != "" {
		regionDisplay += " " + cityName
	}

	// 转换机构类型
	orgType, orgTypeName := convertPsychologicalOrgType(org.Name, org.ConsultWay)

	// 解析服务项目
	services := parsePsychologicalServices(org.Name, org.ConsultWay)

	// 获取服务时间
	serviceTime := getServiceTime(orgType)

	// 构建图片列表 (示例)
	images := []string{
		"https://example.com/orgs/psychological/" + strconv.FormatUint(id, 10) + "_1.jpg",
	}

	// 构建咨询师列表
	counselors := getPsychologicalCounselors(org.ID)

	// 4. 构造响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": PsychologicalOrgDetailResponse{
			ID:           org.ID,
			Name:         org.Name,
			ProvinceCode: org.ProvinceCode,
			CityCode:     org.CityCode,
			DistrictCode: "", // 如果需要 DistrictCode 的值，需处理 org.DistrictCode.String
			ProvinceName: provinceName,
			CityName:     cityName,
			DistrictName: districtName,
			Address:      org.Address,
			ContactPhone: phone,
			ContactUrl:   website,
			IsFree:       isFree,
			ConsultWay:   org.ConsultWay,
			ContentIntro: org.ContentIntro,
			AuditStatus:  org.AuditStatus,
			RejectReason: rejectReasonPtr,

			// 保留原有计算字段
			Type:        orgType,
			TypeName:    orgTypeName,
			Region:      regionDisplay,
			RegionCode:  org.ProvinceCode,
			ServiceTime: serviceTime,
			Description: org.ContentIntro,
			Services:    services,
			Rating:      4.5 + float64(org.ID%10)/10,
			CoverUrl:    "https://example.com/orgs/psychological/" + strconv.FormatUint(id, 10) + ".jpg",
			Status:      "active",
			Images:      images,
			Counselors:  counselors,

			// 新增关联数据
			DiseaseIds:   diseaseIds,
			Diseases:     diseases,
			DiseaseCount: len(diseases),
			CreatedAt:    org.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    org.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

// convertPsychologicalOrgType 转换心理咨询机构类型
func convertPsychologicalOrgType(name, consultWay string) (string, string) {
	if strings.Contains(name, "热线") {
		return "hotline", "心理热线"
	}
	if strings.Contains(name, "中心") {
		return "center", "心理中心"
	}
	if strings.Contains(name, "医院") {
		return "hospital", "心理医院"
	}
	if consultWay == "线上" {
		return "online", "在线咨询"
	}
	return "center", "心理中心"
}

// parsePsychologicalContact 解析心理咨询联系方式
func parsePsychologicalContact(contact string) (phone, email, website string) {
	// 简化处理，实际可根据格式解析
	phone = contact
	if strings.Contains(contact, "@") {
		email = contact
		phone = ""
	}
	if strings.Contains(contact, "http") {
		website = contact
		phone = ""
	}
	return
}

// parsePsychologicalServices 解析心理咨询服务项目
func parsePsychologicalServices(name, consultWay string) []string {
	if strings.Contains(name, "热线") {
		return []string{"心理疏导", "危机干预", "情绪支持"}
	}
	if strings.Contains(name, "中心") {
		return []string{"心理咨询", "心理治疗", "团体辅导", "心理测评"}
	}
	if strings.Contains(name, "医院") {
		return []string{"心理诊断", "心理治疗", "药物治疗", "康复指导"}
	}
	return []string{"心理咨询", "心理支持"}
}

// getServiceTime 获取服务时间
func getServiceTime(orgType string) string {
	if orgType == "hotline" {
		return "24 小时"
	}
	return "周一至周日 08:00-17:00"
}

// getPsychologicalCounselors 获取心理咨询师列表
func getPsychologicalCounselors(orgID uint) []CounselorItem {
	// 实际项目中可从咨询师表查询
	// 这里返回示例数据
	counselorsMap := map[uint][]CounselorItem{
		1: {
			{Name: "李老师", Title: "资深心理咨询师", Specialty: "危机干预、创伤治疗"},
			{Name: "张老师", Title: "心理咨询师", Specialty: "情绪管理、压力疏导"},
		},
		2: {
			{Name: "王医生", Title: "心理治疗师", Specialty: "心理治疗、认知行为疗法"},
			{Name: "赵医生", Title: "主任医师", Specialty: "精神障碍诊断与治疗"},
		},
		3: {
			{Name: "陈老师", Title: "心理咨询师", Specialty: "生命教育、心理支持"},
		},
	}

	if counselors, ok := counselorsMap[orgID]; ok {
		return counselors
	}
	return []CounselorItem{
		{Name: "刘老师", Title: "心理咨询师", Specialty: "心理咨询、情绪疏导"},
	}
}

// TargetItem 目标人群选项项
type TargetItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// TargetResponse 目标人群响应结构
type TargetResponse struct {
	Targets []TargetItem `json:"targets"`
}

// OrgTypeItem 机构类型选项项
type OrgTypeItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// OrgTypeResponse 机构类型响应结构
type OrgTypeResponse struct {
	Types []OrgTypeItem `json:"types"`
}

// GetGuideTargets 获取指南目标人群筛选选项
func GetGuideTargets(c *gin.Context) {
	targets := []TargetItem{
		{Text: "全部人群", Value: "all"},
		{Text: "患者", Value: "patient"},
		{Text: "家属", Value: "family"},
		{Text: "儿童", Value: "child"},
		{Text: "青少年", Value: "teenager"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": TargetResponse{
			Targets: targets,
		},
	})
}

// GetPsychologicalOrgTypes 获取心理咨询机构类型筛选选项
func GetPsychologicalOrgTypes(c *gin.Context) {
	types := []OrgTypeItem{
		{Text: "全部类型", Value: "all"},
		{Text: "心理热线", Value: "hotline"},
		{Text: "心理中心", Value: "center"},
		{Text: "心理医院", Value: "hospital"},
		{Text: "咨询机构", Value: "clinic"},
		{Text: "在线咨询", Value: "online"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": OrgTypeResponse{
			Types: types,
		},
	})
}

// InstitutionOptionsResponse 机构选项响应结构
type InstitutionOptionsResponse struct {
	Regions  []RegionItem  `json:"regions"`
	Types    []OrgTypeItem `json:"types"`
	Diseases []OptionItem  `json:"diseases"`
}

// GetInstitutionOptions 获取康复机构筛选选项
func GetInstitutionOptions(c *gin.Context) {
	// 1. 获取地区选项 (从现有数据中提取不重复的省/市)
	// 注意：实际生产中建议维护一张独立的 region 字典表，这里仅演示从业务表提取
	provinceQuery := "SELECT DISTINCT province FROM rehab_institution WHERE audit_status = 1 ORDER BY province ASC"
	provinceRows, err := db.MySQL.Query(provinceQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询省份选项失败",
		})
		return
	}
	defer provinceRows.Close()

	var regions []RegionItem
	// 添加全部选项
	regions = append(regions, RegionItem{Text: "全部地区", Value: "all"})

	for provinceRows.Next() {
		var province string
		if err := provinceRows.Scan(&province); err != nil {
			continue
		}
		if province != "" {
			regions = append(regions, RegionItem{
				Text:  province,
				Value: convertRegionToCode(province), // 复用现有的转换函数，或根据需要调整
			})
		}
	}

	// 2. 机构类型选项 (硬编码或从字典表获取，SQL中未体现类型字段，暂保留硬编码或根据名称判断的逻辑)
	types := []OrgTypeItem{
		{Text: "全部类型", Value: "all"},
		{Text: "康复医院", Value: "hospital"},
		{Text: "康复中心", Value: "center"},
		{Text: "康复诊所", Value: "clinic"},
		{Text: "社区康复站", Value: "community"},
	}

	// 3. 疾病选项 (从 disease_options 表获取，与原逻辑一致)
	diseaseQuery := `
		SELECT value, name 
		FROM disease_options 
		WHERE is_enabled = 1 
		ORDER BY sort ASC
	`
	diseaseRows, err := db.MySQL.Query(diseaseQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询疾病选项失败",
		})
		return
	}
	defer diseaseRows.Close()

	var diseases []OptionItem
	for diseaseRows.Next() {
		var value int
		var name string
		if err := diseaseRows.Scan(&value, &name); err != nil {
			continue
		}
		diseases = append(diseases, OptionItem{
			Text:  name,
			Value: convertValueToDisease(value),
		})
	}

	if diseases == nil {
		diseases = []OptionItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": InstitutionOptionsResponse{
			Regions:  regions,
			Types:    types,
			Diseases: diseases,
		},
	})
}

// CreateInstitutionRequest 创建康复机构请求结构
type CreateInstitutionRequest struct {
	Name          string `json:"name" binding:"required"`
	ProvinceCode  string `json:"provinceCode" binding:"required"`
	CityCode      string `json:"cityCode" binding:"required"`
	DistrictCode  string `json:"districtCode"`
	ProvinceName  string `json:"provinceName"` // 【新增】接收前端传入的名称
	CityName      string `json:"cityName"`     // 【新增】
	DistrictName  string `json:"districtName"` // 【新增】
	Qualification string `json:"qualification"`
	RehabProjects string `json:"rehabProjects" binding:"required"`
	FeeStandard   string `json:"feeStandard" binding:"required"`
	ContactPhone  string `json:"contactPhone"`
	ContactUrl    string `json:"contactUrl"`
	Address       string `json:"address" binding:"required"`
	DiseaseIds    []int  `json:"diseaseIds"`
	// 注意：创建时通常默认 audit_status 为 0 或 1，由后端控制，或者前端传入
	AuditStatus  int    `json:"auditStatus"`
	RejectReason string `json:"rejectReason"`
}

// CreateInstitution 新增康复机构
func CreateInstitution(c *gin.Context) {
	var req CreateInstitutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务启动失败",
		})
		return
	}
	defer tx.Rollback()

	// 1. 插入主表
	// 【修改】SQL 中增加了 province_name, city_name, district_name, audit_status, reject_reason, sort
	insertQuery := `
		INSERT INTO rehab_institution 
		(name, province_code, city_code, district_code, 
		 province_name, city_name, district_name,
		 qualification, rehab_projects, fee_standard, 
		 contact_phone, contact_url, address, 
		 audit_status, reject_reason,
		 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	// 如果前端没传 auditStatus，默认设为 0 (待审核) 或 1 (已通过)，这里取前端传入值，若为0则默认0
	initialAuditStatus := req.AuditStatus
	if initialAuditStatus == 0 && req.AuditStatus == 0 {
		// 可以根据业务需求调整默认值，例如默认 0
		initialAuditStatus = 0
	}

	result, err := tx.Exec(insertQuery,
		req.Name, req.ProvinceCode, req.CityCode, req.DistrictCode,
		req.ProvinceName, req.CityName, req.DistrictName, // 【新增】插入名称字段
		req.Qualification, req.RehabProjects, req.FeeStandard,
		req.ContactPhone, req.ContactUrl, req.Address,
		initialAuditStatus, req.RejectReason) // 【新增】插入审核状态、驳回原因和排序

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建康复机构失败: " + err.Error(),
		})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取新增ID失败",
		})
		return
	}

	// 2. 插入疾病关联
	if len(req.DiseaseIds) > 0 {
		relQuery := "INSERT INTO rehab_institution_disease_rel (institution_id, disease_id) VALUES (?, ?)"
		for _, did := range req.DiseaseIds {
			if did > 0 {
				_, err := tx.Exec(relQuery, id, did)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    500,
						"message": "关联疾病失败",
					})
					return
				}
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data": gin.H{
			"id": id,
		},
	})
}

// UpdateInstitutionRequest 更新康复机构请求结构
type UpdateInstitutionRequest struct {
	Name          string  `json:"name"`
	ProvinceCode  *string `json:"provinceCode"` // 使用指针以便判断是否传值
	CityCode      *string `json:"cityCode"`     // 使用指针
	DistrictCode  *string `json:"districtCode"` // 使用指针
	Qualification string  `json:"qualification"`
	RehabProjects string  `json:"rehabProjects"`
	FeeStandard   string  `json:"feeStandard"`
	ContactPhone  string  `json:"contactPhone"`
	ContactUrl    string  `json:"contactUrl"`
	Address       string  `json:"address"`
	DiseaseIds    []int   `json:"diseaseIds"`
	Sort          int     `json:"sort"`
	AuditStatus   *int    `json:"auditStatus"`  // 【修改】改为指针，以便区分“未传”和“传了0”
	RejectReason  *string `json:"rejectReason"` // 【新增】如果数据库有 reject_reason 字段，加上这个
	ProvinceName  *string `json:"provinceName"` // 使用指针以便判断是否传值
	CityName      *string `json:"cityName"`     // 使用指针
	DistrictName  *string `json:"districtName"` // 使用指针
}

// UpdateInstitution 更新康复机构
func UpdateInstitution(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的机构 ID",
		})
		return
	}

	var req UpdateInstitutionRequest
	// 绑定 JSON 数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数解析错误: " + err.Error(),
		})
		return
	}

	// 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务启动失败",
		})
		return
	}
	defer tx.Rollback()

	// 1. 构建动态更新主表语句
	updateFields := []string{}
	args := []interface{}{}

	if req.Name != "" {
		updateFields = append(updateFields, "name = ?")
		args = append(args, req.Name)
	}

	// 处理地区 Code 更新 (使用指针判断是否传递)
	if req.ProvinceCode != nil {
		updateFields = append(updateFields, "province_code = ?")
		args = append(args, *req.ProvinceCode)
	}
	if req.CityCode != nil {
		updateFields = append(updateFields, "city_code = ?")
		args = append(args, *req.CityCode)
	}
	if req.DistrictCode != nil {
		updateFields = append(updateFields, "district_code = ?")
		args = append(args, *req.DistrictCode)
	}
	if req.ProvinceName != nil {
		updateFields = append(updateFields, "province_name = ?")
		args = append(args, *req.ProvinceName)
	}
	if req.CityName != nil {
		updateFields = append(updateFields, "city_name = ?")
		args = append(args, *req.CityName)
	}
	if req.DistrictName != nil {
		updateFields = append(updateFields, "district_name= ?")
		args = append(args, *req.DistrictName)
	}

	if req.Qualification != "" {
		updateFields = append(updateFields, "qualification = ?")
		args = append(args, req.Qualification)
	}
	if req.RehabProjects != "" {
		updateFields = append(updateFields, "rehab_projects = ?")
		args = append(args, req.RehabProjects)
	}
	if req.FeeStandard != "" {
		updateFields = append(updateFields, "fee_standard = ?")
		args = append(args, req.FeeStandard)
	}
	if req.ContactPhone != "" {
		updateFields = append(updateFields, "contact_phone = ?")
		args = append(args, req.ContactPhone)
	}
	if req.ContactUrl != "" {
		updateFields = append(updateFields, "contact_url = ?")
		args = append(args, req.ContactUrl)
	}
	if req.Address != "" {
		updateFields = append(updateFields, "address = ?")
		args = append(args, req.Address)
	}

	// 排序权重 (如果业务允许 sort 为 0，建议也改为指针 *int)
	// 这里假设 sort 为 0 代表不更新，或者根据实际需求调整
	if req.Sort != 0 {
		updateFields = append(updateFields, "sort = ?")
		args = append(args, req.Sort)
	}

	// 【修改】审核状态更新：使用指针判断是否传递
	if req.AuditStatus != nil {
		updateFields = append(updateFields, "audit_status = ?")
		args = append(args, *req.AuditStatus)
	}

	// 【新增】驳回原因更新：如果结构体中加了 RejectReason 且数据库有对应字段
	if req.RejectReason != nil {
		updateFields = append(updateFields, "reject_reason = ?")
		args = append(args, *req.RejectReason)
	}

	// 如果有字段需要更新
	if len(updateFields) > 0 {
		updateFields = append(updateFields, "updated_at = NOW()")
		args = append(args, id)

		updateQuery := `UPDATE rehab_institution SET ` + strings.Join(updateFields, ", ") + ` WHERE id = ?`
		_, err = tx.Exec(updateQuery, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新机构信息失败: " + err.Error(),
			})
			return
		}
	}

	// 2. 同步疾病关联 (如果前端传了 DiseaseIds 字段)
	// 注意：JSON 反序列化时，如果前端没传 diseaseIds，req.DiseaseIds 为 nil。
	// 如果前端传了 []，则为空切片，代表清空关联。
	if req.DiseaseIds != nil {
		// 先删除旧关联
		_, err := tx.Exec("DELETE FROM rehab_institution_disease_rel WHERE institution_id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "清理旧疾病关联失败",
			})
			return
		}

		// 再插入新关联
		if len(req.DiseaseIds) > 0 {
			relQuery := "INSERT INTO rehab_institution_disease_rel (institution_id, disease_id) VALUES (?, ?)"
			for _, did := range req.DiseaseIds {
				if did > 0 {
					_, err := tx.Exec(relQuery, id, did)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"code":    500,
							"message": "创建新疾病关联失败",
						})
						return
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    nil,
	})
}

// DeleteInstitution 删除康复机构（物理删除）
func DeleteInstitution(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的机构 ID",
		})
		return
	}

	// 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务启动失败",
		})
		return
	}
	defer tx.Rollback()

	// 1. 先删除关联表数据 (因为外键约束 ON DELETE RESTRICT，必须先删子表)
	_, err = tx.Exec("DELETE FROM rehab_institution_disease_rel WHERE institution_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除关联数据失败",
		})
		return
	}

	// 2. 删除主表数据
	_, err = tx.Exec("DELETE FROM rehab_institution WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除机构失败",
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    nil,
	})
}

// PsychologicalOrgOptionsResponse 心理咨询机构选项响应结构
type PsychologicalOrgOptionsResponse struct {
	Regions  []RegionItem  `json:"regions"`
	Types    []OrgTypeItem `json:"types"`
	Diseases []OptionItem  `json:"diseases"`
}

// CreatePsychologicalOrgRequest 创建心理咨询机构请求结构
type CreatePsychologicalOrgRequest struct {
	Name         string `json:"name" binding:"required"`
	ProvinceCode string `json:"provinceCode" binding:"required"` // 修改为 Code
	CityCode     string `json:"cityCode" binding:"required"`     // 修改为 Code
	DistrictCode string `json:"districtCode"`                    // 修改为 Code
	Address      string `json:"address"`
	ContactPhone string `json:"contactPhone"`
	ContactUrl   string `json:"contactUrl"`
	IsFree       bool   `json:"isFree"`
	ConsultWay   string `json:"consultWay"`
	ContentIntro string `json:"contentIntro" binding:"required"`
	DiseaseIds   []int  `json:"diseaseIds"`
	Sort         int    `json:"sort"`
}

// CreatePsychologicalOrg 新增心理咨询机构
func CreatePsychologicalOrg(c *gin.Context) {
	var req CreatePsychologicalOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务启动失败",
		})
		return
	}
	defer tx.Rollback()

	// 1. 插入主表
	insertQuery := `
		INSERT INTO psych_support_org 
		(name, province_code, city_code, district_code, address, contact_phone, contact_url,
		 is_free, consult_way, content_intro, audit_status,
		 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW())
	`
	isFreeInt := 0
	if req.IsFree {
		isFreeInt = 1
	}

	result, err := tx.Exec(insertQuery,
		req.Name, req.ProvinceCode, req.CityCode, req.DistrictCode, req.Address, // 修改参数顺序和内容
		req.ContactPhone, req.ContactUrl, isFreeInt, req.ConsultWay, req.ContentIntro)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建心理咨询机构失败",
		})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取新增ID失败",
		})
		return
	}

	// 2. 插入疾病关联
	if len(req.DiseaseIds) > 0 {
		relQuery := "INSERT INTO psych_support_org_disease_rel (org_id, disease_id) VALUES (?, ?)"
		for _, did := range req.DiseaseIds {
			_, err := tx.Exec(relQuery, id, did)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "关联疾病失败",
				})
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data": gin.H{
			"id": id,
		},
	})
}

// UpdatePsychologicalOrgRequest 更新心理咨询机构请求结构
type UpdatePsychologicalOrgRequest struct {
	Name         string  `json:"name"`
	ProvinceCode *string `json:"provinceCode"` // 修改为 Code 指针
	CityCode     *string `json:"cityCode"`     // 修改为 Code 指针
	DistrictCode *string `json:"districtCode"` // 修改为 Code 指针
	Address      string  `json:"address"`
	ContactPhone string  `json:"contactPhone"`
	ContactUrl   string  `json:"contactUrl"`
	IsFree       *bool   `json:"isFree"`
	ConsultWay   string  `json:"consultWay"`
	ContentIntro string  `json:"contentIntro"`
	DiseaseIds   []int   `json:"diseaseIds"`
	AuditStatus  int     `json:"auditStatus"`
}

// UpdatePsychologicalOrg 更新心理咨询机构
func UpdatePsychologicalOrg(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的机构 ID",
		})
		return
	}

	var req UpdatePsychologicalOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务启动失败",
		})
		return
	}
	defer tx.Rollback()

	// 1. 构建动态更新主表语句
	updateFields := []string{}
	args := []interface{}{}

	if req.Name != "" {
		updateFields = append(updateFields, "name = ?")
		args = append(args, req.Name)
	}
	// 【修改】处理 Code 字段更新
	if req.ProvinceCode != nil {
		updateFields = append(updateFields, "province_code = ?")
		args = append(args, *req.ProvinceCode)
	}
	if req.CityCode != nil {
		updateFields = append(updateFields, "city_code = ?")
		args = append(args, *req.CityCode)
	}
	if req.DistrictCode != nil {
		updateFields = append(updateFields, "district_code = ?")
		args = append(args, *req.DistrictCode)
	}

	if req.Address != "" {
		updateFields = append(updateFields, "address = ?")
		args = append(args, req.Address)
	}
	if req.ContactPhone != "" {
		updateFields = append(updateFields, "contact_phone = ?")
		args = append(args, req.ContactPhone)
	}
	if req.ContactUrl != "" {
		updateFields = append(updateFields, "contact_url = ?")
		args = append(args, req.ContactUrl)
	}
	if req.IsFree != nil {
		isFreeInt := 0
		if *req.IsFree {
			isFreeInt = 1
		}
		updateFields = append(updateFields, "is_free = ?")
		args = append(args, isFreeInt)
	}
	if req.ConsultWay != "" {
		updateFields = append(updateFields, "consult_way = ?")
		args = append(args, req.ConsultWay)
	}
	if req.ContentIntro != "" {
		updateFields = append(updateFields, "content_intro = ?")
		args = append(args, req.ContentIntro)
	}
	if req.AuditStatus != 0 {
		updateFields = append(updateFields, "audit_status = ?")
		args = append(args, req.AuditStatus)
	}

	updateFields = append(updateFields, "updated_at = NOW()")
	args = append(args, id)

	if len(updateFields) > 1 {
		updateQuery := `UPDATE psych_support_org SET ` + strings.Join(updateFields, ", ") + ` WHERE id = ?`
		_, err = tx.Exec(updateQuery, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新机构信息失败",
			})
			return
		}
	}

	// 2. 同步疾病关联
	if req.DiseaseIds != nil {
		// 先删除旧关联
		_, err := tx.Exec("DELETE FROM psych_support_org_disease_rel WHERE org_id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "清理旧疾病关联失败",
			})
			return
		}

		// 再插入新关联
		if len(req.DiseaseIds) > 0 {
			relQuery := "INSERT INTO psych_support_org_disease_rel (org_id, disease_id) VALUES (?, ?)"
			for _, did := range req.DiseaseIds {
				_, err := tx.Exec(relQuery, id, did)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    500,
						"message": "创建新疾病关联失败",
					})
					return
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    nil,
	})
}

// DeletePsychologicalOrg 删除心理咨询机构（物理删除）
func DeletePsychologicalOrg(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的机构 ID",
		})
		return
	}

	// 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务启动失败",
		})
		return
	}
	defer tx.Rollback()

	// 1. 先删除关联表数据
	_, err = tx.Exec("DELETE FROM psych_support_org_disease_rel WHERE org_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除关联数据失败",
		})
		return
	}

	// 2. 删除主表数据
	_, err = tx.Exec("DELETE FROM psych_support_org WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除机构失败",
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    nil,
	})
}

// TrainingItem 训练指南项响应结构
type TrainingItem struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	Stage    string `json:"stage"`
	Disease  string `json:"disease"`
	Desc     string `json:"desc"`
	CoverUrl string `json:"coverUrl"`
}

// TrainingListResponse 列表响应结构
type TrainingListResponse struct {
	List     []TrainingItem `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

// TrainingDetailResponse 详情响应结构
type TrainingDetailResponse struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	VideoUrl   string `json:"videoUrl"`
	Duration   string `json:"duration"`
	Difficulty string `json:"difficulty"`
	Purpose    string `json:"purpose"`
	Forbidden  string `json:"forbidden"`
	PicUrls    string `json:"picUrls"`
}

// TrainingDiseaseItem 训练指南关联疾病详情项
type TrainingDiseaseItem struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

// TrainingListItem 训练指南列表项响应结构
type TrainingListItem struct {
	ID              uint                  `json:"id"`
	RehabStage      string                `json:"rehabStage"`
	Title           string                `json:"title"`
	TrainPurpose    string                `json:"trainPurpose"`
	TrainContent    string                `json:"trainContent"`
	ForbiddenAction string                `json:"forbiddenAction"`
	PicUrls         []string              `json:"picUrls"` // 前端期望数组，数据库存 JSON 字符串需解析
	GuidePdf        string                `json:"guidePdf"`
	GuideWord       string                `json:"guideWord"`
	AuditStatus     int8                  `json:"auditStatus"`
	RejectReason    string                `json:"rejectReason"`
	Sort            int                   `json:"sort"`
	DiseaseIds      []uint64              `json:"diseaseIds"`
	Diseases        []TrainingDiseaseItem `json:"diseases"`
	CreatedAt       string                `json:"createdAt"`
	UpdatedAt       string                `json:"updatedAt"`
}

type TrainingListDataResponse struct {
	List  []TrainingListItem `json:"list"`
	Total int64              `json:"total"`
}

// GetTrainingList 获取训练指南列表
func GetTrainingList(c *gin.Context) {
	// 获取请求参数
	diseaseStr := c.DefaultQuery("diseaseId", "")
	stage := c.DefaultQuery("rehabStage", "")
	keyword := c.DefaultQuery("keyword", "")

	// 【新增】获取 auditStatus 参数
	auditStatusStr := c.DefaultQuery("auditStatus", "")

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
	// 【修改】初始不固定 audit_status = 1，而是根据参数动态添加
	whereConditions := []string{}
	args := []interface{}{}

	// 【新增】处理 auditStatus 筛选逻辑
	if auditStatusStr != "" {
		auditStatus, err := strconv.Atoi(auditStatusStr)
		if err == nil {
			// 如果前端明确传了审核状态，则按该状态筛选
			whereConditions = append(whereConditions, "g.audit_status = ?")
			args = append(args, auditStatus)
		}
	} else {
		// 【重要】如果前端没传审核状态，默认只展示已通过的指南 (保持原有业务逻辑一致性)
		// 如果希望默认展示所有状态，可以注释掉下面这行
		whereConditions = append(whereConditions, "g.audit_status = 1")
	}

	// 疾病筛选
	if diseaseStr != "" {
		diseaseID, err := strconv.Atoi(diseaseStr)
		if err == nil && diseaseID > 0 {
			whereConditions = append(whereConditions, "EXISTS (SELECT 1 FROM rehab_train_guide_disease_rel r WHERE r.guide_id = g.id AND r.disease_id = ?)")
			args = append(args, diseaseID)
		}
	}

	// 阶段筛选
	if stage != "" {
		whereConditions = append(whereConditions, "g.rehab_stage = ?")
		args = append(args, stage)
	}

	// 关键字模糊筛选
	if keyword != "" {
		whereConditions = append(whereConditions, "g.title LIKE ?")
		args = append(args, "%"+keyword+"%")
	}

	// 拼接 WHERE 子句
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 1. 查询总数
	countQuery := "SELECT COUNT(*) FROM rehab_train_guide g " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败: " + err.Error(),
		})
		return
	}

	// 2. 查询列表主数据
	listQuery := `
		SELECT g.id, g.rehab_stage, g.title, g.train_purpose, g.train_content, 
		       g.forbidden_action, g.pic_urls, g.guide_pdf, g.guide_word, 
		       g.audit_status, g.reject_reason, g.sort, g.created_at, g.updated_at
		FROM rehab_train_guide g
		` + whereClause + `
		ORDER BY g.sort DESC, g.id DESC
		LIMIT ? OFFSET ?
	`

	// 追加分页参数到 args
	listArgs := append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, listArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var listItems []TrainingListItem
	var guideIDs []uint64

	// 临时存储扫描结果
	type tempGuide struct {
		ID              uint
		RehabStage      string
		Title           string
		TrainPurpose    string
		TrainContent    string
		ForbiddenAction sql.NullString
		PicUrls         sql.NullString // JSON 字符串
		GuidePdf        sql.NullString
		GuideWord       sql.NullString
		AuditStatus     int8
		RejectReason    sql.NullString
		Sort            int
		CreatedAt       time.Time
		UpdatedAt       time.Time
	}

	var tempList []tempGuide

	for rows.Next() {
		var t tempGuide
		if err := rows.Scan(
			&t.ID, &t.RehabStage, &t.Title, &t.TrainPurpose, &t.TrainContent,
			&t.ForbiddenAction, &t.PicUrls, &t.GuidePdf, &t.GuideWord,
			&t.AuditStatus, &t.RejectReason, &t.Sort, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			continue
		}
		tempList = append(tempList, t)
		guideIDs = append(guideIDs, uint64(t.ID))
	}

	// 3. 批量查询疾病关联 (解决 N+1 问题)
	diseaseIdsMap := make(map[uint64][]uint64)
	diseaseDetailsMap := make(map[uint64][]TrainingDiseaseItem)

	if len(guideIDs) > 0 {
		placeholders := make([]string, len(guideIDs))
		queryArgs := make([]interface{}, len(guideIDs))
		for i, id := range guideIDs {
			placeholders[i] = "?"
			queryArgs[i] = id
		}
		placeholderStr := strings.Join(placeholders, ",")

		// JOIN disease 表获取详细信息
		relQuery := fmt.Sprintf(`
			SELECT r.guide_id, d.id, d.name, d.alias 
			FROM rehab_train_guide_disease_rel r
			INNER JOIN disease d ON r.disease_id = d.id
			WHERE r.guide_id IN (%s)
			ORDER BY r.guide_id, d.id ASC
		`, placeholderStr)

		relRows, err := db.MySQL.Query(relQuery, queryArgs...)
		if err == nil {
			defer relRows.Close()
			for relRows.Next() {
				var gID uint64
				var dItem TrainingDiseaseItem
				if err := relRows.Scan(&gID, &dItem.ID, &dItem.Name, &dItem.Alias); err == nil {
					diseaseIdsMap[gID] = append(diseaseIdsMap[gID], uint64(dItem.ID))
					diseaseDetailsMap[gID] = append(diseaseDetailsMap[gID], dItem)
				}
			}
		}
	}

	// 4. 组装最终返回数据
	for _, t := range tempList {
		// 处理 PicUrls JSON 字符串转数组
		var picUrls []string
		if t.PicUrls.Valid && t.PicUrls.String != "" {
			picUrlsStr := t.PicUrls.String
			picUrlsStr = strings.TrimPrefix(picUrlsStr, "[")
			picUrlsStr = strings.TrimSuffix(picUrlsStr, "]")
			if picUrlsStr != "" {
				rawUrls := strings.Split(picUrlsStr, ",")
				for _, u := range rawUrls {
					u = strings.TrimSpace(u)
					u = strings.Trim(u, "\"")
					if u != "" {
						picUrls = append(picUrls, u)
					}
				}
			}
		}
		if picUrls == nil {
			picUrls = []string{}
		}

		// 获取疾病数据
		dids := diseaseIdsMap[uint64(t.ID)]
		if dids == nil {
			dids = []uint64{}
		}
		dDetails := diseaseDetailsMap[uint64(t.ID)]
		if dDetails == nil {
			dDetails = []TrainingDiseaseItem{}
		}

		// 处理 Null 字段
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

		item := TrainingListItem{
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
		}

		listItems = append(listItems, item)
	}

	if listItems == nil {
		listItems = []TrainingListItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": TrainingListDataResponse{
			List:  listItems,
			Total: total,
		},
	})
}

// TrainingDetailDataResponse 训练指南详情响应数据结构
type TrainingDetailDataResponse struct {
	ID              uint                  `json:"id"`
	RehabStage      string                `json:"rehabStage"`
	Title           string                `json:"title"`
	TrainPurpose    string                `json:"trainPurpose"`
	TrainContent    string                `json:"trainContent"`
	ForbiddenAction string                `json:"forbiddenAction"`
	PicUrls         []string              `json:"picUrls"`
	GuidePdf        string                `json:"guidePdf"`
	GuideWord       string                `json:"guideWord"`
	AuditStatus     int8                  `json:"auditStatus"`
	RejectReason    *string               `json:"rejectReason"` // 使用指针以便返回 null
	Sort            int                   `json:"sort"`
	DiseaseIds      []uint64              `json:"diseaseIds"`
	Diseases        []TrainingDiseaseItem `json:"diseases"`
	CreatedAt       string                `json:"createdAt"`
	UpdatedAt       string                `json:"updatedAt"`
}

// GetTrainingDetail 获取训练详情
func GetTrainingDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的训练 ID",
		})
		return
	}

	// 1. 查询主表详细信息
	query := `
		SELECT id, rehab_stage, title, train_purpose, train_content, 
		       forbidden_action, pic_urls, guide_pdf, guide_word, 
		       audit_status, reject_reason, sort, created_at, updated_at
		FROM rehab_train_guide
		WHERE id = ?
	`
	// 注意：这里移除了 AND audit_status = 1，通常详情页允许查看待审核或驳回的内容（视业务权限而定）
	// 如果必须只展示已通过的，请加回该条件

	var t struct {
		ID              uint
		RehabStage      string
		Title           string
		TrainPurpose    string
		TrainContent    string
		ForbiddenAction sql.NullString
		PicUrls         sql.NullString
		GuidePdf        sql.NullString
		GuideWord       sql.NullString
		AuditStatus     int8
		RejectReason    sql.NullString
		Sort            int
		CreatedAt       time.Time
		UpdatedAt       time.Time
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&t.ID, &t.RehabStage, &t.Title, &t.TrainPurpose, &t.TrainContent,
		&t.ForbiddenAction, &t.PicUrls, &t.GuidePdf, &t.GuideWord,
		&t.AuditStatus, &t.RejectReason, &t.Sort, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "训练指南不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询训练详情失败: " + err.Error(),
		})
		return
	}

	// 2. 查询关联的疾病详情
	diseaseQuery := `
		SELECT d.id, d.name, d.alias 
		FROM rehab_train_guide_disease_rel r
		INNER JOIN disease d ON r.disease_id = d.id
		WHERE r.guide_id = ?
		ORDER BY d.id ASC
	`
	rows, err := db.MySQL.Query(diseaseQuery, id)
	if err != nil {
		// 记录错误但不中断主流程
		// log.Printf("Query training diseases error: %v", err)
		rows = nil
	}

	var diseaseIds []uint64
	var diseases []TrainingDiseaseItem

	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var d TrainingDiseaseItem
			if err := rows.Scan(&d.ID, &d.Name, &d.Alias); err == nil {
				diseaseIds = append(diseaseIds, uint64(d.ID))
				diseases = append(diseases, d)
			}
		}
	}

	// 确保切片不为 nil
	if diseaseIds == nil {
		diseaseIds = []uint64{}
	}
	if diseases == nil {
		diseases = []TrainingDiseaseItem{}
	}

	// 3. 处理 PicUrls JSON 字符串转数组
	var picUrls []string
	if t.PicUrls.Valid && t.PicUrls.String != "" {
		picUrlsStr := t.PicUrls.String
		// 简单解析 JSON 数组字符串 ["url1", "url2"]
		picUrlsStr = strings.TrimPrefix(picUrlsStr, "[")
		picUrlsStr = strings.TrimSuffix(picUrlsStr, "]")
		if picUrlsStr != "" {
			rawUrls := strings.Split(picUrlsStr, ",")
			for _, u := range rawUrls {
				u = strings.TrimSpace(u)
				u = strings.Trim(u, "\"")
				if u != "" {
					picUrls = append(picUrls, u)
				}
			}
		}
	}
	if picUrls == nil {
		picUrls = []string{}
	}

	// 4. 处理 Null 字段和指针
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

	// 5. 构造响应
	resp := TrainingDetailDataResponse{
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
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    resp,
	})
}

// CreateTrainingRequest 创建训练指南请求结构
type CreateTrainingRequest struct {
	Title           string   `json:"title" binding:"required"`
	TrainContent    string   `json:"trainContent" binding:"required"`
	RehabStage      string   `json:"rehabStage" binding:"required"` // 对应 rehab_stage
	TrainPurpose    string   `json:"trainPurpose"`
	ForbiddenAction string   `json:"forbiddenAction"`
	PicUrls         []string `json:"picUrls"` // 【修改】改为 []string 以接收前端数组
	GuidePDF        string   `json:"guidePdf"`
	GuideWord       string   `json:"guideWord"`
	Sort            int      `json:"sort"`
	AuditStatus     *int     `json:"auditStatus"`  // 【新增】审核状态，使用指针
	RejectReason    *string  `json:"rejectReason"` // 【新增】驳回原因，使用指针
	DiseaseIds      []int    `json:"diseaseIds"`   // 新增：关联的疾病ID列表
}

// CreateTraining 新增训练指南
func CreateTraining(c *gin.Context) {
	var req CreateTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务启动失败",
		})
		return
	}
	defer tx.Rollback()

	// 1. 处理 PicUrls：将 []string 序列化为 JSON 字符串
	picUrlsJson := "[]" // 默认空数组
	if len(req.PicUrls) > 0 {
		jsonBytes, err := json.Marshal(req.PicUrls)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "图片URL序列化失败",
			})
			return
		}
		picUrlsJson = string(jsonBytes)
	}

	// 2. 处理审核状态和驳回原因
	initialAuditStatus := 1 // 默认已通过
	if req.AuditStatus != nil {
		initialAuditStatus = *req.AuditStatus
	}

	initialRejectReason := ""
	if req.RejectReason != nil {
		initialRejectReason = *req.RejectReason
	}

	// 3. 插入主表
	// 【修改】SQL 中增加了 audit_status, reject_reason
	insertQuery := `
		INSERT INTO rehab_train_guide 
		(rehab_stage, title, train_purpose, train_content, forbidden_action,
		 pic_urls, guide_pdf, guide_word, sort, audit_status, reject_reason, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.Exec(insertQuery,
		req.RehabStage, req.Title, req.TrainPurpose, req.TrainContent, req.ForbiddenAction,
		picUrlsJson, req.GuidePDF, req.GuideWord, req.Sort, initialAuditStatus, initialRejectReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建训练指南失败: " + err.Error(),
		})
		return
	}

	// 获取新增的 ID
	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取新增ID失败",
		})
		return
	}

	// 4. 插入疾病关联
	if len(req.DiseaseIds) > 0 {
		relQuery := "INSERT INTO rehab_train_guide_disease_rel (guide_id, disease_id) VALUES (?, ?)"
		for _, did := range req.DiseaseIds {
			if did > 0 {
				_, err := tx.Exec(relQuery, id, did)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    500,
						"message": "关联疾病失败",
					})
					return
				}
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data": gin.H{
			"id": id,
		},
	})
}

// UpdateTrainingRequest 更新训练指南请求结构
type UpdateTrainingRequest struct {
	Title           string   `json:"title"`
	TrainContent    string   `json:"trainContent"`
	RehabStage      string   `json:"rehabStage"`
	TrainPurpose    string   `json:"trainPurpose"`
	ForbiddenAction string   `json:"forbiddenAction"`
	PicUrls         []string `json:"picUrls"` // 【修改】改为 []string 以接收前端数组
	GuidePDF        string   `json:"guidePdf"`
	GuideWord       string   `json:"guideWord"`
	Sort            int      `json:"sort"`
	AuditStatus     *int     `json:"auditStatus"`  // 【修改】改为指针，以便区分“未传”和“传了0”
	RejectReason    *string  `json:"rejectReason"` // 【新增】驳回原因，使用指针
	DiseaseIds      []int    `json:"diseaseIds"`   // 如果传此字段，则更新关联关系
}

// UpdateTraining 更新训练指南
func UpdateTraining(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的训练 ID",
		})
		return
	}

	var req UpdateTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务启动失败",
		})
		return
	}
	defer tx.Rollback()

	// 1. 构建动态更新主表语句
	updateFields := []string{}
	args := []interface{}{}

	if req.Title != "" {
		updateFields = append(updateFields, "title = ?")
		args = append(args, req.Title)
	}
	if req.TrainContent != "" {
		updateFields = append(updateFields, "train_content = ?")
		args = append(args, req.TrainContent)
	}
	if req.RehabStage != "" {
		updateFields = append(updateFields, "rehab_stage = ?")
		args = append(args, req.RehabStage)
	}
	if req.TrainPurpose != "" {
		updateFields = append(updateFields, "train_purpose = ?")
		args = append(args, req.TrainPurpose)
	}
	if req.ForbiddenAction != "" {
		updateFields = append(updateFields, "forbidden_action = ?")
		args = append(args, req.ForbiddenAction)
	}

	// 【修改】处理 PicUrls：将 []string 序列化为 JSON 字符串存入数据库
	if req.PicUrls != nil {
		// 使用 json.Marshal 将切片转换为 JSON 字符串
		picUrlsJson, err := json.Marshal(req.PicUrls)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "图片URL序列化失败",
			})
			return
		}
		updateFields = append(updateFields, "pic_urls = ?")
		args = append(args, string(picUrlsJson))
	}

	if req.GuidePDF != "" {
		updateFields = append(updateFields, "guide_pdf = ?")
		args = append(args, req.GuidePDF)
	}
	if req.GuideWord != "" {
		updateFields = append(updateFields, "guide_word = ?")
		args = append(args, req.GuideWord)
	}

	if req.Sort != 0 {
		updateFields = append(updateFields, "sort = ?")
		args = append(args, req.Sort)
	}

	// 【修改】审核状态更新：使用指针判断是否传递
	if req.AuditStatus != nil {
		updateFields = append(updateFields, "audit_status = ?")
		args = append(args, *req.AuditStatus)

		// 可选：如果审核通过，清空驳回原因
		if *req.AuditStatus == 1 {
			updateFields = append(updateFields, "reject_reason = NULL")
		}
	}

	// 【新增】驳回原因更新：使用指针判断是否传递
	if req.RejectReason != nil {
		updateFields = append(updateFields, "reject_reason = ?")
		args = append(args, *req.RejectReason)
	}

	// 如果有字段需要更新
	if len(updateFields) > 0 {
		updateFields = append(updateFields, "updated_at = NOW()")
		args = append(args, id)

		updateQuery := `UPDATE rehab_train_guide SET ` + strings.Join(updateFields, ", ") + ` WHERE id = ?`
		_, err = tx.Exec(updateQuery, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新训练指南失败: " + err.Error(),
			})
			return
		}
	}

	// 2. 同步疾病关联 (如果前端传了 DiseaseIds)
	if req.DiseaseIds != nil {
		// 先删除旧关联
		_, err := tx.Exec("DELETE FROM rehab_train_guide_disease_rel WHERE guide_id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "清理旧疾病关联失败",
			})
			return
		}

		// 再插入新关联
		if len(req.DiseaseIds) > 0 {
			relQuery := "INSERT INTO rehab_train_guide_disease_rel (guide_id, disease_id) VALUES (?, ?)"
			for _, did := range req.DiseaseIds {
				if did > 0 {
					_, err := tx.Exec(relQuery, id, did)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"code":    500,
							"message": "创建新疾病关联失败",
						})
						return
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    nil,
	})
}

// DeleteTraining 删除训练指南（物理删除）
func DeleteTraining(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的训练 ID",
		})
		return
	}

	// 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务启动失败",
		})
		return
	}
	defer tx.Rollback()

	// 1. 先删除关联表数据 (因为可能存在外键约束 ON DELETE RESTRICT，必须先删子表)
	_, err = tx.Exec("DELETE FROM rehab_train_guide_disease_rel WHERE guide_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除关联疾病数据失败: " + err.Error(),
		})
		return
	}

	// 2. 删除主表数据
	result, err := tx.Exec("DELETE FROM rehab_train_guide WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除训练指南失败: " + err.Error(),
		})
		return
	}

	// 检查是否真的删除了行（可选，用于判断ID是否存在）
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "训练指南不存在",
		})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    nil,
	})
}

// GetTrainingResource 获取资源文件
func GetTrainingResource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的训练 ID",
		})
		return
	}

	resourceType := c.DefaultQuery("type", "pdf")

	query := `
		SELECT title, guide_pdf, guide_word
		FROM rehab_train_guide
		WHERE id = ? AND audit_status = 1
	`

	var training struct {
		Title     string         `db:"title"`
		GuidePDF  sql.NullString `db:"guide_pdf"`
		GuideWord sql.NullString `db:"guide_word"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(&training.Title, &training.GuidePDF, &training.GuideWord)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "训练指南不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询资源文件失败",
		})
		return
	}

	// 根据 type 返回对应资源
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
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "资源文件不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": ResourceResponse{
			DownloadUrl: downloadUrl,
			PreviewUrl:  "https://example.com/preview/" + strconv.FormatUint(id, 10),
			FileName:    fileName,
			FileSize:    "2.5MB", // 实际项目中可从 OSS 获取
		},
	})
}
