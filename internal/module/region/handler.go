package region

import (
	"database/sql"
	"net/http"
	"rare_backend/internal/pkg/db"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RegionItem 行政区划响应结构
type RegionItem struct {
	ID         uint   `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	FullName   string `json:"fullName"`
	ParentCode string `json:"parentCode"`
	Level      int    `json:"level"`
	Sort       int    `json:"sort"`
	IsEnabled  int    `json:"isEnabled"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// CreateRegionRequest 新增行政区划请求
type CreateRegionRequest struct {
	Code       string `json:"code" binding:"required"`
	Name       string `json:"name" binding:"required"`
	FullName   string `json:"fullName" binding:"required"`
	ParentCode string `json:"parentCode"` // 可选，根节点为空或""
	Level      int    `json:"level" binding:"required,min=1,max=3"`
	Sort       int    `json:"sort"`
	IsEnabled  int    `json:"isEnabled"`
}

// UpdateRegionRequest 更新行政区划请求
type UpdateRegionRequest struct {
	Name       *string `json:"name"`
	FullName   *string `json:"fullName"`
	ParentCode *string `json:"parentCode"`
	Level      *int    `json:"level"`
	Sort       *int    `json:"sort"`
	IsEnabled  *int    `json:"isEnabled"`
}

// CreateRegion 新增行政区划
func CreateRegion(c *gin.Context) {
	var req CreateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查编码是否已存在
	checkQuery := "SELECT id FROM region WHERE code = ?"
	var exists uint
	err := db.MySQL.QueryRow(checkQuery, req.Code).Scan(&exists)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "行政区编码已存在",
		})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "检查编码失败",
		})
		return
	}

	// 如果提供了 ParentCode，验证父级是否存在（可选，根据业务严谨度决定）
	if req.ParentCode != "" {
		var parentExists uint
		err := db.MySQL.QueryRow("SELECT id FROM region WHERE code = ?", req.ParentCode).Scan(&parentExists)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "父级行政区不存在",
			})
			return
		}
	}

	// 设置默认值
	if req.IsEnabled != 0 && req.IsEnabled != 1 {
		req.IsEnabled = 1
	}

	// 插入数据库
	insertQuery := `
		INSERT INTO region 
		(code, name, full_name, parent_code, level, sort, is_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := db.MySQL.Exec(insertQuery,
		req.Code, req.Name, req.FullName, req.ParentCode,
		req.Level, req.Sort, req.IsEnabled, now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "新增行政区划失败",
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

// UpdateRegion 更新行政区划
func UpdateRegion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的 ID",
		})
		return
	}

	var req UpdateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 检查记录是否存在
	checkQuery := "SELECT id FROM region WHERE id = ?"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "行政区划不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
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
	if req.FullName != nil {
		updateFields = append(updateFields, "full_name = ?")
		updateArgs = append(updateArgs, *req.FullName)
	}
	if req.ParentCode != nil {
		// 如果修改了 ParentCode，建议验证父级存在性，防止循环引用或无效父级
		if *req.ParentCode != "" {
			var parentExists uint
			err := db.MySQL.QueryRow("SELECT id FROM region WHERE code = ?", *req.ParentCode).Scan(&parentExists)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "父级行政区不存在",
				})
				return
			}
		}
		updateFields = append(updateFields, "parent_code = ?")
		updateArgs = append(updateArgs, *req.ParentCode)
	}
	if req.Level != nil {
		if *req.Level < 1 || *req.Level > 3 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "层级必须在1-3之间",
			})
			return
		}
		updateFields = append(updateFields, "level = ?")
		updateArgs = append(updateArgs, *req.Level)
	}
	if req.Sort != nil {
		updateFields = append(updateFields, "sort = ?")
		updateArgs = append(updateArgs, *req.Sort)
	}
	if req.IsEnabled != nil {
		updateFields = append(updateFields, "is_enabled = ?")
		updateArgs = append(updateArgs, *req.IsEnabled)
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

	updateQuery := "UPDATE region SET " + strings.Join(updateFields, ", ") + " WHERE id = ?"
	_, err = db.MySQL.Exec(updateQuery, updateArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DeleteRegion 删除行政区划（物理删除或逻辑删除，此处采用物理删除，需先检查是否有子级）
func DeleteRegion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的 ID",
		})
		return
	}

	// 获取该记录的 Code
	var code string
	err = db.MySQL.QueryRow("SELECT code FROM region WHERE id = ?", id).Scan(&code)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "行政区划不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	// 检查是否有子级关联
	var childCount int64
	db.MySQL.QueryRow("SELECT COUNT(*) FROM region WHERE parent_code = ?", code).Scan(&childCount)
	if childCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该行政区划下存在子级，无法删除",
		})
		return
	}

	// 执行删除
	_, err = db.MySQL.Exec("DELETE FROM region WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// GetRegionList 获取行政区划列表
// GetRegionListRequest 定义查询参数结构，方便管理
type GetRegionListRequest struct {
	ParentCode string `form:"parentCode"`
	Level      int    `form:"level"`
	IsEnabled  int    `form:"isEnabled"`
	Keyword    string `form:"keyword"`
	Page       int    `form:"page,default=1"`       // 页码，默认第1页
	PageSize   int    `form:"pageSize,default=100"` // 每页数量，默认100条（行政区数据较少，默认大一点可减少请求次数）
}

// GetRegionList 获取行政区划列表（支持分页）
func GetRegionList(c *gin.Context) {
	var req GetRegionListRequest
	// 使用 ShouldBindQuery 绑定 URL 查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 参数校验与默认值处理
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 1000 { // 限制最大页数防止恶意请求
		req.PageSize = 100
	}

	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if req.ParentCode != "" {
		whereClause += " AND parent_code = ?"
		args = append(args, req.ParentCode)
	}
	if req.Level != 0 { // 假设 level 0 表示不筛选层级
		whereClause += " AND level = ?"
		args = append(args, req.Level)
	}
	if req.IsEnabled != 0 { // 假设 isEnabled 0 表示不筛选状态
		whereClause += " AND is_enabled = ?"
		args = append(args, req.IsEnabled)
	}
	if req.Keyword != "" {
		whereClause += " AND (name LIKE ? OR full_name LIKE ? OR code LIKE ?)"
		args = append(args, "%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 1. 查询总数 (用于前端计算总页数)
	countQuery := "SELECT COUNT(*) FROM region " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 2. 查询列表数据 (带 LIMIT 和 OFFSET)
	offset := (req.Page - 1) * req.PageSize
	query := `
		SELECT id, code, name, full_name, parent_code, level, sort, is_enabled, created_at, updated_at
		FROM region
		` + whereClause + `
		ORDER BY sort ASC, code ASC
		LIMIT ? OFFSET ?
	`
	// 注意：LIMIT 和 OFFSET 的参数要追加到 args 后面
	args = append(args, req.PageSize, offset)

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询列表失败",
		})
		return
	}
	defer rows.Close()

	var list []RegionItem
	for rows.Next() {
		var r struct {
			ID         uint      `db:"id"`
			Code       string    `db:"code"`
			Name       string    `db:"name"`
			FullName   string    `db:"full_name"`
			ParentCode string    `db:"parent_code"`
			Level      int       `db:"level"`
			Sort       int       `db:"sort"`
			IsEnabled  int8      `db:"is_enabled"`
			CreatedAt  time.Time `db:"created_at"`
			UpdatedAt  time.Time `db:"updated_at"`
		}
		if err := rows.Scan(
			&r.ID, &r.Code, &r.Name, &r.FullName, &r.ParentCode,
			&r.Level, &r.Sort, &r.IsEnabled, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			continue
		}

		list = append(list, RegionItem{
			ID:         r.ID,
			Code:       r.Code,
			Name:       r.Name,
			FullName:   r.FullName,
			ParentCode: r.ParentCode,
			Level:      r.Level,
			Sort:       r.Sort,
			IsEnabled:  int(r.IsEnabled),
			CreatedAt:  r.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  r.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// 3. 返回结果包含分页信息
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":     list,
			"total":    total,
			"page":     req.Page,
			"pageSize": req.PageSize,
		},
	})
}

// flatRegion 用于接收数据库查询结果的临时结构体
type flatRegion struct {
	Code       string
	Name       string
	Level      int
	ParentCode string
}

// RegionTreeNode 行政区划树节点响应结构 (精简版，匹配前端需求)
type RegionTreeNode struct {
	Code     string            `json:"code"`
	Name     string            `json:"name"`
	Level    int               `json:"level"`
	Children []*RegionTreeNode `json:"children,omitempty"` // omitempty: 如果没有子节点，JSON中不显示该字段或显示为[]，取决于初始化
}

// GetRegionTree 获取行政区划树
func GetRegionTree(c *gin.Context) {
	isEnabledStr := c.DefaultQuery("isEnabled", "1")

	// 处理 isEnabled 参数
	isEnabled := 1
	if isEnabledStr != "" {
		val, err := strconv.Atoi(isEnabledStr)
		if err == nil {
			isEnabled = val
		}
	}

	// 1. 查询数据：必须包含 parent_code 用于构建树
	query := `
		SELECT code, name, level, parent_code
		FROM region
		WHERE is_enabled = ?
		ORDER BY sort ASC, code ASC
	`

	rows, err := db.MySQL.Query(query, isEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "查询失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	// 使用包级别定义的 flatRegion
	var flatList []flatRegion

	for rows.Next() {
		var r flatRegion
		// 注意 Scan 的顺序要与 SELECT 列顺序一致: code, name, level, parent_code
		err := rows.Scan(&r.Code, &r.Name, &r.Level, &r.ParentCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "数据解析失败: " + err.Error(),
			})
			return
		}
		flatList = append(flatList, r)
	}

	// 2. 构建树形结构
	tree := buildRegionTree(flatList)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    tree,
	})
}

// buildRegionTree 将扁平列表转换为树形结构
func buildRegionTree(flatList []flatRegion) []*RegionTreeNode {

	// 1. 创建所有节点的映射，值是指针以便修改 Children
	nodeMap := make(map[string]*RegionTreeNode)
	var rootNodes []*RegionTreeNode

	// 第一遍：初始化所有节点并存入 Map
	for _, r := range flatList {
		node := &RegionTreeNode{
			Code:     r.Code,
			Name:     r.Name,
			Level:    r.Level,
			Children: []*RegionTreeNode{}, // 初始化为空切片，保证 JSON 输出 [] 而不是 null
		}
		nodeMap[r.Code] = node
	}

	// 第二遍：建立父子关系
	for _, r := range flatList {
		node := nodeMap[r.Code]

		// 判断是否为根节点 (ParentCode 为空 或 "0")
		if r.ParentCode == "" || r.ParentCode == "0" {
			rootNodes = append(rootNodes, node)
		} else {
			// 查找父节点
			if parent, exists := nodeMap[r.ParentCode]; exists {
				parent.Children = append(parent.Children, node)
			} else {
				// 如果父节点不存在（例如被过滤掉了），则将其作为根节点处理，避免数据丢失
				rootNodes = append(rootNodes, node)
			}
		}
	}

	return rootNodes
}

// GetProvinceCityTree 获取省市两级行政区划树
func GetProvinceCityTree(c *gin.Context) {
	isEnabledStr := c.DefaultQuery("isEnabled", "1")

	// 处理 isEnabled 参数
	isEnabled := 1
	if isEnabledStr != "" {
		val, err := strconv.Atoi(isEnabledStr)
		if err == nil {
			isEnabled = val
		}
	}

	// 1. 查询数据：只查询 Level 1 (省) 和 Level 2 (市)
	// 注意：这里假设 level=1 是省，level=2 是市。如果你的业务定义不同，请调整 WHERE 条件
	query := `
		SELECT code, name, level, parent_code
		FROM region
		WHERE is_enabled = ? AND level <= 2
		ORDER BY sort ASC, code ASC
	`

	rows, err := db.MySQL.Query(query, isEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "查询失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var flatList []flatRegion

	for rows.Next() {
		var r flatRegion
		err := rows.Scan(&r.Code, &r.Name, &r.Level, &r.ParentCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "数据解析失败: " + err.Error(),
			})
			return
		}
		flatList = append(flatList, r)
	}

	// 2. 构建树形结构 (复用现有的 buildRegionTree 函数)
	// 因为数据源中已经不包含 level 3 的数据，所以生成的树自然只有省市两级
	tree := buildRegionTree(flatList)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    tree,
	})
}

// GetRegionDetail 获取行政区划详情
func GetRegionDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的 ID",
		})
		return
	}

	query := `
		SELECT id, code, name, full_name, parent_code, level, sort, is_enabled, created_at, updated_at
		FROM region
		WHERE id = ?
	`

	var r struct {
		ID         uint      `db:"id"`
		Code       string    `db:"code"`
		Name       string    `db:"name"`
		FullName   string    `db:"full_name"`
		ParentCode string    `db:"parent_code"`
		Level      int       `db:"level"`
		Sort       int       `db:"sort"`
		IsEnabled  int8      `db:"is_enabled"`
		CreatedAt  time.Time `db:"created_at"`
		UpdatedAt  time.Time `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&r.ID, &r.Code, &r.Name, &r.FullName, &r.ParentCode,
		&r.Level, &r.Sort, &r.IsEnabled, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "行政区划不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": RegionItem{
			ID:         r.ID,
			Code:       r.Code,
			Name:       r.Name,
			FullName:   r.FullName,
			ParentCode: r.ParentCode,
			Level:      r.Level,
			Sort:       r.Sort,
			IsEnabled:  int(r.IsEnabled),
			CreatedAt:  r.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  r.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}
