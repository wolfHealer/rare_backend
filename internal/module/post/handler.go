package post

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

// GetCommunityPosts 获取社区帖子列表
// GetCommunityPosts 获取社区帖子列表
func GetCommunityPosts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// 获取筛选参数
	sort := c.DefaultQuery("sort", "latest") // 默认按最新排序
	postType := c.Query("type")              // 帖子类型
	diseaseID := c.Query("disease_id")       // 疾病类型
	categoryID := c.Query("category_id")     // 分类 ID

	// 构建查询条件
	whereClause := "cp.status = 1"
	args := []interface{}{}

	// 按帖子类型筛选
	if postType != "" {
		whereClause += " AND cp.type = ?"
		args = append(args, postType)
	}

	// 按疾病类型筛选
	if diseaseID != "" {
		whereClause += " AND cp.disease_id = ?"
		args = append(args, diseaseID)
	}

	// 按分类 ID 筛选
	if categoryID != "" {
		whereClause += " AND cp.category_id = ?"
		args = append(args, categoryID)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM post cp WHERE " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		fmt.Printf("Count query error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 构建排序逻辑
	orderClause := ""
	switch sort {
	case "latest":
		orderClause = "cp.created_at DESC"
	case "hot":
		orderClause = "cp.like_count DESC"
	default:
		orderClause = "cp.created_at DESC" // 默认按最新排序
	}

	// 查询帖子列表
	listQuery := `
		SELECT 
			cp.id, cp.user_id, u.display_name, cp.content, cp.images, 
			cp.created_at, cp.like_count, cp.comment_count
		FROM post cp
		LEFT JOIN user u ON cp.user_id = u.id
		WHERE ` + whereClause + `
		ORDER BY ` + orderClause + `
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		fmt.Printf("List query error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询帖子列表失败",
		})
		return
	}
	defer rows.Close()

	// 解析结果
	var records []map[string]interface{}
	for rows.Next() {
		var post struct {
			ID           int64          `db:"id"`
			UserID       int64          `db:"user_id"`
			DisplayName  sql.NullString `db:"display_name"`
			Content      string         `db:"content"`
			Images       []byte         `db:"images"`
			CreatedAt    time.Time      `db:"created_at"`
			LikeCount    int            `db:"like_count"`
			CommentCount int            `db:"comment_count"`
		}
		if err := rows.Scan(
			&post.ID, &post.UserID, &post.DisplayName, &post.Content,
			&post.Images, &post.CreatedAt, &post.LikeCount, &post.CommentCount,
		); err != nil {
			fmt.Printf("Scan error: %v\n", err)
			continue
		}

		// 处理 JSON 字段
		var images []string
		if len(post.Images) > 0 {
			json.Unmarshal(post.Images, &images)
		}

		// 默认用户未点赞
		isLiked := false

		records = append(records, map[string]interface{}{
			"id":            post.ID,
			"user_id":       post.UserID,
			"display_name":  post.DisplayName.String,
			"content":       post.Content,
			"images":        images,
			"created_at":    post.CreatedAt.Format(time.RFC3339),
			"like_count":    post.LikeCount,
			"comment_count": post.CommentCount,
			"is_liked":      isLiked,
		})
	}

	// 构造响应数据
	response := gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"records": records,
			"total":   total,
			"page":    page,
			"limit":   limit,
		},
	}

	// 返回结果
	c.JSON(http.StatusOK, response)
}

// CreatePost 创建帖子
// CreatePost 创建帖子
func CreatePost(c *gin.Context) {
	// 定义请求结构体
	var req struct {
		UserID     int64    `json:"user_id" binding:"required"`
		DiseaseID  *int64   `json:"disease_id"`
		CategoryID *int64   `json:"category_id"`             // 新增分类 ID 字段
		Type       string   `json:"type" binding:"required"` // help/experience/emotion/info
		Title      *string  `json:"title"`
		Content    string   `json:"content" binding:"required"`
		Images     []string `json:"images"`
	}

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 验证 type 字段是否合法
	validTypes := map[string]bool{
		"help":       true,
		"experience": true,
		"emotion":    true,
		"info":       true,
	}
	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的帖子类型",
		})
		return
	}

	// 处理 images 字段
	imagesJSON, _ := json.Marshal(req.Images)

	// 插入帖子数据
	query := `
		INSERT INTO post (
			user_id, disease_id, category_id, type, title, content, images, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, 0)
	`
	res, err := db.MySQL.Exec(
		query,
		req.UserID, req.DiseaseID, req.CategoryID, req.Type, req.Title, req.Content, string(imagesJSON),
	)
	if err != nil {
		fmt.Printf("Insert post error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建帖子失败",
		})
		return
	}

	// 获取插入的帖子 ID
	postID, _ := res.LastInsertId()

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data": gin.H{
			"post_id": postID,
		},
	})
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
	likeQuery := "SELECT COUNT(*) FROM post_like WHERE target_id = ? AND user_id = ?"
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

// handler.go

// GetPostComments 获取帖子评论列表
func GetPostComments(c *gin.Context) {
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

	// 查询评论列表
	query := `
		SELECT 
			c.id, c.user_id, u.display_name, c.content, c.created_at
		FROM comment c
		LEFT JOIN user u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`
	rows, err := db.MySQL.Query(query, id)
	if err != nil {
		fmt.Printf("Query comments error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询评论失败",
		})
		return
	}
	defer rows.Close()

	// 解析结果
	var comments []map[string]interface{}
	for rows.Next() {
		var comment struct {
			ID          int64          `db:"id"`
			UserID      int64          `db:"user_id"`
			DisplayName sql.NullString `db:"display_name"`
			Content     string         `db:"content"`
			CreatedAt   time.Time      `db:"created_at"`
		}
		if err := rows.Scan(&comment.ID, &comment.UserID, &comment.DisplayName, &comment.Content, &comment.CreatedAt); err != nil {
			fmt.Printf("Scan comment error: %v\n", err)
			continue
		}

		comments = append(comments, map[string]interface{}{
			"id":           comment.ID,
			"user_id":      comment.UserID,
			"display_name": comment.DisplayName.String,
			"content":      comment.Content,
			"created_at":   comment.CreatedAt.Format(time.RFC3339),
		})
	}

	// 构造响应数据
	response := gin.H{
		"code":    200,
		"message": "success",
		"data":    comments,
	}

	// 返回结果
	c.JSON(http.StatusOK, response)
}

// handler.go

// CreateComment 创建评论
func CreateComment(c *gin.Context) {
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

	// 定义请求结构体
	var req struct {
		UserID     int64  `json:"user_id" binding:"required"`
		Content    string `json:"content" binding:"required"`
		TargetType string `json:"target_type" binding:"required"` // post/comment
	}

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 插入评论数据
	insertQuery := `
		INSERT INTO comment (target_id,target_type, user_id, content, created_at)
		VALUES (?,?,?, ?, ?)
	`
	now := time.Now()
	res, err := db.MySQL.Exec(insertQuery, id, req.TargetType, req.UserID, req.Content, now)
	if err != nil {
		fmt.Printf("Insert comment error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建评论失败",
		})
		return
	}

	// 获取插入的评论 ID
	commentID, _ := res.LastInsertId()

	// 查询用户昵称
	var displayName sql.NullString
	userQuery := "SELECT display_name FROM user WHERE id = ?"
	err = db.MySQL.QueryRow(userQuery, req.UserID).Scan(&displayName)
	if err != nil {
		fmt.Printf("Query user error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询用户信息失败",
		})
		return
	}

	// 构造响应数据
	response := gin.H{
		"code":    201,
		"message": "评论创建成功",
		"data": map[string]interface{}{
			"id":           commentID,
			"post_id":      id,
			"user_id":      req.UserID,
			"display_name": displayName.String,
			"content":      req.Content,
			"created_at":   now.Format(time.RFC3339),
		},
	}

	// 返回结果
	c.JSON(http.StatusCreated, response)
}
