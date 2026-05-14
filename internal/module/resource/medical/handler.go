package medical

import (
	"database/sql"
	"fmt"
	"net/http"
	"rare_backend/internal/pkg/db"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// mapPriceAndDuration 根据检查类型映射价格和时长
func mapPriceAndDuration(category string) (int, string) {
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
		price = 100
	}
	duration, ok := durationMap[category]
	if !ok {
		duration = "1-3 个工作日"
	}
	return price, duration
}

// GetHospitalList 获取医院名录列表
func GetHospitalList(c *gin.Context) {
	keyword := c.DefaultQuery("keyword", "")
	provinceCode := c.DefaultQuery("provinceCode", "")
	cityCode := c.DefaultQuery("cityCode", "")
	districtCode := c.DefaultQuery("districtCode", "")
	level := c.DefaultQuery("level", "")

	// 【新增】获取筛选参数
	isRareNetworkStr := c.DefaultQuery("isRareNetwork", "") // 0:否, 1:是
	auditStatusStr := c.DefaultQuery("auditStatus", "")     // 0:待审核, 1:已通过, 2:已驳回

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

	// 初始化 WHERE 子句
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 关键词搜索
	if keyword != "" {
		whereClause += " AND (name LIKE ? OR treat_scope LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 地区筛选
	if provinceCode != "" {
		whereClause += " AND province_code = ?"
		args = append(args, provinceCode)
	}
	if cityCode != "" {
		whereClause += " AND city_code = ?"
		args = append(args, cityCode)
	}
	if districtCode != "" {
		whereClause += " AND district_code = ?"
		args = append(args, districtCode)
	}

	// 医院等级筛选
	if level != "" {
		whereClause += " AND level = ?"
		args = append(args, level)
	}

	// 【新增】是否罕见病网络成员筛选
	if isRareNetworkStr != "" {
		isRareNetwork, err := strconv.Atoi(isRareNetworkStr)
		if err == nil {
			whereClause += " AND is_rare_network = ?"
			args = append(args, isRareNetwork)
		}
	}

	// 【新增】审核状态筛选
	if auditStatusStr != "" {
		auditStatus, err := strconv.Atoi(auditStatusStr)
		if err == nil {
			whereClause += " AND audit_status = ?"
			args = append(args, auditStatus)
		}
	}

	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM hospital " + whereClause
	if err := db.MySQL.QueryRow(countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询总数失败"})
		return
	}

	// 【修改点 1】查询列表中增加 audit_status 字段
	listQuery := `
		SELECT id, name, province_code, city_code, district_code, 
               province_name, city_name, district_name,
               level, is_rare_network, treat_scope, address, phone, hospital_url, audit_status, created_at
		FROM hospital
		` + whereClause + `
		ORDER BY is_rare_network DESC, created_at DESC
		LIMIT ? OFFSET ?
	`
	// 注意：分页参数最后追加
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询医院列表失败"})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		// 【修改点 2】结构体中增加 AuditStatus 字段
		var h struct {
			ID            uint64         `db:"id"`
			Name          string         `db:"name"`
			ProvinceCode  string         `db:"province_code"`
			CityCode      string         `db:"city_code"`
			DistrictCode  string         `db:"district_code"`
			ProvinceName  string         `db:"province_name"`
			CityName      string         `db:"city_name"`
			DistrictName  string         `db:"district_name"`
			Level         string         `db:"level"`
			IsRareNetwork int8           `db:"is_rare_network"`
			TreatScope    sql.NullString `db:"treat_scope"`
			Address       string         `db:"address"`
			Phone         string         `db:"phone"`
			HospitalURL   sql.NullString `db:"hospital_url"`
			AuditStatus   int8           `db:"audit_status"` // 新增
			CreatedAt     time.Time      `db:"created_at"`
		}

		// 【修改点 3】Scan 中增加 &h.AuditStatus
		if err := rows.Scan(&h.ID, &h.Name, &h.ProvinceCode, &h.CityCode, &h.DistrictCode, &h.ProvinceName, &h.CityName, &h.DistrictName, &h.Level, &h.IsRareNetwork, &h.TreatScope, &h.Address, &h.Phone, &h.HospitalURL, &h.AuditStatus, &h.CreatedAt); err != nil {
			continue
		}

		list = append(list, map[string]interface{}{
			"id":            h.ID,
			"name":          h.Name,
			"provinceCode":  h.ProvinceCode,
			"cityCode":      h.CityCode,
			"districtCode":  h.DistrictCode,
			"provinceName":  h.ProvinceName,
			"cityName":      h.CityName,
			"districtName":  h.DistrictName,
			"level":         h.Level,
			"address":       h.Address,
			"phone":         h.Phone,
			"hospitalUrl":   h.HospitalURL.String,
			"treatScope":    h.TreatScope.String,
			"isRareNetwork": h.IsRareNetwork == 1,
			"auditStatus":   h.AuditStatus, // 【修改点 4】返回数据中增加 auditStatus
			"createdAt":     h.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{"total": total, "list": list}})
}

// GetHospitalDetail 获取医院详情
func GetHospitalDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的医院 ID"})
		return
	}

	// 【修改点】SELECT 中增加 province_name, city_name, district_name
	query := `
		SELECT id, name, province_code, city_code, district_code, 
               province_name, city_name, district_name,
               level, is_rare_network, treat_scope, address, phone, hospital_url, audit_status, reject_reason, created_at
		FROM hospital
		WHERE id = ?
	`
	// 【修改点】结构体增加新字段
	var h struct {
		ID            uint64         `db:"id"`
		Name          string         `db:"name"`
		ProvinceCode  string         `db:"province_code"`
		CityCode      string         `db:"city_code"`
		DistrictCode  string         `db:"district_code"`
		ProvinceName  string         `db:"province_name"` // 新增
		CityName      string         `db:"city_name"`     // 新增
		DistrictName  string         `db:"district_name"` // 新增
		Level         string         `db:"level"`
		IsRareNetwork int8           `db:"is_rare_network"`
		TreatScope    sql.NullString `db:"treat_scope"`
		Address       string         `db:"address"`
		Phone         string         `db:"phone"`
		HospitalURL   sql.NullString `db:"hospital_url"`
		AuditStatus   int8           `db:"audit_status"`
		RejectReason  sql.NullString `db:"reject_reason"`
		CreatedAt     time.Time      `db:"created_at"`
	}

	// 【修改点】Scan 增加新字段
	err = db.MySQL.QueryRow(query, id).Scan(&h.ID, &h.Name, &h.ProvinceCode, &h.CityCode, &h.DistrictCode, &h.ProvinceName, &h.CityName, &h.DistrictName, &h.Level, &h.IsRareNetwork, &h.TreatScope, &h.Address, &h.Phone, &h.HospitalURL, &h.AuditStatus, &h.RejectReason, &h.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "医院不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		}
		return
	}

	// diseaseIDs, _ := getDiseaseIDsByHospital(h.ID)
	diseaseIDs, _ := getDiseasesByHospital(h.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":            h.ID,
			"name":          h.Name,
			"provinceCode":  h.ProvinceCode,
			"cityCode":      h.CityCode,
			"districtCode":  h.DistrictCode,
			"provinceName":  h.ProvinceName, // 新增返回
			"cityName":      h.CityName,     // 新增返回
			"districtName":  h.DistrictName, // 新增返回
			"level":         h.Level,
			"address":       h.Address,
			"phone":         h.Phone,
			"hospitalUrl":   h.HospitalURL.String,
			"treatScope":    h.TreatScope.String,
			"isRareNetwork": h.IsRareNetwork == 1,
			"auditStatus":   h.AuditStatus,
			"rejectReason":  h.RejectReason.String,
			"diseases":      diseaseIDs,
			"createdAt":     h.CreatedAt.Format(time.RFC3339),
		},
	})
}

// CreateHospital 新增医院
func CreateHospital(c *gin.Context) {
	var req struct {
		Name          string   `json:"name" binding:"required"`
		ProvinceCode  string   `json:"provinceCode" binding:"required"`
		CityCode      string   `json:"cityCode" binding:"required"`
		DistrictCode  string   `json:"districtCode"`
		ProvinceName  string   `json:"provinceName"` // 新增：前端传入或后端查表
		CityName      string   `json:"cityName"`     // 新增
		DistrictName  string   `json:"districtName"` // 新增
		Level         string   `json:"level" binding:"required"`
		IsRareNetwork int8     `json:"isRareNetwork"`
		TreatScope    string   `json:"treatScope"`
		Address       string   `json:"address" binding:"required"`
		Phone         string   `json:"phone" binding:"required"`
		HospitalURL   string   `json:"hospitalUrl"`
		DiseaseIDs    []uint64 `json:"diseaseIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 【可选逻辑】如果前端没传 Name，可以根据 Code 去 region 表查。
	// 这里为了代码完整性，假设如果 Name 为空，则尝试查找或报错。
	// 实际生产中建议前端直接传，或者后端统一查表填充。
	if req.ProvinceName == "" {
		// 示例：简单处理，如果没传名称，可能需要查表，这里暂置空或报错，视业务而定
		// 建议：调用 region 服务查询名称
	}

	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "数据库事务失败"})
		return
	}
	defer tx.Rollback()

	// 【修改点】INSERT 语句增加 province_name, city_name, district_name
	insertQuery := `
		INSERT INTO hospital (name, province_code, city_code, district_code, 
                              province_name, city_name, district_name,
                              level, is_rare_network, treat_scope, address, phone, hospital_url, audit_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, NOW(), NOW())
	`
	res, err := tx.Exec(insertQuery,
		req.Name, req.ProvinceCode, req.CityCode, req.DistrictCode,
		req.ProvinceName, req.CityName, req.DistrictName, // 新增参数
		req.Level, req.IsRareNetwork, req.TreatScope, req.Address, req.Phone, req.HospitalURL)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建医院失败: " + err.Error()})
		return
	}

	hospitalID, _ := res.LastInsertId()

	if len(req.DiseaseIDs) > 0 {
		if err := insertHospitalDiseaseRel(tx, uint64(hospitalID), req.DiseaseIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "关联疾病失败"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交事务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{"id": hospitalID}})
}

// UpdateHospital 修改医院信息
func UpdateHospital(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的医院 ID"})
		return
	}

	var req struct {
		Name          string   `json:"name"`
		ProvinceCode  *string  `json:"provinceCode"`
		CityCode      *string  `json:"cityCode"`
		DistrictCode  *string  `json:"districtCode"`
		ProvinceName  *string  `json:"provinceName"` // 新增
		CityName      *string  `json:"cityName"`     // 新增
		DistrictName  *string  `json:"districtName"` // 新增
		Level         string   `json:"level"`
		IsRareNetwork *int8    `json:"isRareNetwork"`
		TreatScope    *string  `json:"treatScope"`
		Address       string   `json:"address"`
		Phone         string   `json:"phone"`
		HospitalURL   *string  `json:"hospitalUrl"`
		AuditStatus   *int8    `json:"auditStatus"`
		RejectReason  *string  `json:"rejectReason"`
		DiseaseIDs    []uint64 `json:"diseaseIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var existID uint64
	if err := db.MySQL.QueryRow("SELECT id FROM hospital WHERE id = ?", id).Scan(&existID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "医院不存在"})
		return
	}

	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "事务失败"})
		return
	}
	defer tx.Rollback()

	fields := []string{}
	values := []interface{}{}

	if req.Name != "" {
		fields = append(fields, "name=?")
		values = append(values, req.Name)
	}
	if req.ProvinceCode != nil {
		fields = append(fields, "province_code=?")
		values = append(values, *req.ProvinceCode)
	}
	if req.CityCode != nil {
		fields = append(fields, "city_code=?")
		values = append(values, *req.CityCode)
	}
	if req.DistrictCode != nil {
		fields = append(fields, "district_code=?")
		values = append(values, *req.DistrictCode)
	}
	// 【修改点】处理新字段的更新
	if req.ProvinceName != nil {
		fields = append(fields, "province_name=?")
		values = append(values, *req.ProvinceName)
	}
	if req.CityName != nil {
		fields = append(fields, "city_name=?")
		values = append(values, *req.CityName)
	}
	if req.DistrictName != nil {
		fields = append(fields, "district_name=?")
		values = append(values, *req.DistrictName)
	}

	if req.Level != "" {
		fields = append(fields, "level=?")
		values = append(values, req.Level)
	}
	if req.IsRareNetwork != nil {
		fields = append(fields, "is_rare_network=?")
		values = append(values, *req.IsRareNetwork)
	}
	if req.TreatScope != nil {
		fields = append(fields, "treat_scope=?")
		values = append(values, *req.TreatScope)
	}
	if req.Address != "" {
		fields = append(fields, "address=?")
		values = append(values, req.Address)
	}
	if req.Phone != "" {
		fields = append(fields, "phone=?")
		values = append(values, req.Phone)
	}
	if req.HospitalURL != nil {
		fields = append(fields, "hospital_url=?")
		values = append(values, *req.HospitalURL)
	}
	if req.AuditStatus != nil {
		fields = append(fields, "audit_status=?")
		values = append(values, *req.AuditStatus)
	}
	if req.RejectReason != nil {
		fields = append(fields, "reject_reason=?")
		values = append(values, *req.RejectReason)
	}

	fields = append(fields, "updated_at=NOW()")
	values = append(values, id)

	if len(fields) > 0 {
		sqlStr := "UPDATE hospital SET " + strings.Join(fields, ", ") + " WHERE id=?"
		if _, err := tx.Exec(sqlStr, values...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
			return
		}
	}

	if req.DiseaseIDs != nil {
		if _, err := tx.Exec("DELETE FROM hospital_disease_rel WHERE hospital_id = ?", id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "清理旧关联失败"})
			return
		}
		if len(req.DiseaseIDs) > 0 {
			if err := insertHospitalDiseaseRel(tx, id, req.DiseaseIDs); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新关联疾病失败"})
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// DeleteHospital 真删除医院
func DeleteHospital(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效ID"})
		return
	}

	// 开启事务以确保数据一致性
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "事务失败"})
		return
	}
	defer tx.Rollback()

	// 1. 检查是否有关联的在职医生 (audit_status = 1)
	// 如果业务允许删除有医生的医院，需要先去删除或转移医生，否则保留此检查
	var count int64
	if err := tx.QueryRow("SELECT COUNT(*) FROM doctor WHERE hospital_id = ? AND audit_status = 1", id).Scan(&count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询医生关联失败"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("该医院下还有 %d 名在职医生，无法删除", count)})
		return
	}

	// 2. 删除关联表数据 (hospital_disease_rel)
	if _, err := tx.Exec("DELETE FROM hospital_disease_rel WHERE hospital_id = ?", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "清理疾病关联失败: " + err.Error()})
		return
	}

	// 3. 删除主表数据
	result, err := tx.Exec("DELETE FROM hospital WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除医院失败: " + err.Error()})
		return
	}

	// 检查是否真的删除了数据
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "医院不存在"})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// --- Doctor Handlers ---

// GetDoctorList 获取医生列表
func GetDoctorList(c *gin.Context) {
	// 1. 获取查询参数
	keyword := c.DefaultQuery("keyword", "")          // 用于模糊搜索姓名和擅长
	diseaseStr := c.DefaultQuery("diseaseId", "0")    // 疾病ID筛选
	title := c.DefaultQuery("title", "")              // 【新增】职称筛选
	hospitalIdStr := c.DefaultQuery("hospitalId", "") // 【新增】医院ID筛选
	level := c.DefaultQuery("level", "")              // 保留原有的医院等级筛选（如果需要）
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")
	auditStatusStr := c.DefaultQuery("auditStatus", "1") // 默认只查已审核通过的医生

	// 2. 参数解析与校验
	diseaseID, _ := strconv.ParseUint(diseaseStr, 10, 64)
	var hospitalID uint64
	if hospitalIdStr != "" {
		hospitalID, _ = strconv.ParseUint(hospitalIdStr, 10, 64)
	}

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 3. 构建 WHERE 子句和参数
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 【新增】处理审核状态筛选
	if auditStatusStr != "" {
		auditStatus, err := strconv.Atoi(auditStatusStr)
		if err == nil {
			whereClause += " AND d.audit_status = ?"
			args = append(args, auditStatus)
		}
	}

	// 模糊搜索：姓名 或 擅长领域
	if keyword != "" {
		whereClause += " AND (d.name LIKE ? OR d.good_at LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 【新增】精确筛选：职称
	if title != "" {
		whereClause += " AND d.title = ?"
		args = append(args, title)
	}

	// 【新增】精确筛选：医院ID
	if hospitalID > 0 {
		whereClause += " AND d.hospital_id = ?"
		args = append(args, hospitalID)
	}

	// 保留原有的医院等级筛选 (基于关联的 hospital 表)
	if level != "" {
		whereClause += " AND h.level = ?"
		args = append(args, level)
	}

	// 4. 构建 JOIN 子句
	// 必须 JOIN hospital 以获取医院名称、等级等信息用于返回
	joinClause := "JOIN hospital h ON d.hospital_id = h.id"

	// 如果按疾病筛选，需要 JOIN 关联表
	if diseaseID > 0 {
		joinClause += " JOIN doctor_disease_rel ddr ON d.id = ddr.doctor_id"
		whereClause += " AND ddr.disease_id = ?"
		args = append(args, diseaseID)
	}

	// 5. 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM doctor d " + joinClause + " " + whereClause
	if err := db.MySQL.QueryRow(countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询总数失败"})
		return
	}

	// 6. 查询列表数据
	// 注意：这里查询了 19 个字段，最后一个是 h.audit_status
	listQuery := `
		SELECT d.id, d.name, d.title, d.department, d.good_at, d.clinic_time, d.contact, d.score, d.comment_num, 
		       h.name as hospital_name, h.province_code, h.city_code, h.district_code, 
               h.province_name, h.city_name, h.district_name, h.level, h.is_rare_network, h.audit_status
		FROM doctor d
		` + joinClause + `
		` + whereClause + `
		ORDER BY d.score DESC, d.comment_num DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var doc struct {
			ID            uint64         `db:"id"`
			Name          string         `db:"name"`
			Title         string         `db:"title"`
			Department    string         `db:"department"`
			GoodAt        sql.NullString `db:"good_at"`
			ClinicTime    sql.NullString `db:"clinic_time"`
			Contact       sql.NullString `db:"contact"`
			Score         float64        `db:"score"`
			CommentNum    int            `db:"comment_num"`
			HospitalName  string         `db:"hospital_name"`
			ProvinceCode  string         `db:"province_code"`
			CityCode      string         `db:"city_code"`
			DistrictCode  string         `db:"district_code"`
			ProvinceName  string         `db:"province_name"`
			CityName      string         `db:"city_name"`
			DistrictName  string         `db:"district_name"`
			Level         string         `db:"level"`
			IsRareNetwork int8           `db:"is_rare_network"`
			AuditStatus   int8           `db:"audit_status"` // 对应 SQL 中的最后一个字段
		}

		// 【关键修复】Scan 参数必须与 SELECT 列一一对应 (共19个参数)
		if err := rows.Scan(
			&doc.ID, &doc.Name, &doc.Title, &doc.Department, &doc.GoodAt, &doc.ClinicTime, &doc.Contact, &doc.Score, &doc.CommentNum,
			&doc.HospitalName, &doc.ProvinceCode, &doc.CityCode, &doc.DistrictCode, &doc.ProvinceName, &doc.CityName, &doc.DistrictName,
			&doc.Level, &doc.IsRareNetwork, &doc.AuditStatus); err != nil {
			// 建议打印错误以便调试
			fmt.Println("Scan error:", err)
			continue
		}

		// 查询该医生关联的疾病ID
		diseaseIDs, _ := getDiseaseIDsByDoctor(doc.ID)

		list = append(list, map[string]interface{}{
			"id":            doc.ID,
			"name":          doc.Name,
			"title":         doc.Title,
			"department":    doc.Department,
			"goodAt":        doc.GoodAt.String,
			"clinicTime":    doc.ClinicTime.String,
			"contact":       doc.Contact.String,
			"rating":        doc.Score,
			"reviewCount":   doc.CommentNum,
			"hospital":      doc.HospitalName,
			"hospitalId":    doc.ID, // 注意：这里返回的是医生ID，如果需要医院ID需在SQL中增加 d.hospital_id
			"provinceCode":  doc.ProvinceCode,
			"cityCode":      doc.CityCode,
			"districtCode":  doc.DistrictCode,
			"provinceName":  doc.ProvinceName,
			"cityName":      doc.CityName,
			"districtName":  doc.DistrictName,
			"level":         doc.Level,
			"isRareNetwork": doc.IsRareNetwork == 1,
			"auditStatus":   doc.AuditStatus, // 返回具体状态值 (0, 1, 2)
			"diseaseIds":    diseaseIDs,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{"total": total, "list": list}})
}

// GetDoctorDetail 获取医生详情
func GetDoctorDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效ID"})
		return
	}

	// 【修改点 1】SELECT 中增加 d.hospital_id 和 h.name as hospital_name
	query := `
		SELECT d.id, d.name, d.title, d.department, d.good_at, d.clinic_time, d.contact, d.score, d.comment_num, d.audit_status, d.reject_reason, d.hospital_id,
		       h.name as hospital_name, h.province_code, h.city_code, h.district_code, 
               h.province_name, h.city_name, h.district_name, h.address, h.phone, h.level, h.is_rare_network
		FROM doctor d
		JOIN hospital h ON d.hospital_id = h.id
		WHERE d.id = ?
	`

	// 【修改点 2】结构体增加 HospitalID 字段
	var doc struct {
		ID            uint64         `db:"id"`
		Name          string         `db:"name"`
		Title         string         `db:"title"`
		Department    string         `db:"department"`
		GoodAt        sql.NullString `db:"good_at"`
		ClinicTime    sql.NullString `db:"clinic_time"`
		Contact       sql.NullString `db:"contact"`
		Score         float64        `db:"score"`
		CommentNum    int            `db:"comment_num"`
		AuditStatus   int8           `db:"audit_status"`
		RejectReason  sql.NullString `db:"reject_reason"`
		HospitalID    uint64         `db:"hospital_id"` // 新增：用于前端 el-select 绑定
		HospitalName  string         `db:"hospital_name"`
		ProvinceCode  string         `db:"province_code"`
		CityCode      string         `db:"city_code"`
		DistrictCode  string         `db:"district_code"`
		ProvinceName  string         `db:"province_name"`
		CityName      string         `db:"city_name"`
		DistrictName  string         `db:"district_name"`
		Address       string         `db:"address"`
		Phone         string         `db:"phone"`
		Level         string         `db:"level"`
		IsRareNetwork int8           `db:"is_rare_network"`
	}

	// 【修改点 3】Scan 增加 &doc.HospitalID (注意顺序要与 SELECT 一致)
	// SELECT 顺序: ..., d.reject_reason, d.hospital_id, h.name ...
	// Scan 顺序:   ..., &doc.RejectReason, &doc.HospitalID, &doc.HospitalName ...
	if err := db.MySQL.QueryRow(query, id).Scan(
		&doc.ID, &doc.Name, &doc.Title, &doc.Department, &doc.GoodAt, &doc.ClinicTime, &doc.Contact, &doc.Score, &doc.CommentNum,
		&doc.AuditStatus, &doc.RejectReason, &doc.HospitalID,
		&doc.HospitalName, &doc.ProvinceCode, &doc.CityCode, &doc.DistrictCode, &doc.ProvinceName, &doc.CityName, &doc.DistrictName,
		&doc.Address, &doc.Phone, &doc.Level, &doc.IsRareNetwork); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "医生不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		}
		return
	}

	// 【修改点 4】获取疾病详情列表 (对象数组)
	diseaseDetails, _ := getDiseaseDetailsByDoctor(doc.ID)

	// 【修改点 5】获取疾病 ID 列表 (整数数组，用于提交)
	diseaseIDs, _ := getDiseaseIDsByDoctor(doc.ID)

	// 【修改点 6】构建返回数据
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":           doc.ID,
			"name":         doc.Name,
			"title":        doc.Title,
			"department":   doc.Department,
			"goodAt":       doc.GoodAt.String,
			"clinicTime":   doc.ClinicTime.String,
			"contact":      doc.Contact.String,
			"score":        doc.Score,
			"commentNum":   doc.CommentNum,
			"auditStatus":  doc.AuditStatus,
			"rejectReason": doc.RejectReason.String,

			// 新增字段
			"hospitalId":   doc.HospitalID,   // 用于 el-select 绑定
			"hospitalName": doc.HospitalName, // 用于展示

			// 疾病相关
			"diseases":   diseaseDetails, // [{id, name, alias}, ...]
			"diseaseIds": diseaseIDs,     // [1, 2, ...]

			// 医院其他信息
			"isRareNetwork": doc.IsRareNetwork == 1,
			"address":       doc.Address,
			"cityCode":      doc.CityCode,
			"cityName":      doc.CityName,
			"districtCode":  doc.DistrictCode,
			"districtName":  doc.DistrictName,
			"level":         doc.Level,
			"phone":         doc.Phone,
			"provinceCode":  doc.ProvinceCode,
			"provinceName":  doc.ProvinceName,
		},
	})
}

// CreateDoctor 新增医生
func CreateDoctor(c *gin.Context) {
	// 【修改点】结构体标签改为驼峰，并增加前端可能传入的其他字段
	var req struct {
		Name       string   `json:"name" binding:"required"`
		Title      string   `json:"title" binding:"required"`
		Department string   `json:"department" binding:"required"`
		GoodAt     string   `json:"goodAt"`     // 对应 good_at
		ClinicTime string   `json:"clinicTime"` // 对应 clinic_time
		Contact    string   `json:"contact"`
		HospitalID uint64   `json:"hospitalId" binding:"required"` // 对应 hospital_id
		DiseaseIDs []uint64 `json:"diseaseIds"`                    // 对应 disease_ids

		// 如果前端希望控制初始状态，可以放开以下字段，否则保持硬编码
		AuditStatus *int8    `json:"auditStatus"`
		Score       *float64 `json:"score"`
		CommentNum  *int     `json:"commentNum"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 检查医院是否存在且已审核
	var hospID uint64
	// 注意：这里检查 audit_status = 1，如果前端传入的是未审核医院ID，会报错。
	// 如果业务允许关联未审核医院，请移除 AND audit_status = 1
	if err := db.MySQL.QueryRow("SELECT id FROM hospital WHERE id = ? AND audit_status = 1", req.HospitalID).Scan(&hospID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "所属医院不存在或未通过审核"})
		return
	}

	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "事务失败"})
		return
	}
	defer tx.Rollback()

	// 默认初始值
	initScore := 0.0
	initCommentNum := 0
	initAuditStatus := 0 // 默认待审核

	// 如果放开了上面的可选字段，可以在这里赋值
	if req.Score != nil {
		initScore = *req.Score
	}
	if req.CommentNum != nil {
		initCommentNum = *req.CommentNum
	}
	if req.AuditStatus != nil {
		initAuditStatus = int(*req.AuditStatus)
	}

	insertQuery := `
		INSERT INTO doctor (hospital_id, name, title, department, good_at, clinic_time, contact, score, comment_num, audit_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	// 【修改点】使用变量而非硬编码的 0
	res, err := tx.Exec(insertQuery,
		req.HospitalID,
		req.Name,
		req.Title,
		req.Department,
		req.GoodAt,
		req.ClinicTime,
		req.Contact,
		initScore,
		initCommentNum,
		initAuditStatus,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建医生失败: " + err.Error()})
		return
	}

	doctorID, _ := res.LastInsertId()

	if len(req.DiseaseIDs) > 0 {
		if err := insertDoctorDiseaseRel(tx, uint64(doctorID), req.DiseaseIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "关联疾病失败: " + err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{"id": doctorID}})
}

// UpdateDoctor 修改医生
func UpdateDoctor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效ID"})
		return
	}

	// 【修改点 1】结构体标签改为驼峰，以匹配前端入参
	var req struct {
		Name         string   `json:"name"`
		Title        string   `json:"title"`
		Department   string   `json:"department"`
		GoodAt       *string  `json:"goodAt"`     // 对应 good_at
		ClinicTime   *string  `json:"clinicTime"` // 对应 clinic_time
		Contact      *string  `json:"contact"`
		HospitalID   *uint64  `json:"hospitalId"`   // 对应 hospital_id
		AuditStatus  *int8    `json:"auditStatus"`  // 对应 audit_status
		RejectReason *string  `json:"rejectReason"` // 对应 reject_reason
		DiseaseIDs   []uint64 `json:"diseaseIds"`   // 对应 disease_ids

		// 如果业务允许前端修改评分和评论数，可以解开以下注释
		// Score      *float64 `json:"score"`
		// CommentNum *int     `json:"commentNum"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "事务失败"})
		return
	}
	defer tx.Rollback()

	fields := []string{}
	values := []interface{}{}

	// 构建动态更新字段
	if req.Name != "" {
		fields = append(fields, "name=?")
		values = append(values, req.Name)
	}
	if req.Title != "" {
		fields = append(fields, "title=?")
		values = append(values, req.Title)
	}
	if req.Department != "" {
		fields = append(fields, "department=?")
		values = append(values, req.Department)
	}

	// 处理指针类型字段（允许设置为空或特定值）
	if req.GoodAt != nil {
		fields = append(fields, "good_at=?")
		values = append(values, *req.GoodAt)
	}
	if req.ClinicTime != nil {
		fields = append(fields, "clinic_time=?")
		values = append(values, *req.ClinicTime)
	}
	if req.Contact != nil {
		fields = append(fields, "contact=?")
		values = append(values, *req.Contact)
	}

	// 处理医院变更
	if req.HospitalID != nil {
		// 检查新医院是否存在且已审核
		var hID uint64
		if err := db.MySQL.QueryRow("SELECT id FROM hospital WHERE id = ? AND audit_status = 1", *req.HospitalID).Scan(&hID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "新医院无效或未通过审核"})
			return
		}
		fields = append(fields, "hospital_id=?")
		values = append(values, *req.HospitalID)
	}

	// 处理审核状态
	if req.AuditStatus != nil {
		fields = append(fields, "audit_status=?")
		values = append(values, *req.AuditStatus)
	}

	// 处理驳回原因
	if req.RejectReason != nil {
		fields = append(fields, "reject_reason=?")
		values = append(values, *req.RejectReason)
	}

	// 如果放开了 Score 和 CommentNum 的修改权限，在此处添加：
	// if req.Score != nil {
	// 	fields = append(fields, "score=?")
	// 	values = append(values, *req.Score)
	// }
	// if req.CommentNum != nil {
	// 	fields = append(fields, "comment_num=?")
	// 	values = append(values, *req.CommentNum)
	// }

	// 始终更新时间
	fields = append(fields, "updated_at=NOW()")
	values = append(values, id)

	// 执行更新
	if len(fields) > 0 {
		sqlStr := "UPDATE doctor SET " + strings.Join(fields, ", ") + " WHERE id=?"
		if _, err := tx.Exec(sqlStr, values...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
			return
		}
	}

	// 更新疾病关联
	// 注意：只有当 frontend 显式传递了 diseaseIds 字段时（即使是空数组），才进行更新操作
	// 在 Go Gin 中，如果 JSON 中缺少该字段，slice 通常为 nil；如果为 []，则不为 nil
	if req.DiseaseIDs != nil {
		// 1. 删除旧关联
		if _, err := tx.Exec("DELETE FROM doctor_disease_rel WHERE doctor_id = ?", id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "清理旧关联失败"})
			return
		}

		// 2. 插入新关联
		if len(req.DiseaseIDs) > 0 {
			if err := insertDoctorDiseaseRel(tx, id, req.DiseaseIDs); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新关联疾病失败: " + err.Error()})
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// DeleteDoctor 真删除医生
func DeleteDoctor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效ID"})
		return
	}

	// 开启事务以确保数据一致性
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "事务失败"})
		return
	}
	defer tx.Rollback()

	// 1. 先删除关联表数据 (doctor_disease_rel)
	// 即使没有关联数据，执行 DELETE 也不会报错，只是影响行数为 0
	if _, err := tx.Exec("DELETE FROM doctor_disease_rel WHERE doctor_id = ?", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "清理疾病关联失败: " + err.Error()})
		return
	}

	// 2. 删除医生主表数据
	result, err := tx.Exec("DELETE FROM doctor WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除医生失败: " + err.Error()})
		return
	}

	// 检查是否真的删除了数据
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "医生不存在"})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// --- Examination Manual Handlers ---

// GetExaminationList 获取检查手册列表
func GetExaminationList(c *gin.Context) {
	// 【新增】获取筛选参数
	keyword := c.DefaultQuery("keyword", "")
	examType := c.DefaultQuery("examType", "")
	auditStatusStr := c.DefaultQuery("auditStatus", "1") // 默认只查已审核通过的

	diseaseStr := c.DefaultQuery("diseaseId", "0")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	diseaseID, _ := strconv.ParseUint(diseaseStr, 10, 64)
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 【新增】初始化 WHERE 子句和参数
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 【新增】处理审核状态筛选
	if auditStatusStr != "" {
		auditStatus, err := strconv.Atoi(auditStatusStr)
		if err == nil {
			whereClause += " AND em.audit_status = ?"
			args = append(args, auditStatus)
		}
	}

	// 【新增】处理关键词筛选 (模糊匹配 exam_name 或 exam_purpose)
	if keyword != "" {
		whereClause += " AND (em.exam_name LIKE ? OR em.exam_purpose LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 【新增】处理检查类型筛选
	if examType != "" {
		whereClause += " AND em.exam_type = ?"
		args = append(args, examType)
	}

	// 处理疾病关联筛选
	joinClause := ""
	if diseaseID > 0 {
		joinClause = "JOIN exam_manual_disease_rel emdr ON em.id = emdr.exam_manual_id"
		whereClause += " AND emdr.disease_id = ?"
		args = append(args, diseaseID)
	}

	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM examination_manual em " + joinClause + " " + whereClause
	if err := db.MySQL.QueryRow(countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询总数失败"})
		return
	}

	// 查询列表数据
	listQuery := `
		SELECT em.id, em.exam_name, em.exam_type, em.exam_purpose, em.sample_notes, em.institution, em.sort, em.created_at, em.updated_at
		FROM examination_manual em
		` + joinClause + `
		` + whereClause + `
		ORDER BY em.sort ASC, em.created_at DESC
		LIMIT ? OFFSET ?
	`
	// 注意：分页参数最后追加
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var item struct {
			ID          uint64    `db:"id"`
			ExamName    string    `db:"exam_name"`
			ExamType    string    `db:"exam_type"`
			ExamPurpose string    `db:"exam_purpose"`
			SampleNotes string    `db:"sample_notes"`
			Institution string    `db:"institution"`
			Sort        int       `db:"sort"`
			CreatedAt   time.Time `db:"created_at"`
			UpdatedAt   time.Time `db:"updated_at"`
		}

		if err := rows.Scan(&item.ID, &item.ExamName, &item.ExamType, &item.ExamPurpose, &item.SampleNotes, &item.Institution, &item.Sort, &item.CreatedAt, &item.UpdatedAt); err != nil {
			continue
		}

		price, duration := mapPriceAndDuration(item.ExamType)
		diseaseIDs, _ := getDiseaseIDsByExamManual(item.ID)

		list = append(list, map[string]interface{}{
			"id":          item.ID,
			"examName":    item.ExamName,
			"examType":    item.ExamType,
			"examPurpose": item.ExamPurpose,
			"sampleNotes": item.SampleNotes,
			"institution": item.Institution,
			"sort":        item.Sort,
			"price":       price,
			"duration":    duration,
			"diseaseIds":  diseaseIDs,
			"createdAt":   item.CreatedAt.Format(time.RFC3339),
			"updatedAt":   item.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{"total": total, "list": list}})
}

// GetExaminationDetail 获取检查手册详情
func GetExaminationDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效ID"})
		return
	}

	query := `
		SELECT id, exam_name, exam_type, exam_purpose, reference_value, abnormal_interpret, sample_notes, institution,
		       template_excel, template_word, compare_template, audit_status, reject_reason, sort, created_at
		FROM examination_manual
		WHERE id = ?
	`
	var item struct {
		ID                uint64         `db:"id"`
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
		AuditStatus       int8           `db:"audit_status"`
		RejectReason      sql.NullString `db:"reject_reason"`
		Sort              int            `db:"sort"`
		CreatedAt         time.Time      `db:"created_at"`
	}

	if err := db.MySQL.QueryRow(query, id).Scan(&item.ID, &item.ExamName, &item.ExamType, &item.ExamPurpose, &item.ReferenceValue, &item.AbnormalInterpret, &item.SampleNotes, &item.Institution, &item.TemplateExcel, &item.TemplateWord, &item.CompareTemplate, &item.AuditStatus, &item.RejectReason, &item.Sort, &item.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		}
		return
	}

	diseaseIDs, _ := getDiseaseIDsByExamManual(item.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":                item.ID,
			"examName":          item.ExamName,
			"examType":          item.ExamType,
			"examPurpose":       item.ExamPurpose,
			"referenceValue":    item.ReferenceValue.String,
			"abnormalInterpret": item.AbnormalInterpret.String,
			"sampleNotes":       item.SampleNotes.String,
			"institution":       item.Institution.String,
			"templates": gin.H{
				"excel":   item.TemplateExcel.String,
				"word":    item.TemplateWord.String,
				"compare": item.CompareTemplate.String,
			},
			"auditStatus":  item.AuditStatus,
			"rejectReason": item.RejectReason.String,
			"sort":         item.Sort,
			"diseaseIds":   diseaseIDs,
			"createdAt":    item.CreatedAt.Format(time.RFC3339),
		},
	})
}

// CreateExamination 新增检查手册
func CreateExamination(c *gin.Context) {
	// 【修改点 1】定义接收 templates 嵌套结构的子结构体
	type TemplatesReq struct {
		Excel   string `json:"excel"`
		Word    string `json:"word"`
		Compare string `json:"compare"`
	}

	// 【修改点 2】修改请求结构体，使用驼峰命名匹配前端入参，并增加 Templates 字段
	var req struct {
		ExamName          string        `json:"examName" binding:"required"`
		ExamType          string        `json:"examType" binding:"required"`
		ExamPurpose       string        `json:"examPurpose" binding:"required"`
		ReferenceValue    string        `json:"referenceValue"`
		AbnormalInterpret string        `json:"abnormalInterpret"`
		SampleNotes       string        `json:"sampleNotes"`
		Institution       string        `json:"institution"`
		Templates         *TemplatesReq `json:"templates"` // 新增：接收嵌套对象
		Sort              int           `json:"sort"`
		DiseaseIDs        []uint64      `json:"diseaseIds"`
		// 注意：通常创建时 audit_status 默认为 0 (待审核)，如果前端强传可以保留，否则建议后端硬编码或忽略
		AuditStatus  *int8  `json:"auditStatus"`
		RejectReason string `json:"rejectReason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "事务失败"})
		return
	}
	defer tx.Rollback()

	// 【修改点 3】处理 templates，如果前端未传则默认为空字符串
	templateExcel := ""
	templateWord := ""
	compareTemplate := ""

	if req.Templates != nil {
		templateExcel = req.Templates.Excel
		templateWord = req.Templates.Word
		compareTemplate = req.Templates.Compare
	}

	// 处理 audit_status，如果前端没传或传了，这里可以做逻辑判断。
	// 通常新建默认为 0 (待审核)。如果前端传了且希望生效：
	initAuditStatus := int8(0)
	if req.AuditStatus != nil {
		initAuditStatus = *req.AuditStatus
	}

	insertQuery := `
		INSERT INTO examination_manual (
			exam_type, exam_name, exam_purpose, reference_value, abnormal_interpret, 
			sample_notes, institution, template_excel, template_word, compare_template, 
			audit_status, sort, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	res, err := tx.Exec(insertQuery,
		req.ExamType,
		req.ExamName,
		req.ExamPurpose,
		req.ReferenceValue,
		req.AbnormalInterpret,
		req.SampleNotes,
		req.Institution,
		templateExcel,   // 映射拆解后的字段
		templateWord,    // 映射拆解后的字段
		compareTemplate, // 映射拆解后的字段
		initAuditStatus,
		req.Sort,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败: " + err.Error()})
		return
	}

	manualID, _ := res.LastInsertId()

	if len(req.DiseaseIDs) > 0 {
		if err := insertExamManualDiseaseRel(tx, uint64(manualID), req.DiseaseIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "关联疾病失败: " + err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{"id": manualID}})
}

// UpdateExamination 修改检查手册
func UpdateExamination(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效ID"})
		return
	}

	// 【修改点 1】定义接收 templates 嵌套结构的子结构体
	type TemplatesReq struct {
		Excel   string `json:"excel"`
		Word    string `json:"word"`
		Compare string `json:"compare"`
	}

	// 【修改点 2】修改请求结构体，增加 Templates 字段
	var req struct {
		ExamName          *string       `json:"examName"`
		ExamType          *string       `json:"examType"`
		ExamPurpose       *string       `json:"examPurpose"`
		ReferenceValue    *string       `json:"referenceValue"`
		AbnormalInterpret *string       `json:"abnormalInterpret"`
		SampleNotes       *string       `json:"sampleNotes"`
		Institution       *string       `json:"institution"`
		Templates         *TemplatesReq `json:"templates"` // 新增：接收嵌套对象
		Sort              *int          `json:"sort"`
		AuditStatus       *int8         `json:"auditStatus"`
		RejectReason      *string       `json:"rejectReason"`
		DiseaseIDs        []uint64      `json:"diseaseIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "事务失败"})
		return
	}
	defer tx.Rollback()

	fields := []string{}
	values := []interface{}{}

	// 基础字段更新逻辑
	if req.ExamName != nil {
		fields = append(fields, "exam_name=?")
		values = append(values, *req.ExamName)
	}
	if req.ExamType != nil {
		fields = append(fields, "exam_type=?")
		values = append(values, *req.ExamType)
	}
	if req.ExamPurpose != nil {
		fields = append(fields, "exam_purpose=?")
		values = append(values, *req.ExamPurpose)
	}
	if req.ReferenceValue != nil {
		fields = append(fields, "reference_value=?")
		values = append(values, *req.ReferenceValue)
	}
	if req.AbnormalInterpret != nil {
		fields = append(fields, "abnormal_interpret=?")
		values = append(values, *req.AbnormalInterpret)
	}
	if req.SampleNotes != nil {
		fields = append(fields, "sample_notes=?")
		values = append(values, *req.SampleNotes)
	}
	if req.Institution != nil {
		fields = append(fields, "institution=?")
		values = append(values, *req.Institution)
	}

	// 【修改点 3】处理 templates 嵌套对象，拆解为数据库字段
	if req.Templates != nil {
		// 如果前端传了 templates 对象，即使内部字段为空字符串，也视为需要更新（覆盖为空）
		// 如果业务逻辑是“不传则不更新”，则需要判断 *req.Templates.Excel != "" 等
		// 这里假设只要传了 templates 对象，就更新对应的列

		fields = append(fields, "template_excel=?")
		values = append(values, req.Templates.Excel)

		fields = append(fields, "template_word=?")
		values = append(values, req.Templates.Word)

		fields = append(fields, "compare_template=?")
		values = append(values, req.Templates.Compare)
	}

	if req.Sort != nil {
		fields = append(fields, "sort=?")
		values = append(values, *req.Sort)
	}
	if req.AuditStatus != nil {
		fields = append(fields, "audit_status=?")
		values = append(values, *req.AuditStatus)
	}
	if req.RejectReason != nil {
		fields = append(fields, "reject_reason=?")
		values = append(values, *req.RejectReason)
	}

	// 始终更新时间
	fields = append(fields, "updated_at=NOW()")
	values = append(values, id)

	// 执行主表更新
	if len(fields) > 0 {
		sqlStr := "UPDATE examination_manual SET " + strings.Join(fields, ", ") + " WHERE id=?"
		if _, err := tx.Exec(sqlStr, values...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
			return
		}
	}

	// 处理疾病关联更新
	if req.DiseaseIDs != nil {
		// 1. 删除旧关联
		if _, err := tx.Exec("DELETE FROM exam_manual_disease_rel WHERE exam_manual_id = ?", id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "清理关联失败"})
			return
		}
		// 2. 插入新关联
		if len(req.DiseaseIDs) > 0 {
			if err := insertExamManualDiseaseRel(tx, id, req.DiseaseIDs); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新关联失败"})
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// DeleteExamination 真删除检查手册
func DeleteExamination(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效ID"})
		return
	}

	// 开启事务以确保数据一致性
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "事务失败"})
		return
	}
	defer tx.Rollback()

	// 1. 先删除关联表数据 (exam_manual_disease_rel)
	// 即使没有关联数据，执行 DELETE 也不会报错，只是影响行数为 0
	if _, err := tx.Exec("DELETE FROM exam_manual_disease_rel WHERE exam_manual_id = ?", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "清理疾病关联失败: " + err.Error()})
		return
	}

	// 2. 删除主表数据
	result, err := tx.Exec("DELETE FROM examination_manual WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除检查手册失败: " + err.Error()})
		return
	}

	// 检查是否真的删除了数据
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "检查手册不存在"})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// --- Helper Functions for Relations ---

// DiseaseSimpleInfo 用于列表展示的疾病简要信息
type DiseaseSimpleInfo struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias,omitempty"` // 别名，如果没有可为空
}

// getDiseasesByHospital 获取医院关联的疾病简要信息（ID, 名称, 别名）
func getDiseasesByHospital(hospitalID uint64) ([]DiseaseSimpleInfo, error) {
	// 假设疾病表名为 disease，关联表为 hospital_disease_rel
	// 请根据你实际的数据库表结构调整 SQL
	query := `
		SELECT d.id, d.name, d.alias 
		FROM hospital_disease_rel hdr
		JOIN disease d ON hdr.disease_id = d.id
		WHERE hdr.hospital_id = ?
	`

	rows, err := db.MySQL.Query(query, hospitalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diseases []DiseaseSimpleInfo
	for rows.Next() {
		var d DiseaseSimpleInfo
		// 注意：如果 alias 字段可能为 NULL，需要使用 sql.NullString 或者确保数据库默认值为空字符串
		// 这里假设 alias 允许为空，使用指针或 NullString 处理更安全
		var alias sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &alias); err != nil {
			continue
		}
		if alias.Valid {
			d.Alias = alias.String
		}
		diseases = append(diseases, d)
	}
	return diseases, nil
}

func getDiseaseIDsByDoctor(doctorID uint64) ([]uint64, error) {
	rows, err := db.MySQL.Query("SELECT disease_id FROM doctor_disease_rel WHERE doctor_id = ?", doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func getDiseaseIDsByExamManual(manualID uint64) ([]uint64, error) {
	rows, err := db.MySQL.Query("SELECT disease_id FROM exam_manual_disease_rel WHERE exam_manual_id = ?", manualID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func insertHospitalDiseaseRel(tx *sql.Tx, hospitalID uint64, diseaseIDs []uint64) error {
	stmt, err := tx.Prepare("INSERT INTO hospital_disease_rel (hospital_id, disease_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, did := range diseaseIDs {
		if _, err := stmt.Exec(hospitalID, did); err != nil {
			return err
		}
	}
	return nil
}

func insertDoctorDiseaseRel(tx *sql.Tx, doctorID uint64, diseaseIDs []uint64) error {
	stmt, err := tx.Prepare("INSERT INTO doctor_disease_rel (doctor_id, disease_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, did := range diseaseIDs {
		if _, err := stmt.Exec(doctorID, did); err != nil {
			return err
		}
	}
	return nil
}

func insertExamManualDiseaseRel(tx *sql.Tx, manualID uint64, diseaseIDs []uint64) error {
	stmt, err := tx.Prepare("INSERT INTO exam_manual_disease_rel (exam_manual_id, disease_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, did := range diseaseIDs {
		if _, err := stmt.Exec(manualID, did); err != nil {
			return err
		}
	}
	return nil
}

// HospitalOptionItem 医院选项响应结构
type HospitalOptionItem struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// GetHospitalOptions 获取医院下拉选项列表
func GetHospitalOptions(c *gin.Context) {
	// 1. 获取查询参数
	keyword := c.DefaultQuery("keyword", "")

	// 2. 构建查询条件
	// 默认只查询已审核通过 (audit_status = 1) 的医院，确保下拉框中的数据是有效的
	whereClause := "WHERE audit_status = 1"
	args := []interface{}{}

	// 如果有关键词，增加模糊搜索
	if keyword != "" {
		whereClause += " AND name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	// 3. 执行查询
	// 只查询 id 和 name，并按名称排序以便前端展示
	query := `
		SELECT id, name 
		FROM hospital 
		` + whereClause + `
		ORDER BY name ASC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询医院选项失败",
		})
		return
	}
	defer rows.Close()

	// 4. 扫描结果
	var options []HospitalOptionItem
	for rows.Next() {
		var item HospitalOptionItem
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			continue
		}
		options = append(options, item)
	}

	// 确保返回空数组而不是 null
	if options == nil {
		options = []HospitalOptionItem{}
	}

	// 5. 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    options,
	})
}

// getDiseaseDetailsByDoctor 获取医生关联的疾病详细信息（ID, Name, Alias）
func getDiseaseDetailsByDoctor(doctorID uint64) ([]DiseaseSimpleInfo, error) {
	query := `
		SELECT d.id, d.name, d.alias 
		FROM doctor_disease_rel ddr
		JOIN disease d ON ddr.disease_id = d.id
		WHERE ddr.doctor_id = ?
	`
	rows, err := db.MySQL.Query(query, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diseases []DiseaseSimpleInfo
	for rows.Next() {
		var d DiseaseSimpleInfo
		var alias sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &alias); err != nil {
			continue
		}
		if alias.Valid {
			d.Alias = alias.String
		}
		diseases = append(diseases, d)
	}
	return diseases, nil
}
