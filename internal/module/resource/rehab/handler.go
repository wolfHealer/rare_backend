package rehab

import (
	"database/sql"
	"fmt"
	"net/http"
	"rare_backend/internal/pkg/db"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

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

// ResourceResponse 资源文件响应结构
type ResourceResponse struct {
	DownloadUrl string `json:"downloadUrl"`
	PreviewUrl  string `json:"previewUrl"`
	FileName    string `json:"fileName"`
	FileSize    string `json:"fileSize"`
}

// DictionaryResponse 字典响应结构
type DictionaryResponse struct {
	Diseases []OptionItem `json:"diseases"`
	Stages   []OptionItem `json:"stages"`
}

// OptionItem 选项项
type OptionItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// GetTrainingList 获取训练指南列表
func GetTrainingList(c *gin.Context) {
	// 获取请求参数
	diseaseStr := c.DefaultQuery("disease", "")
	stage := c.DefaultQuery("stage", "")
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
		// 需要将疾病名称转换为 disease_value
		whereClause += " AND disease_value = (SELECT value FROM disease_options WHERE name = ?)"
		args = append(args, diseaseStr)
	}
	if stage != "" {
		whereClause += " AND illness_stage = ?"
		args = append(args, stage)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM rehab_train_guides " + whereClause
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
		SELECT id, disease_value, illness_stage, title, train_purpose, sort
		FROM rehab_train_guides
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

	var list []TrainingItem
	for rows.Next() {
		var training struct {
			ID           uint   `db:"id"`
			DiseaseValue int    `db:"disease_value"`
			IllnessStage string `db:"illness_stage"`
			Title        string `db:"title"`
			TrainPurpose string `db:"train_purpose"`
			Sort         int    `db:"sort"`
		}
		if err := rows.Scan(
			&training.ID, &training.DiseaseValue, &training.IllnessStage,
			&training.Title, &training.TrainPurpose, &training.Sort,
		); err != nil {
			continue
		}

		// 查询疾病名称
		var diseaseName string
		diseaseQuery := "SELECT name FROM disease_options WHERE value = ?"
		err := db.MySQL.QueryRow(diseaseQuery, training.DiseaseValue).Scan(&diseaseName)
		if err != nil {
			diseaseName = ""
		}

		// 根据病情阶段转换 type
		trainType := convertStageToType(training.IllnessStage)

		list = append(list, TrainingItem{
			ID:       training.ID,
			Title:    training.Title,
			Type:     trainType,
			Stage:    convertStageToValue(training.IllnessStage),
			Disease:  diseaseStr, // 使用请求参数中的 disease
			Desc:     training.TrainPurpose,
			CoverUrl: "https://example.com/cover/" + strconv.FormatUint(uint64(training.ID), 10) + ".jpg",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": TrainingListResponse{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
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

	query := `
		SELECT id, title, train_content, forbidden_action, pic_urls, illness_stage
		FROM rehab_train_guides
		WHERE id = ? AND is_audit = 1
	`

	var training struct {
		ID              uint           `db:"id"`
		Title           string         `db:"title"`
		TrainContent    string         `db:"train_content"`
		ForbiddenAction sql.NullString `db:"forbidden_action"`
		PicUrls         sql.NullString `db:"pic_urls"`
		IllnessStage    string         `db:"illness_stage"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&training.ID, &training.Title, &training.TrainContent,
		&training.ForbiddenAction, &training.PicUrls, &training.IllnessStage,
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
			"message": "查询训练详情失败",
		})
		return
	}

	// 根据病情阶段计算难度和时长
	difficulty, duration := calculateDifficultyAndDuration(training.IllnessStage)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": TrainingDetailResponse{
			ID:         training.ID,
			Title:      training.Title,
			Content:    training.TrainContent,
			VideoUrl:   "https://example.com/video/" + strconv.FormatUint(id, 10) + ".mp4",
			Duration:   duration,
			Difficulty: difficulty,
			Purpose:    "", // 可在 train_content 中提取
			Forbidden:  training.ForbiddenAction.String,
			PicUrls:    training.PicUrls.String,
		},
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
		FROM rehab_train_guides
		WHERE id = ? AND is_audit = 1
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

// GetDictionaries 获取筛选字典
func GetDictionaries(c *gin.Context) {
	// 查询疾病选项
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

	// 病情阶段选项
	stages := []OptionItem{
		{Text: "早期", Value: "early"},
		{Text: "中期", Value: "mid"},
		{Text: "晚期", Value: "late"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": DictionaryResponse{
			Diseases: diseases,
			Stages:   stages,
		},
	})
}

// ========== 辅助函数 ==========

// convertStageToType 将病情阶段转换为训练类型
func convertStageToType(stage string) string {
	typeMap := map[string]string{
		"早期":    "基础训练",
		"中期":    "强化训练",
		"晚期":    "维持训练",
		"early": "基础训练",
		"mid":   "强化训练",
		"late":  "维持训练",
	}
	if val, ok := typeMap[stage]; ok {
		return val
	}
	return "康复训练"
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

// calculateDifficultyAndDuration 根据病情阶段计算难度和时长
func calculateDifficultyAndDuration(stage string) (string, string) {
	switch stage {
	case "早期", "early":
		return "简单", "15 分钟"
	case "中期", "mid":
		return "中等", "30 分钟"
	case "晚期", "late":
		return "困难", "45 分钟"
	default:
		return "中等", "30 分钟"
	}
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// CareManualItem 护理手册项响应结构
type CareManualItem struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	Category   string `json:"category"`
	Content    string `json:"content"`
	Icon       string `json:"icon"`
	Sort       int    `json:"sort"`
	UpdateTime string `json:"updateTime"`
}

// CareManualResponse 护理手册响应结构
type CareManualResponse struct {
	Manuals []CareManualItem `json:"manuals"`
}

// GetCareManuals 获取护理手册列表
func GetCareManuals(c *gin.Context) {
	// 获取请求参数
	disease := c.DefaultQuery("disease", "")
	category := c.DefaultQuery("category", "")

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}

	if disease != "" {
		// 将疾病代码转换为 disease_value
		diseaseValue := convertDiseaseToValue(disease)
		if diseaseValue > 0 {
			whereClause += " AND disease_value = ?"
			args = append(args, diseaseValue)
		}
	}

	// 查询护理手册
	query := `
		SELECT id, disease_value, title, diet_guide, skin_care, oral_care,
		       complication_prevent, bed_care, updated_at
		FROM home_care_manuals
		` + whereClause + `
		ORDER BY id DESC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询护理手册失败",
		})
		return
	}
	defer rows.Close()

	var manuals []CareManualItem

	// 遍历查询结果，构建护理手册列表
	for rows.Next() {
		var manual struct {
			ID                  uint      `db:"id"`
			DiseaseValue        int       `db:"disease_value"`
			Title               string    `db:"title"`
			DietGuide           string    `db:"diet_guide"`
			SkinCare            string    `db:"skin_care"`
			OralCare            string    `db:"oral_care"`
			ComplicationPrevent string    `db:"complication_prevent"`
			BedCare             string    `db:"bed_care"`
			UpdatedAt           time.Time `db:"updated_at"`
		}
		if err := rows.Scan(
			&manual.ID, &manual.DiseaseValue, &manual.Title,
			&manual.DietGuide, &manual.SkinCare, &manual.OralCare,
			&manual.ComplicationPrevent, &manual.BedCare, &manual.UpdatedAt,
		); err != nil {
			continue
		}

		updateTime := manual.UpdatedAt.Format("2006-01-02T15:04:05Z")

		// 定义类别配置（直接在循环内定义）
		categories := []struct {
			Key     string
			Title   string
			Content string
			Icon    string
			Sort    int
			IsFixed bool // 是否为固定内容
		}{
			{"diet", "饮食指导", manual.DietGuide, "food", 1, false},
			{"skin", "皮肤护理", manual.SkinCare, "shield", 2, false},
			{"oral", "口腔护理", manual.OralCare, "smile", 3, false},
			{"rehab", "康复训练", manual.ComplicationPrevent, "replay", 4, false},
			{"bed", "卧床护理", manual.BedCare, "bed", 5, false},
			{"medication", "用药指导", "按时按量服药，注意药物不良反应。具体用药请遵医嘱。", "bag", 6, true},
			{"psychology", "心理支持", "关注患者及家属心理健康，多沟通，给予情感支持。", "heart", 7, true},
		}

		for _, cat := range categories {
			// 按 category 筛选
			if category != "" && category != cat.Key {
				continue
			}
			// 内容为空时跳过（固定内容的除外）
			if !cat.IsFixed && cat.Content == "" {
				continue
			}
			manuals = append(manuals, CareManualItem{
				ID:         manual.ID,
				Title:      cat.Title,
				Category:   cat.Key,
				Content:    cat.Content,
				Icon:       cat.Icon,
				Sort:       cat.Sort,
				UpdateTime: updateTime,
			})
		}
	}

	// 按 sort 排序
	sort.Slice(manuals, func(i, j int) bool {
		return manuals[i].Sort < manuals[j].Sort
	})

	// 确保数组不为 null
	if manuals == nil {
		manuals = []CareManualItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": CareManualResponse{
			Manuals: manuals,
		},
	})
}

// convertDiseaseToValue 将疾病代码转换为 disease_value
func convertDiseaseToValue(diseaseCode string) int {
	diseaseMap := map[string]int{
		"als":            1,
		"huntington":     2,
		"rare":           3,
		"hemophilia":     4,
		"gaucher":        5,
		"pompe":          6,
		"cerebral_palsy": 7,
		"leukemia":       8,
	}
	if value, ok := diseaseMap[diseaseCode]; ok {
		return value
	}
	return 0
}

// ChecklistResponse 护理清单响应结构
type ChecklistResponse struct {
	FileName    string   `json:"fileName"`
	DownloadURL string   `json:"downloadUrl"`
	FileSize    string   `json:"fileSize"`
	UpdateTime  string   `json:"updateTime"`
	Items       []string `json:"items"`
}

// RecordFormResponse 记录表响应结构
type RecordFormResponse struct {
	FileName    string   `json:"fileName"`
	DownloadURL string   `json:"downloadUrl"`
	FileSize    string   `json:"fileSize"`
	UpdateTime  string   `json:"updateTime"`
	Fields      []string `json:"fields"`
}

// CategoryItem 分类项
type CategoryItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
	Icon  string `json:"icon"`
}

// CategoryResponse 分类响应结构
type CategoryResponse struct {
	Categories []CategoryItem `json:"categories"`
}

// GetChecklist 获取日常护理检查清单
func GetChecklist(c *gin.Context) {
	// 获取请求参数
	disease := c.DefaultQuery("disease", "")

	// 根据疾病类型返回针对性清单
	items := getDefaultChecklistItems()
	if disease != "" {
		items = getDiseaseSpecificChecklist(disease)
	}

	// 获取最新更新时间
	updateTime := getManualUpdateTime()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": ChecklistResponse{
			FileName:    "居家护理清单.pdf",
			DownloadURL: "https://example.com/rehab/checklist.pdf",
			FileSize:    "500KB",
			UpdateTime:  updateTime,
			Items:       items,
		},
	})
}

// GetRecordForm 获取病情观察记录表
func GetRecordForm(c *gin.Context) {
	// 获取请求参数
	disease := c.DefaultQuery("disease", "")

	// 根据疾病类型返回针对性记录表
	fields := getDefaultRecordFields()
	if disease != "" {
		fields = getDiseaseSpecificRecordFields(disease)
	}

	// 获取最新更新时间
	updateTime := getManualUpdateTime()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": RecordFormResponse{
			FileName:    "病情观察记录表.xlsx",
			DownloadURL: "https://example.com/rehab/record_form.xlsx",
			FileSize:    "200KB",
			UpdateTime:  updateTime,
			Fields:      fields,
		},
	})
}

// GetCategories 获取护理手册分类
func GetCategories(c *gin.Context) {
	categories := []CategoryItem{
		{Text: "全部", Value: "all", Icon: "apps"},
		{Text: "饮食指导", Value: "diet", Icon: "food"},
		{Text: "皮肤护理", Value: "skin", Icon: "shield"},
		{Text: "口腔护理", Value: "oral", Icon: "smile"},
		{Text: "康复训练", Value: "rehab", Icon: "replay"},
		{Text: "卧床护理", Value: "bed", Icon: "bed"},
		{Text: "用药指导", Value: "medication", Icon: "bag"},
		{Text: "心理支持", Value: "psychology", Icon: "heart"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": CategoryResponse{
			Categories: categories,
		},
	})
}

// getDefaultChecklistItems 获取默认检查清单项目
func getDefaultChecklistItems() []string {
	return []string{
		"□ 晨间护理：洗脸、刷牙、梳头",
		"□ 早餐服药：按医嘱服用",
		"□ 康复训练：上午 10:00",
		"□ 午餐：注意营养搭配",
		"□ 午休：12:00-14:00",
		"□ 康复训练：下午 15:00",
		"□ 晚间护理：擦浴、按摩",
		"□ 晚餐服药：按医嘱服用",
		"□ 记录当日情况",
	}
}

// getDiseaseSpecificChecklist 获取针对特定疾病的检查清单
func getDiseaseSpecificChecklist(disease string) []string {
	// 根据疾病类型返回针对性清单
	checklistMap := map[string][]string{
		"als": {
			"□ 晨间护理：洗脸、刷牙、梳头",
			"□ 呼吸训练：上午 9:00",
			"□ 早餐服药：按医嘱服用",
			"□ 肢体按摩：预防肌肉萎缩",
			"□ 康复训练：上午 10:00",
			"□ 午餐：高蛋白、易消化",
			"□ 午休：12:00-14:00",
			"□ 翻身护理：每 2 小时一次",
			"□ 康复训练：下午 15:00",
			"□ 晚间护理：擦浴、按摩",
			"□ 晚餐服药：按医嘱服用",
			"□ 呼吸监测：睡前检查",
			"□ 记录当日情况",
		},
		"hemophilia": {
			"□ 晨间护理：温和清洁",
			"□ 早餐服药：凝血因子",
			"□ 关节检查：有无肿胀",
			"□ 康复训练：轻度活动",
			"□ 午餐：补充维生素 K",
			"□ 午休：12:00-14:00",
			"□ 避免剧烈运动",
			"□ 康复训练：下午 15:00",
			"□ 晚间护理：检查有无出血",
			"□ 晚餐服药：按医嘱服用",
			"□ 记录当日情况",
		},
		"gaucher": {
			"□ 晨间护理：洗脸、刷牙",
			"□ 早餐服药：酶替代治疗",
			"□ 腹部检查：肝脾大小",
			"□ 康复训练：适度活动",
			"□ 午餐：均衡营养",
			"□ 午休：12:00-14:00",
			"□ 康复训练：下午 15:00",
			"□ 晚间护理：擦浴",
			"□ 晚餐服药：按医嘱服用",
			"□ 血常规监测",
			"□ 记录当日情况",
		},
	}

	if items, ok := checklistMap[disease]; ok {
		return items
	}
	return getDefaultChecklistItems()
}

// getDefaultRecordFields 获取默认记录表字段
func getDefaultRecordFields() []string {
	return []string{
		"日期",
		"体温",
		"血压",
		"心率",
		"呼吸",
		"饮食情况",
		"服药情况",
		"康复训练",
		"异常情况记录",
		"备注",
	}
}

// getDiseaseSpecificRecordFields 获取针对特定疾病的记录表字段
func getDiseaseSpecificRecordFields(disease string) []string {
	// 根据疾病类型返回针对性字段
	fieldsMap := map[string][]string{
		"als": {
			"日期",
			"体温",
			"血压",
			"心率",
			"呼吸频率",
			"血氧饱和度",
			"吞咽功能",
			"肢体活动度",
			"饮食情况",
			"服药情况",
			"康复训练",
			"呼吸训练",
			"异常情况记录",
			"备注",
		},
		"hemophilia": {
			"日期",
			"体温",
			"血压",
			"心率",
			"关节状况",
			"有无出血",
			"出血部位",
			"凝血因子用量",
			"饮食情况",
			"服药情况",
			"康复训练",
			"异常情况记录",
			"备注",
		},
		"gaucher": {
			"日期",
			"体温",
			"血压",
			"心率",
			"腹部状况",
			"肝脾大小",
			"血常规",
			"骨痛情况",
			"饮食情况",
			"服药情况",
			"康复训练",
			"异常情况记录",
			"备注",
		},
	}

	if fields, ok := fieldsMap[disease]; ok {
		return fields
	}
	return getDefaultRecordFields()
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
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	TypeName    string   `json:"typeName"`
	Region      string   `json:"region"`
	RegionCode  string   `json:"regionCode"`
	Address     string   `json:"address"`
	Contact     string   `json:"contact"`
	Phone       string   `json:"phone"`
	Email       string   `json:"email"`
	Website     string   `json:"website"`
	Services    []string `json:"services"`
	Rating      float64  `json:"rating"`
	IsInsurance bool     `json:"isInsurance"`
	Description string   `json:"description"`
	CoverUrl    string   `json:"coverUrl"`
	Status      string   `json:"status"`
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
	region := c.DefaultQuery("region", "")
	instType := c.DefaultQuery("type", "")
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

	if region != "" {
		whereClause += " AND region = ?"
		args = append(args, region)
	}
	if instType != "" {
		// 机构类型筛选（可根据实际需求调整）
		whereClause += " AND name LIKE ?"
		args = append(args, "%"+instType+"%")
	}
	if keyword != "" {
		whereClause += " AND (name LIKE ? OR address LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM rehab_institutions " + whereClause
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
		SELECT id, disease_value, name, region, rehab_projects, 
		       fee_standard, contact, address, created_at
		FROM rehab_institutions
		` + whereClause + `
		ORDER BY id DESC
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

	var list []InstitutionItem
	for rows.Next() {
		var institution struct {
			ID            uint      `db:"id"`
			DiseaseValue  int       `db:"disease_value"`
			Name          string    `db:"name"`
			Region        string    `db:"region"`
			RehabProjects string    `db:"rehab_projects"`
			FeeStandard   string    `db:"fee_standard"`
			Contact       string    `db:"contact"`
			Address       string    `db:"address"`
			CreatedAt     time.Time `db:"created_at"`
		}
		if err := rows.Scan(
			&institution.ID, &institution.DiseaseValue, &institution.Name,
			&institution.Region, &institution.RehabProjects, &institution.FeeStandard,
			&institution.Contact, &institution.Address, &institution.CreatedAt,
		); err != nil {
			continue
		}

		// 解析联系方式
		phone, email, website := parseContact(institution.Contact)

		// 解析服务项目
		services := parseServices(institution.RehabProjects)

		// 转换地区代码
		regionCode := convertRegionToCode(institution.Region)

		// 转换机构类型
		instType, instTypeName := convertInstitutionType(institution.Name)

		list = append(list, InstitutionItem{
			ID:          institution.ID,
			Name:        institution.Name,
			Type:        instType,
			TypeName:    instTypeName,
			Region:      institution.Region,
			RegionCode:  regionCode,
			Address:     institution.Address,
			Contact:     institution.Contact,
			Phone:       phone,
			Email:       email,
			Website:     website,
			Services:    services,
			Rating:      4.5 + float64(institution.ID%10)/10, // 示例评分
			IsInsurance: true,                                // 实际可从数据库字段获取
			Description: institution.FeeStandard,
			CoverUrl:    "https://example.com/institutions/" + strconv.FormatUint(uint64(institution.ID), 10) + ".jpg",
			Status:      "active",
		})
	}

	// 确保数组不为 null
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

	query := `
		SELECT id, disease_value, name, region, qualification, rehab_projects,
		       fee_standard, contact, address, created_at, updated_at
		FROM rehab_institutions
		WHERE id = ? AND is_audit = 1
	`

	var institution struct {
		ID            uint      `db:"id"`
		DiseaseValue  int       `db:"disease_value"`
		Name          string    `db:"name"`
		Region        string    `db:"region"`
		Qualification string    `db:"qualification"`
		RehabProjects string    `db:"rehab_projects"`
		FeeStandard   string    `db:"fee_standard"`
		Contact       string    `db:"contact"`
		Address       string    `db:"address"`
		CreatedAt     time.Time `db:"created_at"`
		UpdatedAt     time.Time `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&institution.ID, &institution.DiseaseValue, &institution.Name,
		&institution.Region, &institution.Qualification, &institution.RehabProjects,
		&institution.FeeStandard, &institution.Contact, &institution.Address,
		&institution.CreatedAt, &institution.UpdatedAt,
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
			"message": "查询机构详情失败",
		})
		return
	}

	// 解析联系方式
	phone, email, website := parseContact(institution.Contact)

	// 解析服务项目
	services := parseServices(institution.RehabProjects)

	// 转换地区代码
	regionCode := convertRegionToCode(institution.Region)

	// 转换机构类型
	instType, instTypeName := convertInstitutionType(institution.Name)

	// 构建图片列表
	images := []string{
		"https://example.com/institutions/" + strconv.FormatUint(id, 10) + "_1.jpg",
		"https://example.com/institutions/" + strconv.FormatUint(id, 10) + "_2.jpg",
	}

	// 构建设施列表
	facilities := []string{"无障碍通道", "停车场", "住院部", "门诊部"}

	// 构建医生列表
	doctors := getInstitutionDoctors(institution.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": InstitutionDetailResponse{
			ID:            institution.ID,
			Name:          institution.Name,
			Type:          instType,
			TypeName:      instTypeName,
			Region:        institution.Region,
			RegionCode:    regionCode,
			Address:       institution.Address,
			Contact:       institution.Contact,
			Phone:         phone,
			Email:         email,
			Website:       website,
			Services:      services,
			Rating:        4.5 + float64(institution.ID%10)/10,
			IsInsurance:   true,
			Description:   institution.FeeStandard,
			CoverUrl:      "https://example.com/institutions/" + strconv.FormatUint(id, 10) + ".jpg",
			Images:        images,
			BusinessHours: "周一至周日 08:00-17:00",
			Facilities:    facilities,
			Doctors:       doctors,
			Status:        "active",
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

// parseContact 解析联系方式
func parseContact(contact string) (phone, email, website string) {
	// 简化处理，实际可根据格式解析
	// 假设 contact 格式为：电话/官网
	parts := strings.Split(contact, "/")
	if len(parts) >= 1 {
		phone = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		website = strings.TrimSpace(parts[1])
	}
	email = "contact@example.com" // 示例
	return
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

// DeviceItem 康复器械项响应结构
type DeviceItem struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	CategoryName string   `json:"categoryName"`
	Desc         string   `json:"desc"`
	SuitableFor  []string `json:"suitableFor"`
	PriceRange   string   `json:"priceRange"`
	IsInsurance  bool     `json:"insuranceCovered"`
	CoverUrl     string   `json:"coverUrl"`
	GuideUrl     string   `json:"guideUrl"`
	VideoUrl     string   `json:"videoUrl"`
	Status       string   `json:"status"`
}

// DeviceListResponse 器械列表响应结构
type DeviceListResponse struct {
	List     []DeviceItem `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
}

// DeviceGuideResponse 器械指南响应结构
type DeviceGuideResponse struct {
	FileName    string `json:"fileName"`
	DownloadURL string `json:"downloadUrl"`
	FileSize    string `json:"fileSize"`
	UpdateTime  string `json:"updateTime"`
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

// DeviceCategoryItem 器械类别项
type DeviceCategoryItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// DeviceCategoryResponse 器械类别响应结构
type DeviceCategoryResponse struct {
	Categories []DeviceCategoryItem `json:"categories"`
}

// GetDevices 获取康复器械列表
func GetDevices(c *gin.Context) {
	// 获取请求参数
	category := c.DefaultQuery("category", "")
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

	if category != "" {
		whereClause += " AND equip_name LIKE ?"
		args = append(args, "%"+category+"%")
	}
	if keyword != "" {
		whereClause += " AND (equip_name LIKE ? OR apply_crowd LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM rehab_equipment_guides " + whereClause
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
		SELECT id, disease_value, equip_name, apply_crowd, buy_suggest,
		       purchase_list, created_at, updated_at
		FROM rehab_equipment_guides
		` + whereClause + `
		ORDER BY id DESC
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

	var list []DeviceItem
	for rows.Next() {
		var device struct {
			ID           uint      `db:"id"`
			DiseaseValue int       `db:"disease_value"`
			EquipName    string    `db:"equip_name"`
			ApplyCrowd   string    `db:"apply_crowd"`
			BuySuggest   string    `db:"buy_suggest"`
			PurchaseList string    `db:"purchase_list"`
			CreatedAt    time.Time `db:"created_at"`
			UpdatedAt    time.Time `db:"updated_at"`
		}
		if err := rows.Scan(
			&device.ID, &device.DiseaseValue, &device.EquipName,
			&device.ApplyCrowd, &device.BuySuggest, &device.PurchaseList,
			&device.CreatedAt, &device.UpdatedAt,
		); err != nil {
			continue
		}

		// 转换器械类别
		category, categoryName := convertDeviceCategory(device.EquipName)

		// 解析适用人群
		suitableFor := parseSuitableFor(device.ApplyCrowd)

		// 获取价格范围
		priceRange := getPriceRange(device.EquipName)

		list = append(list, DeviceItem{
			ID:           device.ID,
			Name:         device.EquipName,
			Category:     category,
			CategoryName: categoryName,
			Desc:         truncateString(device.BuySuggest, 50),
			SuitableFor:  suitableFor,
			PriceRange:   priceRange,
			IsInsurance:  true, // 实际可从数据库字段获取
			CoverUrl:     "https://example.com/devices/" + strconv.FormatUint(uint64(device.ID), 10) + ".jpg",
			GuideUrl:     device.PurchaseList,
			VideoUrl:     "https://example.com/videos/" + category + ".mp4",
			Status:       "active",
		})
	}

	// 确保数组不为 null
	if list == nil {
		list = []DeviceItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": DeviceListResponse{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// GetDeviceGuide 获取器械使用指南
func GetDeviceGuide(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的器械 ID",
		})
		return
	}

	query := `
		SELECT id, equip_name, purchase_list, updated_at
		FROM rehab_equipment_guides
		WHERE id = ? AND is_audit = 1
	`

	var device struct {
		ID           uint      `db:"id"`
		EquipName    string    `db:"equip_name"`
		PurchaseList string    `db:"purchase_list"`
		UpdatedAt    time.Time `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&device.ID, &device.EquipName, &device.PurchaseList, &device.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "器械指南不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询器械指南失败",
		})
		return
	}

	// 检查指南地址是否为空
	if device.PurchaseList == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "该器械暂无指南文件",
		})
		return
	}

	fileName := device.EquipName + "使用指南.pdf"
	fileSize := getFileSize(device.PurchaseList)
	updateTime := device.UpdatedAt.Format("2006-01-02T15:04:05Z")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": DeviceGuideResponse{
			FileName:    fileName,
			DownloadURL: device.PurchaseList,
			FileSize:    fileSize,
			UpdateTime:  updateTime,
		},
	})
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

// GetDeviceCategories 获取器械类别筛选选项
func GetDeviceCategories(c *gin.Context) {
	categories := []DeviceCategoryItem{
		{Text: "全部类别", Value: "all"},
		{Text: "轮椅类", Value: "wheelchair"},
		{Text: "助行类", Value: "walker"},
		{Text: "站立训练类", Value: "standing_frame"},
		{Text: "护理床类", Value: "bed"},
		{Text: "康复训练类", Value: "training"},
		{Text: "生活辅助类", Value: "daily_aid"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": DeviceCategoryResponse{
			Categories: categories,
		},
	})
}

// convertDeviceCategory 转换器械类别
func convertDeviceCategory(name string) (string, string) {
	categoryMap := map[string]struct {
		Category string
		Name     string
	}{
		"轮椅":  {"wheelchair", "轮椅类"},
		"助行":  {"walker", "助行类"},
		"站立":  {"standing_frame", "站立训练类"},
		"护理床": {"bed", "护理床类"},
		"训练":  {"training", "康复训练类"},
		"拐杖":  {"crutch", "助行类"},
		"矫形":  {"orthosis", "康复训练类"},
	}

	for key, val := range categoryMap {
		if strings.Contains(name, key) {
			return val.Category, val.Name
		}
	}
	return "other", "其他器械"
}

// parseSuitableFor 解析适用人群
func parseSuitableFor(crowd string) []string {
	if crowd == "" {
		return []string{"康复患者", "术后恢复", "行动不便"}
	}
	// 按分隔符分割
	suitable := strings.Split(crowd, "，")
	if len(suitable) == 0 {
		suitable = strings.Split(crowd, ",")
	}
	if len(suitable) == 0 {
		return []string{"康复患者", "术后恢复", "行动不便"}
	}
	return suitable
}

// getPriceRange 获取价格范围
func getPriceRange(name string) string {
	priceMap := map[string]string{
		"轮椅":  "500-5000 元",
		"助行器": "200-1000 元",
		"站立架": "1000-8000 元",
		"护理床": "2000-15000 元",
		"拐杖":  "100-500 元",
		"矫形器": "1000-10000 元",
	}

	for key, price := range priceMap {
		if strings.Contains(name, key) {
			return price
		}
	}
	return "价格面议"
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
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	TypeName    string   `json:"typeName"`
	Region      string   `json:"region"`
	RegionCode  string   `json:"regionCode"`
	Address     string   `json:"address"`
	Contact     string   `json:"contact"`
	Phone       string   `json:"phone"`
	Email       string   `json:"email"`
	Website     string   `json:"website"`
	IsFree      bool     `json:"isFree"`
	ServiceTime string   `json:"serviceTime"`
	Description string   `json:"description"`
	Services    []string `json:"services"`
	Rating      float64  `json:"rating"`
	CoverUrl    string   `json:"coverUrl"`
	Status      string   `json:"status"`
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
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	TypeName    string          `json:"typeName"`
	Region      string          `json:"region"`
	RegionCode  string          `json:"regionCode"`
	Address     string          `json:"address"`
	Contact     string          `json:"contact"`
	Phone       string          `json:"phone"`
	Email       string          `json:"email"`
	Website     string          `json:"website"`
	IsFree      bool            `json:"isFree"`
	ServiceTime string          `json:"serviceTime"`
	Description string          `json:"description"`
	Services    []string        `json:"services"`
	Rating      float64         `json:"rating"`
	CoverUrl    string          `json:"coverUrl"`
	Images      []string        `json:"images"`
	Counselors  []CounselorItem `json:"counselors"`
	Status      string          `json:"status"`
}

// GetPsychologicalOrgs 获取心理咨询机构列表
func GetPsychologicalOrgs(c *gin.Context) {
	// 获取请求参数
	region := c.DefaultQuery("region", "")
	orgType := c.DefaultQuery("type", "")
	isFreeStr := c.DefaultQuery("isFree", "")
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
	whereClause := "WHERE is_audit = 1 AND support_type = '咨询机构'"
	args := []interface{}{}

	if region != "" && region != "all" {
		// 将地区代码转换为地区名称
		regionName := convertCodeToRegion(region)
		if regionName != "" {
			whereClause += " AND org_address LIKE ?"
			args = append(args, "%"+regionName+"%")
		}
	}
	if orgType != "" {
		whereClause += " AND name LIKE ?"
		args = append(args, "%"+orgType+"%")
	}
	if isFreeStr != "" {
		isFree := isFreeStr == "true"
		whereClause += " AND is_free = ?"
		args = append(args, isFree)
	}
	if keyword != "" {
		whereClause += " AND (name LIKE ? OR content_intro LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM psychological_supports " + whereClause
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
		SELECT id, disease_value, name, content_intro, org_address,
		       org_contact, is_free, consult_way, created_at, updated_at
		FROM psychological_supports
		` + whereClause + `
		ORDER BY id DESC
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

	var list []PsychologicalOrgItem
	for rows.Next() {
		var org struct {
			ID           uint         `db:"id"`
			DiseaseValue int          `db:"disease_value"`
			Name         string       `db:"name"`
			ContentIntro string       `db:"content_intro"`
			OrgAddress   string       `db:"org_address"`
			OrgContact   string       `db:"org_contact"`
			IsFree       sql.NullBool `db:"is_free"`
			ConsultWay   string       `db:"consult_way"`
			CreatedAt    time.Time    `db:"created_at"`
			UpdatedAt    time.Time    `db:"updated_at"`
		}
		if err := rows.Scan(
			&org.ID, &org.DiseaseValue, &org.Name, &org.ContentIntro,
			&org.OrgAddress, &org.OrgContact, &org.IsFree, &org.ConsultWay,
			&org.CreatedAt, &org.UpdatedAt,
		); err != nil {
			continue
		}

		// 解析联系方式
		phone, email, website := parsePsychologicalContact(org.OrgContact)

		// 提取地区
		region := extractRegionFromAddress(org.OrgAddress)
		regionCode := convertRegionToCode(region)

		// 转换机构类型
		orgType, orgTypeName := convertPsychologicalOrgType(org.Name, org.ConsultWay)

		// 解析服务项目
		services := parsePsychologicalServices(org.Name, org.ConsultWay)

		// 获取服务时间
		serviceTime := getServiceTime(orgType)

		isFree := false
		if org.IsFree.Valid {
			isFree = org.IsFree.Bool
		}

		list = append(list, PsychologicalOrgItem{
			ID:          org.ID,
			Name:        org.Name,
			Type:        orgType,
			TypeName:    orgTypeName,
			Region:      region,
			RegionCode:  regionCode,
			Address:     org.OrgAddress,
			Contact:     org.OrgContact,
			Phone:       phone,
			Email:       email,
			Website:     website,
			IsFree:      isFree,
			ServiceTime: serviceTime,
			Description: org.ContentIntro,
			Services:    services,
			Rating:      4.5 + float64(org.ID%10)/10,
			CoverUrl:    "https://example.com/orgs/psychological/" + strconv.FormatUint(uint64(org.ID), 10) + ".jpg",
			Status:      "active",
		})
	}

	// 确保数组不为 null
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

	query := `
		SELECT id, disease_value, name, content_intro, org_address,
		       org_contact, is_free, consult_way, guide_pdf,
		       manual_patient, manual_family, created_at, updated_at
		FROM psychological_supports
		WHERE id = ? AND is_audit = 1 AND support_type = '咨询机构'
	`

	var org struct {
		ID            uint         `db:"id"`
		DiseaseValue  int          `db:"disease_value"`
		Name          string       `db:"name"`
		ContentIntro  string       `db:"content_intro"`
		OrgAddress    string       `db:"org_address"`
		OrgContact    string       `db:"org_contact"`
		IsFree        sql.NullBool `db:"is_free"`
		ConsultWay    string       `db:"consult_way"`
		GuidePDF      string       `db:"guide_pdf"`
		ManualPatient string       `db:"manual_patient"`
		ManualFamily  string       `db:"manual_family"`
		CreatedAt     time.Time    `db:"created_at"`
		UpdatedAt     time.Time    `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&org.ID, &org.DiseaseValue, &org.Name, &org.ContentIntro,
		&org.OrgAddress, &org.OrgContact, &org.IsFree, &org.ConsultWay,
		&org.GuidePDF, &org.ManualPatient, &org.ManualFamily,
		&org.CreatedAt, &org.UpdatedAt,
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
			"message": "查询机构详情失败",
		})
		return
	}

	// 解析联系方式
	phone, email, website := parsePsychologicalContact(org.OrgContact)

	// 提取地区
	region := extractRegionFromAddress(org.OrgAddress)
	regionCode := convertRegionToCode(region)

	// 转换机构类型
	orgType, orgTypeName := convertPsychologicalOrgType(org.Name, org.ConsultWay)

	// 解析服务项目
	services := parsePsychologicalServices(org.Name, org.ConsultWay)

	// 获取服务时间
	serviceTime := getServiceTime(orgType)

	isFree := false
	if org.IsFree.Valid {
		isFree = org.IsFree.Bool
	}

	// 构建图片列表
	images := []string{
		"https://example.com/orgs/psychological/" + strconv.FormatUint(id, 10) + "_1.jpg",
		"https://example.com/orgs/psychological/" + strconv.FormatUint(id, 10) + "_2.jpg",
	}

	// 构建咨询师列表
	counselors := getPsychologicalCounselors(org.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": PsychologicalOrgDetailResponse{
			ID:          org.ID,
			Name:        org.Name,
			Type:        orgType,
			TypeName:    orgTypeName,
			Region:      region,
			RegionCode:  regionCode,
			Address:     org.OrgAddress,
			Contact:     org.OrgContact,
			Phone:       phone,
			Email:       email,
			Website:     website,
			IsFree:      isFree,
			ServiceTime: serviceTime,
			Description: org.ContentIntro,
			Services:    services,
			Rating:      4.5 + float64(org.ID%10)/10,
			CoverUrl:    "https://example.com/orgs/psychological/" + strconv.FormatUint(id, 10) + ".jpg",
			Images:      images,
			Counselors:  counselors,
			Status:      "active",
		},
	})
}

// convertCodeToRegion 将地区代码转换为地区名称
func convertCodeToRegion(code string) string {
	regionMap := map[string]string{
		"all": "全国",
		"bj":  "北京",
		"sh":  "上海",
		"gz":  "广州",
		"sz":  "深圳",
		"zj":  "浙江",
		"js":  "江苏",
		"sc":  "四川",
		"hb":  "湖北",
		"sd":  "山东",
	}
	if name, ok := regionMap[code]; ok {
		return name
	}
	return ""
}

// extractRegionFromAddress 从地址中提取地区
func extractRegionFromAddress(address string) string {
	if address == "" {
		return "全国"
	}
	if strings.Contains(address, "北京") {
		return "北京"
	}
	if strings.Contains(address, "上海") {
		return "上海"
	}
	if strings.Contains(address, "广州") {
		return "广州"
	}
	if strings.Contains(address, "深圳") {
		return "深圳"
	}
	return "全国"
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

// PsychologicalGuideItem 心理疏导指南项响应结构
type PsychologicalGuideItem struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Target        string `json:"target"`
	TargetName    string `json:"targetName"`
	Desc          string `json:"desc"`
	CoverUrl      string `json:"coverUrl"`
	DownloadURL   string `json:"downloadUrl"`
	FileSize      string `json:"fileSize"`
	ViewCount     int    `json:"viewCount"`
	DownloadCount int    `json:"downloadCount"`
	UpdateTime    string `json:"updateTime"`
	Status        string `json:"status"`
}

// PsychologicalGuideListResponse 指南列表响应结构
type PsychologicalGuideListResponse struct {
	List     []PsychologicalGuideItem `json:"list"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
}

// PsychologicalGuideDownloadResponse 指南下载响应结构
type PsychologicalGuideDownloadResponse struct {
	FileName    string `json:"fileName"`
	DownloadURL string `json:"downloadUrl"`
	FileSize    string `json:"fileSize"`
	ExpireTime  int    `json:"expireTime"`
}

// GetPsychologicalGuides 获取心理疏导指南列表
// GetPsychologicalGuides 获取心理疏导指南列表
func GetPsychologicalGuides(c *gin.Context) {
	// 获取请求参数
	target := c.DefaultQuery("target", "")
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
	whereClause := "WHERE is_audit = 1 AND support_type IN ('疏导指南', '心理手册')"
	args := []interface{}{}

	if keyword != "" {
		whereClause += " AND (name LIKE ? OR content_intro LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM psychological_supports " + whereClause
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
		SELECT id, disease_value, name, content_intro, guide_pdf,
		       manual_patient, manual_family, created_at, updated_at
		FROM psychological_supports
		` + whereClause + `
		ORDER BY id DESC
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

	var list []PsychologicalGuideItem
	for rows.Next() {
		var guide struct {
			ID            uint      `db:"id"`
			DiseaseValue  int       `db:"disease_value"`
			Name          string    `db:"name"`
			ContentIntro  string    `db:"content_intro"`
			GuidePDF      string    `db:"guide_pdf"`
			ManualPatient string    `db:"manual_patient"`
			ManualFamily  string    `db:"manual_family"`
			CreatedAt     time.Time `db:"created_at"`
			UpdatedAt     time.Time `db:"updated_at"`
		}
		if err := rows.Scan(
			&guide.ID, &guide.DiseaseValue, &guide.Name, &guide.ContentIntro,
			&guide.GuidePDF, &guide.ManualPatient, &guide.ManualFamily,
			&guide.CreatedAt, &guide.UpdatedAt,
		); err != nil {
			continue
		}

		// 根据 target 筛选（内联 matchTarget 逻辑）
		if target != "" {
			skip := false
			if target == "patient" {
				if guide.ManualPatient == "" && !strings.Contains(guide.Name, "患者") {
					skip = true
				}
			} else if target == "family" {
				if guide.ManualFamily == "" && !strings.Contains(guide.Name, "家属") {
					skip = true
				}
			} else if target == "child" {
				if !strings.Contains(guide.Name, "儿童") {
					skip = true
				}
			}
			if skip {
				continue
			}
		}

		// 转换目标人群
		target, targetName := convertGuideTarget(guide.Name, guide.ManualPatient, guide.ManualFamily)

		// 获取下载链接
		var downloadURL string
		if target == "family" && guide.ManualFamily != "" {
			downloadURL = guide.ManualFamily
		} else if target == "patient" && guide.ManualPatient != "" {
			downloadURL = guide.ManualPatient
		} else if guide.GuidePDF != "" {
			downloadURL = guide.GuidePDF
		} else if guide.ManualPatient != "" {
			downloadURL = guide.ManualPatient
		} else if guide.ManualFamily != "" {
			downloadURL = guide.ManualFamily
		}

		if downloadURL == "" {
			continue
		}

		// 获取文件大小
		fileSize := getFileSize(downloadURL)

		list = append(list, PsychologicalGuideItem{
			ID:            guide.ID,
			Title:         guide.Name,
			Target:        target,
			TargetName:    targetName,
			Desc:          truncateString(guide.ContentIntro, 50),
			CoverUrl:      "https://example.com/guides/psychological/" + strconv.FormatUint(uint64(guide.ID), 10) + ".jpg",
			DownloadURL:   downloadURL,
			FileSize:      fileSize,
			ViewCount:     1000 + int(guide.ID%10)*100,
			DownloadCount: 300 + int(guide.ID%10)*30,
			UpdateTime:    guide.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Status:        "published",
		})
	}

	// 确保数组不为 null
	if list == nil {
		list = []PsychologicalGuideItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": PsychologicalGuideListResponse{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// GetPsychologicalGuideDownload 下载心理疏导指南
func GetPsychologicalGuideDownload(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的指南 ID",
		})
		return
	}

	query := `
		SELECT id, name, guide_pdf, manual_patient, manual_family, updated_at
		FROM psychological_supports
		WHERE id = ? AND is_audit = 1 AND support_type IN ('疏导指南', '心理手册')
	`

	var guide struct {
		ID            uint      `db:"id"`
		Name          string    `db:"name"`
		GuidePDF      string    `db:"guide_pdf"`
		ManualPatient string    `db:"manual_patient"`
		ManualFamily  string    `db:"manual_family"`
		UpdatedAt     time.Time `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&guide.ID, &guide.Name, &guide.GuidePDF,
		&guide.ManualPatient, &guide.ManualFamily, &guide.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "指南不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询指南失败",
		})
		return
	}

	// 获取下载链接（优先使用 guide_pdf）
	downloadURL := guide.GuidePDF
	if downloadURL == "" {
		downloadURL = guide.ManualPatient
	}
	if downloadURL == "" {
		downloadURL = guide.ManualFamily
	}

	if downloadURL == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "该指南暂无下载文件",
		})
		return
	}

	fileName := guide.Name + ".pdf"
	fileSize := getFileSize(downloadURL)
	expireTime := 3600 // 链接有效期 1 小时

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": PsychologicalGuideDownloadResponse{
			FileName:    fileName,
			DownloadURL: downloadURL,
			FileSize:    fileSize,
			ExpireTime:  expireTime,
		},
	})
}

// convertGuideTarget 转换指南目标人群
func convertGuideTarget(name, manualPatient, manualFamily string) (string, string) {
	if strings.Contains(name, "家属") || manualFamily != "" {
		return "family", "家属"
	}
	if strings.Contains(name, "儿童") {
		return "child", "儿童患者"
	}
	if strings.Contains(name, "患者") || manualPatient != "" {
		return "patient", "患者"
	}
	return "general", "通用"
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

// GetPsychologicalOrgRegions 获取心理咨询机构地区筛选选项
func GetPsychologicalOrgRegions(c *gin.Context) {
	regions := []RegionItem{
		{Text: "全部地区", Value: "all"},
		{Text: "全国", Value: "all"},
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
