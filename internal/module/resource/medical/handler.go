package medical

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"rare_backend/internal/pkg/db"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

// GetGuideList 获取指南列表
func GetGuideList(c *gin.Context) {
	// 获取请求参数
	diseaseStr := c.DefaultQuery("disease", "0")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	disease, err := strconv.Atoi(diseaseStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的疾病 ID",
		})
		return
	}

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := ""
	args := []interface{}{}
	if disease != 0 {
		whereClause = " AND g.disease_value = ?"
		args = append(args, disease)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM guides g WHERE g.is_audit = 1" + whereClause
	var total int64
	err = db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 查询列表
	listQuery := `
        SELECT g.id, g.title, g.org, g.year, g.summary, g.created_at
        FROM guides g
        WHERE g.is_audit = 1` + whereClause + `
        ORDER BY g.is_latest DESC, g.sort ASC, g.created_at DESC
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

	var list []map[string]interface{}
	for rows.Next() {
		var guide struct {
			ID        int64  `db:"id"`
			Title     string `db:"title"`
			Org       string `db:"org"`
			Year      string `db:"year"`
			Summary   string `db:"summary"`
			CreatedAt string `db:"created_at"`
		}
		if err := rows.Scan(&guide.ID, &guide.Title, &guide.Org, &guide.Year, &guide.Summary, &guide.CreatedAt); err != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id":        guide.ID,
			"title":     guide.Title,
			"org":       guide.Org,
			"year":      guide.Year,
			"summary":   guide.Summary,
			"createdAt": guide.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total": total,
			"list":  list,
		},
	})
}

// GetGuideDetail 获取指南详情
func GetGuideDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的指南 ID",
		})
		return
	}

	// 查询 guides 表
	query := `
		SELECT g.id, g.title, g.org, g.year, g.version, g.guide_type,
		       g.is_latest, g.summary, g.content, g.pdf_url_hd,
		       g.pdf_url_zip, g.word_url, g.preview_url, g.explain_url,
		       g.is_audit, g.sort, g.created_at
		FROM guides g
		WHERE g.id = ? AND g.is_audit = 1
	`

	var guide struct {
		ID         int64          `db:"id"`
		Title      string         `db:"title"`
		Org        string         `db:"org"`
		Year       string         `db:"year"`
		Version    string         `db:"version"`
		GuideType  string         `db:"guide_type"`
		IsLatest   int8           `db:"is_latest"`
		Summary    sql.NullString `db:"summary"`
		Content    sql.NullString `db:"content"`
		PdfUrlHd   string         `db:"pdf_url_hd"`
		PdfUrlZip  string         `db:"pdf_url_zip"`
		WordUrl    string         `db:"word_url"`
		PreviewUrl sql.NullString `db:"preview_url"`
		ExplainUrl sql.NullString `db:"explain_url"`
		IsAudit    int8           `db:"is_audit"`
		Sort       int            `db:"sort"`
		CreatedAt  string         `db:"created_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&guide.ID, &guide.Title, &guide.Org, &guide.Year, &guide.Version,
		&guide.GuideType, &guide.IsLatest, &guide.Summary, &guide.Content,
		&guide.PdfUrlHd, &guide.PdfUrlZip, &guide.WordUrl, &guide.PreviewUrl,
		&guide.ExplainUrl, &guide.IsAudit, &guide.Sort, &guide.CreatedAt,
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
			"message": "查询指南详情失败",
		})
		return
	}

	// 构建期望的返回格式
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":        guide.ID,
			"title":     guide.Title,
			"org":       guide.Org,
			"year":      guide.Year,
			"summary":   guide.Summary.String,
			"content":   guide.Content.String,
			"pdfUrl":    guide.PdfUrlHd, // 使用高清版 PDF 地址
			"createdAt": guide.CreatedAt,
		},
	})
}

// GetInspectionDetail 获取检查检验手册详情
func GetInspectionDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的检查 ID",
		})
		return
	}

	query := `
		SELECT id, exam_name, exam_type, exam_purpose, reference_value,
		       abnormal_interpret, sample_notes, institution,
		       template_excel, template_word, compare_template, created_at
		FROM examination_manuals
		WHERE id = ? AND is_audit = 1
	`

	var item struct {
		ID                int64          `db:"id"`
		ExamName          string         `db:"exam_name"`
		ExamType          string         `db:"exam_type"`
		ExamPurpose       string         `db:"exam_purpose"`
		ReferenceValue    sql.NullString `db:"reference_value"`
		AbnormalInterpret sql.NullString `db:"abnormal_interpret"`
		SampleNotes       sql.NullString `db:"sample_notes"`
		Institution       sql.NullString `db:"institution"`
		TemplateExcel     sql.NullString `db:"template_excel"`
		TemplateWord      sql.NullString `db:"template_word"`
		CompareTemplate   sql.NullString `db:"compare_template"`
		CreatedAt         string         `db:"created_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&item.ID, &item.ExamName, &item.ExamType, &item.ExamPurpose,
		&item.ReferenceValue, &item.AbnormalInterpret, &item.SampleNotes,
		&item.Institution, &item.TemplateExcel, &item.TemplateWord,
		&item.CompareTemplate, &item.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "检查手册不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询检查手册详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":                item.ID,
			"name":              item.ExamName,
			"category":          item.ExamType,
			"desc":              item.ExamPurpose,
			"referenceValue":    item.ReferenceValue.String,
			"abnormalInterpret": item.AbnormalInterpret.String,
			"preparation":       item.SampleNotes.String,
			"institution":       item.Institution.String,
			"templates": gin.H{
				"excel":   item.TemplateExcel.String,
				"word":    item.TemplateWord.String,
				"compare": item.CompareTemplate.String,
			},
			"createdAt": item.CreatedAt,
		},
	})
}

// GetInspectionList 获取检查检验手册列表
func GetInspectionList(c *gin.Context) {
	// 获取请求参数
	diseaseStr := c.DefaultQuery("disease", "0")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	disease, err := strconv.Atoi(diseaseStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的疾病 ID",
		})
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := ""
	args := []interface{}{}
	if disease != 0 {
		whereClause = " AND disease_value = ?"
		args = append(args, disease)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM examination_manuals WHERE is_audit = 1" + whereClause
	var total int64
	err = db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 查询列表
	listQuery := `
		SELECT id, exam_name, exam_purpose, exam_type, sample_notes, created_at
		FROM examination_manuals
		WHERE is_audit = 1` + whereClause + `
		ORDER BY sort ASC, created_at DESC
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

	var list []map[string]interface{}
	for rows.Next() {
		var item struct {
			ID          int64  `db:"id"`
			Name        string `db:"exam_name"`
			Desc        string `db:"exam_purpose"`
			Category    string `db:"exam_type"`
			Preparation string `db:"sample_notes"`
			CreatedAt   string `db:"created_at"`
		}
		if err := rows.Scan(&item.ID, &item.Name, &item.Desc, &item.Category, &item.Preparation, &item.CreatedAt); err != nil {
			continue
		}

		// 根据检查类型映射价格和时长
		price, duration := mapPriceAndDuration(item.Category)

		list = append(list, map[string]interface{}{
			"id":          item.ID,
			"name":        item.Name,
			"desc":        item.Desc,
			"category":    item.Category,
			"price":       price,
			"duration":    duration,
			"preparation": item.Preparation,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total": total,
			"list":  list,
		},
	})
}

// mapPriceAndDuration 根据检查类型映射价格和时长
func mapPriceAndDuration(category string) (int, string) {
	// 可根据实际需求扩展映射规则
	priceMap := map[string]int{
		"基因检测":  299,
		"血液检查":  50,
		"影像学检查": 200,
		"病理检查":  150,
		"生化检查":  80,
	}

	durationMap := map[string]string{
		"基因检测":  "3-5 个工作日",
		"血液检查":  "当天出结果",
		"影像学检查": "1-2 个工作日",
		"病理检查":  "5-7 个工作日",
		"生化检查":  "当天出结果",
	}

	price, ok := priceMap[category]
	if !ok {
		price = 100 // 默认价格
	}

	duration, ok := durationMap[category]
	if !ok {
		duration = "1-3 个工作日" // 默认时长
	}

	return price, duration
}

// GetDirectoryList 获取医生/医院名录列表
func GetDirectoryList(c *gin.Context) {
	// 获取请求参数
	keyword := c.DefaultQuery("keyword", "")
	diseaseStr := c.DefaultQuery("disease", "0")
	region := c.DefaultQuery("region", "")
	level := c.DefaultQuery("level", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")

	disease, _ := strconv.Atoi(diseaseStr)
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 查询医生总数
	doctorWhere := "WHERE d.is_audit = 1"
	doctorArgs := []interface{}{}
	if keyword != "" {
		doctorWhere += " AND (d.name LIKE ? OR d.good_at LIKE ?)"
		doctorArgs = append(doctorArgs, "%"+keyword+"%", "%"+keyword+"%")
	}
	if disease != 0 {
		doctorWhere += " AND d.disease_value = ?"
		doctorArgs = append(doctorArgs, disease)
	}
	if region != "" {
		doctorWhere += " AND h.region = ?"
		doctorArgs = append(doctorArgs, region)
	}
	if level != "" {
		doctorWhere += " AND h.level = ?"
		doctorArgs = append(doctorArgs, level)
	}

	var doctorTotal int64
	doctorCountQuery := "SELECT COUNT(*) FROM doctors d JOIN hospitals h ON d.hospital_id = h.id " + doctorWhere
	db.MySQL.QueryRow(doctorCountQuery, doctorArgs...).Scan(&doctorTotal)

	// 查询医院总数
	hospitalWhere := "WHERE is_audit = 1"
	hospitalArgs := []interface{}{}
	if keyword != "" {
		hospitalWhere += " AND (name LIKE ? OR treat_scope LIKE ?)"
		hospitalArgs = append(hospitalArgs, "%"+keyword+"%", "%"+keyword+"%")
	}
	if region != "" {
		hospitalWhere += " AND region = ?"
		hospitalArgs = append(hospitalArgs, region)
	}
	if level != "" {
		hospitalWhere += " AND level = ?"
		hospitalArgs = append(hospitalArgs, level)
	}

	var hospitalTotal int64
	hospitalCountQuery := "SELECT COUNT(*) FROM hospitals " + hospitalWhere
	db.MySQL.QueryRow(hospitalCountQuery, hospitalArgs...).Scan(&hospitalTotal)

	total := doctorTotal + hospitalTotal

	// 查询医生列表
	doctorListQuery := `
		SELECT d.id, d.name, d.title, d.department, d.good_at, d.clinic_time,
		       d.contact, d.score, d.comment_num, h.name as hospital_name,
		       h.region, h.is_rare_network
		FROM doctors d
		JOIN hospitals h ON d.hospital_id = h.id
		` + doctorWhere + `
		ORDER BY d.score DESC, d.comment_num DESC
		LIMIT ? OFFSET ?
	`
	doctorArgs = append(doctorArgs, pageSize, offset)

	doctorRows, err := db.MySQL.Query(doctorListQuery, doctorArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医生列表失败",
		})
		return
	}
	defer doctorRows.Close()

	var list []map[string]interface{}

	// 扫描医生数据
	for doctorRows.Next() {
		var doctor struct {
			ID            int64          `db:"id"`
			Name          string         `db:"name"`
			Title         string         `db:"title"`
			Department    string         `db:"department"`
			GoodAt        sql.NullString `db:"good_at"`
			ClinicTime    sql.NullString `db:"clinic_time"`
			Contact       sql.NullString `db:"contact"`
			Score         float64        `db:"score"`
			CommentNum    int            `db:"comment_num"`
			HospitalName  string         `db:"hospital_name"`
			Region        string         `db:"region"`
			IsRareNetwork int8           `db:"is_rare_network"`
		}
		if err := doctorRows.Scan(
			&doctor.ID, &doctor.Name, &doctor.Title, &doctor.Department,
			&doctor.GoodAt, &doctor.ClinicTime, &doctor.Contact,
			&doctor.Score, &doctor.CommentNum, &doctor.HospitalName,
			&doctor.Region, &doctor.IsRareNetwork,
		); err != nil {
			continue
		}

		list = append(list, map[string]interface{}{
			"id":              doctor.ID,
			"type":            "doctor",
			"name":            doctor.Name,
			"hospital":        doctor.HospitalName,
			"department":      doctor.Department,
			"title":           doctor.Title,
			"specialty":       doctor.GoodAt.String,
			"region":          doctor.Region,
			"rating":          doctor.Score,
			"reviewCount":     doctor.CommentNum,
			"isNetworkMember": doctor.IsRareNetwork == 1,
		})
	}

	// 查询医院列表
	hospitalListQuery := `
		SELECT id, name, level, is_rare_network, treat_scope, address, region, phone
		FROM hospitals
		` + hospitalWhere + `
		ORDER BY is_rare_network DESC, created_at DESC
		LIMIT ? OFFSET ?
	`
	hospitalArgs = append(hospitalArgs, pageSize, offset)

	hospitalRows, err := db.MySQL.Query(hospitalListQuery, hospitalArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医院列表失败",
		})
		return
	}
	defer hospitalRows.Close()

	// 扫描医院数据
	for hospitalRows.Next() {
		var hospital struct {
			ID            int64          `db:"id"`
			Name          string         `db:"name"`
			Level         string         `db:"level"`
			IsRareNetwork int8           `db:"is_rare_network"`
			TreatScope    sql.NullString `db:"treat_scope"`
			Address       string         `db:"address"`
			Region        string         `db:"region"`
			Phone         string         `db:"phone"`
		}
		if err := hospitalRows.Scan(
			&hospital.ID, &hospital.Name, &hospital.Level, &hospital.IsRareNetwork,
			&hospital.TreatScope, &hospital.Address, &hospital.Region, &hospital.Phone,
		); err != nil {
			continue
		}

		list = append(list, map[string]interface{}{
			"id":              hospital.ID,
			"type":            "hospital",
			"name":            hospital.Name,
			"level":           hospital.Level,
			"address":         hospital.Address,
			"specialty":       hospital.TreatScope.String,
			"region":          hospital.Region,
			"rating":          4.5, // 医院默认评分
			"reviewCount":     0,   // 医院暂无评价数
			"isNetworkMember": hospital.IsRareNetwork == 1,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total": total,
			"list":  list,
		},
	})
}

// ExportDoctors 导出医生名录 Excel

// ExportDoctors 导出医生名录 Excel
// ExportDoctors 导出医生名录 Excel
func ExportDoctors(c *gin.Context) {
	// 获取筛选参数（与列表接口相同）
	keyword := c.DefaultQuery("keyword", "")
	diseaseStr := c.DefaultQuery("disease", "0")
	region := c.DefaultQuery("region", "")
	level := c.DefaultQuery("level", "")

	disease, _ := strconv.Atoi(diseaseStr)

	// 构建查询条件
	whereClause := "WHERE d.is_audit = 1"
	args := []interface{}{}
	if keyword != "" {
		whereClause += " AND (d.name LIKE ? OR d.good_at LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	if disease != 0 {
		whereClause += " AND d.disease_value = ?"
		args = append(args, disease)
	}
	if region != "" {
		whereClause += " AND h.region = ?"
		args = append(args, region)
	}
	if level != "" {
		whereClause += " AND h.level = ?"
		args = append(args, level)
	}

	// 查询医生数据
	query := `
		SELECT d.name, h.name as hospital_name, d.department, d.title,
		       d.good_at, h.region, d.clinic_time, d.contact, d.score
		FROM doctors d
		JOIN hospitals h ON d.hospital_id = h.id
		` + whereClause + `
		ORDER BY d.score DESC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医生数据失败",
		})
		return
	}
	defer rows.Close()

	// 创建 Excel 文件
	excel := excelize.NewFile()
	sheetName := "医生名录"
	// 创建新工作表并检查错误
	index, err := excel.NewSheet(sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建 Excel 工作表失败",
		})
		return
	}
	excel.SetActiveSheet(index)
	// 删除默认创建的 Sheet1，保持文件整洁
	excel.DeleteSheet("Sheet1")

	// 设置表头
	headers := []string{"姓名", "医院", "科室", "职称", "擅长领域", "地区", "出诊时间", "联系方式", "评价评分"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		excel.SetCellValue(sheetName, cell, header)
	}

	// 填充数据
	rowNum := 2
	for rows.Next() {
		var doctor struct {
			Name         string         `db:"name"`
			HospitalName string         `db:"hospital_name"`
			Department   string         `db:"department"`
			Title        string         `db:"title"`
			GoodAt       sql.NullString `db:"good_at"`
			Region       string         `db:"region"`
			ClinicTime   sql.NullString `db:"clinic_time"`
			Contact      sql.NullString `db:"contact"`
			Score        float64        `db:"score"`
		}
		if err := rows.Scan(
			&doctor.Name, &doctor.HospitalName, &doctor.Department, &doctor.Title,
			&doctor.GoodAt, &doctor.Region, &doctor.ClinicTime, &doctor.Contact, &doctor.Score,
		); err != nil {
			continue
		}

		data := []interface{}{
			doctor.Name,
			doctor.HospitalName,
			doctor.Department,
			doctor.Title,
			doctor.GoodAt.String,
			doctor.Region,
			doctor.ClinicTime.String,
			maskContact(doctor.Contact.String), // 加密联系方式
			doctor.Score,
		}

		for i, value := range data {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowNum)
			excel.SetCellValue(sheetName, cell, value)
		}
		rowNum++
	}

	// 设置列宽
	for i := 1; i <= len(headers); i++ {
		// 修复：v2 版本使用 ColumnNumberToName 将列号转换为字母
		col, _ := excelize.ColumnNumberToName(i)
		excel.SetColWidth(sheetName, col, col, 15)
	}

	// 设置响应头
	filename := "医生名录.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// 输出文件
	if err := excel.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成 Excel 文件失败",
		})
		return
	}
}

// ExportHospitals 导出医院名录 PDF
func ExportHospitals(c *gin.Context) {
	// 获取筛选参数
	keyword := c.DefaultQuery("keyword", "")
	region := c.DefaultQuery("region", "")
	level := c.DefaultQuery("level", "")

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}
	if keyword != "" {
		whereClause += " AND (name LIKE ? OR treat_scope LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	if region != "" {
		whereClause += " AND region = ?"
		args = append(args, region)
	}
	if level != "" {
		whereClause += " AND level = ?"
		args = append(args, level)
	}

	// 查询医院数据
	query := `
		SELECT name, level, is_rare_network, treat_scope, address, phone, region
		FROM hospitals
		` + whereClause + `
		ORDER BY is_rare_network DESC, created_at DESC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医院数据失败",
		})
		return
	}
	defer rows.Close()

	// 收集医院数据
	var hospitals []struct {
		Name          string `db:"name"`
		Level         string `db:"level"`
		IsRareNetwork int8   `db:"is_rare_network"`
		TreatScope    string `db:"treat_scope"`
		Address       string `db:"address"`
		Phone         string `db:"phone"`
		Region        string `db:"region"`
	}

	for rows.Next() {
		var h struct {
			Name          string         `db:"name"`
			Level         string         `db:"level"`
			IsRareNetwork int8           `db:"is_rare_network"`
			TreatScope    sql.NullString `db:"treat_scope"`
			Address       string         `db:"address"`
			Phone         string         `db:"phone"`
			Region        string         `db:"region"`
		}
		if err := rows.Scan(&h.Name, &h.Level, &h.IsRareNetwork, &h.TreatScope, &h.Address, &h.Phone, &h.Region); err != nil {
			continue
		}
		hospitals = append(hospitals, struct {
			Name          string `db:"name"`
			Level         string `db:"level"`
			IsRareNetwork int8   `db:"is_rare_network"`
			TreatScope    string `db:"treat_scope"`
			Address       string `db:"address"`
			Phone         string `db:"phone"`
			Region        string `db:"region"`
		}{
			Name:          h.Name,
			Level:         h.Level,
			IsRareNetwork: h.IsRareNetwork,
			TreatScope:    h.TreatScope.String,
			Address:       h.Address,
			Phone:         h.Phone,
			Region:        h.Region,
		})
	}

	// 创建 PDF（使用支持中文的字体）
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// 注册中文字体（需要提前下载字体文件放到 font 目录）
	pdf.AddFont("SimSun", "", "simSun.json")
	pdf.SetFont("SimSun", "", 12)

	// 标题
	pdf.SetFontSize(16)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 10, "罕见病诊疗医院名录")
	pdf.Ln(15)

	// 表头
	pdf.SetFontSize(10)
	headers := []string{"医院名称", "医院等级", "协作网成员", "诊疗范围", "地址", "咨询电话", "地区"}
	widths := []float64{40, 15, 20, 40, 40, 25, 20}

	// 绘制表头
	pdf.SetFillColor(200, 200, 200)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.1)

	x := pdf.GetX()
	y := pdf.GetY()
	for i, header := range headers {
		pdf.Rect(x, y, widths[i], 8, "FD")
		pdf.CellFormat(widths[i], 8, header, "1", 0, "CM", true, 0, "")
		x += widths[i]
	}
	pdf.Ln(8)

	// 绘制数据行
	pdf.SetFillColor(255, 255, 255)
	pdf.SetFontSize(9)

	for _, h := range hospitals {
		x = pdf.GetX()
		y = pdf.GetY()

		// 计算最高行高（处理多行文本）
		rowHeight := 8.0
		treatScopeLines := len(h.TreatScope) / 20
		if treatScopeLines > 1 {
			rowHeight = float64(treatScopeLines) * 8
		}

		// 绘制单元格
		data := []string{
			h.Name,
			h.Level,
			mapNetworkMember(h.IsRareNetwork),
			truncateString(h.TreatScope, 20),
			truncateString(h.Address, 20),
			h.Phone,
			h.Region,
		}

		for i, value := range data {
			pdf.Rect(x, y, widths[i], rowHeight, "D")
			pdf.CellFormat(widths[i], rowHeight, value, "0", 0, "LM", false, 0, "")
			x += widths[i]
		}
		pdf.Ln(rowHeight)

		// 检查是否需要分页
		if pdf.GetY() > 250 {
			pdf.AddPage()
			pdf.SetFont("SimSun", "", 9)
		}
	}

	// 设置响应头
	filename := "医院名录.pdf"
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// 输出 PDF 到响应
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.String(http.StatusInternalServerError, "PDF生成失败: %v", err)
		return
	}
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// mapNetworkMember 将协作网成员状态转换为文本
func mapNetworkMember(isRareNetwork int8) string {
	if isRareNetwork == 1 {
		return "是"
	}
	return "否"
}

// truncateString 截断字符串并添加省略号
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// maskContact 加密联系方式
func maskContact(contact string) string {
	if len(contact) <= 4 {
		return "****"
	}
	return contact[:2] + "****" + contact[len(contact)-2:]
}

// GetDoctorDetail 获取医生详情
func GetDoctorDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的医生 ID",
		})
		return
	}

	// 查询医生基本信息（关联医院表）
	query := `
		SELECT d.id, d.name, d.title, d.department, d.good_at, d.clinic_time,
		       d.contact, d.score, d.comment_num, d.is_audit,
		       h.name as hospital_name, h.is_rare_network
		FROM doctors d
		JOIN hospitals h ON d.hospital_id = h.id
		WHERE d.id = ? AND d.is_audit = 1
	`

	var doctor struct {
		ID            uint    `db:"id"`
		Name          string  `db:"name"`
		Title         string  `db:"title"`
		Department    string  `db:"department"`
		GoodAt        string  `db:"good_at"`
		ClinicTime    string  `db:"clinic_time"`
		Contact       string  `db:"contact"`
		Score         float64 `db:"score"`
		CommentNum    int     `db:"comment_num"`
		IsAudit       int8    `db:"is_audit"`
		HospitalName  string  `db:"hospital_name"`
		IsRareNetwork int8    `db:"is_rare_network"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&doctor.ID, &doctor.Name, &doctor.Title, &doctor.Department,
		&doctor.GoodAt, &doctor.ClinicTime, &doctor.Contact,
		&doctor.Score, &doctor.CommentNum, &doctor.IsAudit,
		&doctor.HospitalName, &doctor.IsRareNetwork,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "医生不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医生详情失败",
		})
		return
	}

	// 查询医生评价列表（如果存在 doctor_reviews 表）
	// 如果暂时没有评价表，可以返回空数组
	var reviews []map[string]interface{}
	reviewQuery := `
		SELECT id, content, rating, created_at
		FROM doctor_reviews
		WHERE doctor_id = ? AND is_audit = 1
		ORDER BY created_at DESC
		LIMIT 10
	`

	reviewRows, err := db.MySQL.Query(reviewQuery, id)
	if err == nil {
		defer reviewRows.Close()
		for reviewRows.Next() {
			var review struct {
				ID        uint   `db:"id"`
				Content   string `db:"content"`
				Rating    int    `db:"rating"`
				CreatedAt string `db:"created_at"`
			}
			if err := reviewRows.Scan(&review.ID, &review.Content, &review.Rating, &review.CreatedAt); err != nil {
				continue
			}
			reviews = append(reviews, map[string]interface{}{
				"id":      review.ID,
				"content": review.Content,
				"date":    review.CreatedAt[:10],
				"rating":  review.Rating,
			})
		}
	}
	// 如果查询失败（表不存在），reviews 保持为空数组

	// 构建响应数据
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":              doctor.ID,
			"name":            doctor.Name,
			"hospital":        doctor.HospitalName,
			"department":      doctor.Department,
			"title":           doctor.Title,
			"specialty":       doctor.GoodAt,
			"schedule":        doctor.ClinicTime,
			"phone":           doctor.Contact, // 数据库中已加密存储
			"rating":          doctor.Score,
			"reviewCount":     doctor.CommentNum,
			"reviews":         reviews,
			"isNetworkMember": doctor.IsRareNetwork == 1,
		},
	})
}

// GetHospitalDetail 获取医院详情
func GetHospitalDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的医院 ID",
		})
		return
	}

	// 查询医院基本信息
	query := `
		SELECT id, name, level, treat_scope, address, phone,
		       hospital_url, is_rare_network, is_audit, created_at
		FROM hospitals
		WHERE id = ? AND is_audit = 1
	`

	var hospital struct {
		ID            uint           `db:"id"`
		Name          string         `db:"name"`
		Level         string         `db:"level"`
		TreatScope    sql.NullString `db:"treat_scope"`
		Address       string         `db:"address"`
		Phone         string         `db:"phone"`
		HospitalURL   sql.NullString `db:"hospital_url"`
		IsRareNetwork int8           `db:"is_rare_network"`
		IsAudit       int8           `db:"is_audit"`
		CreatedAt     string         `db:"created_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&hospital.ID, &hospital.Name, &hospital.Level,
		&hospital.TreatScope, &hospital.Address, &hospital.Phone,
		&hospital.HospitalURL, &hospital.IsRareNetwork,
		&hospital.IsAudit, &hospital.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "医院不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医院详情失败",
		})
		return
	}

	// 查询医院评价列表（如果存在 hospital_reviews 表）
	var reviews []map[string]interface{}
	var reviewCount int
	var rating float64 = 4.9 // 默认评分

	reviewQuery := `
		SELECT id, content, rating, created_at
		FROM hospital_reviews
		WHERE hospital_id = ? AND is_audit = 1
		ORDER BY created_at DESC
		LIMIT 10
	`

	reviewRows, err := db.MySQL.Query(reviewQuery, id)
	if err == nil {
		defer reviewRows.Close()
		for reviewRows.Next() {
			var review struct {
				ID        uint   `db:"id"`
				Content   string `db:"content"`
				Rating    int    `db:"rating"`
				CreatedAt string `db:"created_at"`
			}
			if err := reviewRows.Scan(&review.ID, &review.Content, &review.Rating, &review.CreatedAt); err != nil {
				continue
			}
			reviews = append(reviews, map[string]interface{}{
				"id":      review.ID,
				"content": review.Content,
				"date":    review.CreatedAt[:10],
				"rating":  review.Rating,
			})
		}
		reviewCount = len(reviews)

		// 计算平均评分
		if len(reviews) > 0 {
			totalRating := 0
			for _, r := range reviews {
				totalRating += r["rating"].(int)
			}
			rating = float64(totalRating) / float64(len(reviews))
		}
	}
	// 如果查询失败（表不存在），使用默认值

	// 构建响应数据
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":              hospital.ID,
			"name":            hospital.Name,
			"level":           hospital.Level,
			"specialty":       hospital.TreatScope.String,
			"address":         hospital.Address,
			"phone":           hospital.Phone,
			"hospitalUrl":     hospital.HospitalURL.String,
			"rating":          rating,
			"reviewCount":     reviewCount,
			"reviews":         reviews,
			"isNetworkMember": hospital.IsRareNetwork == 1,
		},
	})
}

// GetDoctorList 获取医生名录列表
func GetDoctorList(c *gin.Context) {
	// 获取请求参数
	keyword := c.DefaultQuery("keyword", "")
	diseaseStr := c.DefaultQuery("disease", "0")
	region := c.DefaultQuery("region", "")
	level := c.DefaultQuery("level", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")

	disease, _ := strconv.Atoi(diseaseStr)
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := "WHERE d.is_audit = 1"
	args := []interface{}{}
	if keyword != "" {
		whereClause += " AND (d.name LIKE ? OR d.good_at LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	if disease != 0 {
		whereClause += " AND d.disease_value = ?"
		args = append(args, disease)
	}
	if region != "" {
		whereClause += " AND h.region = ?"
		args = append(args, region)
	}
	if level != "" {
		whereClause += " AND h.level = ?"
		args = append(args, level)
	}

	// 查询医生总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM doctors d JOIN hospitals h ON d.hospital_id = h.id " + whereClause
	db.MySQL.QueryRow(countQuery, args...).Scan(&total)

	// 查询医生列表
	listQuery := `
		SELECT d.id, d.name, d.title, d.department, d.good_at, d.clinic_time,
		       d.contact, d.score, d.comment_num, h.name as hospital_name,
		       h.region, h.is_rare_network
		FROM doctors d
		JOIN hospitals h ON d.hospital_id = h.id
		` + whereClause + `
		ORDER BY d.score DESC, d.comment_num DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医生列表失败",
		})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var doctor struct {
			ID            int64          `db:"id"`
			Name          string         `db:"name"`
			Title         string         `db:"title"`
			Department    string         `db:"department"`
			GoodAt        sql.NullString `db:"good_at"`
			ClinicTime    sql.NullString `db:"clinic_time"`
			Contact       sql.NullString `db:"contact"`
			Score         float64        `db:"score"`
			CommentNum    int            `db:"comment_num"`
			HospitalName  string         `db:"hospital_name"`
			Region        string         `db:"region"`
			IsRareNetwork int8           `db:"is_rare_network"`
		}
		if err := rows.Scan(
			&doctor.ID, &doctor.Name, &doctor.Title, &doctor.Department,
			&doctor.GoodAt, &doctor.ClinicTime, &doctor.Contact,
			&doctor.Score, &doctor.CommentNum, &doctor.HospitalName,
			&doctor.Region, &doctor.IsRareNetwork,
		); err != nil {
			continue
		}

		list = append(list, map[string]interface{}{
			"id":              doctor.ID,
			"name":            doctor.Name,
			"hospital":        doctor.HospitalName,
			"department":      doctor.Department,
			"title":           doctor.Title,
			"specialty":       doctor.GoodAt.String,
			"region":          doctor.Region,
			"rating":          doctor.Score,
			"reviewCount":     doctor.CommentNum,
			"isNetworkMember": doctor.IsRareNetwork == 1,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total": total,
			"list":  list,
		},
	})
}

// GetHospitalList 获取医院名录列表
func GetHospitalList(c *gin.Context) {
	// 获取请求参数
	keyword := c.DefaultQuery("keyword", "")
	region := c.DefaultQuery("region", "")
	level := c.DefaultQuery("level", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}
	if keyword != "" {
		whereClause += " AND (name LIKE ? OR treat_scope LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	if region != "" {
		whereClause += " AND region = ?"
		args = append(args, region)
	}
	if level != "" {
		whereClause += " AND level = ?"
		args = append(args, level)
	}

	// 查询医院总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM hospitals " + whereClause
	db.MySQL.QueryRow(countQuery, args...).Scan(&total)

	// 查询医院列表
	listQuery := `
		SELECT id, name, level, is_rare_network, treat_scope, address, region, phone
		FROM hospitals
		` + whereClause + `
		ORDER BY is_rare_network DESC, created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医院列表失败",
		})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var hospital struct {
			ID            int64          `db:"id"`
			Name          string         `db:"name"`
			Level         string         `db:"level"`
			IsRareNetwork int8           `db:"is_rare_network"`
			TreatScope    sql.NullString `db:"treat_scope"`
			Address       string         `db:"address"`
			Region        string         `db:"region"`
			Phone         string         `db:"phone"`
		}
		if err := rows.Scan(
			&hospital.ID, &hospital.Name, &hospital.Level, &hospital.IsRareNetwork,
			&hospital.TreatScope, &hospital.Address, &hospital.Region, &hospital.Phone,
		); err != nil {
			continue
		}

		list = append(list, map[string]interface{}{
			"id":              hospital.ID,
			"name":            hospital.Name,
			"level":           hospital.Level,
			"address":         hospital.Address,
			"specialty":       hospital.TreatScope.String,
			"region":          hospital.Region,
			"rating":          4.5, // 医院默认评分
			"reviewCount":     0,   // 医院暂无评价数
			"isNetworkMember": hospital.IsRareNetwork == 1,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total": total,
			"list":  list,
		},
	})
}

// UpdateDoctor 修改医生信息
func UpdateDoctor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的医生 ID",
		})
		return
	}

	// 解析请求体
	var req struct {
		Name       string `json:"name"`
		Title      string `json:"title"`
		Department string `json:"department"`
		GoodAt     string `json:"good_at"`
		ClinicTime string `json:"clinic_time"`
		Contact    string `json:"contact"`
		HospitalID uint   `json:"hospital_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 参数校验
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "医生姓名不能为空",
		})
		return
	}

	// 检查医生是否存在
	checkQuery := "SELECT id FROM doctors WHERE id = ? AND is_audit = 1"
	var existID uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&existID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "医生不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医生失败",
		})
		return
	}

	// 构建更新语句
	updateQuery := `
		UPDATE doctors 
		SET name = ?, title = ?, department = ?, good_at = ?, 
		    clinic_time = ?, contact = ?, hospital_id = ?,
		    updated_at = NOW()
		WHERE id = ?
	`

	result, err := db.MySQL.Exec(
		updateQuery,
		req.Name, req.Title, req.Department, req.GoodAt,
		req.ClinicTime, req.Contact, req.HospitalID, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新医生信息失败",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新医生信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": id,
		},
	})
}

// DeleteDoctor 删除医生（软删除）
func DeleteDoctor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的医生 ID",
		})
		return
	}

	// 检查医生是否存在
	checkQuery := "SELECT id FROM doctors WHERE id = ? AND is_audit = 1"
	var existID uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&existID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "医生不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医生失败",
		})
		return
	}

	// 软删除：将 is_audit 设为 0
	deleteQuery := "UPDATE doctors SET is_audit = 0, updated_at = NOW() WHERE id = ?"
	result, err := db.MySQL.Exec(deleteQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除医生失败",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除医生失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": id,
		},
	})
}

// CreateDoctor 新增医生
func CreateDoctor(c *gin.Context) {
	// 解析请求体
	var req struct {
		Name         string `json:"name"`
		Title        string `json:"title"`
		Department   string `json:"department"`
		GoodAt       string `json:"good_at"`
		ClinicTime   string `json:"clinic_time"`
		Contact      string `json:"contact"`
		HospitalID   uint   `json:"hospital_id"`
		DiseaseValue int    `json:"disease_value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 参数校验
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "医生姓名不能为空",
		})
		return
	}

	if req.HospitalID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "所属医院不能为空",
		})
		return
	}

	// 检查医院是否存在
	hospitalCheckQuery := "SELECT id FROM hospitals WHERE id = ? AND is_audit = 1"
	var hospitalID uint
	err := db.MySQL.QueryRow(hospitalCheckQuery, req.HospitalID).Scan(&hospitalID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "所属医院不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医院失败",
		})
		return
	}

	// 检查医生是否已存在（同一医院下姓名不能重复）
	checkQuery := "SELECT id FROM doctors WHERE name = ? AND hospital_id = ? AND is_audit = 1"
	var existID uint
	err = db.MySQL.QueryRow(checkQuery, req.Name, req.HospitalID).Scan(&existID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": "该医院下已存在同名医生",
		})
		return
	}

	// 插入医生数据
	insertQuery := `
		INSERT INTO doctors (
			name, title, department, good_at, clinic_time, 
			contact, hospital_id, disease_value, is_audit, 
			score, comment_num, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, 0, 0, NOW(), NOW())
	`

	result, err := db.MySQL.Exec(
		insertQuery,
		req.Name, req.Title, req.Department, req.GoodAt,
		req.ClinicTime, req.Contact, req.HospitalID, req.DiseaseValue,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建医生失败",
		})
		return
	}

	// 获取插入的 ID
	doctorID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取医生 ID 失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": doctorID,
		},
	})
}

// CreateHospital 新增医院
func CreateHospital(c *gin.Context) {
	// 解析请求体
	var req struct {
		Name          string `json:"name"`
		Level         string `json:"level"`
		TreatScope    string `json:"treat_scope"`
		Address       string `json:"address"`
		Phone         string `json:"phone"`
		HospitalURL   string `json:"hospital_url"`
		Region        string `json:"region"`
		IsRareNetwork int8   `json:"is_rare_network"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 参数校验
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "医院名称不能为空",
		})
		return
	}

	// 检查医院是否已存在（同名医院不能重复）
	checkQuery := "SELECT id FROM hospitals WHERE name = ? AND is_audit = 1"
	var existID uint
	err := db.MySQL.QueryRow(checkQuery, req.Name).Scan(&existID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": "医院名称已存在",
		})
		return
	}

	// 插入医院数据
	insertQuery := `
		INSERT INTO hospitals (
			name, level, treat_scope, address, phone, 
			hospital_url, region, is_rare_network, is_audit, 
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW())
	`

	result, err := db.MySQL.Exec(
		insertQuery,
		req.Name, req.Level, req.TreatScope, req.Address,
		req.Phone, req.HospitalURL, req.Region, req.IsRareNetwork,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建医院失败",
		})
		return
	}

	// 获取插入的 ID
	hospitalID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取医院 ID 失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": hospitalID,
		},
	})
}

// UpdateHospital 修改医院信息
func UpdateHospital(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的医院 ID",
		})
		return
	}

	// 解析请求体
	var req struct {
		Name          string `json:"name"`
		Level         string `json:"level"`
		TreatScope    string `json:"treat_scope"`
		Address       string `json:"address"`
		Phone         string `json:"phone"`
		HospitalURL   string `json:"hospital_url"`
		Region        string `json:"region"`
		IsRareNetwork int8   `json:"is_rare_network"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 参数校验
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "医院名称不能为空",
		})
		return
	}

	// 检查医院是否存在
	checkQuery := "SELECT id FROM hospitals WHERE id = ? AND is_audit = 1"
	var existID uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&existID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "医院不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医院失败",
		})
		return
	}

	// 检查医院名称是否与其他医院重复（排除当前医院）
	nameCheckQuery := "SELECT id FROM hospitals WHERE name = ? AND id != ? AND is_audit = 1"
	var duplicateID uint
	err = db.MySQL.QueryRow(nameCheckQuery, req.Name, id).Scan(&duplicateID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": "医院名称已存在",
		})
		return
	}

	// 构建更新语句
	updateQuery := `
		UPDATE hospitals 
		SET name = ?, level = ?, treat_scope = ?, address = ?, 
		    phone = ?, hospital_url = ?, region = ?, is_rare_network = ?,
		    updated_at = NOW()
		WHERE id = ?
	`

	result, err := db.MySQL.Exec(
		updateQuery,
		req.Name, req.Level, req.TreatScope, req.Address,
		req.Phone, req.HospitalURL, req.Region, req.IsRareNetwork, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新医院信息失败",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新医院信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": id,
		},
	})
}

// DeleteHospital 删除医院（软删除）
func DeleteHospital(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的医院 ID",
		})
		return
	}

	// 检查医院是否存在
	checkQuery := "SELECT id FROM hospitals WHERE id = ? AND is_audit = 1"
	var existID uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&existID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "医院不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医院失败",
		})
		return
	}

	// 检查该医院下是否还有医生
	doctorCheckQuery := "SELECT COUNT(*) FROM doctors WHERE hospital_id = ? AND is_audit = 1"
	var doctorCount int64
	err = db.MySQL.QueryRow(doctorCheckQuery, id).Scan(&doctorCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询关联医生失败",
		})
		return
	}

	if doctorCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("该医院下还有 %d 名医生，无法删除", doctorCount),
		})
		return
	}

	// 软删除：将 is_audit 设为 0
	deleteQuery := "UPDATE hospitals SET is_audit = 0, updated_at = NOW() WHERE id = ?"
	result, err := db.MySQL.Exec(deleteQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除医院失败",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除医院失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": id,
		},
	})
}

// CreateGuide 新增诊疗指南
func CreateGuide(c *gin.Context) {
	// 解析请求体
	var req struct {
		Title        string `json:"title"`
		Org          string `json:"org"`
		Year         string `json:"year"`
		Version      string `json:"version"`
		GuideType    string `json:"guide_type"`
		Summary      string `json:"summary"`
		Content      string `json:"content"`
		PdfUrlHd     string `json:"pdf_url_hd"`
		PdfUrlZip    string `json:"pdf_url_zip"`
		WordUrl      string `json:"word_url"`
		PreviewUrl   string `json:"preview_url"`
		ExplainUrl   string `json:"explain_url"`
		DiseaseValue int    `json:"disease_value"`
		Sort         int    `json:"sort"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 参数校验
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "指南标题不能为空",
		})
		return
	}

	// 检查指南是否已存在（同标题同年份不能重复）
	checkQuery := "SELECT id FROM guides WHERE title = ? AND year = ? AND is_audit = 1"
	var existID uint
	err := db.MySQL.QueryRow(checkQuery, req.Title, req.Year).Scan(&existID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": "该指南已存在",
		})
		return
	}

	// 插入指南数据
	insertQuery := `
		INSERT INTO guides (
			title, org, year, version, guide_type, summary, content,
			pdf_url_hd, pdf_url_zip, word_url, preview_url, explain_url,
			disease_value, is_latest, is_audit, sort, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1, ?, NOW(), NOW())
	`

	result, err := db.MySQL.Exec(
		insertQuery,
		req.Title, req.Org, req.Year, req.Version, req.GuideType,
		req.Summary, req.Content, req.PdfUrlHd, req.PdfUrlZip,
		req.WordUrl, req.PreviewUrl, req.ExplainUrl,
		req.DiseaseValue, req.Sort,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建指南失败",
		})
		return
	}

	// 获取插入的 ID
	guideID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取指南 ID 失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": guideID,
		},
	})
}

// UpdateGuide 修改诊疗指南
func UpdateGuide(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的指南 ID",
		})
		return
	}

	// 解析请求体
	var req struct {
		Title        string `json:"title"`
		Org          string `json:"org"`
		Year         string `json:"year"`
		Version      string `json:"version"`
		GuideType    string `json:"guide_type"`
		Summary      string `json:"summary"`
		Content      string `json:"content"`
		PdfUrlHd     string `json:"pdf_url_hd"`
		PdfUrlZip    string `json:"pdf_url_zip"`
		WordUrl      string `json:"word_url"`
		PreviewUrl   string `json:"preview_url"`
		ExplainUrl   string `json:"explain_url"`
		DiseaseValue int    `json:"disease_value"`
		IsLatest     int8   `json:"is_latest"`
		Sort         int    `json:"sort"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 参数校验
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "指南标题不能为空",
		})
		return
	}

	// 检查指南是否存在
	checkQuery := "SELECT id FROM guides WHERE id = ? AND is_audit = 1"
	var existID uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&existID)
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

	// 检查指南标题是否与其他指南重复（排除当前指南）
	nameCheckQuery := "SELECT id FROM guides WHERE title = ? AND year = ? AND id != ? AND is_audit = 1"
	var duplicateID uint
	err = db.MySQL.QueryRow(nameCheckQuery, req.Title, req.Year, id).Scan(&duplicateID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": "该指南已存在",
		})
		return
	}

	// 构建更新语句
	updateQuery := `
		UPDATE guides 
		SET title = ?, org = ?, year = ?, version = ?, guide_type = ?,
		    summary = ?, content = ?, pdf_url_hd = ?, pdf_url_zip = ?,
		    word_url = ?, preview_url = ?, explain_url = ?,
		    disease_value = ?, is_latest = ?, sort = ?, updated_at = NOW()
		WHERE id = ?
	`

	result, err := db.MySQL.Exec(
		updateQuery,
		req.Title, req.Org, req.Year, req.Version, req.GuideType,
		req.Summary, req.Content, req.PdfUrlHd, req.PdfUrlZip,
		req.WordUrl, req.PreviewUrl, req.ExplainUrl,
		req.DiseaseValue, req.IsLatest, req.Sort, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新指南信息失败",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新指南信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": id,
		},
	})
}

// DeleteGuide 删除诊疗指南（软删除）
func DeleteGuide(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的指南 ID",
		})
		return
	}

	// 检查指南是否存在
	checkQuery := "SELECT id FROM guides WHERE id = ? AND is_audit = 1"
	var existID uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&existID)
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

	// 软删除：将 is_audit 设为 0
	deleteQuery := "UPDATE guides SET is_audit = 0, updated_at = NOW() WHERE id = ?"
	result, err := db.MySQL.Exec(deleteQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除指南失败",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除指南失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": id,
		},
	})
}

// CreateInspection 新增检查检验手册
func CreateInspection(c *gin.Context) {
	// 解析请求体
	var req struct {
		ExamName          string `json:"exam_name"`
		ExamType          string `json:"exam_type"`
		ExamPurpose       string `json:"exam_purpose"`
		ReferenceValue    string `json:"reference_value"`
		AbnormalInterpret string `json:"abnormal_interpret"`
		SampleNotes       string `json:"sample_notes"`
		Institution       string `json:"institution"`
		TemplateExcel     string `json:"template_excel"`
		TemplateWord      string `json:"template_word"`
		CompareTemplate   string `json:"compare_template"`
		DiseaseValue      int    `json:"disease_value"`
		Sort              int    `json:"sort"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 参数校验
	if req.ExamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "检查名称不能为空",
		})
		return
	}

	// 检查检查手册是否已存在（同名称不能重复）
	checkQuery := "SELECT id FROM examination_manuals WHERE exam_name = ? AND is_audit = 1"
	var existID uint
	err := db.MySQL.QueryRow(checkQuery, req.ExamName).Scan(&existID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": "该检查手册已存在",
		})
		return
	}

	// 插入检查手册数据
	insertQuery := `
		INSERT INTO examination_manuals (
			exam_name, exam_type, exam_purpose, reference_value,
			abnormal_interpret, sample_notes, institution,
			template_excel, template_word, compare_template,
			disease_value, is_audit, sort, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, NOW(), NOW())
	`

	result, err := db.MySQL.Exec(
		insertQuery,
		req.ExamName, req.ExamType, req.ExamPurpose, req.ReferenceValue,
		req.AbnormalInterpret, req.SampleNotes, req.Institution,
		req.TemplateExcel, req.TemplateWord, req.CompareTemplate,
		req.DiseaseValue, req.Sort,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建检查手册失败",
		})
		return
	}

	// 获取插入的 ID
	inspectionID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取检查手册 ID 失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": inspectionID,
		},
	})
}

// UpdateInspection 修改检查检验手册
func UpdateInspection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的检查 ID",
		})
		return
	}

	// 解析请求体
	var req struct {
		ExamName          string `json:"exam_name"`
		ExamType          string `json:"exam_type"`
		ExamPurpose       string `json:"exam_purpose"`
		ReferenceValue    string `json:"reference_value"`
		AbnormalInterpret string `json:"abnormal_interpret"`
		SampleNotes       string `json:"sample_notes"`
		Institution       string `json:"institution"`
		TemplateExcel     string `json:"template_excel"`
		TemplateWord      string `json:"template_word"`
		CompareTemplate   string `json:"compare_template"`
		DiseaseValue      int    `json:"disease_value"`
		Sort              int    `json:"sort"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 参数校验
	if req.ExamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "检查名称不能为空",
		})
		return
	}

	// 检查检查手册是否存在
	checkQuery := "SELECT id FROM examination_manuals WHERE id = ? AND is_audit = 1"
	var existID uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&existID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "检查手册不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询检查手册失败",
		})
		return
	}

	// 检查检查手册名称是否与其他手册重复（排除当前手册）
	nameCheckQuery := "SELECT id FROM examination_manuals WHERE exam_name = ? AND id != ? AND is_audit = 1"
	var duplicateID uint
	err = db.MySQL.QueryRow(nameCheckQuery, req.ExamName, id).Scan(&duplicateID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": "该检查手册已存在",
		})
		return
	}

	// 构建更新语句
	updateQuery := `
		UPDATE examination_manuals 
		SET exam_name = ?, exam_type = ?, exam_purpose = ?, reference_value = ?,
		    abnormal_interpret = ?, sample_notes = ?, institution = ?,
		    template_excel = ?, template_word = ?, compare_template = ?,
		    disease_value = ?, sort = ?, updated_at = NOW()
		WHERE id = ?
	`

	result, err := db.MySQL.Exec(
		updateQuery,
		req.ExamName, req.ExamType, req.ExamPurpose, req.ReferenceValue,
		req.AbnormalInterpret, req.SampleNotes, req.Institution,
		req.TemplateExcel, req.TemplateWord, req.CompareTemplate,
		req.DiseaseValue, req.Sort, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新检查手册失败",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新检查手册失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": id,
		},
	})
}

// DeleteInspection 删除检查手册（软删除）
func DeleteInspection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的检查 ID",
		})
		return
	}

	// 检查检查手册是否存在
	checkQuery := "SELECT id FROM examination_manuals WHERE id = ? AND is_audit = 1"
	var existID uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&existID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "检查手册不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询检查手册失败",
		})
		return
	}

	// 软删除：将 is_audit 设为 0
	deleteQuery := "UPDATE examination_manuals SET is_audit = 0, updated_at = NOW() WHERE id = ?"
	result, err := db.MySQL.Exec(deleteQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除检查手册失败",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除检查手册失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": id,
		},
	})
}
