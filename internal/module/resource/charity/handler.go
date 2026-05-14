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

// ProjectListResponse 列表响应结构
type ProjectListResponse struct {
	List     []ProjectItem `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

// ProjectItem 救助项目项响应结构
type ProjectItem struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Organizer       string `json:"organizer"`
	ApplyCondition  string `json:"applyCondition"`
	Status          string `json:"status"`
	ReliefType      string `json:"reliefType"`
	DiseaseIds      []int  `json:"diseaseIds"`
	ReliefStandard  string `json:"reliefStandard"`
	ApplyDifficulty string `json:"applyDifficulty"`

	// 新增字段
	AuditStatus  int    `json:"auditStatus"`  // 审核状态：0待审核 1已通过 2已驳回
	RejectReason string `json:"rejectReason"` // 驳回原因
	Sort         int    `json:"sort"`         // 排序权重
	UpdatedAt    string `json:"updatedAt"`    // 更新时间，格式化为字符串
}

// ListProjects 获取救助项目列表
func ListProjects(c *gin.Context) {
	// 获取请求参数
	typeFilter := c.DefaultQuery("reliefType", "")
	diseaseStr := c.DefaultQuery("diseaseId", "")
	difficulty := c.DefaultQuery("applyDifficulty", "")

	// 【新增】获取 auditStatus 和 keyword 参数
	auditStatusStr := c.DefaultQuery("auditStatus", "")
	keyword := c.DefaultQuery("keyword", "")

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	diseaseID := 0
	if diseaseStr != "" {
		diseaseID, _ = strconv.Atoi(diseaseStr)
	}

	// 【新增】处理 auditStatus 参数
	var auditStatus int
	if auditStatusStr != "" {
		status, err := strconv.Atoi(auditStatusStr)
		if err == nil && (status == 0 || status == 1 || status == 2) {
			auditStatus = status
		} else {
			// 如果传递了非法状态值，可以选择返回错误或者忽略该筛选条件
			// 这里选择忽略，保持默认行为（或者你可以返回 400 错误）
			auditStatus = -1 // 标记为无效，后续不加入 WHERE 条件
		}
	} else {
		auditStatus = -1 // 未传递，不筛选
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
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 【修改】处理 audit_status 筛选
	// 如果前端传了 auditStatus，则按传入值筛选；否则，默认只展示已通过的 (audit_status = 1)
	// 注意：如果是管理后台接口，通常希望默认看到所有状态，或者由前端明确控制。
	// 这里保留原逻辑的“安全默认值”：如果没有指定 auditStatus，默认只看通过的。
	if auditStatus != -1 {
		whereClause += " AND p.audit_status = ?"
		args = append(args, auditStatus)
	} else {
		// 如果业务要求默认只看通过的，保留此行；如果希望默认看全部，注释掉此行
		whereClause += " AND p.audit_status = 1"
	}

	if typeFilter != "" {
		whereClause += " AND p.relief_type = ?"
		args = append(args, typeFilter)
	}

	if diseaseID != 0 {
		whereClause += " AND EXISTS (SELECT 1 FROM relief_project_disease_rel r WHERE r.project_id = p.id AND r.disease_id = ?)"
		args = append(args, diseaseID)
	}

	if difficulty != "" {
		whereClause += " AND p.apply_difficulty = ?"
		args = append(args, difficulty)
	}

	// 【新增】处理 keyword 关键字筛选 (匹配 name 和 organizer)
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		whereClause += " AND (p.name LIKE ? OR p.organizer LIKE ?)"
		args = append(args, likeKeyword, likeKeyword)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM relief_project p " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 查询列表基础信息
	listQuery := `
		SELECT p.id, p.name, p.organizer, p.apply_condition, p.audit_status, p.reject_reason, p.sort,
		       p.relief_type, p.relief_standard, p.apply_difficulty, p.updated_at
		FROM relief_project p
		` + whereClause + `
		ORDER BY p.sort DESC, p.id DESC
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
			ID              uint      `db:"id"`
			Name            string    `db:"name"`
			Organizer       string    `db:"organizer"`
			ApplyCondition  string    `db:"apply_condition"`
			AuditStatus     int       `db:"audit_status"`
			RejectReason    string    `db:"reject_reason"`
			Sort            int       `db:"sort"`
			ReliefType      string    `db:"relief_type"`
			ReliefStandard  string    `db:"relief_standard"`
			ApplyDifficulty string    `db:"apply_difficulty"`
			UpdatedAt       time.Time `db:"updated_at"` // 建议改为 time.Time 以便格式化
		}

		// 扫描数据
		if err := rows.Scan(
			&project.ID, &project.Name, &project.Organizer, &project.ApplyCondition,
			&project.AuditStatus, &project.RejectReason, &project.Sort,
			&project.ReliefType, &project.ReliefStandard, &project.ApplyDifficulty,
			&project.UpdatedAt,
		); err != nil {
			continue
		}

		// 转换状态显示文本
		statusText := "closed"
		if project.AuditStatus == 1 {
			statusText = "open"
		} else if project.AuditStatus == 0 {
			statusText = "pending"
		} else if project.AuditStatus == 2 {
			statusText = "rejected"
		}

		// 查询该项目关联的疾病ID列表
		var diseaseIds []int
		relRows, err := db.MySQL.Query("SELECT disease_id FROM relief_project_disease_rel WHERE project_id = ?", project.ID)
		if err == nil {
			for relRows.Next() {
				var did int
				if err := relRows.Scan(&did); err == nil {
					diseaseIds = append(diseaseIds, did)
				}
			}
			relRows.Close()
		}
		if diseaseIds == nil {
			diseaseIds = []int{}
		}

		// 格式化更新时间
		updatedAtStr := project.UpdatedAt.Format("2006-01-02T15:04:05Z")

		list = append(list, ProjectItem{
			ID:              project.ID,
			Name:            project.Name,
			Organizer:       project.Organizer,
			ApplyCondition:  project.ApplyCondition,
			Status:          statusText,
			ReliefType:      project.ReliefType,
			DiseaseIds:      diseaseIds,
			ReliefStandard:  project.ReliefStandard,
			ApplyDifficulty: project.ApplyDifficulty,
			AuditStatus:     project.AuditStatus,
			RejectReason:    project.RejectReason,
			Sort:            project.Sort,
			UpdatedAt:       updatedAtStr,
		})
	}

	if list == nil {
		list = []ProjectItem{}
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

// ProjectDiseaseItem 项目关联疾病项
type ProjectDiseaseItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ProjectDiseaseDetail 项目关联疾病详情结构
type ProjectDiseaseDetail struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
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

	// 1. 查询项目基础信息
	// 【修复】去掉了多余的 '}'
	query := `
		SELECT p.id, p.name, p.apply_process, p.apply_condition, p.apply_deadline,
		       p.contact_phone, p.contact_url, p.apply_form, p.apply_guide, p.material_list,
			   p.relief_type, p.relief_standard, p.apply_difficulty, p.organizer,
			   p.created_at, p.updated_at, p.sort, p.audit_status, p.reject_reason
		FROM relief_project p
		WHERE p.id = ?
	`

	var project struct {
		ID              uint       `db:"id"`
		Name            string     `db:"name"`
		ApplyProcess    string     `db:"apply_process"`
		ApplyCondition  string     `db:"apply_condition"`
		ApplyDeadline   *time.Time `db:"apply_deadline"`
		ContactPhone    string     `db:"contact_phone"`
		ContactUrl      string     `db:"contact_url"` // 注意：结构体字段名需与 JSON tag 或数据库列对应，这里保持 ContactUrl
		ApplyForm       string     `db:"apply_form"`
		ApplyGuide      string     `db:"apply_guide"`
		MaterialList    string     `db:"material_list"`
		ReliefType      string     `db:"relief_type"`
		ReliefStandard  string     `db:"relief_standard"`
		ApplyDifficulty string     `db:"apply_difficulty"`
		Organizer       string     `db:"organizer"`
		CreatedAt       time.Time  `db:"created_at"`
		UpdatedAt       time.Time  `db:"updated_at"`
		Sort            int        `db:"sort"`
		AuditStatus     int        `db:"audit_status"`
		RejectReason    string     `db:"reject_reason"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&project.ID, &project.Name, &project.ApplyProcess, &project.ApplyCondition,
		&project.ApplyDeadline, &project.ContactPhone, &project.ContactUrl,
		&project.ApplyForm, &project.ApplyGuide, &project.MaterialList,
		&project.ReliefType, &project.ReliefStandard, &project.ApplyDifficulty, &project.Organizer,
		&project.CreatedAt, &project.UpdatedAt, &project.Sort, &project.AuditStatus, &project.RejectReason,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "项目不存在",
			})
			return
		}
		// 【调试建议】如果依然报错，可以打印具体错误日志
		// log.Printf("GetProjectDetail Query Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询项目详情失败: " + err.Error(),
		})
		return
	}

	// 2. 查询该项目关联的疾病ID列表和详细信息
	diseaseQuery := `
		SELECT d.id, d.name, d.alias 
		FROM relief_project_disease_rel r
		INNER JOIN disease d ON r.disease_id = d.id
		WHERE r.project_id = ?
	`

	rows, err := db.MySQL.Query(diseaseQuery, id)
	if err != nil {
		// 记录错误但不中断主流程，疾病列表返回空
		// log.Printf("GetProjectDetail Disease Query Error: %v", err)
		rows = nil
	}

	var diseaseIds []int
	var diseases []ProjectDiseaseDetail

	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var d struct {
				ID    int    `db:"id"`
				Name  string `db:"name"`
				Alias string `db:"alias"`
			}
			if err := rows.Scan(&d.ID, &d.Name, &d.Alias); err == nil {
				diseaseIds = append(diseaseIds, d.ID)
				diseases = append(diseases, ProjectDiseaseDetail{
					ID:    d.ID,
					Name:  d.Name,
					Alias: d.Alias,
				})
			}
		}
	}

	// 确保切片不为 nil
	if diseaseIds == nil {
		diseaseIds = []int{}
	}
	if diseases == nil {
		diseases = []ProjectDiseaseDetail{}
	}

	// 3. 构造返回数据
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":              project.ID,
			"name":            project.Name,
			"applyProcess":    project.ApplyProcess,
			"applyCondition":  project.ApplyCondition,
			"applyDeadline":   project.ApplyDeadline,
			"contactPhone":    project.ContactPhone,
			"contactUrl":      project.ContactUrl, // 对应前端期望的 contactUrl
			"applyForm":       project.ApplyForm,
			"applyGuide":      project.ApplyGuide,
			"materialList":    project.MaterialList,
			"reliefType":      project.ReliefType,
			"reliefStandard":  project.ReliefStandard,
			"applyDifficulty": project.ApplyDifficulty,
			"organizer":       project.Organizer,
			"diseaseIds":      diseaseIds,
			"diseases":        diseases,
			"createdAt":       project.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updatedAt":       project.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			"sort":            project.Sort,
			"auditStatus":     project.AuditStatus,
			"rejectReason":    project.RejectReason,
		},
	})
}

// CreateProjectRequest 创建项目请求结构
type CreateProjectRequest struct {
	Name            string  `json:"name" binding:"required"`
	ApplyProcess    string  `json:"applyProcess" binding:"required"` // apply_process
	ApplyCondition  string  `json:"applyCondition"`                  // apply_condition
	ReliefType      string  `json:"reliefType" binding:"required"`
	DiseaseIDs      []int   `json:"diseaseIds"`
	ReliefStandard  string  `json:"reliefStandard"`
	ApplyDifficulty string  `json:"applyDifficulty"`
	ApplyDeadline   *string `json:"applyDeadline"`
	ContactPhone    string  `json:"contactPhone"`
	ContactURL      string  `json:"contactUrl"`
	ApplyForm       string  `json:"applyForm"`
	ApplyGuide      string  `json:"applyGuide"`
	MaterialList    string  `json:"materialList"`
	Organizer       string  `json:"organizer"`
	Sort            int     `json:"sort"`
}

// CreateProject 新增救助项目
func CreateProject(c *gin.Context) {
	var req CreateProjectRequest
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

	// 插入主表
	insertQuery := `
		INSERT INTO relief_project 
		(name, apply_process, apply_condition, relief_type,
		 relief_standard, apply_difficulty, apply_deadline, contact_phone, contact_url,
		 apply_form, apply_guide, material_list, organizer, sort,
		 audit_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW())
	`

	var deadline interface{}
	if req.ApplyDeadline != nil && *req.ApplyDeadline != "" {
		deadline = *req.ApplyDeadline
	} else {
		deadline = nil
	}

	result, err := tx.Exec(insertQuery,
		req.Name, req.ApplyProcess, req.ApplyCondition, req.ReliefType,
		req.ReliefStandard, req.ApplyDifficulty, deadline, req.ContactPhone, req.ContactURL,
		req.ApplyForm, req.ApplyGuide, req.MaterialList, req.Organizer, req.Sort)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建项目失败",
		})
		return
	}

	projectID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取项目ID失败",
		})
		return
	}

	// 插入疾病关联
	if len(req.DiseaseIDs) > 0 {
		relQuery := "INSERT INTO relief_project_disease_rel (project_id, disease_id) VALUES (?, ?)"
		for _, did := range req.DiseaseIDs {
			if did > 0 {
				_, err := tx.Exec(relQuery, projectID, did)
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
			"id": projectID,
		},
	})
}

// UpdateProjectRequest 更新项目请求结构
type UpdateProjectRequest struct {
	Name            string  `json:"name"`
	ApplyProcess    string  `json:"applyProcess"`
	ApplyCondition  string  `json:"applyCondition"`
	ReliefType      string  `json:"reliefType"` // 标准字段
	Type            string  `json:"type"`       // 【新增】兼容前端可能传的 type 字段
	DiseaseIDs      []int   `json:"diseaseIds"`
	ReliefStandard  string  `json:"reliefStandard"`
	ApplyDifficulty string  `json:"applyDifficulty"`
	ApplyDeadline   *string `json:"applyDeadline"`
	ContactPhone    string  `json:"contactPhone"`
	ContactURL      string  `json:"contactUrl"`
	ApplyForm       string  `json:"applyForm"`
	ApplyGuide      string  `json:"applyGuide"`
	MaterialList    string  `json:"materialList"`
	Organizer       string  `json:"organizer"`
	Sort            *int    `json:"sort"`
	AuditStatus     *int    `json:"auditStatus"`
	RejectReason    *string `json:"rejectReason"` // 【新增】驳回原因
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
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 【兼容处理】如果前端传了 type 但没传 reliefType，使用 type 的值
	if req.ReliefType == "" && req.Type != "" {
		req.ReliefType = req.Type
	}

	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务启动失败",
		})
		return
	}
	defer tx.Rollback()

	// 构建动态更新语句
	updateFields := []string{}
	args := []interface{}{}

	if req.Name != "" {
		updateFields = append(updateFields, "name = ?")
		args = append(args, req.Name)
	}
	if req.ApplyProcess != "" {
		updateFields = append(updateFields, "apply_process = ?")
		args = append(args, req.ApplyProcess)
	}
	if req.ApplyCondition != "" {
		updateFields = append(updateFields, "apply_condition = ?")
		args = append(args, req.ApplyCondition)
	}
	if req.ReliefType != "" {
		updateFields = append(updateFields, "relief_type = ?")
		args = append(args, req.ReliefType)
	}
	if req.ReliefStandard != "" {
		updateFields = append(updateFields, "relief_standard = ?")
		args = append(args, req.ReliefStandard)
	}
	if req.ApplyDifficulty != "" {
		updateFields = append(updateFields, "apply_difficulty = ?")
		args = append(args, req.ApplyDifficulty)
	}

	// 【修复】处理 apply_deadline 格式问题
	// 数据库字段类型为 DATE，需要确保传入的值格式为 YYYY-MM-DD
	if req.ApplyDeadline != nil {
		deadlineStr := *req.ApplyDeadline
		// 如果字符串包含 'T'，说明是 ISO 8601 格式，截取日期部分
		if strings.Contains(deadlineStr, "T") {
			parts := strings.Split(deadlineStr, "T")
			deadlineStr = parts[0]
		}
		// 可选：进一步验证格式是否为 YYYY-MM-DD，这里简单处理直接传入
		updateFields = append(updateFields, "apply_deadline = ?")
		args = append(args, deadlineStr)
	}

	if req.ContactPhone != "" {
		updateFields = append(updateFields, "contact_phone = ?")
		args = append(args, req.ContactPhone)
	}
	if req.ContactURL != "" {
		updateFields = append(updateFields, "contact_url = ?")
		args = append(args, req.ContactURL)
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
	if req.Sort != nil {
		updateFields = append(updateFields, "sort = ?")
		args = append(args, *req.Sort)
	}
	if req.AuditStatus != nil {
		updateFields = append(updateFields, "audit_status = ?")
		args = append(args, *req.AuditStatus)
	}
	// 【新增】处理 RejectReason 更新
	if req.RejectReason != nil {
		updateFields = append(updateFields, "reject_reason = ?")
		args = append(args, *req.RejectReason)
	}

	if len(updateFields) > 0 {
		updateFields = append(updateFields, "updated_at = NOW()")
		args = append(args, id)

		updateQuery := `UPDATE relief_project SET ` + strings.Join(updateFields, ", ") + ` WHERE id = ?`
		_, err = tx.Exec(updateQuery, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新项目失败: " + err.Error(),
			})
			return
		}
	}

	// 处理疾病关联更新
	if req.DiseaseIDs != nil {
		// 1. 删除旧关联
		_, err := tx.Exec("DELETE FROM relief_project_disease_rel WHERE project_id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "清理旧疾病关联失败",
			})
			return
		}

		// 2. 插入新关联
		if len(req.DiseaseIDs) > 0 {
			relQuery := "INSERT INTO relief_project_disease_rel (project_id, disease_id) VALUES (?, ?)"
			for _, did := range req.DiseaseIDs {
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

// DeleteProject 删除救助项目（真删除/物理删除）
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

	// 开启事务，确保主表和关联表数据同时删除
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务启动失败",
		})
		return
	}
	defer tx.Rollback()

	// 1. 先删除关联表中的数据 (relief_project_disease_rel)
	// 注意：如果外键设置了 ON DELETE CASCADE，这一步可以省略，但显式删除更安全
	_, err = tx.Exec("DELETE FROM relief_project_disease_rel WHERE project_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "清理项目疾病关联失败",
		})
		return
	}

	// 2. 删除主表数据 (relief_project)
	deleteQuery := `DELETE FROM relief_project WHERE id = ?`
	result, err := tx.Exec(deleteQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除项目失败: " + err.Error(),
		})
		return
	}

	// 检查是否真的删除了记录
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取删除结果失败",
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "项目不存在",
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

// ChannelItem 求助渠道项 (修改后以匹配期望返回值)
type ChannelItem struct {
	ID                   uint    `json:"id"`
	ChannelType          string  `json:"channelType"`
	Name                 string  `json:"name"`
	ApplyCondition       string  `json:"applyCondition"`
	ResponseTime         string  `json:"responseTime"`
	ContactPhone         string  `json:"contactPhone"`
	ContactUrl           string  `json:"contactUrl"`
	HelpLetterTemplate   string  `json:"helpLetterTemplate"`
	CrowdfundingTemplate string  `json:"crowdfundingTemplate"`
	AuditStatus          int     `json:"auditStatus"`
	RejectReason         *string `json:"rejectReason"` // 使用指针以支持 null
	Sort                 int     `json:"sort"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
}

// ChannelListResponse 渠道列表响应结构
type ChannelListResponse struct {
	List     []ChannelItem `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`     // 如果不需要分页信息可注释掉
	PageSize int           `json:"pageSize"` // 如果不需要分页信息可注释掉
}

// ChannelResponse 渠道响应结构
type ChannelResponse struct {
	Channels []ChannelItem `json:"channels"`
}

// convertChannelType 转换渠道类型
func convertChannelType(channelType string) string {
	typeMap := map[string]string{
		"emergency_help":     "urgent",       // 紧急求助
		"crowdfunding":       "crowdfunding", // 众筹求助
		"charity_consulting": "consulting",   // 公益咨询 (新增映射)
		"founding_support":   "foundation",   // 基金会支持 (新增映射，注意拼写是否为 foundation_support)

		// 兼容旧的中文映射，如果数据库中存的是中文
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

// GetChannels 获取求助渠道列表
func GetChannels(c *gin.Context) {
	// 1. 获取请求参数
	channelType := c.DefaultQuery("channelType", "")
	keyword := c.DefaultQuery("keyword", "")

	// 【新增】获取 auditStatus 参数
	auditStatusStr := c.DefaultQuery("auditStatus", "")

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	// 2. 处理分页参数
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 3. 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 【新增】处理 auditStatus 筛选逻辑
	if auditStatusStr != "" {
		// 如果前端传了 auditStatus，尝试解析并加入筛选
		status, err := strconv.Atoi(auditStatusStr)
		if err == nil && (status == 0 || status == 1 || status == 2) {
			whereClause += " AND audit_status = ?"
			args = append(args, status)
		}
		// 如果传入非法值，可以选择忽略该筛选条件（即查询所有状态），或者返回错误。
		// 这里选择忽略非法值，不加入 WHERE 条件，相当于查询所有状态。
	} else {
		// 【重要】如果前端没传 auditStatus，默认只展示已通过的 (audit_status = 1)
		// 如果业务需求是默认展示所有，请注释掉下面这一行
		whereClause += " AND audit_status = 1"
	}

	if channelType != "" {
		whereClause += " AND channel_type = ?"
		args = append(args, channelType)
	}

	if keyword != "" {
		whereClause += " AND name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	// 4. 查询总数
	countQuery := "SELECT COUNT(*) FROM help_channel " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 5. 查询列表数据 (查询所有必要字段)
	listQuery := `
		SELECT id, channel_type, name, apply_condition, response_time, 
		       contact_phone, contact_url, help_letter_template, crowdfunding_template,
		       audit_status, reject_reason, sort, created_at, updated_at
		FROM help_channel 
		` + whereClause + `
		ORDER BY sort DESC, id ASC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
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
			ID                   uint           `db:"id"`
			ChannelType          string         `db:"channel_type"`
			Name                 string         `db:"name"`
			ApplyCondition       string         `db:"apply_condition"`
			ResponseTime         string         `db:"response_time"`
			ContactPhone         string         `db:"contact_phone"`
			ContactUrl           string         `db:"contact_url"`
			HelpLetterTemplate   string         `db:"help_letter_template"`
			CrowdfundingTemplate string         `db:"crowdfunding_template"`
			AuditStatus          int            `db:"audit_status"`
			RejectReason         sql.NullString `db:"reject_reason"`
			Sort                 int            `db:"sort"`
			CreatedAt            time.Time      `db:"created_at"`
			UpdatedAt            time.Time      `db:"updated_at"`
		}

		if err := rows.Scan(
			&channel.ID, &channel.ChannelType, &channel.Name,
			&channel.ApplyCondition, &channel.ResponseTime, &channel.ContactPhone, &channel.ContactUrl,
			&channel.HelpLetterTemplate, &channel.CrowdfundingTemplate,
			&channel.AuditStatus, &channel.RejectReason, &channel.Sort,
			&channel.CreatedAt, &channel.UpdatedAt,
		); err != nil {
			continue
		}

		// 处理 RejectReason null 值
		var rejectReasonPtr *string
		if channel.RejectReason.Valid {
			rejectReasonPtr = &channel.RejectReason.String
		}

		channels = append(channels, ChannelItem{
			ID:                   channel.ID,
			ChannelType:          channel.ChannelType,
			Name:                 channel.Name,
			ApplyCondition:       channel.ApplyCondition,
			ResponseTime:         channel.ResponseTime,
			ContactPhone:         channel.ContactPhone,
			ContactUrl:           channel.ContactUrl,
			HelpLetterTemplate:   channel.HelpLetterTemplate,
			CrowdfundingTemplate: channel.CrowdfundingTemplate,
			AuditStatus:          channel.AuditStatus,
			RejectReason:         rejectReasonPtr,
			Sort:                 channel.Sort,
			CreatedAt:            channel.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:            channel.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// 确保数组不为 null
	if channels == nil {
		channels = []ChannelItem{}
	}

	// 6. 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":  channels,
			"total": total,
		},
	})
}

// GetChannelDetail 获取求助渠道详情
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

	// 【修改】查询渠道详情，增加 audit_status 和 reject_reason
	query := `
		SELECT id, channel_type, name, apply_condition, response_time, 
		       contact_phone, contact_url, help_letter_template, crowdfunding_template,
		       audit_status, reject_reason, sort, created_at, updated_at
		FROM help_channel
		WHERE id = ?
	`
	// 注意：这里去掉了 AND audit_status = 1，以便管理员能查看未通过的详情。
	// 如果业务要求只有已通过的才能看详情，请加回该条件。

	var channel struct {
		ID                   uint           `db:"id"`
		ChannelType          string         `db:"channel_type"`
		Name                 string         `db:"name"`
		ApplyCondition       string         `db:"apply_condition"`
		ResponseTime         string         `db:"response_time"`
		ContactPhone         string         `db:"contact_phone"`
		ContactUrl           string         `db:"contact_url"`
		HelpLetterTemplate   string         `db:"help_letter_template"`
		CrowdfundingTemplate string         `db:"crowdfunding_template"`
		AuditStatus          int            `db:"audit_status"`
		RejectReason         sql.NullString `db:"reject_reason"`
		Sort                 int            `db:"sort"`
		CreatedAt            time.Time      `db:"created_at"`
		UpdatedAt            time.Time      `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&channel.ID, &channel.ChannelType, &channel.Name,
		&channel.ApplyCondition, &channel.ResponseTime, &channel.ContactPhone, &channel.ContactUrl,
		&channel.HelpLetterTemplate, &channel.CrowdfundingTemplate,
		&channel.AuditStatus, &channel.RejectReason, &channel.Sort,
		&channel.CreatedAt, &channel.UpdatedAt,
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

	// 处理 RejectReason null 值
	var rejectReasonPtr *string
	if channel.RejectReason.Valid {
		rejectReasonPtr = &channel.RejectReason.String
	}

	// 【修改】构造返回数据，直接使用 ChannelItem 结构体或匿名结构体以匹配期望格式
	// 这里我们构造一个符合期望 JSON 结构的匿名结构体或直接赋值给 ChannelItem
	data := ChannelItem{
		ID:                   channel.ID,
		ChannelType:          channel.ChannelType, // 直接返回原始类型，不再转换
		Name:                 channel.Name,
		ApplyCondition:       channel.ApplyCondition,
		ResponseTime:         channel.ResponseTime,
		ContactPhone:         channel.ContactPhone,
		ContactUrl:           channel.ContactUrl,
		HelpLetterTemplate:   channel.HelpLetterTemplate,
		CrowdfundingTemplate: channel.CrowdfundingTemplate,
		AuditStatus:          channel.AuditStatus,
		RejectReason:         rejectReasonPtr,
		Sort:                 channel.Sort,
		CreatedAt:            channel.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:            channel.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}

// CaseItem 案例项响应结构
type CaseItem struct {
	ID               uint    `json:"id"`
	DiseaseID        int64   `json:"diseaseId"`
	DiseaseName      string  `json:"diseaseName"`
	ProjectID        uint    `json:"projectId"`
	ProjectName      string  `json:"projectName"`
	CaseTitle        string  `json:"caseTitle"`
	PatientDesc      string  `json:"patientDesc"`
	ApplyCycle       string  `json:"applyCycle"`
	ActualRelief     string  `json:"actualRelief"`
	Experience       string  `json:"experience"`
	PitfallGuide     string  `json:"pitfallGuide"`
	CasePdf          *string `json:"casePdf"`          // 使用指针以支持 null
	MaterialTemplate *string `json:"materialTemplate"` // 使用指针以支持 null
	AuditStatus      int     `json:"auditStatus"`
	RejectReason     *string `json:"rejectReason"` // 使用指针以支持 null
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}

// CaseListResponse 案例列表响应结构
type CaseListResponse struct {
	List     []CaseItem `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

// CaseDetailResponse 案例详情响应结构
// CaseDetailResponse 案例详情响应结构
type CaseDetailResponse struct {
	ID               uint    `json:"id"`
	DiseaseID        int64   `json:"diseaseId"`
	DiseaseName      string  `json:"diseaseName"` // 新增：直接从数据库获取或关联查询
	ProjectID        uint    `json:"projectId"`
	ProjectName      string  `json:"projectName"`
	CaseTitle        string  `json:"caseTitle"`
	PatientDesc      string  `json:"patientDesc"`
	ApplyCycle       string  `json:"applyCycle"`
	ActualRelief     string  `json:"actualRelief"`
	Experience       string  `json:"experience"`
	PitfallGuide     string  `json:"pitfallGuide"`
	CasePdf          *string `json:"casePdf"`          // 修改为指针以支持 null
	MaterialTemplate *string `json:"materialTemplate"` // 修改为指针以支持 null
	AuditStatus      int     `json:"auditStatus"`
	RejectReason     *string `json:"rejectReason"` // 修改为指针以支持 null
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`

	// 保留部分原有字段以兼容旧前端，如果不需要可以删除
	// Content      string `json:"content"`
	// Summary      string `json:"summary"`
}

// GetCases 获取救助案例列表
func GetCases(c *gin.Context) {
	// 1. 获取请求参数
	diseaseStr := c.DefaultQuery("diseaseId", "")
	keyword := c.DefaultQuery("keyword", "")

	// 【关键修改】获取 auditStatus 参数
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

	// 2. 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 【关键修改】处理 auditStatus 筛选
	if auditStatusStr != "" {
		auditStatus, err := strconv.Atoi(auditStatusStr)
		// 验证状态值是否合法 (0:待审核, 1:已通过, 2:已驳回)
		if err == nil && (auditStatus == 0 || auditStatus == 1 || auditStatus == 2) {
			whereClause += " AND rc.audit_status = ?"
			args = append(args, auditStatus)
		}
		// 如果传递了非法值，忽略该筛选条件，查询所有状态
	}

	// 其他筛选条件
	if diseaseStr != "" {
		whereClause += " AND rc.disease_id = ?"
		args = append(args, diseaseStr)
	}
	if keyword != "" {
		whereClause += " AND (rc.case_title LIKE ? OR rc.patient_desc LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 3. 查询总数
	countQuery := "SELECT COUNT(*) FROM relief_case rc " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 4. 查询列表
	listQuery := `
		SELECT rc.id, rc.disease_id, d.name as disease_name, 
		       rc.project_id, rp.name as project_name,
		       rc.case_title, rc.patient_desc, rc.apply_cycle, 
		       rc.actual_relief, rc.experience, rc.pitfall_guide, 
		       rc.case_pdf, rc.material_template,
		       rc.audit_status, rc.reject_reason,
		       rc.created_at, rc.updated_at
		FROM relief_case rc
		LEFT JOIN relief_project rp ON rc.project_id = rp.id
		LEFT JOIN disease d ON rc.disease_id = d.id
		` + whereClause + `
		ORDER BY rc.created_at DESC
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
			ID               uint           `db:"id"`
			DiseaseID        int64          `db:"disease_id"`
			DiseaseName      string         `db:"disease_name"`
			ProjectID        uint           `db:"project_id"`
			ProjectName      string         `db:"project_name"`
			CaseTitle        string         `db:"case_title"`
			PatientDesc      string         `db:"patient_desc"`
			ApplyCycle       string         `db:"apply_cycle"`
			ActualRelief     string         `db:"actual_relief"`
			Experience       string         `db:"experience"`
			PitfallGuide     string         `db:"pitfall_guide"`
			CasePdf          sql.NullString `db:"case_pdf"`
			MaterialTemplate sql.NullString `db:"material_template"`
			AuditStatus      int            `db:"audit_status"`
			RejectReason     sql.NullString `db:"reject_reason"`
			CreatedAt        time.Time      `db:"created_at"`
			UpdatedAt        time.Time      `db:"updated_at"`
		}

		if err := rows.Scan(
			&caseItem.ID, &caseItem.DiseaseID, &caseItem.DiseaseName,
			&caseItem.ProjectID, &caseItem.ProjectName,
			&caseItem.CaseTitle, &caseItem.PatientDesc, &caseItem.ApplyCycle,
			&caseItem.ActualRelief, &caseItem.Experience, &caseItem.PitfallGuide,
			&caseItem.CasePdf, &caseItem.MaterialTemplate,
			&caseItem.AuditStatus, &caseItem.RejectReason,
			&caseItem.CreatedAt, &caseItem.UpdatedAt,
		); err != nil {
			continue
		}

		// 处理 NullString 转为 *string
		var casePdfPtr *string
		if caseItem.CasePdf.Valid {
			casePdfPtr = &caseItem.CasePdf.String
		}

		var materialTemplatePtr *string
		if caseItem.MaterialTemplate.Valid {
			materialTemplatePtr = &caseItem.MaterialTemplate.String
		}

		var rejectReasonPtr *string
		if caseItem.RejectReason.Valid {
			rejectReasonPtr = &caseItem.RejectReason.String
		}

		// 格式化时间
		createdAtStr := caseItem.CreatedAt.Format("2006-01-02 15:04:05")
		updatedAtStr := caseItem.UpdatedAt.Format("2006-01-02 15:04:05")

		list = append(list, CaseItem{
			ID:               caseItem.ID,
			DiseaseID:        caseItem.DiseaseID,
			DiseaseName:      caseItem.DiseaseName,
			ProjectID:        caseItem.ProjectID,
			ProjectName:      caseItem.ProjectName,
			CaseTitle:        caseItem.CaseTitle,
			PatientDesc:      caseItem.PatientDesc,
			ApplyCycle:       caseItem.ApplyCycle,
			ActualRelief:     caseItem.ActualRelief,
			Experience:       caseItem.Experience,
			PitfallGuide:     caseItem.PitfallGuide,
			CasePdf:          casePdfPtr,
			MaterialTemplate: materialTemplatePtr,
			AuditStatus:      caseItem.AuditStatus,
			RejectReason:     rejectReasonPtr,
			CreatedAt:        createdAtStr,
			UpdatedAt:        updatedAtStr,
		})
	}

	if list == nil {
		list = []CaseItem{}
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

// GetCaseDetail 获取案例详情
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

	// 修改 SQL：关联 disease 表获取疾病名称，增加 audit_status, reject_reason, updated_at
	query := `
		SELECT rc.id, rc.disease_id, d.name as disease_name,
		       rc.project_id, rp.name as project_name, 
			   rc.case_title, rc.patient_desc, rc.apply_cycle, 
		       rc.actual_relief, rc.experience, rc.pitfall_guide, 
		       rc.case_pdf, rc.material_template,
		       rc.audit_status, rc.reject_reason,
		       rc.created_at, rc.updated_at
		FROM relief_case rc
		LEFT JOIN relief_project rp ON rc.project_id = rp.id
		LEFT JOIN disease d ON rc.disease_id = d.id
		WHERE rc.id = ?
	`
	// 注意：这里去掉了 AND rc.audit_status = 1，以便管理员或特定场景下能查看未通过的案例详情。
	// 如果业务要求只有已通过的才能看详情，请加回该条件。

	var caseItem struct {
		ID               uint           `db:"id"`
		DiseaseID        int64          `db:"disease_id"`
		DiseaseName      string         `db:"disease_name"`
		ProjectID        uint           `db:"project_id"`
		ProjectName      string         `db:"project_name"`
		CaseTitle        string         `db:"case_title"`
		PatientDesc      string         `db:"patient_desc"`
		ApplyCycle       string         `db:"apply_cycle"`
		ActualRelief     string         `db:"actual_relief"`
		Experience       string         `db:"experience"`
		PitfallGuide     string         `db:"pitfall_guide"`
		CasePdf          sql.NullString `db:"case_pdf"`
		MaterialTemplate sql.NullString `db:"material_template"`
		AuditStatus      int            `db:"audit_status"`
		RejectReason     sql.NullString `db:"reject_reason"`
		CreatedAt        time.Time      `db:"created_at"`
		UpdatedAt        time.Time      `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&caseItem.ID, &caseItem.DiseaseID, &caseItem.DiseaseName,
		&caseItem.ProjectID, &caseItem.ProjectName,
		&caseItem.CaseTitle, &caseItem.PatientDesc, &caseItem.ApplyCycle,
		&caseItem.ActualRelief, &caseItem.Experience, &caseItem.PitfallGuide,
		&caseItem.CasePdf, &caseItem.MaterialTemplate,
		&caseItem.AuditStatus, &caseItem.RejectReason,
		&caseItem.CreatedAt, &caseItem.UpdatedAt,
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

	// 处理 NullString 转为 *string
	var casePdfPtr *string
	if caseItem.CasePdf.Valid {
		casePdfPtr = &caseItem.CasePdf.String
	}

	var materialTemplatePtr *string
	if caseItem.MaterialTemplate.Valid {
		materialTemplatePtr = &caseItem.MaterialTemplate.String
	}

	var rejectReasonPtr *string
	if caseItem.RejectReason.Valid {
		rejectReasonPtr = &caseItem.RejectReason.String
	}

	// 格式化时间
	createdAtStr := caseItem.CreatedAt.Format("2006-01-02 15:04:05")
	updatedAtStr := caseItem.UpdatedAt.Format("2006-01-02 15:04:05")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": CaseDetailResponse{
			ID:               caseItem.ID,
			DiseaseID:        caseItem.DiseaseID,
			DiseaseName:      caseItem.DiseaseName,
			ProjectID:        caseItem.ProjectID,
			ProjectName:      caseItem.ProjectName,
			CaseTitle:        caseItem.CaseTitle,
			PatientDesc:      caseItem.PatientDesc,
			ApplyCycle:       caseItem.ApplyCycle,
			ActualRelief:     caseItem.ActualRelief,
			Experience:       caseItem.Experience,
			PitfallGuide:     caseItem.PitfallGuide,
			CasePdf:          casePdfPtr,
			MaterialTemplate: materialTemplatePtr,
			AuditStatus:      caseItem.AuditStatus,
			RejectReason:     rejectReasonPtr,
			CreatedAt:        createdAtStr,
			UpdatedAt:        updatedAtStr,
		},
	})
}

// getDiseaseInfo 获取疾病信息（名称和代码）
// 保持原有实现，但确保 map 的 key 是字符串形式的 ID
func getDiseaseInfo(diseaseValue string) (string, string) {
	diseaseMap := map[string]struct {
		Name string
		Code string
	}{
		"1": {Name: "渐冻症", Code: "als"},
		"2": {Name: "血友病", Code: "hemophilia"},
		// ... 确保这里的 key 与数据库中的 disease_id 对应
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
		FROM relief_case 
		WHERE id = ? AND audit_status = 1
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
	fileSize := getFileSize(casePDF.CasePDF)
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
	// 从数据库动态查询
	// 注意：表名 relief_cases -> relief_case, 字段 disease_value -> disease_id
	query := `
		SELECT DISTINCT rc.disease_id, do.name, do.code
		FROM relief_case rc
		INNER JOIN disease_options do ON rc.disease_id = do.value
		WHERE rc.audit_status = 1
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
			DiseaseID int64  `db:"disease_id"`
			Name      string `db:"name"`
			Code      string `db:"code"`
		}
		if err := rows.Scan(&disease.DiseaseID, &disease.Name, &disease.Code); err != nil {
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

// CreateCaseRequest 创建案例请求结构
type CreateCaseRequest struct {
	DiseaseID        int64  `json:"diseaseId" binding:"required"`
	ProjectID        uint   `json:"projectId" binding:"required"`
	CaseTitle        string `json:"caseTitle" binding:"required"`
	PatientDesc      string `json:"patientDesc" binding:"required"`
	ApplyCycle       string `json:"applyCycle" binding:"required"`
	ActualRelief     string `json:"actualRelief" binding:"required"`
	Experience       string `json:"experience" binding:"required"`
	PitfallGuide     string `json:"pitfallGuide" binding:"required"`
	CasePdf          string `json:"casePdf"`
	MaterialTemplate string `json:"materialTemplate"`
	AuditStatus      *int   `json:"auditStatus"`  // 审核状态：0待审核 1已通过 2已驳回，不传默认0
	RejectReason     string `json:"rejectReason"` // 驳回原因
}

// CreateCase 新增救助案例
func CreateCase(c *gin.Context) {
	var req CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 参数校验
	if req.DiseaseID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "疾病ID不能为空",
		})
		return
	}
	if req.ProjectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "关联项目ID不能为空",
		})
		return
	}
	if req.CaseTitle == "" {
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
	if req.ApplyCycle == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "申请周期不能为空",
		})
		return
	}
	if req.ActualRelief == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "实际救助金额不能为空",
		})
		return
	}
	if req.Experience == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "申请经验分享不能为空",
		})
		return
	}
	if req.PitfallGuide == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "申请避坑要点不能为空",
		})
		return
	}

	// 处理审核状态，默认0（待审核）
	auditStatus := 0
	if req.AuditStatus != nil {
		// 验证审核状态值是否合法
		if *req.AuditStatus >= 0 && *req.AuditStatus <= 2 {
			auditStatus = *req.AuditStatus
		}
	}

	// 插入数据库
	insertQuery := `
		INSERT INTO relief_case 
		(disease_id, project_id, case_title, patient_desc, apply_cycle, actual_relief, experience,
		 pitfall_guide, case_pdf, material_template, audit_status, reject_reason,
		 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := db.MySQL.Exec(insertQuery,
		req.DiseaseID, req.ProjectID, req.CaseTitle, req.PatientDesc, req.ApplyCycle, req.ActualRelief, req.Experience,
		req.PitfallGuide, req.CasePdf, req.MaterialTemplate, auditStatus, req.RejectReason)

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
	DiseaseID        *int64 `json:"diseaseId"` // 指针以便判断是否传递
	ActualRelief     string `json:"actualRelief"`
	Experience       string `json:"experience"`
	PitfallGuide     string `json:"pitfallGuide"`
	ApplyCycle       string `json:"applyCycle"`
	CasePdf          string `json:"casePdf"`
	MaterialTemplate string `json:"materialTemplate"`
	ProjectID        *uint  `json:"projectId"` // 指针
	AuditStatus      *int   `json:"auditStatus"`
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
	if req.DiseaseID != nil {
		updateFields = append(updateFields, "disease_id = ?")
		args = append(args, *req.DiseaseID)
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
	if req.ApplyCycle != "" {
		updateFields = append(updateFields, "apply_cycle = ?")
		args = append(args, req.ApplyCycle)
	}
	if req.CasePdf != "" {
		updateFields = append(updateFields, "case_pdf = ?")
		args = append(args, req.CasePdf)
	}
	if req.MaterialTemplate != "" {
		updateFields = append(updateFields, "material_template = ?")
		args = append(args, req.MaterialTemplate)
	}
	if req.ProjectID != nil {
		updateFields = append(updateFields, "project_id = ?")
		args = append(args, *req.ProjectID)
	}
	if req.AuditStatus != nil {
		updateFields = append(updateFields, "audit_status = ?")
		args = append(args, *req.AuditStatus)
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "未提供更新字段",
		})
		return
	}

	updateFields = append(updateFields, "updated_at = NOW()")
	args = append(args, id)

	updateQuery := `UPDATE relief_case SET ` + strings.Join(updateFields, ", ") + ` WHERE id = ?`
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
// DeleteCase 删除救助案例（真删除/物理删除）
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

	// 【修改点】使用 DELETE 语句进行物理删除
	// 注意：如果 relief_case 表有其他外键关联（如评论、点赞等），可能需要先清理关联数据或设置级联删除
	deleteQuery := `DELETE FROM relief_case WHERE id = ?`
	result, err := db.MySQL.Exec(deleteQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除案例失败",
		})
		return
	}

	// 检查是否真的删除了记录（防止ID不存在时返回成功）
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取删除结果失败",
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "案例不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    nil,
	})
}

// ProjectOptionItem 项目选项项（简化结构）
type ProjectOptionItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// ProjectOptionsResponse 项目选项响应结构
type ProjectOptionsResponse struct {
	List []ProjectOptionItem `json:"list"`
}

// GetProjectOptions 获取项目选项列表（用于下拉选择）
func GetProjectOptions(c *gin.Context) {
	// 获取请求参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")
	keyword := c.DefaultQuery("keyword", "")
	auditStatusStr := c.DefaultQuery("auditStatus", "")

	// 处理分页参数
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
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 处理 auditStatus 筛选
	if auditStatusStr != "" {
		auditStatus, err := strconv.Atoi(auditStatusStr)
		if err == nil && (auditStatus == 0 || auditStatus == 1 || auditStatus == 2) {
			whereClause += " AND audit_status = ?"
			args = append(args, auditStatus)
		}
	}

	// 处理 keyword 关键字筛选（匹配项目名称）
	if keyword != "" {
		whereClause += " AND name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	// 查询列表
	listQuery := `
        SELECT id, name
        FROM relief_project
        ` + whereClause + `
        ORDER BY sort DESC, id DESC
        LIMIT ? OFFSET ?
    `
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询项目选项失败",
		})
		return
	}
	defer rows.Close()

	var list []ProjectOptionItem
	for rows.Next() {
		var item ProjectOptionItem
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			continue
		}
		list = append(list, item)
	}

	// 确保数组不为 null
	if list == nil {
		list = []ProjectOptionItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": ProjectOptionsResponse{
			List: list,
		},
	})
}

// CreateChannelRequest 创建渠道请求结构
type CreateChannelRequest struct {
	ChannelType          string `json:"channelType" binding:"required"`
	Name                 string `json:"name" binding:"required"`
	ApplyCondition       string `json:"applyCondition" binding:"required"`
	ResponseTime         string `json:"responseTime" binding:"required"`
	ContactPhone         string `json:"contactPhone"`
	ContactUrl           string `json:"contactUrl"`
	HelpLetterTemplate   string `json:"helpLetterTemplate"`
	CrowdfundingTemplate string `json:"crowdfundingTemplate"`
	Sort                 int    `json:"sort"`
	AuditStatus          *int   `json:"auditStatus"` // 可选，默认0
}

// UpdateChannelRequest 更新渠道请求结构
type UpdateChannelRequest struct {
	ChannelType          string  `json:"channelType"`
	Name                 string  `json:"name"`
	ApplyCondition       string  `json:"applyCondition"`
	ResponseTime         string  `json:"responseTime"`
	ContactPhone         string  `json:"contactPhone"`
	ContactUrl           string  `json:"contactUrl"`
	HelpLetterTemplate   string  `json:"helpLetterTemplate"`
	CrowdfundingTemplate string  `json:"crowdfundingTemplate"`
	Sort                 *int    `json:"sort"`
	AuditStatus          *int    `json:"auditStatus"`
	RejectReason         *string `json:"rejectReason"`
}

// CreateChannel 新增求助渠道
func CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 处理默认审核状态
	auditStatus := 0
	if req.AuditStatus != nil {
		if *req.AuditStatus >= 0 && *req.AuditStatus <= 2 {
			auditStatus = *req.AuditStatus
		}
	}

	insertQuery := `
		INSERT INTO help_channel 
		(channel_type, name, apply_condition, response_time, contact_phone, contact_url,
		 help_letter_template, crowdfunding_template, audit_status, reject_reason, sort, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?, NOW(), NOW())
	`
	result, err := db.MySQL.Exec(insertQuery,
		req.ChannelType, req.Name, req.ApplyCondition, req.ResponseTime,
		req.ContactPhone, req.ContactUrl, req.HelpLetterTemplate, req.CrowdfundingTemplate,
		auditStatus, req.Sort)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建渠道失败",
		})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data": gin.H{
			"id": id,
		},
	})
}

// UpdateChannel 更新求助渠道
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

	// 构建动态更新语句
	updateFields := []string{}
	args := []interface{}{}

	if req.ChannelType != "" {
		updateFields = append(updateFields, "channel_type = ?")
		args = append(args, req.ChannelType)
	}
	if req.Name != "" {
		updateFields = append(updateFields, "name = ?")
		args = append(args, req.Name)
	}
	if req.ApplyCondition != "" {
		updateFields = append(updateFields, "apply_condition = ?")
		args = append(args, req.ApplyCondition)
	}
	if req.ResponseTime != "" {
		updateFields = append(updateFields, "response_time = ?")
		args = append(args, req.ResponseTime)
	}
	if req.ContactPhone != "" {
		updateFields = append(updateFields, "contact_phone = ?")
		args = append(args, req.ContactPhone)
	}
	if req.ContactUrl != "" {
		updateFields = append(updateFields, "contact_url = ?")
		args = append(args, req.ContactUrl)
	}
	if req.HelpLetterTemplate != "" {
		updateFields = append(updateFields, "help_letter_template = ?")
		args = append(args, req.HelpLetterTemplate)
	}
	if req.CrowdfundingTemplate != "" {
		updateFields = append(updateFields, "crowdfunding_template = ?")
		args = append(args, req.CrowdfundingTemplate)
	}
	if req.Sort != nil {
		updateFields = append(updateFields, "sort = ?")
		args = append(args, *req.Sort)
	}
	if req.AuditStatus != nil {
		updateFields = append(updateFields, "audit_status = ?")
		args = append(args, *req.AuditStatus)
	}
	if req.RejectReason != nil {
		updateFields = append(updateFields, "reject_reason = ?")
		args = append(args, *req.RejectReason)
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "未提供更新字段",
		})
		return
	}

	updateFields = append(updateFields, "updated_at = NOW()")
	args = append(args, id)

	updateQuery := `UPDATE help_channel SET ` + strings.Join(updateFields, ", ") + ` WHERE id = ?`
	_, err = db.MySQL.Exec(updateQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新渠道失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    nil,
	})
}

// DeleteChannel 删除求助渠道（物理删除）
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

	deleteQuery := `DELETE FROM help_channel WHERE id = ?`
	result, err := db.MySQL.Exec(deleteQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除渠道失败",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取删除结果失败",
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "渠道不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    nil,
	})
}
