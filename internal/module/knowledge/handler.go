package knowledge

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"rare_backend/internal/pkg/db"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateDisease 创建疾病
func CreateDisease(c *gin.Context) {
	var req struct {
		Name         string   `json:"name" binding:"required"`
		Alias        string   `json:"alias"`
		CategoryID   int64    `json:"category_id" binding:"required"`
		Introduction string   `json:"introduction"`
		Symptoms     string   `json:"symptoms"`
		Guidelines   string   `json:"guidelines"`
		Medications  string   `json:"medications"`
		Experiences  string   `json:"experiences"`
		Images       []string `json:"images"`
		Status       int      `json:"status" default:"1"`
		CreatorID    int64    `json:"creator_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 检查分类是否存在
	var count int
	checkQuery := "SELECT COUNT(*) FROM category WHERE id = ? AND status = 1"
	err := db.MySQL.QueryRow(checkQuery, req.CategoryID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "无效的分类ID",
		})
		return
	}

	// 插入疾病数据
	imagesJSON, _ := json.Marshal(req.Images)
	query := `
		INSERT INTO disease (
			name, alias, category_id, introduction, symptoms, guidelines, medications, experiences, images, status, creator_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := db.MySQL.Exec(
		query,
		req.Name, req.Alias, req.CategoryID, req.Introduction, req.Symptoms,
		req.Guidelines, req.Medications, req.Experiences, string(imagesJSON),
		req.Status, req.CreatorID,
	)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "创建疾病失败",
		})
		return
	}

	diseaseID, _ := res.LastInsertId()
	c.JSON(200, gin.H{
		"code":    200,
		"message": "创建成功",
		"data": gin.H{
			"disease_id": diseaseID,
		},
	})
}

// UpdateDisease 更新疾病
func UpdateDisease(c *gin.Context) {
	// 获取疾病 ID
	diseaseID := c.Param("id")
	id, err := strconv.ParseInt(diseaseID, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "无效的疾病 ID",
		})
		return
	}

	// 定义请求结构体（支持部分字段更新）
	var req struct {
		Name         *string   `json:"name"`
		Alias        *string   `json:"alias"`
		CategoryID   *int64    `json:"category_id"`
		Introduction *string   `json:"introduction"`
		Symptoms     *string   `json:"symptoms"`
		Guidelines   *string   `json:"guidelines"`
		Medications  *string   `json:"medications"`
		Experiences  *string   `json:"experiences"`
		Images       *[]string `json:"images"`
		Status       *int      `json:"status"`
	}

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 验证 category_id 是否存在（如果提供了该字段）
	if req.CategoryID != nil {
		var count int
		checkQuery := "SELECT COUNT(*) FROM category WHERE id = ? AND status = 1"
		err := db.MySQL.QueryRow(checkQuery, *req.CategoryID).Scan(&count)
		if err != nil || count == 0 {
			c.JSON(400, gin.H{
				"code":    400,
				"message": "无效的分类ID",
			})
			return
		}
	}

	// 构建动态更新 SQL
	setClause := ""
	args := []interface{}{}

	if req.Name != nil {
		setClause += "name = ?, "
		args = append(args, *req.Name)
	}
	if req.Alias != nil {
		setClause += "alias = ?, "
		args = append(args, *req.Alias)
	}
	if req.CategoryID != nil {
		setClause += "category_id = ?, "
		args = append(args, *req.CategoryID)
	}
	if req.Introduction != nil {
		setClause += "introduction = ?, "
		args = append(args, *req.Introduction)
	}
	if req.Symptoms != nil {
		setClause += "symptoms = ?, "
		args = append(args, *req.Symptoms)
	}
	if req.Guidelines != nil {
		setClause += "guidelines = ?, "
		args = append(args, *req.Guidelines)
	}
	if req.Medications != nil {
		setClause += "medications = ?, "
		args = append(args, *req.Medications)
	}
	if req.Experiences != nil {
		setClause += "experiences = ?, "
		args = append(args, *req.Experiences)
	}
	if req.Images != nil {
		imagesJSON, _ := json.Marshal(*req.Images)
		setClause += "images = ?, "
		args = append(args, string(imagesJSON))
	}
	if req.Status != nil {
		setClause += "status = ?, "
		args = append(args, *req.Status)
	}

	// 移除末尾逗号并拼接完整 SQL
	if len(setClause) > 0 {
		setClause = setClause[:len(setClause)-2]
	}
	query := fmt.Sprintf("UPDATE disease SET %s WHERE id = ?", setClause)
	args = append(args, id)

	// 执行更新
	_, err = db.MySQL.Exec(query, args...)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "更新疾病失败",
		})
		return
	}

	// 返回成功响应
	c.JSON(200, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// GetDiseases 获取疾病列表
func GetDiseases(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	offset := (page - 1) * size

	// 获取筛选参数
	categoryID := c.Query("category_id")
	status := c.Query("status")

	// 构建查询条件
	whereClause := ""
	args := []interface{}{}

	if categoryID != "" {
		whereClause += " AND d.category_id = ?"
		args = append(args, categoryID)
	}
	if status != "" {
		whereClause += " AND d.status = ?"
		args = append(args, status)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM disease d WHERE 1=1" + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 查询列表（关联 category 表获取分类名称）
	listQuery := `
		SELECT 
			d.id, d.name, d.alias, d.category_id, c.name AS category_name,
			d.introduction, d.symptoms, d.guidelines, d.medications, d.experiences,
			d.images, d.status, d.creator_id, d.created_at, d.updated_at
		FROM disease d
		LEFT JOIN category c ON d.category_id = c.id
		WHERE 1=1` + whereClause + `
		ORDER BY d.created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, size, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "查询列表失败",
		})
		return
	}
	defer rows.Close()

	// 解析结果
	var diseases []map[string]interface{}
	for rows.Next() {
		var disease struct {
			ID           int64          `db:"id"`
			Name         string         `db:"name"`
			Alias        sql.NullString `db:"alias"`
			CategoryID   sql.NullInt64  `db:"category_id"`
			CategoryName sql.NullString `db:"category_name"`
			Introduction sql.NullString `db:"introduction"`
			Symptoms     sql.NullString `db:"symptoms"`
			Guidelines   sql.NullString `db:"guidelines"`
			Medications  sql.NullString `db:"medications"`
			Experiences  sql.NullString `db:"experiences"`
			Images       []byte         `db:"images"`
			Status       int            `db:"status"`
			CreatorID    sql.NullInt64  `db:"creator_id"`
			CreatedAt    time.Time      `db:"created_at"`
			UpdatedAt    time.Time      `db:"updated_at"`
		}

		if err := rows.Scan(
			&disease.ID, &disease.Name, &disease.Alias, &disease.CategoryID,
			&disease.CategoryName, &disease.Introduction, &disease.Symptoms,
			&disease.Guidelines, &disease.Medications, &disease.Experiences,
			&disease.Images, &disease.Status, &disease.CreatorID,
			&disease.CreatedAt, &disease.UpdatedAt,
		); err != nil {
			continue
		}

		// 处理 JSON 字段
		var images []string
		json.Unmarshal(disease.Images, &images)

		diseases = append(diseases, map[string]interface{}{
			"id":            disease.ID,
			"name":          disease.Name,
			"alias":         disease.Alias.String,
			"category_id":   disease.CategoryID.Int64,
			"category_name": disease.CategoryName.String,
			"introduction":  disease.Introduction.String,
			"symptoms":      disease.Symptoms.String,
			"guidelines":    disease.Guidelines.String,
			"medications":   disease.Medications.String,
			"experiences":   disease.Experiences.String,
			"images":        images,
			"status":        disease.Status,
			"creator_id":    disease.CreatorID.Int64,
			"created_at":    disease.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":    disease.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// 返回响应
	c.JSON(200, gin.H{
		"code":    200,
		"message": "查询成功",
		"data": gin.H{
			"list":  diseases,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// CreateCategory 创建分类
func CreateCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IconURL     string `json:"icon_url"`
		SortOrder   int    `json:"sort_order"`
		Status      int    `json:"status" default:"1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	query := `
		INSERT INTO category (name, description, icon_url, sort_order, status)
		VALUES (?, ?, ?, ?, ?)
	`
	res, err := db.MySQL.Exec(query, req.Name, req.Description, req.IconURL, req.SortOrder, req.Status)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "创建分类失败",
		})
		return
	}

	categoryID, _ := res.LastInsertId()
	c.JSON(200, gin.H{
		"code":    200,
		"message": "创建成功",
		"data": gin.H{
			"category_id": categoryID,
		},
	})
}

// GetCategories 获取分类列表
func GetCategories(c *gin.Context) {
	query := `
		SELECT id, name, description, icon_url, sort_order, status
		FROM category
		WHERE status = 1
		ORDER BY sort_order ASC
	`
	rows, err := db.MySQL.Query(query)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "查询分类失败",
		})
		return
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var category struct {
			ID          int64  `db:"id"`
			Name        string `db:"name"`
			Description string `db:"description"`
			IconURL     string `db:"icon_url"`
			SortOrder   int    `db:"sort_order"`
			Status      int    `db:"status"`
		}
		if err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.IconURL, &category.SortOrder, &category.Status); err != nil {
			continue
		}
		categories = append(categories, map[string]interface{}{
			"id":          category.ID,
			"name":        category.Name,
			"description": category.Description,
			"icon_url":    category.IconURL,
			"sort_order":  category.SortOrder,
			"status":      category.Status,
		})
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "查询成功",
		"data": gin.H{
			"categories": categories,
		},
	})
}

func GetDiseasesByCategory(c *gin.Context) {
	// 从 URL 参数中获取 categoryId
	categoryId := c.Param("categoryId")

	// 验证 categoryId 是否为有效整数
	_, err := strconv.ParseInt(categoryId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的分类ID",
		})
		return
	}

	// 查询分类名称
	var categoryName string
	categoryQuery := "SELECT name FROM category WHERE id = ? AND status = 1"
	err = db.MySQL.QueryRow(categoryQuery, categoryId).Scan(&categoryName)
	if err != nil {
		fmt.Printf("Category query error: %v\n", err) // 打印错误日志
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询分类失败",
		})
		return
	}

	// 查询该分类下的疾病列表
	diseaseQuery := `
		SELECT id, name, introduction
		FROM disease
		WHERE category_id = ? AND status = 1
		ORDER BY created_at DESC
	`
	rows, err := db.MySQL.Query(diseaseQuery, categoryId)
	if err != nil {
		fmt.Printf("Disease query error: %v\n", err) // 打印错误日志
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询疾病列表失败",
		})
		return
	}
	defer rows.Close()

	// 解析结果
	var diseases []map[string]interface{}
	for rows.Next() {
		var disease struct {
			ID           int64  `db:"id"`
			Name         string `db:"name"`
			Introduction string `db:"introduction"`
		}
		if err := rows.Scan(&disease.ID, &disease.Name, &disease.Introduction); err != nil {
			fmt.Printf("Scan error: %v\n", err) // 打印扫描错误
			continue
		}
		diseases = append(diseases, map[string]interface{}{
			"id":           disease.ID,
			"name":         disease.Name,
			"introduction": disease.Introduction,
		})
	}

	// 构造响应数据
	response := gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"categoryName": categoryName,
			"diseases":     diseases,
		},
	}

	// 返回结果
	c.JSON(http.StatusOK, response)
}

// GetDiseaseByID 根据疾病 ID 获取疾病详情
func GetDiseaseByID(c *gin.Context) {
	// 从 URL 参数中获取疾病 ID
	diseaseID := c.Param("id")

	// 验证疾病 ID 是否为有效整数
	id, err := strconv.ParseInt(diseaseID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的疾病 ID",
		})
		return
	}

	// 查询疾病详情
	query := `
		SELECT name, introduction, symptoms, guidelines, medications, experiences, images
		FROM disease
		WHERE id = ? AND status = 1
	`
	var disease struct {
		Name         string         `db:"name"`
		Introduction sql.NullString `db:"introduction"`
		Symptoms     sql.NullString `db:"symptoms"`
		Guidelines   sql.NullString `db:"guidelines"`
		Medications  sql.NullString `db:"medications"`
		Experiences  sql.NullString `db:"experiences"`
		Images       []byte         `db:"images"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&disease.Name,
		&disease.Introduction,
		&disease.Symptoms,
		&disease.Guidelines,
		&disease.Medications,
		&disease.Experiences,
		&disease.Images,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "疾病不存在",
			})
			return
		}
		fmt.Printf("Disease query error: %v\n", err) // 打印错误日志
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询疾病详情失败",
		})
		return
	}

	// 处理 JSON 字段
	var images []string
	if len(disease.Images) > 0 {
		json.Unmarshal(disease.Images, &images)
	}

	// 构造响应数据
	response := gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"name":         disease.Name,
			"introduction": disease.Introduction.String,
			"symptoms":     disease.Symptoms.String,
			"guidelines":   disease.Guidelines.String,
			"medications":  disease.Medications.String,
			"experiences":  disease.Experiences.String,
			"images":       images,
		},
	}

	// 返回结果
	c.JSON(http.StatusOK, response)
}

// LikePost 点赞帖子
func LikePost(c *gin.Context) {
	// 从 URL 参数中获取帖子 ID
	postID := c.Param("id")

	// 验证帖子 ID 是否为有效整数
	id, err := strconv.ParseInt(postID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的帖子 ID",
		})
		return
	}

	// 查询帖子是否存在
	var count int
	checkQuery := "SELECT COUNT(*) FROM post WHERE id = ? AND status = 1"
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "帖子不存在",
		})
		return
	}

	// 假设用户 ID 从 JWT 或上下文中获取（此处简化为固定值）
	userID := int64(1001) // 实际开发中应从认证中间件获取

	// 查询用户是否已点赞
	var isLiked bool
	likeQuery := "SELECT COUNT(*) FROM post_like WHERE post_id = ? AND user_id = ?"
	err = db.MySQL.QueryRow(likeQuery, id, userID).Scan(&count)
	if err != nil {
		fmt.Printf("Like query error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询点赞状态失败",
		})
		return
	}
	isLiked = count > 0

	// 更新点赞状态
	if isLiked {
		// 取消点赞
		_, err = db.MySQL.Exec("DELETE FROM post_like WHERE post_id = ? AND user_id = ?", id, userID)
		if err != nil {
			fmt.Printf("Unlike error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "取消点赞失败",
			})
			return
		}
		isLiked = false
	} else {
		// 点赞
		_, err = db.MySQL.Exec("INSERT INTO post_like (post_id, user_id) VALUES (?, ?)", id, userID)
		if err != nil {
			fmt.Printf("Like error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "点赞失败",
			})
			return
		}
		isLiked = true
	}

	// 更新帖子点赞数
	var likeCount int
	updateLikeCountQuery := `
		UPDATE post 
		SET like_count = (SELECT COUNT(*) FROM post_like WHERE post_id = ?)
		WHERE id = ?
	`
	_, err = db.MySQL.Exec(updateLikeCountQuery, id, id)
	if err != nil {
		fmt.Printf("Update like count error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新点赞数失败",
		})
		return
	}

	// 查询最新的点赞数
	countQuery := "SELECT like_count FROM post WHERE id = ?"
	err = db.MySQL.QueryRow(countQuery, id).Scan(&likeCount)
	if err != nil {
		fmt.Printf("Query like count error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询点赞数失败",
		})
		return
	}

	// 构造响应数据
	response := gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"is_liked":   isLiked,
			"like_count": likeCount,
		},
	}

	// 返回结果
	c.JSON(http.StatusOK, response)
}

// GetDiseaseTree 获取疾病分类树
func GetDiseaseTree(c *gin.Context) {
	// 查询所有分类
	categoryQuery := "SELECT id, name FROM category WHERE status = 1"
	categoryRows, err := db.MySQL.Query(categoryQuery)
	if err != nil {
		fmt.Printf("Category query error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询分类失败",
		})
		return
	}
	defer categoryRows.Close()

	// 解析分类数据
	var categories []map[string]interface{}
	for categoryRows.Next() {
		var category struct {
			ID   int64  `db:"id"`
			Name string `db:"name"`
		}
		if err := categoryRows.Scan(&category.ID, &category.Name); err != nil {
			fmt.Printf("Category scan error: %v\n", err)
			continue
		}

		// 查询该分类下的疾病
		diseaseQuery := "SELECT id, name FROM disease WHERE category_id = ? AND status = 1"
		diseaseRows, err := db.MySQL.Query(diseaseQuery, category.ID)
		if err != nil {
			fmt.Printf("Disease query error: %v\n", err)
			continue
		}
		defer diseaseRows.Close()

		// 解析疾病数据
		var children []map[string]interface{}
		for diseaseRows.Next() {
			var disease struct {
				ID   int64  `db:"id"`
				Name string `db:"name"`
			}
			if err := diseaseRows.Scan(&disease.ID, &disease.Name); err != nil {
				fmt.Printf("Disease scan error: %v\n", err)
				continue
			}
			children = append(children, map[string]interface{}{
				"id":   disease.ID,
				"name": disease.Name,
			})
		}

		// 构造分类数据
		categories = append(categories, map[string]interface{}{
			"id":       category.ID,
			"name":     category.Name,
			"children": children,
		})
	}

	// 构造响应数据
	response := gin.H{
		"code":    200,
		"message": "success",
		"data":    categories,
	}

	// 返回结果
	c.JSON(http.StatusOK, response)
}
