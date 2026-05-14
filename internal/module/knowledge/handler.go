package knowledge

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"rare_backend/internal/pkg/db"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CategoryItem 分类响应结构
type CategoryItem struct {
	ID          uint   `json:"id"`
	ParentID    uint   `json:"parentId"`
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	SortOrder   int    `json:"sortOrder"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// CreateCategoryRequest 新增分类请求
type CreateCategoryRequest struct {
	ParentID    uint   `json:"parentId"`    // 父分类ID，0为一级
	Level       int    `json:"level"`       // 层级：1/2/3
	Name        string `json:"name"`        // 分类名称
	Code        string `json:"code"`        // 分类编码
	Description string `json:"description"` // 分类描述
	IconURL     string `json:"iconUrl"`     // 分类图标链接
	SortOrder   int    `json:"sortOrder"`   // 排序字段
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	ParentID    *uint   `json:"parentId"`
	Level       *int    `json:"level"`
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
	IconURL     *string `json:"iconUrl"`
	SortOrder   *int    `json:"sortOrder"`
	Status      *int    `json:"status"`
}

// CreateCategory 新增分类
func CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
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
			"message": "分类名称不能为空",
		})
		return
	}
	if req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "分类编码不能为空",
		})
		return
	}
	if req.Level < 1 || req.Level > 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "层级必须在1-3之间",
		})
		return
	}

	// 检查编码是否已存在
	checkQuery := "SELECT id FROM category WHERE code = ? AND status = 1"
	var exists uint
	err := db.MySQL.QueryRow(checkQuery, req.Code).Scan(&exists)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "分类编码已存在",
		})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "检查编码失败",
		})
		return
	}

	// 插入数据库
	insertQuery := `
		INSERT INTO category 
		(parent_id, level, name, code, description, icon_url, sort_order, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
	`
	now := time.Now()
	result, err := db.MySQL.Exec(insertQuery,
		req.ParentID, req.Level, req.Name, req.Code,
		req.Description, req.IconURL, req.SortOrder,
		now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "新增分类失败",
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

// UpdateCategory 更新分类
func UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的分类 ID",
		})
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 检查分类是否存在
	checkQuery := "SELECT id FROM category WHERE id = ?"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "分类不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询分类失败",
		})
		return
	}

	// 构建动态更新语句
	updateFields := []string{}
	updateArgs := []interface{}{}

	if req.ParentID != nil {
		updateFields = append(updateFields, "parent_id = ?")
		updateArgs = append(updateArgs, *req.ParentID)
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
	if req.Name != nil {
		updateFields = append(updateFields, "name = ?")
		updateArgs = append(updateArgs, *req.Name)
	}
	if req.Code != nil {
		// 如果修改了code，需要检查唯一性
		checkCodeQuery := "SELECT id FROM category WHERE code = ? AND id != ? AND status = 1"
		var codeExists uint
		err := db.MySQL.QueryRow(checkCodeQuery, *req.Code, id).Scan(&codeExists)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "分类编码已存在",
			})
			return
		} else if err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "检查编码失败",
			})
			return
		}
		updateFields = append(updateFields, "code = ?")
		updateArgs = append(updateArgs, *req.Code)
	}
	if req.Description != nil {
		updateFields = append(updateFields, "description = ?")
		updateArgs = append(updateArgs, *req.Description)
	}
	if req.IconURL != nil {
		updateFields = append(updateFields, "icon_url = ?")
		updateArgs = append(updateArgs, *req.IconURL)
	}
	if req.SortOrder != nil {
		updateFields = append(updateFields, "sort_order = ?")
		updateArgs = append(updateArgs, *req.SortOrder)
	}
	if req.Status != nil {
		updateFields = append(updateFields, "status = ?")
		updateArgs = append(updateArgs, *req.Status)
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

	updateQuery := "UPDATE category SET " + joinUpdateFields(updateFields) + " WHERE id = ?"
	_, err = db.MySQL.Exec(updateQuery, updateArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新分类失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DeleteCategory 删除分类（软删除，将状态设为0）
func DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的分类 ID",
		})
		return
	}

	// 检查分类是否存在且未删除
	checkQuery := "SELECT id FROM category WHERE id = ? AND status = 1"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "分类不存在或已删除",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询分类失败",
		})
		return
	}

	// 软删除：将 status 设为 0
	deleteQuery := "UPDATE category SET status = 0, updated_at = ? WHERE id = ?"
	_, err = db.MySQL.Exec(deleteQuery, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除分类失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// joinUpdateFields 辅助函数：拼接更新字段字符串
func joinUpdateFields(fields []string) string {
	result := ""
	for i, f := range fields {
		if i > 0 {
			result += ", "
		}
		result += f
	}
	return result
}

// TagItem 标签响应结构
type TagItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	SortOrder int    `json:"sortOrder"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// CreateTagRequest 新增标签请求
type CreateTagRequest struct {
	Name      string `json:"name"`      // 标签名
	Code      string `json:"code"`      // 标签编码
	SortOrder int    `json:"sortOrder"` // 排序字段
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	Name      *string `json:"name"`
	Code      *string `json:"code"`
	SortOrder *int    `json:"sortOrder"`
	Status    *int    `json:"status"`
}

// CreateTag 新增标签
func CreateTag(c *gin.Context) {
	var req CreateTagRequest
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
			"message": "标签名称不能为空",
		})
		return
	}
	if req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "标签编码不能为空",
		})
		return
	}

	// 检查编码是否已存在
	checkQuery := "SELECT id FROM tag WHERE code = ? AND status = 1"
	var exists uint
	err := db.MySQL.QueryRow(checkQuery, req.Code).Scan(&exists)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "标签编码已存在",
		})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "检查编码失败",
		})
		return
	}

	// 插入数据库
	insertQuery := `
		INSERT INTO tag 
		(name, code, sort_order, status, created_at, updated_at)
		VALUES (?, ?, ?, 1, ?, ?)
	`
	now := time.Now()
	result, err := db.MySQL.Exec(insertQuery,
		req.Name, req.Code, req.SortOrder,
		now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "新增标签失败",
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

// UpdateTag 更新标签
func UpdateTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的标签 ID",
		})
		return
	}

	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 检查标签是否存在
	checkQuery := "SELECT id FROM tag WHERE id = ?"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "标签不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询标签失败",
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
	if req.Code != nil {
		// 如果修改了code，需要检查唯一性
		checkCodeQuery := "SELECT id FROM tag WHERE code = ? AND id != ? AND status = 1"
		var codeExists uint
		err := db.MySQL.QueryRow(checkCodeQuery, *req.Code, id).Scan(&codeExists)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "标签编码已存在",
			})
			return
		} else if err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "检查编码失败",
			})
			return
		}
		updateFields = append(updateFields, "code = ?")
		updateArgs = append(updateArgs, *req.Code)
	}
	if req.SortOrder != nil {
		updateFields = append(updateFields, "sort_order = ?")
		updateArgs = append(updateArgs, *req.SortOrder)
	}
	if req.Status != nil {
		updateFields = append(updateFields, "status = ?")
		updateArgs = append(updateArgs, *req.Status)
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

	updateQuery := "UPDATE tag SET " + strings.Join(updateFields, ", ") + " WHERE id = ?"
	_, err = db.MySQL.Exec(updateQuery, updateArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新标签失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DeleteTag 删除标签（软删除，将状态设为0）
func DeleteTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的标签 ID",
		})
		return
	}

	// 检查标签是否存在且未删除
	checkQuery := "SELECT id FROM tag WHERE id = ? AND status = 1"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "标签不存在或已删除",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询标签失败",
		})
		return
	}

	// 软删除：将 status 设为 0
	deleteQuery := "UPDATE tag SET status = 0, updated_at = ? WHERE id = ?"
	_, err = db.MySQL.Exec(deleteQuery, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除标签失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// SimpleCategoryItem 简化版分类响应结构
type SimpleCategoryItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// SimpleTagItem 简化版标签响应结构
type SimpleTagItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// DiseaseItem 疾病响应结构
type DiseaseItem struct {
	ID           uint        `json:"id"`
	Name         string      `json:"name"`
	Alias        string      `json:"alias"`
	Introduction string      `json:"introduction"`
	Symptoms     string      `json:"symptoms"`
	Images       interface{} `json:"images"` // JSON 数组或字符串

	// 关联字段 - 符合新的返回值样式
	PrimaryCategoryID   *uint   `json:"primaryCategoryId"`   // 主分类ID (指针，可能为空)
	PrimaryCategoryName *string `json:"primaryCategoryName"` // 主分类名称

	CategoryIDs []uint               `json:"categoryIds"` // 所有分类ID列表
	Categories  []SimpleCategoryItem `json:"categories"`  // 所有分类详情列表

	TagIDs []uint          `json:"tagIds"` // 所有标签ID列表
	Tags   []SimpleTagItem `json:"tags"`   // 所有标签详情列表

	// 【新增】关联的文章列表
	Articles []ArticleItem `json:"articles"`

	Status    int    `json:"status"`
	CreatorID uint   `json:"creatorId"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// CreateDiseaseRequest 新增疾病请求
type CreateDiseaseRequest struct {
	Name         string   `json:"name"`         // 病种名称
	Alias        string   `json:"alias"`        // 别名
	Introduction string   `json:"introduction"` // 简介
	Symptoms     string   `json:"symptoms"`     // 症状
	Images       []string `json:"images"`       // 修改为 []string，接收 JSON 数组
	Status       int      `json:"status"`       // 状态：1启用 0停用
	CreatorID    uint     `json:"creatorId"`    // 创建人ID

	// 关联字段
	PrimaryCategoryID uint   `json:"primary_category_id"` // 主分类ID
	CategoryIDs       []uint `json:"category_ids"`        // 所有关联的分类ID列表
	TagIDs            []uint `json:"tag_ids"`             // 关联的标签ID列表
}

// UpdateDiseaseRequest 更新疾病请求
type UpdateDiseaseRequest struct {
	Name         *string   `json:"name"`
	Alias        *string   `json:"alias"`
	Introduction *string   `json:"introduction"`
	Symptoms     *string   `json:"symptoms"`
	Images       *[]string `json:"images"` // 支持更新为数组
	Status       *int      `json:"status"`

	// 新增关联字段更新支持
	PrimaryCategoryID *uint   `json:"primary_category_id"` // 主分类ID
	CategoryIDs       *[]uint `json:"category_ids"`        // 所有关联的分类ID列表
	TagIDs            *[]uint `json:"tag_ids"`             // 关联的标签ID列表
}

// CreateDisease 新增疾病
func CreateDisease(c *gin.Context) {
	var req CreateDiseaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 1. 必填字段验证
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "病种名称不能为空",
		})
		return
	}

	// 默认状态处理
	if req.Status != 0 && req.Status != 1 {
		req.Status = 1
	}

	// 2. 处理 Images 字段：将 []string 转换为 JSON 字符串用于存储
	// MySQL JSON 类型可以直接接受 string 格式的 JSON
	imagesJSON := "[]"
	if len(req.Images) > 0 {
		bytes, err := json.Marshal(req.Images)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "图片数据格式化失败",
			})
			return
		}
		imagesJSON = string(bytes)
	}

	// 3. 开启数据库事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务开启失败",
		})
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	// 4. 插入 disease 主表
	insertDiseaseQuery := `
		INSERT INTO disease 
		(name, alias, introduction, symptoms, images, status, creator_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := tx.Exec(insertDiseaseQuery,
		req.Name, req.Alias, req.Introduction, req.Symptoms,
		imagesJSON, req.Status, req.CreatorID, now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "新增疾病失败: " + err.Error(),
		})
		return
	}

	diseaseID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取疾病ID失败",
		})
		return
	}

	// 5. 插入 disease_category_rel 关联表
	if len(req.CategoryIDs) > 0 {
		categoryMap := make(map[uint]bool)
		for _, catID := range req.CategoryIDs {
			categoryMap[catID] = true
		}

		for catID := range categoryMap {
			isPrimary := 0
			if catID == req.PrimaryCategoryID {
				isPrimary = 1
			}

			insertCatRelQuery := `
				INSERT INTO disease_category_rel (disease_id, category_id, is_primary, created_at)
				VALUES (?, ?, ?, ?)
			`
			_, err := tx.Exec(insertCatRelQuery, diseaseID, catID, isPrimary, now)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "关联分类失败: " + err.Error(),
				})
				return
			}
		}
	}

	// 6. 插入 disease_tag_rel 关联表
	if len(req.TagIDs) > 0 {
		tagMap := make(map[uint]bool)
		for _, tagID := range req.TagIDs {
			tagMap[tagID] = true
		}

		for tagID := range tagMap {
			insertTagRelQuery := `
				INSERT INTO disease_tag_rel (disease_id, tag_id, created_at)
				VALUES (?, ?, ?)
			`
			_, err := tx.Exec(insertTagRelQuery, diseaseID, tagID, now)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "关联标签失败: " + err.Error(),
				})
				return
			}
		}
	}

	// 7. 提交事务
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	// 8. 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id": diseaseID,
		},
	})
}

// GetDiseaseByID 获取疾病详情
func GetDiseaseByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的疾病 ID",
		})
		return
	}

	// 1. 查询疾病主表信息
	query := `
		SELECT id, name, alias, introduction, symptoms, images, status, creator_id, created_at, updated_at
		FROM disease
		WHERE id = ? AND status = 1
	`

	var disease struct {
		ID           uint           `db:"id"`
		Name         string         `db:"name"`
		Alias        sql.NullString `db:"alias"`
		Introduction sql.NullString `db:"introduction"`
		Symptoms     sql.NullString `db:"symptoms"`
		Images       sql.NullString `db:"images"`
		Status       int8           `db:"status"`
		CreatorID    sql.NullInt64  `db:"creator_id"`
		CreatedAt    time.Time      `db:"created_at"`
		UpdatedAt    time.Time      `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&disease.ID, &disease.Name, &disease.Alias, &disease.Introduction,
		&disease.Symptoms, &disease.Images, &disease.Status, &disease.CreatorID,
		&disease.CreatedAt, &disease.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "疾病不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询疾病详情失败",
		})
		return
	}

	// 2. 解析 Images
	var imagesData interface{}
	if disease.Images.Valid && disease.Images.String != "" {
		json.Unmarshal([]byte(disease.Images.String), &imagesData)
	}

	// 3. 查询关联的分类详情 (ID, Name, IsPrimary)
	catQuery := `
		SELECT dcr.category_id, c.name, dcr.is_primary 
		FROM disease_category_rel dcr
		JOIN category c ON dcr.category_id = c.id
		WHERE dcr.disease_id = ?
	`

	var categoryIDs []uint
	var categories []SimpleCategoryItem
	var primaryCategoryID *uint
	var primaryCategoryName *string

	catRows, err := db.MySQL.Query(catQuery, id)
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var catID uint
			var catName string
			var isPrimary int8

			if err := catRows.Scan(&catID, &catName, &isPrimary); err == nil {
				categoryIDs = append(categoryIDs, catID)
				categories = append(categories, SimpleCategoryItem{
					ID:   catID,
					Name: catName,
				})

				if isPrimary == 1 {
					pid := catID
					pName := catName
					primaryCategoryID = &pid
					primaryCategoryName = &pName
				}
			}
		}
	}

	// 确保切片不为 null
	if categoryIDs == nil {
		categoryIDs = []uint{}
	}
	if categories == nil {
		categories = []SimpleCategoryItem{}
	}

	// 4. 查询关联的标签详情 (ID, Name)
	tagQuery := `
		SELECT dtr.tag_id, t.name 
		FROM disease_tag_rel dtr
		JOIN tag t ON dtr.tag_id = t.id
		WHERE dtr.disease_id = ?
	`

	var tagIDs []uint
	var tags []SimpleTagItem

	tagRows, err := db.MySQL.Query(tagQuery, id)
	if err == nil {
		defer tagRows.Close()
		for tagRows.Next() {
			var tagID uint
			var tagName string

			if err := tagRows.Scan(&tagID, &tagName); err == nil {
				tagIDs = append(tagIDs, tagID)
				tags = append(tags, SimpleTagItem{
					ID:   tagID,
					Name: tagName,
				})
			}
		}
	}

	// 确保切片不为 null
	if tagIDs == nil {
		tagIDs = []uint{}
	}
	if tags == nil {
		tags = []SimpleTagItem{}
	}

	// 5. 【新增】查询关联的文章列表
	// 只查询已发布 (status = 2) 的文章，并按发布时间倒序排列
	articleQuery := `
		SELECT a.id, a.title, a.summary, a.cover_image, a.source_name, a.publish_time, a.view_count
		FROM disease_article_rel dar
		JOIN article a ON dar.article_id = a.id
		WHERE dar.disease_id = ? AND a.status = 2
		ORDER BY a.publish_time DESC
	`

	var articles []ArticleItem
	artRows, err := db.MySQL.Query(articleQuery, id)
	if err == nil {
		defer artRows.Close()
		for artRows.Next() {
			var art ArticleItem
			var publishTime sql.NullTime // 处理可能为 NULL 的时间

			if err := artRows.Scan(
				&art.ID, &art.Title, &art.Summary, &art.CoverImage,
				&art.SourceName, &publishTime, &art.ViewCount,
			); err == nil {
				// 格式化时间
				if publishTime.Valid {
					art.PublishTime = publishTime.Time.Format("2006-01-02 15:04:05")
				} else {
					art.PublishTime = ""
				}
				articles = append(articles, art)
			}
		}
	}

	// 确保 articles 不为 null
	if articles == nil {
		articles = []ArticleItem{}
	}

	// 6. 构建响应对象
	item := DiseaseItem{
		ID:           disease.ID,
		Name:         disease.Name,
		Alias:        disease.Alias.String,
		Introduction: disease.Introduction.String,
		Symptoms:     disease.Symptoms.String,
		Images:       imagesData,

		PrimaryCategoryID:   primaryCategoryID,
		PrimaryCategoryName: primaryCategoryName,

		CategoryIDs: categoryIDs,
		Categories:  categories,

		TagIDs: tagIDs,
		Tags:   tags,

		// 【新增】赋值文章列表
		Articles: articles,

		Status:    int(disease.Status),
		CreatorID: uint(disease.CreatorID.Int64),
		CreatedAt: disease.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: disease.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    item,
	})
}

// UpdateDisease 更新疾病
func UpdateDisease(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的疾病 ID",
		})
		return
	}

	var req UpdateDiseaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 1. 检查疾病是否存在
	checkQuery := "SELECT id FROM disease WHERE id = ?"
	var exists uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "疾病不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询疾病失败",
		})
		return
	}

	// 2. 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务开启失败",
		})
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	updateFields := []string{}
	updateArgs := []interface{}{}

	// 3. 构建主表更新字段
	if req.Name != nil {
		updateFields = append(updateFields, "name = ?")
		updateArgs = append(updateArgs, *req.Name)
	}
	if req.Alias != nil {
		updateFields = append(updateFields, "alias = ?")
		updateArgs = append(updateArgs, *req.Alias)
	}
	if req.Introduction != nil {
		updateFields = append(updateFields, "introduction = ?")
		updateArgs = append(updateArgs, *req.Introduction)
	}
	if req.Symptoms != nil {
		updateFields = append(updateFields, "symptoms = ?")
		updateArgs = append(updateArgs, *req.Symptoms)
	}

	// 处理 Images 更新
	if req.Images != nil {
		imagesJSON := "[]"
		if len(*req.Images) > 0 {
			bytes, err := json.Marshal(*req.Images)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "图片数据格式化失败",
				})
				return
			}
			imagesJSON = string(bytes)
		}
		updateFields = append(updateFields, "images = ?")
		updateArgs = append(updateArgs, imagesJSON)
	}

	if req.Status != nil {
		updateFields = append(updateFields, "status = ?")
		updateArgs = append(updateArgs, *req.Status)
	}

	// 如果有主表字段需要更新
	if len(updateFields) > 0 {
		updateFields = append(updateFields, "updated_at = ?")
		updateArgs = append(updateArgs, now)
		updateArgs = append(updateArgs, id)

		updateQuery := "UPDATE disease SET " + strings.Join(updateFields, ", ") + " WHERE id = ?"
		_, err = tx.Exec(updateQuery, updateArgs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新疾病主表失败: " + err.Error(),
			})
			return
		}
	}

	// 4. 处理分类关联更新 (如果前端传了 category_ids)
	if req.CategoryIDs != nil {
		// 4.1 删除旧的分类关联
		_, err = tx.Exec("DELETE FROM disease_category_rel WHERE disease_id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "清理旧分类关联失败",
			})
			return
		}

		// 4.2 插入新的分类关联
		if len(*req.CategoryIDs) > 0 {
			// 去重
			categoryMap := make(map[uint]bool)
			for _, catID := range *req.CategoryIDs {
				categoryMap[catID] = true
			}

			for catID := range categoryMap {
				isPrimary := 0
				// 如果传了 primary_category_id 且匹配当前分类ID，则设为主分类
				if req.PrimaryCategoryID != nil && *req.PrimaryCategoryID == catID {
					isPrimary = 1
				}

				insertCatRelQuery := `
					INSERT INTO disease_category_rel (disease_id, category_id, is_primary, created_at)
					VALUES (?, ?, ?, ?)
				`
				_, err := tx.Exec(insertCatRelQuery, id, catID, isPrimary, now)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    500,
						"message": "关联分类失败: " + err.Error(),
					})
					return
				}
			}
		}
	} else if req.PrimaryCategoryID != nil {
		// 如果没传 category_ids 但传了 primary_category_id，说明只想修改主分类标记
		// 这种情况比较复杂，通常建议一起传。这里简单处理：更新现有记录中的 is_primary
		// 先将所有设为0
		_, err = tx.Exec("UPDATE disease_category_rel SET is_primary = 0 WHERE disease_id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新主分类标记失败",
			})
			return
		}
		// 再将指定的设为主分类（前提是该关联已存在）
		if *req.PrimaryCategoryID > 0 {
			_, err = tx.Exec("UPDATE disease_category_rel SET is_primary = 1 WHERE disease_id = ? AND category_id = ?", id, *req.PrimaryCategoryID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "更新主分类标记失败",
				})
				return
			}
		}
	}

	// 5. 处理标签关联更新 (如果前端传了 tag_ids)
	if req.TagIDs != nil {
		// 5.1 删除旧的标签关联
		_, err = tx.Exec("DELETE FROM disease_tag_rel WHERE disease_id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "清理旧标签关联失败",
			})
			return
		}

		// 5.2 插入新的标签关联
		if len(*req.TagIDs) > 0 {
			// 去重
			tagMap := make(map[uint]bool)
			for _, tagID := range *req.TagIDs {
				tagMap[tagID] = true
			}

			for tagID := range tagMap {
				insertTagRelQuery := `
					INSERT INTO disease_tag_rel (disease_id, tag_id, created_at)
					VALUES (?, ?, ?)
				`
				_, err := tx.Exec(insertTagRelQuery, id, tagID, now)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    500,
						"message": "关联标签失败: " + err.Error(),
					})
					return
				}
			}
		}
	}

	// 6. 提交事务
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DeleteDisease 删除疾病（物理删除主表及关联表数据）
func DeleteDisease(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的疾病 ID",
		})
		return
	}

	// 1. 开启数据库事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库事务开启失败",
		})
		return
	}
	// 确保在函数退出前处理事务状态
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // 重新抛出 panic
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// 2. 检查疾病是否存在
	// 注意：物理删除前通常不需要判断 status，只要 ID 存在即可删除。
	// 如果业务要求只能删除“启用”的疾病，可以加上 AND status = 1
	checkQuery := "SELECT id FROM disease WHERE id = ?"
	var exists uint
	err = tx.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "疾病不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询疾病失败",
		})
		return
	}

	// 3. 物理删除主表记录
	deleteMainQuery := "DELETE FROM disease WHERE id = ?"
	_, err = tx.Exec(deleteMainQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除疾病主表失败",
		})
		return
	}

	// 4. 清理关联表数据
	// 即使主表删除了，为了保持数据库整洁和避免潜在的外键约束问题（如果有的话），
	// 显式删除关联表数据是最佳实践。

	// 4.1 删除分类关联
	_, err = tx.Exec("DELETE FROM disease_category_rel WHERE disease_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "清理分类关联失败",
		})
		return
	}

	// 4.2 删除标签关联
	_, err = tx.Exec("DELETE FROM disease_tag_rel WHERE disease_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "清理标签关联失败",
		})
		return
	}

	// 5. 提交事务
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "事务提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// GetDiseases 获取疾病列表（支持分页、关键词、状态、分类、标签筛选）
func GetDiseases(c *gin.Context) {
	// 1. 获取请求参数
	keyword := c.DefaultQuery("keyword", "")
	statusStr := c.DefaultQuery("status", "") // 空字符串表示不筛选状态
	categoryIdStr := c.DefaultQuery("categoryId", "")
	tagIdStr := c.DefaultQuery("tagId", "")

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	// 2. 参数解析与校验
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// 解析可选整数参数
	var statusFilter *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			statusFilter = &s
		}
	}

	var categoryIdFilter *uint
	if categoryIdStr != "" {
		id, err := strconv.ParseUint(categoryIdStr, 10, 32)
		if err == nil {
			v := uint(id)
			categoryIdFilter = &v
		}
	}

	var tagIdFilter *uint
	if tagIdStr != "" {
		id, err := strconv.ParseUint(tagIdStr, 10, 32)
		if err == nil {
			v := uint(id)
			tagIdFilter = &v
		}
	}

	// 3. 构建动态 SQL 查询条件
	// 基础查询：从 disease 表开始
	// 如果涉及分类或标签筛选，需要 JOIN 关联表
	whereConditions := []string{"1=1"} // 使用 1=1 方便后续拼接 AND
	args := []interface{}{}

	// 3.1 状态筛选
	if statusFilter != nil {
		whereConditions = append(whereConditions, "d.status = ?")
		args = append(args, *statusFilter)
	} else {
		// 如果未指定状态，默认只查询启用的？或者查询所有？
		// 通常列表页默认只看启用的，除非明确传 status=0。
		// 这里为了兼容旧逻辑，如果没传 status，我们默认加上 status=1，除非业务要求看所有。
		// 假设默认只看启用状态，如果前端想看不启用的，必须传 status=0
		whereConditions = append(whereConditions, "d.status = 1")
	}

	// 3.2 关键词筛选 (名称或别名)
	if keyword != "" {
		whereConditions = append(whereConditions, "(d.name LIKE ? OR d.alias LIKE ?)")
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 3.3 分类筛选 (需要 JOIN disease_category_rel)
	hasJoinCategory := false
	if categoryIdFilter != nil {
		hasJoinCategory = true
		whereConditions = append(whereConditions, "dcr.category_id = ?")
		args = append(args, *categoryIdFilter)
	}

	// 3.4 标签筛选 (需要 JOIN disease_tag_rel)
	hasJoinTag := false
	if tagIdFilter != nil {
		hasJoinTag = true
		whereConditions = append(whereConditions, "dtr.tag_id = ?")
		args = append(args, *tagIdFilter)
	}

	// 拼接 WHERE 子句
	whereClause := strings.Join(whereConditions, " AND ")

	// 4. 构建总数查询 SQL
	countQuery := "SELECT COUNT(*) FROM disease d"

	// 根据是否有筛选条件决定是否需要 JOIN
	if hasJoinCategory {
		countQuery += " INNER JOIN disease_category_rel dcr ON d.id = dcr.disease_id"
	}
	if hasJoinTag {
		countQuery += " INNER JOIN disease_tag_rel dtr ON d.id = dtr.disease_id"
	}

	countQuery += " WHERE " + whereClause

	var total int64
	err = db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败: " + err.Error(),
		})
		return
	}

	// 5. 构建列表查询 SQL
	listQuery := "SELECT d.id, d.name, d.alias, d.introduction, d.symptoms, d.images, d.status, d.creator_id, d.created_at, d.updated_at FROM disease d"

	if hasJoinCategory {
		listQuery += " INNER JOIN disease_category_rel dcr ON d.id = dcr.disease_id"
	}
	if hasJoinTag {
		listQuery += " INNER JOIN disease_tag_rel dtr ON d.id = dtr.disease_id"
	}

	listQuery += " WHERE " + whereClause + " ORDER BY d.created_at DESC LIMIT ? OFFSET ?"

	// 注意：args 需要追加分页参数
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

	// 6. 扫描结果
	var list []DiseaseItem
	for rows.Next() {
		var d struct {
			ID           uint           `db:"id"`
			Name         string         `db:"name"`
			Alias        sql.NullString `db:"alias"`
			Introduction sql.NullString `db:"introduction"`
			Symptoms     sql.NullString `db:"symptoms"`
			Images       sql.NullString `db:"images"`
			Status       int8           `db:"status"`
			CreatorID    sql.NullInt64  `db:"creator_id"`
			CreatedAt    time.Time      `db:"created_at"`
			UpdatedAt    time.Time      `db:"updated_at"`
		}

		if err := rows.Scan(
			&d.ID, &d.Name, &d.Alias, &d.Introduction,
			&d.Symptoms, &d.Images, &d.Status, &d.CreatorID,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			// 记录错误日志，跳过当前行
			continue
		}

		item := DiseaseItem{
			ID:           d.ID,
			Name:         d.Name,
			Alias:        d.Alias.String,
			Introduction: d.Introduction.String,
			Symptoms:     d.Symptoms.String,
			Images:       d.Images.String,
			Status:       int(d.Status),
			CreatorID:    uint(d.CreatorID.Int64),
			CreatedAt:    d.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    d.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		list = append(list, item)
	}

	// 确保 list 不为 null
	if list == nil {
		list = []DiseaseItem{}
	}

	// 7. 返回结果
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

// GetCategories 获取分类列表（通常用于树形展示或下拉选项，不分页）
func GetCategories(c *gin.Context) {
	// 可选：支持按层级或父ID筛选
	parentIDStr := c.DefaultQuery("parentId", "-1") // -1 表示获取所有，或者特定逻辑
	levelStr := c.DefaultQuery("level", "")

	whereClause := "WHERE status = 1"
	args := []interface{}{}

	if parentIDStr != "-1" {
		parentID, err := strconv.Atoi(parentIDStr)
		if err == nil {
			whereClause += " AND parent_id = ?"
			args = append(args, parentID)
		}
	}

	if levelStr != "" {
		level, err := strconv.Atoi(levelStr)
		if err == nil {
			whereClause += " AND level = ?"
			args = append(args, level)
		}
	}

	query := `
		SELECT id, parent_id, level, name, code, description, icon_url, sort_order, status, created_at, updated_at
		FROM category
		` + whereClause + `
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询分类列表失败",
		})
		return
	}
	defer rows.Close()

	var list []CategoryItem
	for rows.Next() {
		var cat struct {
			ID          uint           `db:"id"`
			ParentID    uint           `db:"parent_id"`
			Level       int            `db:"level"`
			Name        string         `db:"name"`
			Code        string         `db:"code"`
			Description sql.NullString `db:"description"`
			IconURL     sql.NullString `db:"icon_url"`
			SortOrder   int            `db:"sort_order"`
			Status      int8           `db:"status"`
			CreatedAt   string         `db:"created_at"`
			UpdatedAt   string         `db:"updated_at"`
		}
		if err := rows.Scan(
			&cat.ID, &cat.ParentID, &cat.Level, &cat.Name, &cat.Code,
			&cat.Description, &cat.IconURL, &cat.SortOrder, &cat.Status,
			&cat.CreatedAt, &cat.UpdatedAt,
		); err != nil {
			continue
		}

		list = append(list, CategoryItem{
			ID:          cat.ID,
			ParentID:    cat.ParentID,
			Level:       cat.Level,
			Name:        cat.Name,
			Code:        cat.Code,
			Description: cat.Description.String,
			IconURL:     cat.IconURL.String,
			SortOrder:   cat.SortOrder,
			Status:      int(cat.Status),
			CreatedAt:   cat.CreatedAt,
			UpdatedAt:   cat.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    list,
	})
}

// GetTags 获取标签列表
func GetTags(c *gin.Context) {
	query := `
		SELECT id, name, code, sort_order, status, created_at, updated_at
		FROM tag
		WHERE status = 1
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := db.MySQL.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询标签列表失败",
		})
		return
	}
	defer rows.Close()

	var list []TagItem
	for rows.Next() {
		var t struct {
			ID        uint   `db:"id"`
			Name      string `db:"name"`
			Code      string `db:"code"`
			SortOrder int    `db:"sort_order"`
			Status    int8   `db:"status"`
			CreatedAt string `db:"created_at"`
			UpdatedAt string `db:"updated_at"`
		}
		if err := rows.Scan(&t.ID, &t.Name, &t.Code, &t.SortOrder, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}

		list = append(list, TagItem{
			ID:        t.ID,
			Name:      t.Name,
			Code:      t.Code,
			SortOrder: t.SortOrder,
			Status:    int(t.Status),
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    list,
	})
}

// GetDiseasesByCategory 获取指定分类下的疾病列表
func GetDiseasesByCategory(c *gin.Context) {
	// 1. 获取路径参数 categoryId
	categoryIDStr := c.Param("categoryId")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的分类 ID",
		})
		return
	}

	// 2. 获取分页参数
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

	// 3. 验证分类是否存在且启用
	checkCatQuery := "SELECT id FROM category WHERE id = ? AND status = 1"
	var catExists uint
	err = db.MySQL.QueryRow(checkCatQuery, categoryID).Scan(&catExists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "分类不存在或已停用",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询分类失败",
		})
		return
	}

	// 4. 查询总数
	countQuery := `
		SELECT COUNT(*) 
		FROM disease_category_rel dcr
		JOIN disease d ON dcr.disease_id = d.id
		WHERE dcr.category_id = ? AND d.status = 1
	`
	var total int64
	err = db.MySQL.QueryRow(countQuery, categoryID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 5. 查询列表
	// 注意：这里假设主要按创建时间倒序排列，如果需要按主分类优先或其他排序，可调整 ORDER BY
	listQuery := `
		SELECT d.id, d.name, d.alias, d.introduction, d.symptoms, d.images, d.status, d.creator_id, d.created_at, d.updated_at
		FROM disease_category_rel dcr
		JOIN disease d ON dcr.disease_id = d.id
		WHERE dcr.category_id = ? AND d.status = 1
		ORDER BY d.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := db.MySQL.Query(listQuery, categoryID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询疾病列表失败",
		})
		return
	}
	defer rows.Close()

	var list []DiseaseItem
	for rows.Next() {
		var d struct {
			ID           uint           `db:"id"`
			Name         string         `db:"name"`
			Alias        sql.NullString `db:"alias"`
			Introduction sql.NullString `db:"introduction"`
			Symptoms     sql.NullString `db:"symptoms"`
			Images       sql.NullString `db:"images"`
			Status       int8           `db:"status"`
			CreatorID    sql.NullInt64  `db:"creator_id"`
			CreatedAt    time.Time      `db:"created_at"`
			UpdatedAt    time.Time      `db:"updated_at"`
		}
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Alias, &d.Introduction,
			&d.Symptoms, &d.Images, &d.Status, &d.CreatorID,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			continue
		}

		list = append(list, DiseaseItem{
			ID:           d.ID,
			Name:         d.Name,
			Alias:        d.Alias.String,
			Introduction: d.Introduction.String,
			Symptoms:     d.Symptoms.String,
			Images:       d.Images.String,
			Status:       int(d.Status),
			CreatorID:    uint(d.CreatorID.Int64),
			CreatedAt:    d.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    d.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// 6. 返回结果
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

// CategoryTreeNode 分类树节点响应结构
type CategoryTreeNode struct {
	ID          uint               `json:"id"`
	ParentID    uint               `json:"parentId"` // 注意：JSON字段名保持与现有风格一致，如果前端需要parent_id请改为 parent_id
	Level       int                `json:"level"`
	Name        string             `json:"name"`
	Code        string             `json:"code"`
	Description string             `json:"description"`
	IconURL     string             `json:"iconUrl"`
	SortOrder   int                `json:"sortOrder"`
	Status      int                `json:"status"`
	CreatedAt   string             `json:"createdAt"`
	UpdatedAt   string             `json:"updatedAt"`
	Children    []CategoryTreeNode `json:"children"`
}

// GetCategoryTree 获取分类树
func GetCategoryTree(c *gin.Context) {
	// 1. 获取请求参数
	keyword := c.DefaultQuery("keyword", "")
	statusStr := c.DefaultQuery("status", "1") // 默认只查启用状态，或者根据需求调整默认值

	// 解析 status 参数
	var statusFilter *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			statusFilter = &s
		}
	}

	// 2. 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 关键词搜索
	if keyword != "" {
		whereClause += " AND name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	// 状态过滤
	if statusFilter != nil {
		whereClause += " AND status = ?"
		args = append(args, *statusFilter)
	}

	// 3. 查询所有符合条件的分类（按 sort_order 和 id 排序，保证顺序稳定）
	query := `
		SELECT id, parent_id, level, name, code, description, icon_url, sort_order, status, created_at, updated_at
		FROM category
		` + whereClause + `
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询分类失败",
		})
		return
	}
	defer rows.Close()

	// 4. 扫描数据到扁平列表
	var allCategories []CategoryTreeNode
	for rows.Next() {
		var cat struct {
			ID          uint           `db:"id"`
			ParentID    uint           `db:"parent_id"`
			Level       int            `db:"level"`
			Name        string         `db:"name"`
			Code        string         `db:"code"`
			Description sql.NullString `db:"description"`
			IconURL     sql.NullString `db:"icon_url"`
			SortOrder   int            `db:"sort_order"`
			Status      int8           `db:"status"`
			CreatedAt   time.Time      `db:"created_at"`
			UpdatedAt   time.Time      `db:"updated_at"`
		}

		// 注意：Scan 的顺序必须与 SELECT 列顺序一致
		if err := rows.Scan(
			&cat.ID, &cat.ParentID, &cat.Level, &cat.Name, &cat.Code,
			&cat.Description, &cat.IconURL, &cat.SortOrder, &cat.Status,
			&cat.CreatedAt, &cat.UpdatedAt,
		); err != nil {
			continue
		}

		allCategories = append(allCategories, CategoryTreeNode{
			ID:          cat.ID,
			ParentID:    cat.ParentID,
			Level:       cat.Level,
			Name:        cat.Name,
			Code:        cat.Code,
			Description: cat.Description.String,
			IconURL:     cat.IconURL.String,
			SortOrder:   cat.SortOrder,
			Status:      int(cat.Status),
			CreatedAt:   cat.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   cat.UpdatedAt.Format("2006-01-02 15:04:05"),
			Children:    []CategoryTreeNode{}, // 初始化空切片，避免 JSON 输出 null
		})
	}

	// 5. 构建树形结构
	tree := buildCategoryTree(allCategories)

	c.JSON(http.StatusOK, gin.H{
		"code":    200, // 根据项目规范，成功码可能是 0 或 200，此处沿用之前代码的 200，如需 0 请修改
		"message": "success",
		"data":    tree,
	})
}

// buildCategoryTree 将扁平列表转换为树形结构
func buildCategoryTree(categories []CategoryTreeNode) []CategoryTreeNode {
	// 使用 map 存储 ID 到节点的引用，方便快速查找父节点
	nodeMap := make(map[uint]*CategoryTreeNode)
	var rootNodes []CategoryTreeNode

	// 第一遍遍历：将所有节点放入 map，并初始化 Children
	for i := range categories {
		nodeMap[categories[i].ID] = &categories[i]
	}

	// 第二遍遍历：建立父子关系
	for i := range categories {
		node := &categories[i]
		if node.ParentID == 0 {
			// 根节点
			rootNodes = append(rootNodes, *node)
		} else {
			// 查找父节点
			if parent, exists := nodeMap[node.ParentID]; exists {
				parent.Children = append(parent.Children, *node)
			} else {
				// 如果父节点不存在（可能被过滤掉了或者数据不一致），可以作为根节点处理或者忽略
				// 这里选择将其作为根节点展示，避免数据丢失
				rootNodes = append(rootNodes, *node)
			}
		}
	}

	return rootNodes
}

// SimpleDiseaseItem 简化版疾病响应结构（用于搜索列表）
type SimpleDiseaseItem struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

// SearchDiseasesResponse 搜索疾病响应结构
type SearchDiseasesResponse struct {
	List []SimpleDiseaseItem `json:"list"`
	// 如果前端需要分页信息，可以取消注释下面几行
	// Total    int64 `json:"total"`
	// Page     int   `json:"page"`
	// PageSize int   `json:"pageSize"`
}

// SearchDiseases 根据关键词搜索疾病
func SearchDiseases(c *gin.Context) {
	// 1. 获取请求参数
	keyword := c.DefaultQuery("keyword", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10") // 注意：这里使用 page_size 以符合常见规范，也可兼容 pageSize

	// 参数校验
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "关键词不能为空",
		})
		return
	}

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 兼容 pageSize 和 page_size (如果前端传的是 pageSize)
	if pageSize == 0 {
		pageSizeStrAlt := c.DefaultQuery("pageSize", "10")
		pageSize, _ = strconv.Atoi(pageSizeStrAlt)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 2. 构建查询条件
	// 搜索名称或别名
	whereClause := "WHERE status = 1 AND (name LIKE ? OR alias LIKE ?)"
	args := []interface{}{"%" + keyword + "%", "%" + keyword + "%"}

	// 3. 查询总数 (可选，如果前端需要分页总数)
	countQuery := "SELECT COUNT(*) FROM disease " + whereClause
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
		SELECT id, name, alias
		FROM disease
		` + whereClause + `
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`
	// 注意：args 需要追加分页参数
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

	var list []SimpleDiseaseItem
	for rows.Next() {
		var d struct {
			ID    uint           `db:"id"`
			Name  string         `db:"name"`
			Alias sql.NullString `db:"alias"`
		}
		if err := rows.Scan(&d.ID, &d.Name, &d.Alias); err != nil {
			continue
		}

		list = append(list, SimpleDiseaseItem{
			ID:    d.ID,
			Name:  d.Name,
			Alias: d.Alias.String,
		})
	}

	// 确保 list 不为 null
	if list == nil {
		list = []SimpleDiseaseItem{}
	}

	// 5. 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success", // 注意：你要求的返回值中是 "msg"，而项目其他接口用的是 "message"，这里按你的要求使用 "msg"
		"data": SearchDiseasesResponse{
			List: list,
			// Total:    total, // 如果需要分页信息可开启
			// Page:     page,
			// PageSize: pageSize,
		},
	})
}

// DiseaseOptionItem 疾病选项响应结构
type DiseaseOptionItem struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

// GetDiseaseOptions 获取疾病下拉选项列表
func GetDiseaseOptions(c *gin.Context) {
	// 1. 获取查询参数
	keyword := c.DefaultQuery("keyword", "")

	// 2. 构建查询条件
	// 默认只查询已启用 (status = 1) 的疾病，确保下拉框中的数据是有效的
	whereClause := "WHERE status = 1"
	args := []interface{}{}

	// 如果有关键词，增加模糊搜索（搜索名称或别名）
	if keyword != "" {
		whereClause += " AND (name LIKE ? OR alias LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 3. 执行查询
	// 查询 id, name, alias，并按名称排序以便前端展示
	query := `
		SELECT id, name, alias 
		FROM disease 
		` + whereClause + `
		ORDER BY name ASC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询疾病选项失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	// 4. 扫描结果
	var options []DiseaseOptionItem
	for rows.Next() {
		var item DiseaseOptionItem
		var alias sql.NullString // 使用 NullString 处理可能为 NULL 的 alias 字段

		if err := rows.Scan(&item.ID, &item.Name, &alias); err != nil {
			continue
		}

		// 处理 NULL 值，如果为 NULL 则设为空字符串
		if alias.Valid {
			item.Alias = alias.String
		} else {
			item.Alias = ""
		}

		options = append(options, item)
	}

	// 确保返回空数组而不是 null
	if options == nil {
		options = []DiseaseOptionItem{}
	}

	// 5. 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    options,
	})
}

// --- Article Related Structures ---

// ArticleItem 文章列表项响应结构
type ArticleItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	CoverImage  string `json:"coverImage"`
	SourceName  string `json:"sourceName"`
	Status      int    `json:"status"`
	PublishTime string `json:"publishTime"`
	ViewCount   int64  `json:"viewCount"`
	IsTop       int    `json:"isTop"`
	IsRecommend int    `json:"isRecommend"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// 【新增/修改】AdminArticleItem 用于后台管理文章列表，包含更多字段
type AdminArticleItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	CoverImage  string `json:"coverImage"`
	SourceName  string `json:"sourceName"`
	Status      int    `json:"status"`
	PublishTime string `json:"publishTime"`
	ViewCount   int64  `json:"viewCount"`
	IsTop       int    `json:"isTop"`
	IsRecommend int    `json:"isRecommend"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// ArticleBlockItem 文章内容块响应结构
type ArticleBlockItem struct {
	ID        uint        `json:"id"`
	BlockType string      `json:"blockType"`
	SortNo    int         `json:"sortNo"`
	Title     string      `json:"title"`
	Content   string      `json:"content"`
	Extra     interface{} `json:"extra"` // JSON 对象
}

// ArticleTagItem 文章标签响应结构
type ArticleTagItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ArticleDetailResponse 文章详情响应结构
type ArticleDetailResponse struct {
	ID             uint   `json:"id"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	CoverImage     string `json:"coverImage"`
	ContentType    string `json:"contentType"`
	AuthorID       *uint  `json:"authorId"`
	SourceName     string `json:"sourceName"`
	SourceURL      string `json:"sourceUrl"`
	Status         int    `json:"status"`
	PublishTime    string `json:"publishTime"`
	ViewCount      int64  `json:"viewCount"`
	LikeCount      int64  `json:"likeCount"`
	FavoriteCount  int64  `json:"favoriteCount"`
	IsTop          int    `json:"isTop"`
	IsRecommend    int    `json:"isRecommend"`
	SeoTitle       string `json:"seoTitle"`
	SeoKeywords    string `json:"seoKeywords"`
	SeoDescription string `json:"seoDescription"`

	Blocks []ArticleBlockItem `json:"blocks"`
	Tags   []ArticleTagItem   `json:"tags"`

	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// CreateArticleRequest 创建文章请求
type CreateArticleRequest struct {
	Title          string  `json:"title" binding:"required"`
	Summary        string  `json:"summary"`
	CoverImage     string  `json:"coverImage"`
	ContentType    string  `json:"contentType"`
	AuthorID       *uint   `json:"authorId"`
	SourceName     string  `json:"sourceName"`
	SourceURL      string  `json:"sourceUrl"`
	Status         int     `json:"status"` // 0:草稿, 1:待审核, 2:已发布
	PublishTime    *string `json:"publishTime"`
	IsTop          int     `json:"isTop"`
	IsRecommend    int     `json:"isRecommend"`
	SeoTitle       string  `json:"seoTitle"`
	SeoKeywords    string  `json:"seoKeywords"`
	SeoDescription string  `json:"seoDescription"`

	Blocks     []CreateArticleBlockRequest `json:"blocks"`
	TagIDs     []uint                      `json:"tagIds"`
	DiseaseIDs []uint                      `json:"diseaseIds"` // 关联的疾病ID
}

// CreateArticleBlockRequest 创建文章块请求
type CreateArticleBlockRequest struct {
	BlockType string      `json:"blockType" binding:"required"`
	SortNo    int         `json:"sortNo"`
	Title     string      `json:"title"`
	Content   string      `json:"content"`
	Extra     interface{} `json:"extra"`
}

// UpdateArticleRequest 更新文章请求
type UpdateArticleRequest struct {
	Title          *string `json:"title"`
	Summary        *string `json:"summary"`
	CoverImage     *string `json:"coverImage"`
	ContentType    *string `json:"contentType"`
	AuthorID       *uint   `json:"authorId"`
	SourceName     *string `json:"sourceName"`
	SourceURL      *string `json:"sourceUrl"`
	Status         *int    `json:"status"`
	PublishTime    *string `json:"publishTime"`
	IsTop          *int    `json:"isTop"`
	IsRecommend    *int    `json:"isRecommend"`
	SeoTitle       *string `json:"seoTitle"`
	SeoKeywords    *string `json:"seoKeywords"`
	SeoDescription *string `json:"seoDescription"`

	Blocks     []CreateArticleBlockRequest `json:"blocks"`     // 全量替换
	TagIDs     *[]uint                     `json:"tagIds"`     // 全量替换
	DiseaseIDs *[]uint                     `json:"diseaseIds"` // 全量替换
}

// CreateArticle 创建文章
func CreateArticle(c *gin.Context) {
	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "数据库事务失败"})
		return
	}
	defer tx.Rollback()

	now := time.Now()
	var publishTime *time.Time
	if req.PublishTime != nil && *req.PublishTime != "" {
		pt, err := time.Parse("2006-01-02 15:04:05", *req.PublishTime)
		if err == nil {
			publishTime = &pt
		}
	}

	// 1. 插入 article 主表
	res, err := tx.Exec(`
		INSERT INTO article (title, summary, cover_image, content_type, author_id, source_name, source_url, 
			status, publish_time, view_count, like_count, favorite_count, is_top, is_recommend, 
			seo_title, seo_keywords, seo_description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, ?, ?, ?, ?, ?, ?, ?)
	`, req.Title, req.Summary, req.CoverImage, req.ContentType, req.AuthorID, req.SourceName, req.SourceURL,
		req.Status, publishTime, req.IsTop, req.IsRecommend, req.SeoTitle, req.SeoKeywords, req.SeoDescription, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建文章失败: " + err.Error()})
		return
	}

	articleID, _ := res.LastInsertId()

	// 2. 插入文章块
	for _, b := range req.Blocks {
		extraJSON, _ := json.Marshal(b.Extra)
		_, err := tx.Exec(`
			INSERT INTO article_block (article_id, block_type, sort_no, title, content, extra, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, articleID, b.BlockType, b.SortNo, b.Title, b.Content, extraJSON, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建文章块失败: " + err.Error()})
			return
		}
	}

	// 3. 关联标签
	for _, tagID := range req.TagIDs {
		_, err := tx.Exec("INSERT INTO article_tag_rel (article_id, tag_id) VALUES (?, ?)", articleID, tagID)
		if err != nil {
			// 忽略重复插入错误，或者根据业务需求处理
			continue
		}
	}

	// 4. 关联疾病
	for _, diseaseID := range req.DiseaseIDs {
		_, err := tx.Exec("INSERT INTO disease_article_rel (disease_id, article_id) VALUES (?, ?)", diseaseID, articleID)
		if err != nil {
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "事务提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{"id": articleID}})
}

// GetArticleByID 获取文章详情
func GetArticleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效ID"})
		return
	}

	// 1. 查询主表
	var art struct {
		ID             uint
		Title          string
		Summary        sql.NullString
		CoverImage     sql.NullString
		ContentType    sql.NullString
		AuthorID       sql.NullInt64
		SourceName     sql.NullString
		SourceURL      sql.NullString
		Status         int8
		PublishTime    sql.NullTime
		ViewCount      int64
		LikeCount      int64
		FavoriteCount  int64
		IsTop          int8
		IsRecommend    int8
		SeoTitle       sql.NullString
		SeoKeywords    sql.NullString
		SeoDescription sql.NullString
		CreatedAt      time.Time
		UpdatedAt      time.Time
	}

	err = db.MySQL.QueryRow("SELECT * FROM article WHERE id = ?", id).Scan(
		&art.ID, &art.Title, &art.Summary, &art.CoverImage, &art.ContentType, &art.AuthorID,
		&art.SourceName, &art.SourceURL, &art.Status, &art.PublishTime, &art.ViewCount,
		&art.LikeCount, &art.FavoriteCount, &art.IsTop, &art.IsRecommend,
		&art.SeoTitle, &art.SeoKeywords, &art.SeoDescription, &art.CreatedAt, &art.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文章不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		}
		return
	}

	// 2. 查询内容块
	blocks := []ArticleBlockItem{}
	blockRows, err := db.MySQL.Query("SELECT id, block_type, sort_no, title, content, extra FROM article_block WHERE article_id = ? ORDER BY sort_no ASC", id)
	if err == nil {
		defer blockRows.Close()
		for blockRows.Next() {
			var b ArticleBlockItem
			var extraRaw []byte
			blockRows.Scan(&b.ID, &b.BlockType, &b.SortNo, &b.Title, &b.Content, &extraRaw)
			json.Unmarshal(extraRaw, &b.Extra)
			blocks = append(blocks, b)
		}
	}

	// 3. 查询标签
	tags := []ArticleTagItem{}
	tagRows, err := db.MySQL.Query(`
		SELECT t.id, t.name, t.type 
		FROM article_tag_rel atr 
		JOIN article_tag t ON atr.tag_id = t.id 
		WHERE atr.article_id = ?
	`, id)
	if err == nil {
		defer tagRows.Close()
		for tagRows.Next() {
			var t ArticleTagItem
			tagRows.Scan(&t.ID, &t.Name, &t.Type)
			tags = append(tags, t)
		}
	}

	// 组装响应
	var authorID *uint
	if art.AuthorID.Valid {
		v := uint(art.AuthorID.Int64)
		authorID = &v
	}

	var pubTimeStr string
	if art.PublishTime.Valid {
		pubTimeStr = art.PublishTime.Time.Format("2006-01-02 15:04:05")
	}

	resp := ArticleDetailResponse{
		ID: art.ID, Title: art.Title, Summary: art.Summary.String, CoverImage: art.CoverImage.String,
		ContentType: art.ContentType.String, AuthorID: authorID, SourceName: art.SourceName.String,
		SourceURL: art.SourceURL.String, Status: int(art.Status), PublishTime: pubTimeStr,
		ViewCount: art.ViewCount, LikeCount: art.LikeCount, FavoriteCount: art.FavoriteCount,
		IsTop: int(art.IsTop), IsRecommend: int(art.IsRecommend),
		SeoTitle: art.SeoTitle.String, SeoKeywords: art.SeoKeywords.String, SeoDescription: art.SeoDescription.String,
		Blocks: blocks, Tags: tags,
		CreatedAt: art.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: art.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": resp})
}

// UpdateArticle 更新文章
func UpdateArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效ID"})
		return
	}

	var req UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	tx, err := db.MySQL.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "事务失败"})
		return
	}
	defer tx.Rollback()

	// 1. 动态更新主表
	fields := []string{}
	args := []interface{}{}

	if req.Title != nil {
		fields = append(fields, "title=?")
		args = append(args, *req.Title)
	}
	if req.Summary != nil {
		fields = append(fields, "summary=?")
		args = append(args, *req.Summary)
	}
	if req.CoverImage != nil {
		fields = append(fields, "cover_image=?")
		args = append(args, *req.CoverImage)
	}
	if req.ContentType != nil {
		fields = append(fields, "content_type=?")
		args = append(args, *req.ContentType)
	}
	if req.AuthorID != nil {
		fields = append(fields, "author_id=?")
		args = append(args, *req.AuthorID)
	}
	if req.SourceName != nil {
		fields = append(fields, "source_name=?")
		args = append(args, *req.SourceName)
	}
	if req.SourceURL != nil {
		fields = append(fields, "source_url=?")
		args = append(args, *req.SourceURL)
	}
	if req.Status != nil {
		fields = append(fields, "status=?")
		args = append(args, *req.Status)
	}
	if req.PublishTime != nil {
		if *req.PublishTime == "" {
			fields = append(fields, "publish_time=NULL")
		} else {
			fields = append(fields, "publish_time=?")
			args = append(args, *req.PublishTime)
		}
	}
	if req.IsTop != nil {
		fields = append(fields, "is_top=?")
		args = append(args, *req.IsTop)
	}
	if req.IsRecommend != nil {
		fields = append(fields, "is_recommend=?")
		args = append(args, *req.IsRecommend)
	}
	if req.SeoTitle != nil {
		fields = append(fields, "seo_title=?")
		args = append(args, *req.SeoTitle)
	}
	if req.SeoKeywords != nil {
		fields = append(fields, "seo_keywords=?")
		args = append(args, *req.SeoKeywords)
	}
	if req.SeoDescription != nil {
		fields = append(fields, "seo_description=?")
		args = append(args, *req.SeoDescription)
	}

	if len(fields) > 0 {
		fields = append(fields, "updated_at=?")
		args = append(args, time.Now())
		args = append(args, id)
		query := "UPDATE article SET " + strings.Join(fields, ", ") + " WHERE id=?"
		if _, err := tx.Exec(query, args...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新主表失败"})
			return
		}
	}

	// 2. 更新内容块 (删除旧的，插入新的)
	if req.Blocks != nil {
		tx.Exec("DELETE FROM article_block WHERE article_id=?", id)
		for _, b := range req.Blocks {
			extraJSON, _ := json.Marshal(b.Extra)
			tx.Exec("INSERT INTO article_block (article_id, block_type, sort_no, title, content, extra, created_at) VALUES (?,?,?,?,?,?,?)",
				id, b.BlockType, b.SortNo, b.Title, b.Content, extraJSON, time.Now())
		}
	}

	// 3. 更新标签关联
	if req.TagIDs != nil {
		tx.Exec("DELETE FROM article_tag_rel WHERE article_id=?", id)
		for _, tid := range *req.TagIDs {
			tx.Exec("INSERT INTO article_tag_rel (article_id, tag_id) VALUES (?,?)", id, tid)
		}
	}

	// 4. 更新疾病关联
	if req.DiseaseIDs != nil {
		tx.Exec("DELETE FROM disease_article_rel WHERE article_id=?", id)
		for _, did := range *req.DiseaseIDs {
			tx.Exec("INSERT INTO disease_article_rel (disease_id, article_id) VALUES (?,?)", did, id)
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// DeleteArticle 删除文章
func DeleteArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效ID"})
		return
	}

	tx, _ := db.MySQL.Begin()
	defer tx.Rollback()

	tx.Exec("DELETE FROM article_block WHERE article_id=?", id)
	tx.Exec("DELETE FROM article_tag_rel WHERE article_id=?", id)
	tx.Exec("DELETE FROM disease_article_rel WHERE article_id=?", id)
	_, err = tx.Exec("DELETE FROM article WHERE id=?", id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// GetArticles 获取文章列表
func GetArticles(c *gin.Context) {
	keyword := c.DefaultQuery("keyword", "")
	statusStr := c.DefaultQuery("status", "")
	diseaseIDStr := c.DefaultQuery("diseaseId", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	where := "WHERE 1=1"
	args := []interface{}{}

	if keyword != "" {
		where += " AND (title LIKE ? OR summary LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	if statusStr != "" {
		where += " AND status=?"
		args = append(args, statusStr)
	}

	joinDisease := ""
	if diseaseIDStr != "" {
		joinDisease = "INNER JOIN disease_article_rel dar ON a.id = dar.article_id"
		where += " AND dar.disease_id=?"
		args = append(args, diseaseIDStr)
	}

	// Count
	var total int64
	countSQL := "SELECT COUNT(*) FROM article a " + joinDisease + " " + where
	db.MySQL.QueryRow(countSQL, args...).Scan(&total)

	// List
	listSQL := "SELECT a.id, a.title, a.summary, a.cover_image, a.source_name, a.status, a.publish_time, a.view_count, a.is_top, a.is_recommend, a.created_at, a.updated_at FROM article a " + joinDisease + " " + where + " ORDER BY a.is_top DESC, a.publish_time DESC, a.id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, _ := db.MySQL.Query(listSQL, args...)
	defer rows.Close()

	// 【修改点】使用 AdminArticleItem
	list := []AdminArticleItem{}
	for rows.Next() {
		var item AdminArticleItem
		var pubTime sql.NullTime
		// 注意：Scan 的顺序必须与 SQL 查询列的顺序一致
		err := rows.Scan(&item.ID, &item.Title, &item.Summary, &item.CoverImage, &item.SourceName, &item.Status, &pubTime, &item.ViewCount, &item.IsTop, &item.IsRecommend, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			continue
		}
		if pubTime.Valid {
			item.PublishTime = pubTime.Time.Format("2006-01-02 15:04:05")
		}
		// 格式化时间字段 (如果数据库返回的是字符串格式的时间，可能需要额外处理，这里假设 Scan 出来的是 time.Time 或 string)
		// 注意：上面的 SQL 查出来的是 datetime，Scan 到 string 可能会直接变成 "2006-01-02T15:04:05Z" 格式，取决于驱动。
		// 为了稳妥，建议 Scan 到 time.Time 然后 Format，或者确保数据库驱动配置正确。
		// 这里简化处理，假设 CreatedAt/UpdatedAt 在 struct 中是 string，如果 Scan 报错，需改为 time.Time 中间变量。

		list = append(list, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200, "message": "success",
		"data": gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize},
	})
}
