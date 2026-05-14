package drug

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"rare_backend/internal/pkg/db"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// convertDrugType 转换药品类型枚举
func convertDrugType(drugType string) string {
	typeMap := map[string]string{
		"进口":  "imported",
		"国产":  "domestic",
		"仿制药": "generic",
	}
	if val, ok := typeMap[drugType]; ok {
		return val
	}
	return drugType
}

// mapYesNo 布尔值转中文
func mapYesNo(val int8) string {
	if val == 1 {
		return "是"
	}
	return "否"
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// boolToInt 布尔转整数
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// DrugItem 药品项响应结构
type DrugItem struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	Indication       string `json:"indication"`
	Type             string `json:"type"`
	Insurance        bool   `json:"insurance"`
	Desc             string `json:"desc"`
	ManualURL        string `json:"manualUrl"`
	DosageForm       string `json:"dosageForm"`
	Spec             string `json:"spec"`
	RefPrice         string `json:"refPrice"`
	HasRelief        bool   `json:"hasRelief"`
	IsLaunched       bool   `json:"isLaunched"`
	NeedPrescription bool   `json:"needPrescription"`
}

// DrugListResponse 列表响应结构
type DrugListResponse struct {
	List     []DrugItem `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

// DrugOptionsResponse 筛选选项响应
type DrugOptionsResponse struct {
	Types      []OptionItem `json:"types"`
	Insurances []OptionItem `json:"insurances"`
}

// OptionItem 选项项
type OptionItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// CreateDrugRequest 新增药品请求
type CreateDrugRequest struct {
	GenericName      string   `json:"genericName"`      // 通用名
	BrandName        string   `json:"brandName"`        // 商品名
	Indication       string   `json:"indication"`       // 适应症
	DrugType         string   `json:"drugType"`         // 药品类型
	IsInsurance      bool     `json:"isInsurance"`      // 是否医保
	DosageForm       string   `json:"dosageForm"`       // 剂型
	Spec             string   `json:"spec"`             // 规格
	RefPrice         float64  `json:"refPrice"`         // 参考价格 (建议前端传字符串，或者后端改为float64)
	HasRelief        bool     `json:"hasRelief"`        // 是否有赠药援助
	IsLaunched       bool     `json:"isLaunched"`       // 是否国内上市
	NeedPrescription bool     `json:"needPrescription"` // 是否需要处方
	ManualOriginal   string   `json:"manualOriginal"`   // 说明书原版链接
	ManualPopular    string   `json:"manualPopular"`    // 说明书通俗版链接
	AuditStatus      int8     `json:"auditStatus"`      // 【新增】审核状态，默认0待审核
	RejectReason     string   `json:"rejectReason"`     // 【新增】驳回原因
	DiseaseIds       []uint64 `json:"diseaseIds"`       // 【新增】疾病ID列表
}

// UpdateDrugRequest 更新药品请求
type UpdateDrugRequest struct {
	GenericName      *string  `json:"genericName"` // 【新增】如果需要允许修改通用名
	BrandName        *string  `json:"brandName"`
	Indication       *string  `json:"indication"`
	DrugType         *string  `json:"drugType"`
	IsInsurance      *bool    `json:"isInsurance"`
	DosageForm       *string  `json:"dosageForm"`
	Spec             *string  `json:"spec"`
	RefPrice         *float64 `json:"refPrice"`
	HasRelief        *bool    `json:"hasRelief"`
	IsLaunched       *bool    `json:"isLaunched"`
	NeedPrescription *bool    `json:"needPrescription"`
	ManualOriginal   *string  `json:"manualOriginal"`
	ManualPopular    *string  `json:"manualPopular"`
	AuditStatus      *int8    `json:"auditStatus"`  // 【新增】审核状态
	RejectReason     *string  `json:"rejectReason"` // 【新增】驳回原因
	DiseaseIds       []uint64 `json:"diseaseIds"`   // 【新增】疾病ID列表
}

// CreateDrug 新增药品
func CreateDrug(c *gin.Context) {
	var req CreateDrugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 必填字段验证
	if req.GenericName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "通用名不能为空",
		})
		return
	}

	// 验证药品类型非空
	if req.DrugType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "药品类型不能为空",
		})
		return
	}

	// 插入主表数据库
	// 【修改点 1】SQL 中增加 audit_status, reject_reason
	insertQuery := `
		INSERT INTO rare_drug 
		(generic_name, brand_name, indication, drug_type, is_insurance,
		 dosage_form, spec, ref_price, has_relief, is_launched, need_prescription,
		 manual_original, manual_popular, audit_status, reject_reason, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()

	// 如果前端没传 auditStatus，默认设为 0 (待审核)
	auditStatus := req.AuditStatus

	result, err := db.MySQL.Exec(insertQuery,
		req.GenericName, req.BrandName, req.Indication, req.DrugType,
		boolToInt(req.IsInsurance), req.DosageForm, req.Spec, req.RefPrice,
		boolToInt(req.HasRelief), boolToInt(req.IsLaunched), boolToInt(req.NeedPrescription),
		req.ManualOriginal, req.ManualPopular, auditStatus, req.RejectReason, now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "新增药品失败: " + err.Error(),
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

	// 【修改点 2】处理疾病关联 (DiseaseIds)
	if len(req.DiseaseIds) > 0 {
		// 构建批量插入语句
		insertPlaceholders := make([]string, 0, len(req.DiseaseIds))
		insertArgs := make([]interface{}, 0, len(req.DiseaseIds)*2)

		for _, diseaseID := range req.DiseaseIds {
			insertPlaceholders = append(insertPlaceholders, "(?, ?)")
			insertArgs = append(insertArgs, id, diseaseID)
		}

		if len(insertPlaceholders) > 0 {
			insertRelQuery := fmt.Sprintf("INSERT INTO drug_disease_rel (drug_id, disease_id) VALUES %s", strings.Join(insertPlaceholders, ","))
			_, err := db.MySQL.Exec(insertRelQuery, insertArgs...)
			if err != nil {
				// 如果关联插入失败，可以考虑回滚主表删除，或者记录日志。
				// 简单起见，这里返回错误，前端需要知道虽然药品创建了但关联失败了
				fmt.Printf("[CreateDrug] Insert disease relation error: %v\n", err)

				// 可选：为了数据一致性，删除刚才创建的主表记录
				// db.MySQL.Exec("DELETE FROM rare_drug WHERE id = ?", id)

				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "新增药品成功，但疾病关联失败: " + err.Error(),
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": id,
		},
	})
}

// UpdateDrug 更新药品
func UpdateDrug(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的药品 ID",
		})
		return
	}

	var req UpdateDrugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查药品是否存在
	checkQuery := "SELECT id FROM rare_drug WHERE id = ?"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "药品不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询药品失败",
		})
		return
	}

	// --- 1. 构建主表动态更新语句 ---
	updateFields := []string{}
	updateArgs := []interface{}{}

	if req.GenericName != nil {
		updateFields = append(updateFields, "generic_name = ?")
		updateArgs = append(updateArgs, *req.GenericName)
	}
	if req.BrandName != nil {
		updateFields = append(updateFields, "brand_name = ?")
		updateArgs = append(updateArgs, *req.BrandName)
	}
	if req.Indication != nil {
		updateFields = append(updateFields, "indication = ?")
		updateArgs = append(updateArgs, *req.Indication)
	}
	if req.DrugType != nil {
		updateFields = append(updateFields, "drug_type = ?")
		updateArgs = append(updateArgs, *req.DrugType)
	}
	if req.IsInsurance != nil {
		updateFields = append(updateFields, "is_insurance = ?")
		updateArgs = append(updateArgs, boolToInt(*req.IsInsurance))
	}
	if req.DosageForm != nil {
		updateFields = append(updateFields, "dosage_form = ?")
		updateArgs = append(updateArgs, *req.DosageForm)
	}
	if req.Spec != nil {
		updateFields = append(updateFields, "spec = ?")
		updateArgs = append(updateArgs, *req.Spec)
	}
	if req.RefPrice != nil {
		updateFields = append(updateFields, "ref_price = ?")
		// 【修改】将 float64 格式化为字符串，保留两位小数
		updateArgs = append(updateArgs, fmt.Sprintf("%.2f", *req.RefPrice))
	}
	if req.HasRelief != nil {
		updateFields = append(updateFields, "has_relief = ?")
		updateArgs = append(updateArgs, boolToInt(*req.HasRelief))
	}
	if req.IsLaunched != nil {
		updateFields = append(updateFields, "is_launched = ?")
		updateArgs = append(updateArgs, boolToInt(*req.IsLaunched))
	}
	if req.NeedPrescription != nil {
		updateFields = append(updateFields, "need_prescription = ?")
		updateArgs = append(updateArgs, boolToInt(*req.NeedPrescription))
	}
	if req.ManualOriginal != nil {
		updateFields = append(updateFields, "manual_original = ?")
		updateArgs = append(updateArgs, *req.ManualOriginal)
	}
	if req.ManualPopular != nil {
		updateFields = append(updateFields, "manual_popular = ?")
		updateArgs = append(updateArgs, *req.ManualPopular)
	}

	// 【新增】处理 AuditStatus
	if req.AuditStatus != nil {
		updateFields = append(updateFields, "audit_status = ?")
		updateArgs = append(updateArgs, *req.AuditStatus)
	}

	// 【新增】处理 RejectReason
	if req.RejectReason != nil {
		updateFields = append(updateFields, "reject_reason = ?")
		updateArgs = append(updateArgs, *req.RejectReason)
	}

	// 只有当有字段需要更新时才执行主表更新
	if len(updateFields) > 0 {
		// 添加更新时间和 ID
		updateFields = append(updateFields, "updated_at = ?")
		updateArgs = append(updateArgs, time.Now())
		updateArgs = append(updateArgs, id)

		updateQuery := "UPDATE rare_drug SET " + strings.Join(updateFields, ", ") + " WHERE id = ?"
		_, err = db.MySQL.Exec(updateQuery, updateArgs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新药品主表失败: " + err.Error(),
			})
			return
		}
	}

	// --- 2. 处理疾病关联 (DiseaseIds) ---
	if req.DiseaseIds != nil {
		// 2.1 删除旧的关联
		deleteRelQuery := "DELETE FROM drug_disease_rel WHERE drug_id = ?"
		_, err := db.MySQL.Exec(deleteRelQuery, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "清理旧疾病关联失败",
			})
			return
		}

		// 2.2 插入新的关联
		if len(req.DiseaseIds) > 0 {
			insertPlaceholders := make([]string, 0, len(req.DiseaseIds))
			insertArgs := make([]interface{}, 0, len(req.DiseaseIds)*2)

			for _, diseaseID := range req.DiseaseIds {
				insertPlaceholders = append(insertPlaceholders, "(?, ?)")
				insertArgs = append(insertArgs, id, diseaseID)
			}

			if len(insertPlaceholders) > 0 {
				insertRelQuery := fmt.Sprintf("INSERT INTO drug_disease_rel (drug_id, disease_id) VALUES %s", strings.Join(insertPlaceholders, ","))
				_, err := db.MySQL.Exec(insertRelQuery, insertArgs...)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    500,
						"message": "新增疾病关联失败: " + err.Error(),
					})
					return
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DeleteDrug 删除药品（真删除）
func DeleteDrug(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的药品 ID",
		})
		return
	}

	// 1. 检查药品是否存在
	// 【修改】移除 audit_status 的限制，只要 ID 存在即视为可删除
	checkQuery := "SELECT id FROM rare_drug WHERE id = ?"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "药品不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询药品失败",
		})
		return
	}

	// 2. 清理关联数据
	// 【新增】删除 drug_disease_rel 表中的关联记录，防止孤儿数据或外键冲突
	deleteRelQuery := "DELETE FROM drug_disease_rel WHERE drug_id = ?"
	_, err = db.MySQL.Exec(deleteRelQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "清理疾病关联失败: " + err.Error(),
		})
		return
	}

	// 3. 执行物理删除
	// 【修改】使用 DELETE 语句直接删除记录
	deleteQuery := "DELETE FROM rare_drug WHERE id = ?"
	_, err = db.MySQL.Exec(deleteQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除药品失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// GetDrugList 获取药品列表
func GetDrugList(c *gin.Context) {
	// 获取请求参数
	keyword := c.DefaultQuery("keyword", "")
	typeFilter := c.DefaultQuery("drugType", "")
	insuranceStr := c.DefaultQuery("isInsurance", "")

	// 【新增】获取 hasRelief 和 auditStatus 参数
	hasReliefStr := c.DefaultQuery("hasRelief", "")
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
	// 注意：默认只查询审核通过的药品，除非前端显式传递了 auditStatus
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 【新增】处理 auditStatus 筛选
	// 如果前端传了 auditStatus，则按传入值筛选；否则默认只查 audit_status = 1 (通过)
	if auditStatusStr != "" {
		auditStatus, err := strconv.Atoi(auditStatusStr)
		if err == nil {
			whereClause += " AND d.audit_status = ?"
			args = append(args, auditStatus)
		}
	} else {
		// 默认行为：只展示审核通过的药品
		whereClause += " AND d.audit_status = 1"
	}

	if keyword != "" {
		whereClause += " AND (d.generic_name LIKE ? OR d.brand_name LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	if typeFilter != "" {
		whereClause += " AND d.drug_type = ?"
		args = append(args, typeFilter)
	}
	if insuranceStr != "" {
		isInsurance := 0
		if insuranceStr == "true" || insuranceStr == "1" {
			isInsurance = 1
		}
		whereClause += " AND d.is_insurance = ?"
		args = append(args, isInsurance)
	}

	// 【新增】处理 hasRelief 筛选
	if hasReliefStr != "" {
		hasRelief := 0
		// 支持 "true"/"1" 或 "false"/"0"
		if hasReliefStr == "true" || hasReliefStr == "1" {
			hasRelief = 1
		}
		whereClause += " AND d.has_relief = ?"
		args = append(args, hasRelief)
	}

	// 1. 查询总数
	countQuery := "SELECT COUNT(*) FROM rare_drug d " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		fmt.Println("[GetDrugList] Count query error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败: " + err.Error(),
		})
		return
	}

	// 2. 查询药品基本列表
	listQuery := `
		SELECT 
			d.id, 
			d.generic_name, 
			d.brand_name, 
			d.indication, 
			d.drug_type, 
			d.is_insurance,
			d.dosage_form, 
			d.spec, 
			d.ref_price, 
			d.has_relief, 
			d.is_launched, 
			d.need_prescription,
			d.manual_original, 
			d.manual_popular, 
			d.audit_status,
			d.created_at, 
			d.updated_at
		FROM rare_drug d
		` + whereClause + `
		ORDER BY d.is_launched DESC, d.created_at DESC
		LIMIT ? OFFSET ?
	`
	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, listArgs...)
	if err != nil {
		fmt.Println("[GetDrugList] List query error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	// 定义返回结构体
	type DrugListItem struct {
		ID               uint      `json:"id"`
		GenericName      string    `json:"generic_name"`
		BrandName        string    `json:"brand_name"`
		Indication       string    `json:"indication"`
		DrugType         string    `json:"drug_type"`
		IsInsurance      bool      `json:"is_insurance"`
		DosageForm       string    `json:"dosage_form"`
		Spec             string    `json:"spec"`
		RefPrice         string    `json:"ref_price"`
		HasRelief        bool      `json:"has_relief"`
		IsLaunched       bool      `json:"is_launched"`
		NeedPrescription bool      `json:"need_prescription"`
		ManualOriginal   string    `json:"manual_original"`
		ManualPopular    string    `json:"manual_popular"`
		AuditStatus      int8      `json:"audit_status"`
		CreatedAt        time.Time `json:"created_at"`
		UpdatedAt        time.Time `json:"updated_at"`
		DiseaseIds       []uint64  `json:"disease_ids"`
	}

	var drugs []DrugListItem
	var drugIDs []uint64
	drugMap := make(map[uint64]int)

	for rows.Next() {
		var drug struct {
			ID               uint           `db:"id"`
			GenericName      string         `db:"generic_name"`
			BrandName        sql.NullString `db:"brand_name"`
			Indication       string         `db:"indication"`
			DrugType         string         `db:"drug_type"`
			IsInsurance      int8           `db:"is_insurance"`
			DosageForm       string         `db:"dosage_form"`
			Spec             string         `db:"spec"`
			RefPrice         sql.NullString `db:"ref_price"`
			HasRelief        int8           `db:"has_relief"`
			IsLaunched       int8           `db:"is_launched"`
			NeedPrescription int8           `db:"need_prescription"`
			ManualOriginal   sql.NullString `db:"manual_original"`
			ManualPopular    sql.NullString `db:"manual_popular"`
			AuditStatus      int8           `db:"audit_status"`
			CreatedAt        time.Time      `db:"created_at"`
			UpdatedAt        time.Time      `db:"updated_at"`
		}

		err := rows.Scan(
			&drug.ID,
			&drug.GenericName,
			&drug.BrandName,
			&drug.Indication,
			&drug.DrugType,
			&drug.IsInsurance,
			&drug.DosageForm,
			&drug.Spec,
			&drug.RefPrice,
			&drug.HasRelief,
			&drug.IsLaunched,
			&drug.NeedPrescription,
			&drug.ManualOriginal,
			&drug.ManualPopular,
			&drug.AuditStatus,
			&drug.CreatedAt,
			&drug.UpdatedAt,
		)

		if err != nil {
			fmt.Printf("[GetDrugList] Scan error: %v\n", err)
			continue
		}

		index := len(drugs)
		item := DrugListItem{
			ID:               drug.ID,
			GenericName:      drug.GenericName,
			BrandName:        drug.BrandName.String,
			Indication:       drug.Indication,
			DrugType:         drug.DrugType,
			IsInsurance:      drug.IsInsurance == 1,
			DosageForm:       drug.DosageForm,
			Spec:             drug.Spec,
			RefPrice:         drug.RefPrice.String,
			HasRelief:        drug.HasRelief == 1,
			IsLaunched:       drug.IsLaunched == 1,
			NeedPrescription: drug.NeedPrescription == 1,
			ManualOriginal:   drug.ManualOriginal.String,
			ManualPopular:    drug.ManualPopular.String,
			AuditStatus:      drug.AuditStatus,
			CreatedAt:        drug.CreatedAt,
			UpdatedAt:        drug.UpdatedAt,
			DiseaseIds:       []uint64{},
		}

		drugs = append(drugs, item)
		drugIDs = append(drugIDs, uint64(drug.ID))
		drugMap[uint64(drug.ID)] = index
	}

	if err = rows.Err(); err != nil {
		fmt.Println("[GetDrugList] Rows iteration error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据遍历失败",
		})
		return
	}

	// 3. 批量查询疾病关联信息
	if len(drugIDs) > 0 {
		placeholders := make([]string, len(drugIDs))
		queryArgs := make([]interface{}, len(drugIDs))
		for i, id := range drugIDs {
			placeholders[i] = "?"
			queryArgs[i] = id
		}

		relQuery := fmt.Sprintf("SELECT drug_id, disease_id FROM drug_disease_rel WHERE drug_id IN (%s)", strings.Join(placeholders, ","))

		relRows, err := db.MySQL.Query(relQuery, queryArgs...)
		if err != nil {
			fmt.Println("[GetDrugList] Relation query error:", err)
		} else {
			defer relRows.Close()
			for relRows.Next() {
				var dID uint64
				var disID uint64
				if err := relRows.Scan(&dID, &disID); err != nil {
					continue
				}
				if idx, ok := drugMap[dID]; ok {
					drugs[idx].DiseaseIds = append(drugs[idx].DiseaseIds, disID)
				}
			}
		}
	}

	if drugs == nil {
		drugs = []DrugListItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":     drugs,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// GetDrugDetail 获取药品详情
func GetDrugDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的药品 ID",
		})
		return
	}

	// 1. 查询药品基本信息
	// 【修改点 1】SELECT 增加 audit_status, created_at, updated_at 等字段，确保返回完整信息
	query := `
		SELECT id, generic_name, brand_name, indication, drug_type, is_insurance,
		       dosage_form, spec, ref_price, has_relief, is_launched, need_prescription,
		       manual_original, manual_popular, audit_status, created_at, updated_at
		FROM rare_drug
		WHERE id = ?
	`

	var drug struct {
		ID               uint           `db:"id"`
		GenericName      string         `db:"generic_name"`
		BrandName        sql.NullString `db:"brand_name"`
		Indication       string         `db:"indication"`
		DrugType         string         `db:"drug_type"`
		IsInsurance      int8           `db:"is_insurance"`
		DosageForm       string         `db:"dosage_form"`
		Spec             string         `db:"spec"`
		RefPrice         sql.NullString `db:"ref_price"`
		HasRelief        int8           `db:"has_relief"`
		IsLaunched       int8           `db:"is_launched"`
		NeedPrescription int8           `db:"need_prescription"`
		ManualOriginal   sql.NullString `db:"manual_original"`
		ManualPopular    sql.NullString `db:"manual_popular"`
		AuditStatus      int8           `db:"audit_status"`
		CreatedAt        time.Time      `db:"created_at"`
		UpdatedAt        time.Time      `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&drug.ID, &drug.GenericName, &drug.BrandName, &drug.Indication,
		&drug.DrugType, &drug.IsInsurance, &drug.DosageForm, &drug.Spec,
		&drug.RefPrice, &drug.HasRelief, &drug.IsLaunched, &drug.NeedPrescription,
		&drug.ManualOriginal, &drug.ManualPopular, &drug.AuditStatus,
		&drug.CreatedAt, &drug.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "药品不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询药品详情失败: " + err.Error(),
		})
		return
	}

	// 2. 查询关联的疾病 ID 和名称
	// 【修改点 2】联查 disease 表获取 name
	diseaseQuery := `
		SELECT d.id, d.name 
		FROM drug_disease_rel rel
		JOIN disease d ON rel.disease_id = d.id
		WHERE rel.drug_id = ?
	`
	diseaseRows, err := db.MySQL.Query(diseaseQuery, id)
	if err != nil {
		fmt.Println("[GetDrugDetail] Query disease error:", err)
		// 即使疾病查询失败，也不应阻断主流程，返回空列表即可
	}

	// 定义疾病简单结构体用于返回
	type DiseaseSimple struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
	}

	var diseases []DiseaseSimple
	if diseaseRows != nil {
		defer diseaseRows.Close()
		for diseaseRows.Next() {
			var dis DiseaseSimple
			if err := diseaseRows.Scan(&dis.ID, &dis.Name); err != nil {
				continue
			}
			diseases = append(diseases, dis)
		}
	}

	// 确保返回空数组而不是 null
	if diseases == nil {
		diseases = []DiseaseSimple{}
	}

	// 3. 构建返回数据
	// 【修改点 3】使用 gin.H 或自定义结构体，确保 JSON Key 为蛇形命名（原样返回）
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":               drug.ID,
			"genericName":      drug.GenericName,
			"brandName":        drug.BrandName.String,
			"indication":       drug.Indication,
			"drugType":         drug.DrugType,
			"isInsurance":      drug.IsInsurance == 1,
			"dosageForm":       drug.DosageForm,
			"spec":             drug.Spec,
			"refPrice":         drug.RefPrice.String,
			"hasRelief":        drug.HasRelief == 1,
			"isLaunched":       drug.IsLaunched == 1,
			"needPrescription": drug.NeedPrescription == 1,
			"manualOriginal":   drug.ManualOriginal.String,
			"manualPopular":    drug.ManualPopular.String,
			"auditStatus":      drug.AuditStatus,
			"createdAt":        drug.CreatedAt.Format(time.RFC3339),
			"updatedAt":        drug.UpdatedAt.Format(time.RFC3339),
			"diseases":         diseases, // 返回包含 id 和 name 的疾病列表
		},
	})
}

// DownloadManual 下载药品说明书
func DownloadManual(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的药品 ID",
		})
		return
	}

	query := `SELECT manual_original, manual_popular FROM rare_drug WHERE id = ? AND audit_status = 1`

	var manualOriginal, manualPopular string
	err = db.MySQL.QueryRow(query, id).Scan(&manualOriginal, &manualPopular)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "药品不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询说明书失败",
		})
		return
	}

	// 返回说明书下载地址
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"original": manualOriginal,
			"popular":  manualPopular,
			"url":      manualOriginal, // 默认返回官方版
		},
	})
}

// ExportDrugs 导出药品名录 Excel
func ExportDrugs(c *gin.Context) {
	// 获取筛选参数
	diseaseStr := c.DefaultQuery("disease", "0")
	keyword := c.DefaultQuery("keyword", "")
	typeFilter := c.DefaultQuery("type", "")
	insuranceStr := c.DefaultQuery("insurance", "")

	diseaseID, _ := strconv.Atoi(diseaseStr)

	// 构建查询条件
	whereClause := "WHERE d.audit_status = 1"
	args := []interface{}{}

	if diseaseID != 0 {
		whereClause += " AND EXISTS (SELECT 1 FROM drug_disease_rel r WHERE r.drug_id = d.id AND r.disease_id = ?)"
		args = append(args, diseaseID)
	}
	if keyword != "" {
		whereClause += " AND (d.generic_name LIKE ? OR d.brand_name LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	if typeFilter != "" {
		whereClause += " AND d.drug_type = ?"
		args = append(args, typeFilter)
	}
	if insuranceStr != "" {
		isInsurance := 0
		if insuranceStr == "true" || insuranceStr == "1" {
			isInsurance = 1
		}
		whereClause += " AND d.is_insurance = ?"
		args = append(args, isInsurance)
	}

	// 查询药品数据
	query := `
		SELECT d.generic_name, d.brand_name, d.indication, d.drug_type, d.is_insurance,
		       d.dosage_form, d.spec, d.ref_price, d.has_relief, d.is_launched
		FROM rare_drug d
		` + whereClause + `
		ORDER BY d.is_launched DESC, d.created_at DESC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询药品数据失败",
		})
		return
	}
	defer rows.Close()

	// 创建 Excel 文件
	excel := excelize.NewFile()
	sheetName := "药品名录"
	index, err := excel.NewSheet(sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建 Excel 工作表失败",
		})
		return
	}
	excel.SetActiveSheet(index)
	excel.DeleteSheet("Sheet1")

	// 设置表头
	headers := []string{"通用名", "商品名", "适应症", "类型", "医保", "剂型", "规格", "参考价格", "赠药援助", "国内上市"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		excel.SetCellValue(sheetName, cell, header)
	}

	// 填充数据
	rowNum := 2
	for rows.Next() {
		var drug struct {
			GenericName string         `db:"generic_name"`
			BrandName   sql.NullString `db:"brand_name"`
			Indication  string         `db:"indication"`
			DrugType    string         `db:"drug_type"`
			IsInsurance int8           `db:"is_insurance"`
			DosageForm  string         `db:"dosage_form"`
			Spec        string         `db:"spec"`
			RefPrice    sql.NullString `db:"ref_price"`
			HasRelief   int8           `db:"has_relief"`
			IsLaunched  int8           `db:"is_launched"`
		}
		if err := rows.Scan(
			&drug.GenericName, &drug.BrandName, &drug.Indication,
			&drug.DrugType, &drug.IsInsurance, &drug.DosageForm, &drug.Spec,
			&drug.RefPrice, &drug.HasRelief, &drug.IsLaunched,
		); err != nil {
			continue
		}

		data := []interface{}{
			drug.GenericName,
			drug.BrandName.String,
			truncateString(drug.Indication, 50),
			convertDrugType(drug.DrugType),
			mapYesNo(drug.IsInsurance),
			drug.DosageForm,
			drug.Spec,
			drug.RefPrice.String,
			mapYesNo(drug.HasRelief),
			mapYesNo(drug.IsLaunched),
		}

		for i, value := range data {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowNum)
			excel.SetCellValue(sheetName, cell, value)
		}
		rowNum++
	}

	// 设置列宽
	for i := 1; i <= len(headers); i++ {
		col, _ := excelize.ColumnNumberToName(i)
		excel.SetColWidth(sheetName, col, col, 20)
	}

	// 设置响应头
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("药品名录_%s.xlsx", timestamp)
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

// GetDrugOptions 获取筛选选项
func GetDrugOptions(c *gin.Context) {
	// 药品类型选项
	types := []OptionItem{
		{Label: "进口药", Value: "origin_import"},
		{Label: "国产药", Value: "origin_domestic"},
		{Label: "仿制药", Value: "generic"},
		{Label: "其他", Value: "other"},
	}

	// 医保选项
	insurances := []OptionItem{
		{Label: "医保", Value: "true"},
		{Label: "非医保", Value: "false"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": DrugOptionsResponse{
			Types:      types,
			Insurances: insurances,
		},
	})
}

// ChannelItem 渠道项响应结构
type ChannelItem struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Address       string `json:"address"`
	Desc          string `json:"desc"`
	Region        string `json:"region"`
	Contact       string `json:"contact"`
	IsInsurance   bool   `json:"isInsurance"`
	DeliveryScope string `json:"deliveryScope"`
	DeliveryCycle string `json:"deliveryCycle"`
}

// ChannelContactResponse 渠道联系方式响应
type ChannelContactResponse struct {
	Phone  string `json:"phone"`
	Wechat string `json:"wechat"`
	Email  string `json:"email"`
}

// GetChannelList 获取渠道列表
func GetChannelList(c *gin.Context) {
	// 获取请求参数
	keyword := c.DefaultQuery("keyword", "")
	provinceCode := c.DefaultQuery("provinceCode", "")
	cityCode := c.DefaultQuery("cityCode", "")
	districtCode := c.DefaultQuery("districtCode", "")
	channelType := c.DefaultQuery("channelType", "")
	delivery := c.DefaultQuery("deliveryScope", "")

	// 【新增】获取 auditStatus 筛选参数
	auditStatusStr := c.DefaultQuery("auditStatus", "")
	isInsuranceSettleStr := c.DefaultQuery("isInsuranceSettle", "")

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
	// 给主表起别名 c，避免关联查询时字段歧义
	// 【修改】初始条件改为 1=1，以便灵活拼接 auditStatus
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 【新增】处理 auditStatus 筛选
	if auditStatusStr != "" {
		auditStatus, err := strconv.Atoi(auditStatusStr)
		if err == nil {
			whereClause += " AND c.audit_status = ?"
			args = append(args, auditStatus)
		}
	} else {
		// 默认行为：如果未指定，通常后台管理可能希望看到所有状态，或者默认只看通过的。
		// 这里保持原有逻辑，如果不传则默认查审核通过的 (audit_status = 1)
		// 如果希望默认查所有，可以注释掉下面这行
		whereClause += " AND c.audit_status = 1"
	}

	// 【新增】处理 isInsuranceSettle 筛选
	if isInsuranceSettleStr != "" {
		isInsuranceSettle := 0
		// 支持 "true"/"1" 或 "false"/"0"
		if isInsuranceSettleStr == "true" || isInsuranceSettleStr == "1" {
			isInsuranceSettle = 1
		}
		whereClause += " AND c.is_insurance_settle = ?"
		args = append(args, isInsuranceSettle)
	}

	// 【新增】按渠道名称模糊搜索
	if keyword != "" {
		whereClause += " AND c.name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	// 按省份 Code 筛选
	if provinceCode != "" {
		whereClause += " AND c.province_code = ?"
		args = append(args, provinceCode)
	}
	// 按城市 Code 筛选
	if cityCode != "" {
		whereClause += " AND c.city_code = ?"
		args = append(args, cityCode)
	}
	// 按区县 Code 筛选
	if districtCode != "" {
		whereClause += " AND c.district_code = ?"
		args = append(args, districtCode)
	}
	// 按渠道类型筛选
	if channelType != "" {
		whereClause += " AND c.channel_type = ?"
		args = append(args, channelType)
	}

	// 根据配送方式筛选
	if delivery != "" {
		switch delivery {
		case "pickup":
			whereClause += " AND c.delivery_scope LIKE ?"
			args = append(args, "%仅门店自提%")
		case "delivery":
			whereClause += " AND c.delivery_scope LIKE ?"
			args = append(args, "%全国%")
		case "local":
			whereClause += " AND c.delivery_scope LIKE ?"
			args = append(args, "%同城%")
		}
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM drug_channel c " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败: " + err.Error(),
		})
		return
	}

	// 查询列表
	// 【修改】SELECT 中增加 c.audit_status
	listQuery := `
		SELECT 
			c.id, c.name, c.channel_type, c.province_code, c.city_code, c.district_code, 
			c.address, c.contact_phone, c.contact_url,
			c.delivery_scope, c.delivery_cycle, c.is_insurance_settle, c.qualification, 
			c.created_at, c.updated_at,
			c.province_name, c.city_name, c.district_name,
			c.audit_status -- 新增：审核状态
		FROM drug_channel c
		` + whereClause + `
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		fmt.Printf("[GetChannelList] Query error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	// 使用 map[string]interface{} 存储每一行数据
	var list []map[string]interface{}
	for rows.Next() {
		// 定义变量接收所有字段
		var (
			id                uint
			name              string
			channelType       string
			provinceCode      string
			cityCode          string
			districtCode      sql.NullString
			address           sql.NullString
			contactPhone      sql.NullString
			contactUrl        sql.NullString
			deliveryScope     string
			deliveryCycle     string
			isInsuranceSettle int8
			qualification     sql.NullString
			createdAt         time.Time
			updatedAt         time.Time

			// 地区名称变量
			provinceName sql.NullString
			cityName     sql.NullString
			districtName sql.NullString

			// 【新增】审核状态变量
			auditStatus int8
		)

		// 【修改】Scan 增加 auditStatus
		err := rows.Scan(
			&id, &name, &channelType, &provinceCode, &cityCode, &districtCode,
			&address, &contactPhone, &contactUrl,
			&deliveryScope, &deliveryCycle, &isInsuranceSettle,
			&qualification, &createdAt, &updatedAt,
			&provinceName, &cityName, &districtName,
			&auditStatus, // 新增 Scan
		)
		if err != nil {
			fmt.Printf("[GetChannelList] Scan error: %v\n", err)
			continue
		}

		// 构建 map
		item := map[string]interface{}{
			"id":                id,
			"name":              name,
			"channelType":       channelType,
			"provinceCode":      provinceCode,
			"cityCode":          cityCode,
			"districtCode":      districtCode.String,
			"address":           address.String,
			"contactPhone":      contactPhone.String,
			"contactUrl":        contactUrl.String,
			"deliveryScope":     deliveryScope,
			"deliveryCycle":     deliveryCycle,
			"isInsuranceSettle": isInsuranceSettle == 1,
			"qualification":     qualification.String,
			"createdAt":         createdAt.Format(time.RFC3339),
			"updatedAt":         updatedAt.Format(time.RFC3339),

			"provinceName": provinceName.String,
			"cityName":     cityName.String,
			"districtName": districtName.String,

			// 【新增】返回审核状态
			"auditStatus": auditStatus,
		}

		list = append(list, item)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据遍历失败",
		})
		return
	}

	if list == nil {
		list = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":     list,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// GetChannelDetail 获取渠道详情及联系方式
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

	// 1. 查询渠道基本信息
	query := `
		SELECT 
			id, name, channel_type, province_code, city_code, district_code, address, contact_phone, contact_url,
			delivery_scope, delivery_cycle, is_insurance_settle, audit_status, qualification,
			province_name, city_name, district_name
		FROM drug_channel
		WHERE id = ? AND audit_status = 1
	`

	var channel struct {
		ID            uint           `db:"id"`
		Name          string         `db:"name"`
		ChannelType   string         `db:"channel_type"`
		ProvinceCode  string         `db:"province_code"`
		CityCode      string         `db:"city_code"`
		DistrictCode  sql.NullString `db:"district_code"`
		Address       sql.NullString `db:"address"`
		ContactPhone  sql.NullString `db:"contact_phone"`
		ContactURL    sql.NullString `db:"contact_url"`
		DeliveryScope string         `db:"delivery_scope"`
		DeliveryCycle string         `db:"delivery_cycle"`
		IsInsurance   int8           `db:"is_insurance_settle"`
		AuditStatus   int8           `db:"audit_status"`
		Qualification sql.NullString `db:"qualification"`
		ProvinceName  sql.NullString `db:"province_name"`
		CityName      sql.NullString `db:"city_name"`
		DistrictName  sql.NullString `db:"district_name"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&channel.ID, &channel.Name, &channel.ChannelType, &channel.ProvinceCode, &channel.CityCode, &channel.DistrictCode,
		&channel.Address, &channel.ContactPhone, &channel.ContactURL,
		&channel.DeliveryScope, &channel.DeliveryCycle, &channel.IsInsurance, &channel.AuditStatus, &channel.Qualification,
		&channel.ProvinceName, &channel.CityName, &channel.DistrictName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "渠道不存在",
			})
			return
		}
		fmt.Printf("[GetChannelDetail] Query channel error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询渠道详情失败: " + err.Error(),
		})
		return
	}

	// 2. 查询关联的药品列表
	// 【新增】通过 drug_channel_drug_rel 关联 rare_drug
	drugQuery := `
		SELECT d.id, d.generic_name, d.brand_name
		FROM drug_channel_drug_rel rel
		JOIN rare_drug d ON rel.drug_id = d.id
		WHERE rel.channel_id = ?
	`
	drugRows, err := db.MySQL.Query(drugQuery, id)
	if err != nil {
		fmt.Printf("[GetChannelDetail] Query drugs error: %v\n", err)
		// 药品查询失败不应阻断主流程，返回空列表即可
	}

	// 定义药品简单结构体用于返回
	type ChannelDrugItem struct {
		ID          uint   `json:"id"`
		GenericName string `json:"genericName"`
		BrandName   string `json:"brandName"`
	}

	var drugs []ChannelDrugItem
	if drugRows != nil {
		defer drugRows.Close()
		for drugRows.Next() {
			var d struct {
				ID          uint           `db:"id"`
				GenericName string         `db:"generic_name"`
				BrandName   sql.NullString `db:"brand_name"`
			}
			if err := drugRows.Scan(&d.ID, &d.GenericName, &d.BrandName); err != nil {
				continue
			}

			// 构建显示名称逻辑，或者分别返回通用名和商品名
			brandName := ""
			if d.BrandName.Valid {
				brandName = d.BrandName.String
			}

			drugs = append(drugs, ChannelDrugItem{
				ID:          d.ID,
				GenericName: d.GenericName,
				BrandName:   brandName,
			})
		}
	}

	// 确保返回空数组而不是 null
	if drugs == nil {
		drugs = []ChannelDrugItem{}
	}

	// 3. 构建描述信息
	descParts := []string{}
	if channel.DeliveryScope != "" {
		descParts = append(descParts, channel.DeliveryScope)
	}
	if channel.IsInsurance == 1 {
		descParts = append(descParts, "支持医保")
	}
	if channel.DeliveryCycle != "" {
		descParts = append(descParts, channel.DeliveryCycle)
	}
	desc := ""
	if len(descParts) > 0 {
		desc = strings.Join(descParts, "，")
	}

	// 4. 构建返回数据
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":            channel.ID,
			"name":          channel.Name,
			"channelType":   channel.ChannelType,
			"provinceCode":  channel.ProvinceCode,
			"cityCode":      channel.CityCode,
			"districtCode":  channel.DistrictCode.String,
			"address":       channel.Address.String,
			"desc":          desc,
			"contactPhone":  channel.ContactPhone.String,
			"contactUrl":    channel.ContactURL.String,
			"qualification": channel.Qualification.String,
			"deliveryScope": channel.DeliveryScope,
			"auditStatus":   channel.AuditStatus,
			"provinceName":  channel.ProvinceName.String,
			"cityName":      channel.CityName.String,
			"districtName":  channel.DistrictName.String,

			// 【新增】返回关联药品列表
			"drugs": drugs,
		},
	})
}

// CreateChannelRequest 新增渠道请求
type CreateChannelRequest struct {
	DrugID            uint   `json:"drugId"`
	Name              string `json:"name"`
	ChannelType       string `json:"channelType"`
	ProvinceCode      string `json:"provinceCode"` // 修改为 Code
	CityCode          string `json:"cityCode"`     // 修改为 Code
	DistrictCode      string `json:"districtCode"` // 修改为 Code
	Address           string `json:"address"`
	ContactPhone      string `json:"contactPhone"`
	ContactURL        string `json:"contactUrl"`
	DeliveryScope     string `json:"deliveryScope"`
	DeliveryCycle     string `json:"deliveryCycle"`
	IsInsuranceSettle bool   `json:"isInsuranceSettle"`
	Qualification     string `json:"qualification"`
}

// CreateChannel 新增渠道
func CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 必填字段验证
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "渠道名称不能为空",
		})
		return
	}
	// 验证 Code 是否为空
	if req.ProvinceCode == "" || req.CityCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "省份和城市 Code 不能为空",
		})
		return
	}
	if req.ContactPhone == "" && req.ContactURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "至少提供一种联系方式",
		})
		return
	}

	// 插入数据库
	insertQuery := `
		INSERT INTO drug_channel 
		(drug_id, name, channel_type, province_code, city_code, district_code, address, contact_phone, contact_url,
		 delivery_scope, delivery_cycle, is_insurance_settle, qualification, audit_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
	`
	now := time.Now()
	result, err := db.MySQL.Exec(insertQuery,
		req.DrugID, req.Name, req.ChannelType, req.ProvinceCode, req.CityCode, req.DistrictCode, req.Address,
		req.ContactPhone, req.ContactURL,
		req.DeliveryScope, req.DeliveryCycle, boolToInt(req.IsInsuranceSettle),
		req.Qualification, now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "新增渠道失败",
		})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": id,
		},
	})
}

// UpdateChannelRequest 更新渠道请求
type UpdateChannelRequest struct {
	Name              *string `json:"name"`
	ChannelType       *string `json:"channelType"`
	ProvinceCode      *string `json:"provinceCode"` // 修改为 Code
	CityCode          *string `json:"cityCode"`     // 修改为 Code
	DistrictCode      *string `json:"districtCode"` // 修改为 Code
	Address           *string `json:"address"`
	ContactPhone      *string `json:"contactPhone"`
	ContactURL        *string `json:"contactUrl"`
	DeliveryScope     *string `json:"deliveryScope"`
	DeliveryCycle     *string `json:"deliveryCycle"`
	IsInsuranceSettle *bool   `json:"isInsuranceSettle"`
	Qualification     *string `json:"qualification"`
	DrugID            *uint   `json:"drugId"`
}

// UpdateChannel 更新渠道
func UpdateChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道 ID",
		})
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 检查渠道是否存在
	checkQuery := "SELECT id FROM drug_channel WHERE id = ?"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
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
			"message": "查询渠道失败",
		})
		return
	}

	// 构建动态更新语句
	updateFields := []string{}
	updateArgs := []interface{}{}

	if req.Name != nil {
		updateFields = append(updateFields, "name = ?")
		updateArgs = append(updateArgs, *req.Name)
	}
	if req.ChannelType != nil {
		updateFields = append(updateFields, "channel_type = ?")
		updateArgs = append(updateArgs, *req.ChannelType)
	}
	if req.ProvinceCode != nil {
		updateFields = append(updateFields, "province_code = ?")
		updateArgs = append(updateArgs, *req.ProvinceCode)
	}
	if req.CityCode != nil {
		updateFields = append(updateFields, "city_code = ?")
		updateArgs = append(updateArgs, *req.CityCode)
	}
	if req.DistrictCode != nil {
		updateFields = append(updateFields, "district_code = ?")
		updateArgs = append(updateArgs, *req.DistrictCode)
	}
	if req.Address != nil {
		updateFields = append(updateFields, "address = ?")
		updateArgs = append(updateArgs, *req.Address)
	}
	if req.ContactPhone != nil {
		updateFields = append(updateFields, "contact_phone = ?")
		updateArgs = append(updateArgs, *req.ContactPhone)
	}
	if req.ContactURL != nil {
		updateFields = append(updateFields, "contact_url = ?")
		updateArgs = append(updateArgs, *req.ContactURL)
	}
	if req.DeliveryScope != nil {
		updateFields = append(updateFields, "delivery_scope = ?")
		updateArgs = append(updateArgs, *req.DeliveryScope)
	}
	if req.DeliveryCycle != nil {
		updateFields = append(updateFields, "delivery_cycle = ?")
		updateArgs = append(updateArgs, *req.DeliveryCycle)
	}
	if req.IsInsuranceSettle != nil {
		updateFields = append(updateFields, "is_insurance_settle = ?")
		updateArgs = append(updateArgs, boolToInt(*req.IsInsuranceSettle))
	}
	if req.Qualification != nil {
		updateFields = append(updateFields, "qualification = ?")
		updateArgs = append(updateArgs, *req.Qualification)
	}
	if req.DrugID != nil {
		updateFields = append(updateFields, "drug_id = ?")
		updateArgs = append(updateArgs, *req.DrugID)
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "未提供更新字段",
		})
		return
	}

	// 添加更新时间和 ID
	updateFields = append(updateFields, "updated_at = ?")
	updateArgs = append(updateArgs, time.Now())
	updateArgs = append(updateArgs, id)

	updateQuery := "UPDATE drug_channel SET " + strings.Join(updateFields, ", ") + " WHERE id = ?"
	_, err = db.MySQL.Exec(updateQuery, updateArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新渠道失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DeleteChannel 删除渠道（软删除）
func DeleteChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道 ID",
		})
		return
	}

	// 检查渠道是否存在
	checkQuery := "SELECT id FROM drug_channel WHERE id = ? AND audit_status = 1"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
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
			"message": "查询渠道失败",
		})
		return
	}

	// 软删除：将 audit_status 设为 2 (驳回/删除)
	deleteQuery := "UPDATE drug_channel SET audit_status = 2, updated_at = ? WHERE id = ?"
	_, err = db.MySQL.Exec(deleteQuery, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除渠道失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// ContactChannel 获取联系方式
func ContactChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道 ID",
		})
		return
	}

	// 使用匿名结构体解析请求体（如果前端还需要传 contactType 来区分返回电话还是URL，可以保留，否则直接全返回）
	var req struct {
		ContactType string `json:"contactType"`
	}
	// 忽略绑定错误，因为可能是 GET 请求或者不需要 body
	_ = c.ShouldBindJSON(&req)

	// 查询渠道信息
	query := `SELECT contact_phone, contact_url, name FROM drug_channel WHERE id = ? AND audit_status = 1`
	var contactPhone, contactURL, name string
	err = db.MySQL.QueryRow(query, id).Scan(&contactPhone, &contactURL, &name)
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
			"message": "查询联系方式失败",
		})
		return
	}

	// 根据 contactType 决定返回内容，或者默认返回所有
	contactData := ChannelContactResponse{
		Phone:  contactPhone,
		Wechat: "", // 新表中没有微信字段，如果有需求可能需要从 contact_url 或备注中解析，这里留空
		Email:  "", // 新表中没有邮箱字段
	}

	// 如果前端指定要 url (例如在线表单链接)，可以放在 Wechat 或 Email 字段复用，或者扩展结构体
	// 这里假设 contact_url 可能是微信号链接或在线客服链接
	if req.ContactType == "url" {
		contactData.Wechat = contactURL
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    contactData,
	})
}

// FeedbackChannel 提交反馈评价
func FeedbackChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道 ID",
		})
		return
	}

	// 使用匿名结构体解析请求体
	var req struct {
		Rating  int    `json:"rating"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 验证评分范围
	if req.Rating < 1 || req.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "评分必须在 1-5 之间",
		})
		return
	}

	// 验证渠道是否存在
	checkQuery := "SELECT id FROM drug_channel WHERE id = ? AND audit_status = 1"
	var channelID uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&channelID)
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
			"message": "验证渠道失败",
		})
		return
	}

	// 插入反馈记录
	insertQuery := `
		INSERT INTO drug_channel_feedback (channel_id, rating, content, created_at) 
		VALUES (?, ?, ?, ?)
	`
	_, err = db.MySQL.Exec(insertQuery, id, req.Rating, req.Content, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交反馈失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DonationItem 赠药项目响应结构
type DonationItem struct {
	ID            uint   `json:"id"`
	DrugID        uint   `json:"drugId"`
	DiseaseValue  int    `json:"diseaseValue"`
	Name          string `json:"name"`
	Organizer     string `json:"organizer"`
	Condition     string `json:"condition"`
	Period        string `json:"period"`
	Dosage        string `json:"dosage"`
	ApplyForm     string `json:"applyForm"`
	ApplyGuide    string `json:"applyGuide"`
	MaterialList  string `json:"materialList"`
	ProgressQuery string `json:"progressQuery"`
}

// GetDonationList 获取赠药援助项目列表
func GetDonationList(c *gin.Context) {
	// 获取请求参数
	keyword := c.DefaultQuery("keyword", "")
	diseaseStr := c.DefaultQuery("diseaseId", "0")
	drugIDStr := c.DefaultQuery("drugId", "0")

	// 【新增】获取 auditStatus 筛选参数
	auditStatusStr := c.DefaultQuery("auditStatus", "")
	organizer := c.DefaultQuery("organizer", "") // 新增：主办方筛选

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	diseaseID, _ := strconv.Atoi(diseaseStr)
	drugID, _ := strconv.Atoi(drugIDStr)
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 构建基础查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 【新增】处理 auditStatus 筛选
	if auditStatusStr != "" {
		auditStatus, err := strconv.Atoi(auditStatusStr)
		if err == nil {
			whereClause += " AND p.audit_status = ?"
			args = append(args, auditStatus)
		}
	} else {
		// 默认行为：如果未指定，可以根据业务需求决定默认查哪个状态。
		// 通常后台管理列表可能希望看到所有状态，或者默认只看通过的。
		// 这里假设如果不传，默认只查审核通过的 (audit_status = 1)，保持与之前逻辑一致或根据需求调整
		// 如果希望默认查所有，可以注释掉下面这行
		whereClause += " AND p.audit_status = 1"
	}

	// 【新增】处理 organizer 筛选 (模糊匹配)
	if organizer != "" {
		whereClause += " AND p.organizer LIKE ?"
		args = append(args, "%"+organizer+"%")
	}

	// 【修改】关键字模糊筛选：匹配项目名称、药品商品名、药品通用名
	if keyword != "" {
		whereClause += " AND (p.name LIKE ? OR d.brand_name LIKE ? OR d.generic_name LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if diseaseID != 0 {
		whereClause += " AND p.disease_id = ?"
		args = append(args, diseaseID)
	}

	if drugID != 0 {
		whereClause += " AND p.drug_id = ?"
		args = append(args, drugID)
	}

	// 【修改】查询总数：需要加入 JOIN 以支持 keyword 过滤
	// 注意：Count 查询也需要包含 audit_status 的条件
	countQuery := "SELECT COUNT(*) FROM drug_relief_project p LEFT JOIN rare_drug d ON p.drug_id = d.id " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败: " + err.Error(),
		})
		return
	}

	// 【核心修改】查询列表：
	// 1. 增加 p.audit_status, p.updated_at
	// 2. 增加 LEFT JOIN disease
	listQuery := `
		SELECT 
			p.id, p.drug_id, p.disease_id, p.name, p.organizer, p.apply_condition,
		       p.relief_cycle, p.relief_dosage_desc, p.apply_form, p.apply_guide, p.material_list, p.progress_query,
			p.audit_status, p.updated_at, -- 新增：审核状态和更新时间
			d.generic_name, d.brand_name,
			dis.name as disease_name
		FROM drug_relief_project p
		LEFT JOIN rare_drug d ON p.drug_id = d.id
		LEFT JOIN disease dis ON p.disease_id = dis.id
		` + whereClause + `
		ORDER BY p.created_at DESC
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

	// 【修改】定义接收结构体，增加新字段
	type ProjectRow struct {
		ID               uint           `db:"id"`
		DrugID           uint           `db:"drug_id"`
		DiseaseID        int            `db:"disease_id"`
		Name             string         `db:"name"`
		Organizer        string         `db:"organizer"`
		ApplyCondition   string         `db:"apply_condition"`
		ReliefCycle      string         `db:"relief_cycle"`
		ReliefDosageDesc string         `db:"relief_dosage_desc"`
		ApplyForm        sql.NullString `db:"apply_form"`
		ApplyGuide       sql.NullString `db:"apply_guide"`
		MaterialList     sql.NullString `db:"material_list"`
		ProgressQuery    sql.NullString `db:"progress_query"`

		// 新增字段
		AuditStatus int8      `db:"audit_status"`
		UpdatedAt   time.Time `db:"updated_at"`

		GenericName string         `db:"generic_name"`
		BrandName   sql.NullString `db:"brand_name"`
		DiseaseName sql.NullString `db:"disease_name"`
	}

	var list []gin.H
	for rows.Next() {
		var project ProjectRow

		// 【修改】Scan 所有字段
		if err := rows.Scan(
			&project.ID, &project.DrugID, &project.DiseaseID, &project.Name,
			&project.Organizer, &project.ApplyCondition, &project.ReliefCycle,
			&project.ReliefDosageDesc, &project.ApplyForm, &project.ApplyGuide,
			&project.MaterialList, &project.ProgressQuery,
			&project.AuditStatus, &project.UpdatedAt, // 新增 Scan
			&project.GenericName, &project.BrandName, &project.DiseaseName,
		); err != nil {
			fmt.Printf("[GetDonationList] Scan error: %v\n", err)
			continue
		}

		// 构建返回的药品显示名称
		drugName := project.GenericName
		if project.BrandName.Valid && project.BrandName.String != "" {
			drugName = project.BrandName.String + " (" + project.GenericName + ")"
		} else if project.GenericName == "" {
			drugName = "未知药品"
		}

		// 构建返回的疾病名称
		diseaseName := ""
		if project.DiseaseName.Valid {
			diseaseName = project.DiseaseName.String
		} else {
			diseaseName = "未知疾病"
		}

		list = append(list, gin.H{
			"id":               project.ID,
			"drugId":           project.DrugID,
			"diseaseId":        project.DiseaseID,
			"name":             project.Name,
			"organizer":        project.Organizer,
			"applyCondition":   project.ApplyCondition,
			"reliefCycle":      project.ReliefCycle,
			"reliefDosageDesc": project.ReliefDosageDesc,
			"applyForm":        project.ApplyForm.String,
			"applyGuide":       project.ApplyGuide.String,
			"materialList":     project.MaterialList.String,
			"progressQuery":    project.ProgressQuery.String,

			// 【新增】返回关联名称及状态时间
			"drugName":    drugName,
			"diseaseName": diseaseName,
			"auditStatus": project.AuditStatus,
			"updatedAt":   project.UpdatedAt.Format(time.RFC3339), // 格式化时间
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// GetDonationDetail 获取赠药项目详情
func GetDonationDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目 ID",
		})
		return
	}

	// 【核心修改】查询详情：
	// 1. 增加 LEFT JOIN rare_drug 和 disease
	// 2. 增加 p.reject_reason, p.updated_at
	query := `
		SELECT 
			p.id, p.drug_id, p.disease_id, p.name, p.organizer, p.apply_condition,
		       p.relief_cycle, p.relief_dosage_desc, p.apply_form, p.apply_guide, p.material_list, p.progress_query,
		       p.audit_status, p.reject_reason, p.created_at, p.updated_at,
			d.generic_name, d.brand_name,
			dis.name as disease_name
		FROM drug_relief_project p
		LEFT JOIN rare_drug d ON p.drug_id = d.id
		LEFT JOIN disease dis ON p.disease_id = dis.id
		WHERE p.id = ?
	`

	// 【修改】定义接收结构体，增加新字段
	var project struct {
		ID               uint           `db:"id"`
		DrugID           uint           `db:"drug_id"`
		DiseaseID        int            `db:"disease_id"`
		Name             string         `db:"name"`
		Organizer        string         `db:"organizer"`
		ApplyCondition   string         `db:"apply_condition"`
		ReliefCycle      string         `db:"relief_cycle"`
		ReliefDosageDesc string         `db:"relief_dosage_desc"`
		ApplyForm        sql.NullString `db:"apply_form"`
		ApplyGuide       sql.NullString `db:"apply_guide"`
		MaterialList     sql.NullString `db:"material_list"`
		ProgressQuery    sql.NullString `db:"progress_query"`

		// 新增字段
		AuditStatus  int8           `db:"audit_status"`
		RejectReason sql.NullString `db:"reject_reason"` // 假设表中存在此字段
		CreatedAt    time.Time      `db:"created_at"`
		UpdatedAt    time.Time      `db:"updated_at"`

		GenericName string         `db:"generic_name"`
		BrandName   sql.NullString `db:"brand_name"`
		DiseaseName sql.NullString `db:"disease_name"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&project.ID, &project.DrugID, &project.DiseaseID, &project.Name,
		&project.Organizer, &project.ApplyCondition, &project.ReliefCycle,
		&project.ReliefDosageDesc, &project.ApplyForm, &project.ApplyGuide,
		&project.MaterialList, &project.ProgressQuery,
		&project.AuditStatus, &project.RejectReason, &project.CreatedAt, &project.UpdatedAt,
		&project.GenericName, &project.BrandName, &project.DiseaseName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "赠药项目不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询项目详情失败: " + err.Error(),
		})
		return
	}

	// 构建返回的药品显示名称
	drugName := project.GenericName
	if project.BrandName.Valid && project.BrandName.String != "" {
		drugName = project.BrandName.String + " (" + project.GenericName + ")"
	} else if project.GenericName == "" {
		drugName = "未知药品"
	}

	// 构建返回的疾病名称
	diseaseName := ""
	if project.DiseaseName.Valid {
		diseaseName = project.DiseaseName.String
	} else {
		diseaseName = "未知疾病"
	}

	// 构建驳回原因
	rejectReason := ""
	if project.RejectReason.Valid {
		rejectReason = project.RejectReason.String
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":            project.ID,
			"drugId":        project.DrugID,
			"diseaseId":     project.DiseaseID,
			"name":          project.Name,
			"organizer":     project.Organizer,
			"condition":     project.ApplyCondition,
			"period":        project.ReliefCycle,
			"dosage":        project.ReliefDosageDesc,
			"applyForm":     project.ApplyForm.String,
			"applyGuide":    project.ApplyGuide.String,
			"materialList":  project.MaterialList.String,
			"progressQuery": project.ProgressQuery.String,

			// 【新增】返回关联名称及额外信息
			"drugName":     drugName,
			"diseaseName":  diseaseName,
			"auditStatus":  project.AuditStatus,
			"rejectReason": rejectReason,
			"createdAt":    project.CreatedAt.Format(time.RFC3339),
			"updatedAt":    project.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// CreateDonationRequest 新增赠药项目请求
type CreateDonationRequest struct {
	DrugID         uint   `json:"drugId"`         // 关联药品 ID
	DiseaseValue   int    `json:"diseaseValue"`   // 疾病分类
	Name           string `json:"name"`           // 项目名称
	Organizer      string `json:"organizer"`      // 主办方
	ApplyCondition string `json:"applyCondition"` // 申请条件
	ReliefCycle    string `json:"reliefCycle"`    // 援助周期
	DrugDosage     string `json:"drugDosage"`     // 药品剂量
	ApplyForm      string `json:"applyForm"`      // 申请表格链接
	ApplyGuide     string `json:"applyGuide"`     // 申请指南链接
	MaterialList   string `json:"materialList"`   // 材料清单
	ProgressQuery  string `json:"progressQuery"`  // 进度查询方式
}

// CreateDonation 新增赠药项目
func CreateDonation(c *gin.Context) {
	var req CreateDonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 必填字段验证
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "项目名称不能为空",
		})
		return
	}
	if req.Organizer == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "主办方不能为空",
		})
		return
	}
	if req.DrugID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "关联药品 ID 不能为空",
		})
		return
	}

	// 验证药品是否存在
	checkQuery := "SELECT id FROM rare_drug WHERE id = ? AND audit_status = 1"
	var exists uint
	err := db.MySQL.QueryRow(checkQuery, req.DrugID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "关联药品不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "验证药品失败",
		})
		return
	}

	// 插入数据库
	// 注意：字段名变更，is_audit 变为 audit_status
	insertQuery := `
		INSERT INTO drug_relief_project 
		(drug_id, disease_id, name, organizer, apply_condition,
		 relief_cycle, relief_dosage_desc, apply_form, apply_guide, material_list, progress_query,
		 audit_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
	`
	now := time.Now()
	result, err := db.MySQL.Exec(insertQuery,
		req.DrugID, req.DiseaseValue, req.Name, req.Organizer, req.ApplyCondition,
		req.ReliefCycle, req.DrugDosage, req.ApplyForm, req.ApplyGuide,
		req.MaterialList, req.ProgressQuery, now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "新增项目失败",
		})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": id,
		},
	})
}

// UpdateDonationRequest 更新赠药项目请求
type UpdateDonationRequest struct {
	DrugID         *uint   `json:"drugId"`
	DiseaseValue   *int    `json:"diseaseValue"`
	Name           *string `json:"name"`
	Organizer      *string `json:"organizer"`
	ApplyCondition *string `json:"applyCondition"`
	ReliefCycle    *string `json:"reliefCycle"`
	DrugDosage     *string `json:"drugDosage"`
	ApplyForm      *string `json:"applyForm"`
	ApplyGuide     *string `json:"applyGuide"`
	MaterialList   *string `json:"materialList"`
	ProgressQuery  *string `json:"progressQuery"`
}

// UpdateDonation 更新赠药项目
func UpdateDonation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目 ID",
		})
		return
	}

	var req UpdateDonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 检查项目是否存在
	checkQuery := "SELECT id FROM drug_relief_project WHERE id = ?"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
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
			"message": "查询项目失败",
		})
		return
	}

	// 构建动态更新语句
	updateFields := []string{}
	updateArgs := []interface{}{}

	if req.DrugID != nil {
		// 验证药品是否存在
		drugCheckQuery := "SELECT id FROM rare_drug WHERE id = ? AND audit_status = 1"
		var drugExists uint
		if err := db.MySQL.QueryRow(drugCheckQuery, *req.DrugID).Scan(&drugExists); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "关联药品不存在",
			})
			return
		}
		updateFields = append(updateFields, "drug_id = ?")
		updateArgs = append(updateArgs, *req.DrugID)
	}
	if req.DiseaseValue != nil {
		updateFields = append(updateFields, "disease_id = ?")
		updateArgs = append(updateArgs, *req.DiseaseValue)
	}
	if req.Name != nil {
		updateFields = append(updateFields, "name = ?")
		updateArgs = append(updateArgs, *req.Name)
	}
	if req.Organizer != nil {
		updateFields = append(updateFields, "organizer = ?")
		updateArgs = append(updateArgs, *req.Organizer)
	}
	if req.ApplyCondition != nil {
		updateFields = append(updateFields, "apply_condition = ?")
		updateArgs = append(updateArgs, *req.ApplyCondition)
	}
	if req.ReliefCycle != nil {
		updateFields = append(updateFields, "relief_cycle = ?")
		updateArgs = append(updateArgs, *req.ReliefCycle)
	}
	if req.DrugDosage != nil {
		updateFields = append(updateFields, "relief_dosage_desc = ?")
		updateArgs = append(updateArgs, *req.DrugDosage)
	}
	if req.ApplyForm != nil {
		updateFields = append(updateFields, "apply_form = ?")
		updateArgs = append(updateArgs, *req.ApplyForm)
	}
	if req.ApplyGuide != nil {
		updateFields = append(updateFields, "apply_guide = ?")
		updateArgs = append(updateArgs, *req.ApplyGuide)
	}
	if req.MaterialList != nil {
		updateFields = append(updateFields, "material_list = ?")
		updateArgs = append(updateArgs, *req.MaterialList)
	}
	if req.ProgressQuery != nil {
		updateFields = append(updateFields, "progress_query = ?")
		updateArgs = append(updateArgs, *req.ProgressQuery)
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "未提供更新字段",
		})
		return
	}

	// 添加更新时间和 ID
	updateFields = append(updateFields, "updated_at = ?")
	updateArgs = append(updateArgs, time.Now())
	updateArgs = append(updateArgs, id)

	updateQuery := "UPDATE drug_relief_project SET " + strings.Join(updateFields, ", ") + " WHERE id = ?"
	_, err = db.MySQL.Exec(updateQuery, updateArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新项目失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DeleteDonation 删除赠药项目（软删除）
func DeleteDonation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目 ID",
		})
		return
	}

	// 检查项目是否存在
	checkQuery := "SELECT id FROM drug_relief_project WHERE id = ? AND audit_status = 1"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
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
			"message": "查询项目失败",
		})
		return
	}

	// 软删除：将 audit_status 设为 2 (驳回/删除)
	deleteQuery := "UPDATE drug_relief_project SET audit_status = 2, updated_at = ? WHERE id = ?"
	_, err = db.MySQL.Exec(deleteQuery, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除项目失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// ApplyDonation 提交赠药援助申请
func ApplyDonation(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目 ID",
		})
		return
	}

	// 使用匿名结构体解析请求体
	var req struct {
		UserID         uint   `json:"userId"`
		PatientName    string `json:"patientName"`
		PatientIdCard  string `json:"patientIdCard"` // 明文身份证号
		DiagnosisProof string `json:"diagnosisProof"`
		IncomeProof    string `json:"incomeProof"`
		ContactPhone   string `json:"contactPhone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 验证必填字段
	if req.PatientName == "" || req.PatientIdCard == "" || req.ContactPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请填写完整信息",
		})
		return
	}

	// 验证身份证号格式
	if len(req.PatientIdCard) != 18 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "身份证号格式错误",
		})
		return
	}

	// 验证赠药项目是否存在
	checkQuery := "SELECT id, name FROM drug_relief_project WHERE id = ? AND audit_status = 1"
	var projectName string
	err = db.MySQL.QueryRow(checkQuery, projectID).Scan(&projectID, &projectName)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "赠药项目不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "验证项目失败",
		})
		return
	}

	// --- 敏感数据处理开始 ---
	// 1. 加密存储 (生产环境请使用 AES-256 等强加密算法)
	// 这里仅做示例，使用 Base64 模拟加密，实际不可逆或需密钥
	idCardEnc := base64.StdEncoding.EncodeToString([]byte(req.PatientIdCard))

	// 2. 脱敏展示 (例如: 110101********1234)
	maskedID := req.PatientIdCard[:6] + "********" + req.PatientIdCard[14:]
	// --- 敏感数据处理结束 ---

	// 生成申请编号
	applicationID := fmt.Sprintf("RELIEF%s%03d", time.Now().Format("20060102"), projectID)

	// 插入申请记录
	// 注意：表名变更，字段变更为 patient_id_card_enc, patient_id_card_mask
	insertQuery := `
		INSERT INTO drug_relief_application 
		(project_id, user_id, patient_name, patient_id_card_enc, patient_id_card_mask,
		 diagnosis_proof, income_proof, contact_phone, status, submit_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?)
	`
	now := time.Now()
	result, err := db.MySQL.Exec(insertQuery,
		projectID, req.UserID, req.PatientName, idCardEnc, maskedID,
		req.DiagnosisProof, req.IncomeProof, req.ContactPhone, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交申请失败",
		})
		return
	}

	// 获取最后插入的 ID 用于日志
	appID, err := result.LastInsertId()
	if err != nil {
		// 如果获取 ID 失败，记录日志但不一定阻断主流程，视业务需求而定
		// 这里为了严谨，如果拿不到 ID 就无法记录日志，但申请已经成功
		fmt.Println("Failed to get last insert ID:", err)
	} else {
		// 插入进度日志
		// 注意：表名变更，字段 desc 变为 action_desc
		logQuery := `
			INSERT INTO drug_relief_log (application_id, status, action_desc, created_at)
			VALUES (?, 'pending', '申请已提交', ?)
		`
		_, _ = db.MySQL.Exec(logQuery, appID, now)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "申请提交成功",
		"data": gin.H{
			"applicationId": applicationID,
			"status":        "pending",
		},
	})
}

// GetDonationProgress 查询申请进度
func GetDonationProgress(c *gin.Context) {
	// 这里的 id 是 application_id (业务编号，如 RELIEF20231027001)
	applicationID := c.Param("id")

	// 查询申请信息
	// 注意：表名变更，status_text 字段移除，只返回 status
	query := `
		SELECT a.id, a.project_id, p.name, a.status,
		       a.submit_time, a.updated_at
		FROM drug_relief_application a
		JOIN drug_relief_project p ON a.project_id = p.id
		WHERE a.id = ? -- 如果前端传的是主键ID
		-- 如果前端传的是业务编号 RELIEFxxx，则改为: WHERE a.application_id = ? 
	`

	// 修正：根据路由 /donations/:id/progress，通常 :id 指的是资源ID。
	// 在原代码中，它被当作 application_id (varchar) 处理。
	// 但在新表中，我们有自增 id 和 application_id (varchar)。
	// 为了兼容性和安全性，建议前端传递自增 ID 或者我们同时支持。
	// 这里假设前端传递的是数据库主键 ID (uint)，因为这样更高效且唯一。
	// 如果前端必须传递 RELIEFxxx，请将下面的 query 中的 a.id = ? 改为 a.application_id = ?

	// 尝试将 param 转为 uint，如果是数字则查主键，否则查业务号
	appIDUint, err := strconv.ParseUint(applicationID, 10, 64)

	var progress struct {
		ApplicationID uint64    `db:"id"` // 数据库主键
		ProjectID     uint      `db:"project_id"`
		ProjectName   string    `db:"name"`
		Status        string    `db:"status"`
		SubmitTime    time.Time `db:"submit_time"`
		UpdateTime    time.Time `db:"updated_at"`
	}

	if err == nil {
		// 是数字，查主键
		query = `
			SELECT a.id, a.project_id, p.name, a.status, a.submit_time, a.updated_at
			FROM drug_relief_application a
			JOIN drug_relief_project p ON a.project_id = p.id
			WHERE a.id = ?
		`
		err = db.MySQL.QueryRow(query, appIDUint).Scan(
			&progress.ApplicationID, &progress.ProjectID, &progress.ProjectName,
			&progress.Status, &progress.SubmitTime, &progress.UpdateTime,
		)
	} else {
		// 不是数字，查业务编号 application_id
		query = `
			SELECT a.id, a.project_id, p.name, a.status, a.submit_time, a.updated_at
			FROM drug_relief_application a
			JOIN drug_relief_project p ON a.project_id = p.id
			WHERE a.application_id = ?
		`
		// 临时变量接收
		var tempID uint64
		err = db.MySQL.QueryRow(query, applicationID).Scan(
			&tempID, &progress.ProjectID, &progress.ProjectName,
			&progress.Status, &progress.SubmitTime, &progress.UpdateTime,
		)
		progress.ApplicationID = tempID
	}

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "申请记录不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询进度失败",
		})
		return
	}

	// 查询进度日志
	// 注意：表名变更，desc 变为 action_desc
	logQuery := `
		SELECT status, action_desc, created_at
		FROM drug_relief_log
		WHERE application_id = ?
		ORDER BY created_at ASC
	`

	logRows, err := db.MySQL.Query(logQuery, progress.ApplicationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询日志失败",
		})
		return
	}
	defer logRows.Close()

	var logs []gin.H
	for logRows.Next() {
		var log struct {
			Status     string    `db:"status"`
			ActionDesc string    `db:"action_desc"`
			CreatedAt  time.Time `db:"created_at"`
		}
		if err := logRows.Scan(&log.Status, &log.ActionDesc, &log.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, gin.H{
			"time":   log.CreatedAt.Format("2006-01-02 15:04:05"),
			"status": log.Status,
			"desc":   log.ActionDesc,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"applicationId": progress.ApplicationID,
			"projectId":     progress.ProjectID,
			"projectName":   progress.ProjectName,
			"status":        progress.Status,
			"submitTime":    progress.SubmitTime.Format("2006-01-02 15:04:05"),
			"updateTime":    progress.UpdateTime.Format("2006-01-02 15:04:05"),
			"logs":          logs,
		},
	})
}

// DownloadDonationGuide 下载赠药指南
func DownloadDonationGuide(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目 ID",
		})
		return
	}

	// 查询指南文件路径
	// 注意：表名变更，audit_status 变更
	query := `SELECT apply_guide, name FROM drug_relief_project WHERE id = ? AND audit_status = 1`
	var guideURL, name string
	err = db.MySQL.QueryRow(query, projectID).Scan(&guideURL, &name)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "赠药项目不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询指南失败",
		})
		return
	}

	if guideURL == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "指南文件不存在",
		})
		return
	}

	// 如果 guide_url 是本地文件路径
	if strings.HasPrefix(guideURL, "/") || strings.HasPrefix(guideURL, "./") {
		filePath := guideURL
		if !strings.HasPrefix(filePath, "/") {
			filePath = "./" + filePath
		}

		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "指南文件不存在",
			})
			return
		}

		// 设置响应头
		filename := fmt.Sprintf("%s_申请指南.pdf", name)
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

		// 输出文件
		c.File(filePath)
		return
	}

	// 如果 guide_url 是远程 URL，重定向到该 URL
	c.Redirect(http.StatusTemporaryRedirect, guideURL)
}

// DonationOptionsResponse 赠药筛选选项响应
type DonationOptionsResponse struct {
	Diseases []OptionItem `json:"diseases"`
	Drugs    []OptionItem `json:"drugs"`
}

// GetDonationOptions 获取赠药项目筛选选项
func GetDonationOptions(c *gin.Context) {
	// 获取所有不重复的疾病分类
	// 注意：表名变更，字段 disease_value -> disease_id
	diseaseQuery := "SELECT DISTINCT disease_id FROM drug_relief_project WHERE audit_status = 1 AND disease_id != 0"
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
		var diseaseID int
		if err := diseaseRows.Scan(&diseaseID); err != nil {
			continue
		}
		diseases = append(diseases, OptionItem{
			Label: fmt.Sprintf("疾病分类%d", diseaseID),
			Value: strconv.Itoa(diseaseID),
		})
	}

	// 获取所有药品选项
	// 注意：表名 rare_drugs -> rare_drug，is_audit -> audit_status
	drugQuery := "SELECT id, generic_name, brand_name FROM rare_drug WHERE audit_status = 1"
	drugRows, err := db.MySQL.Query(drugQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询药品选项失败",
		})
		return
	}
	defer drugRows.Close()

	var drugs []OptionItem
	for drugRows.Next() {
		var drug struct {
			ID          uint           `db:"id"`
			GenericName string         `db:"generic_name"`
			BrandName   sql.NullString `db:"brand_name"`
		}
		if err := drugRows.Scan(&drug.ID, &drug.GenericName, &drug.BrandName); err != nil {
			continue
		}
		name := drug.GenericName
		if drug.BrandName.Valid && drug.BrandName.String != "" {
			name = drug.BrandName.String
		}
		drugs = append(drugs, OptionItem{
			Label: name,
			Value: strconv.FormatUint(uint64(drug.ID), 10),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": DonationOptionsResponse{
			Diseases: diseases,
			Drugs:    drugs,
		},
	})
}
