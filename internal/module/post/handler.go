package post

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

// GetPostComments 获取帖子评论树
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

	// 查询所有评论（包括一级和子级评论）
	query := `
		SELECT 
			c.id, c.user_id, u.display_name, c.content, c.parent_id, c.root_id, c.created_at
		FROM comment c
		LEFT JOIN user u ON c.user_id = u.id
		WHERE c.target_type = 'post' AND c.target_id = ?
		ORDER BY c.root_id ASC, c.created_at ASC
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

	// 定义评论节点结构
	type CommentNode struct {
		ID          int64          `json:"id"`
		UserID      int64          `json:"user_id"`
		DisplayName string         `json:"display_name"`
		Content     string         `json:"content"`
		ParentID    int64          `json:"parent_id"`
		RootID      int64          `json:"root_id"`
		CreatedAt   time.Time      `json:"created_at"`
		Children    []*CommentNode `json:"children,omitempty"`
	}

	// 存储所有评论节点
	allComments := make(map[int64]*CommentNode)
	var rootComments []*CommentNode

	// 扫描数据库结果
	for rows.Next() {
		var comment struct {
			ID          int64          `db:"id"`
			UserID      int64          `db:"user_id"`
			DisplayName sql.NullString `db:"display_name"`
			Content     string         `db:"content"`
			ParentID    int64          `db:"parent_id"`
			RootID      int64          `db:"root_id"`
			CreatedAt   time.Time      `db:"created_at"`
		}
		if err := rows.Scan(&comment.ID, &comment.UserID, &comment.DisplayName, &comment.Content, &comment.ParentID, &comment.RootID, &comment.CreatedAt); err != nil {
			fmt.Printf("Scan comment error: %v\n", err)
			continue
		}

		// 创建评论节点
		node := &CommentNode{
			ID:          comment.ID,
			UserID:      comment.UserID,
			DisplayName: comment.DisplayName.String,
			Content:     comment.Content,
			ParentID:    comment.ParentID,
			RootID:      comment.RootID,
			CreatedAt:   comment.CreatedAt,
		}

		// 将节点存储到映射中
		allComments[comment.ID] = node

		// 如果是一级评论，则加入根评论列表
		if comment.ParentID == 0 {
			rootComments = append(rootComments, node)
		}
	}

	// 构建评论树
	for _, node := range allComments {
		if node.ParentID != 0 {
			parentNode, exists := allComments[node.ParentID]
			if exists {
				parentNode.Children = append(parentNode.Children, node)
			}
		}
	}

	// 构造响应数据
	response := gin.H{
		"code":    200,
		"message": "success",
		"data":    rootComments, // 返回嵌套的树形结构
	}

	// 返回结果
	c.JSON(http.StatusOK, response)
}

// 创建评论
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
		UserID   int64  `json:"user_id" binding:"required"`
		Content  string `json:"content" binding:"required"`
		ParentID *int64 `json:"parent_id"` // 可选字段，表示回复的评论 ID
	}

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 初始化默认值
	parentID := int64(0)
	rootID := int64(0)

	// 如果指定了 parent_id，则查询父评论是否存在并确定 root_id
	if req.ParentID != nil && *req.ParentID > 0 {
		parentID = *req.ParentID

		// 查询父评论是否存在且属于当前帖子
		var parentComment struct {
			ID     int64 `db:"id"`
			RootID int64 `db:"root_id"`
		}
		parentQuery := "SELECT id, root_id FROM comment WHERE id = ? AND target_type = 'post' AND target_id = ?"
		err = db.MySQL.QueryRow(parentQuery, parentID, id).Scan(&parentComment.ID, &parentComment.RootID)
		if err != nil || parentComment.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "父评论不存在或不属于当前帖子",
			})
			return
		}

		// 设置 root_id
		if parentComment.RootID == 0 {
			rootID = parentID // 父评论是一级评论
		} else {
			rootID = parentComment.RootID // 父评论是子评论
		}
	}

	// 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		fmt.Printf("Begin transaction error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "开启事务失败",
		})
		return
	}
	defer tx.Rollback() // 确保事务回滚

	// 插入评论数据
	insertQuery := `
		INSERT INTO comment (
			target_id, target_type, user_id, content, parent_id, root_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	res, err := tx.Exec(insertQuery, id, "post", req.UserID, req.Content, parentID, rootID, now)
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

	// 更新父评论或帖子的 reply_count
	if parentID > 0 {
		// 更新父评论的 reply_count
		_, err = tx.Exec("UPDATE comment SET reply_count = reply_count + 1 WHERE id = ?", parentID)
	} else {
		// 更新帖子的 comment_count
		_, err = tx.Exec("UPDATE post SET comment_count = comment_count + 1 WHERE id = ?", id)
	}
	if err != nil {
		fmt.Printf("Update reply/comment count error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新评论计数失败",
		})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		fmt.Printf("Commit transaction error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败",
		})
		return
	}

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
			"parent_id":    parentID,
			"root_id":      rootID,
			"created_at":   now.Format(time.RFC3339),
		},
	}

	// 返回结果
	c.JSON(http.StatusCreated, response)
}

// GetPostDetail 获取帖子详情
func GetPostDetail(c *gin.Context) {
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

	// 查询帖子详情
	query := `
		SELECT 
			p.id, p.title, p.content, p.user_id, u.display_name, p.created_at, 
			p.images, p.like_count, p.comment_count
		FROM post p
		LEFT JOIN user u ON p.user_id = u.id
		WHERE p.id = ? AND p.status = 1
	`
	var post struct {
		ID           int64          `db:"id"`
		Title        sql.NullString `db:"title"`
		Content      string         `db:"content"`
		UserID       int64          `db:"user_id"`
		DisplayName  sql.NullString `db:"display_name"`
		CreatedAt    time.Time      `db:"created_at"`
		Images       []byte         `db:"images"`
		LikeCount    int            `db:"like_count"`
		CommentCount int            `db:"comment_count"`
	}
	err = db.MySQL.QueryRow(query, id).Scan(
		&post.ID, &post.Title, &post.Content, &post.UserID, &post.DisplayName,
		&post.CreatedAt, &post.Images, &post.LikeCount, &post.CommentCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "帖子不存在",
			})
			return
		}
		fmt.Printf("Query post detail error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询帖子详情失败",
		})
		return
	}

	// 处理 JSON 字段（图片）
	var images []string
	if len(post.Images) > 0 {
		json.Unmarshal(post.Images, &images)
	}

	// 默认用户未点赞（实际开发中应从 JWT 或上下文中获取用户 ID 进行判断）
	isLiked := false

	// 构造响应数据
	response := gin.H{
		"code":    200,
		"message": "success",
		"data": map[string]interface{}{
			"id":            post.ID,
			"title":         post.Title.String,
			"content":       post.Content,
			"user_id":       post.UserID,
			"display_name":  post.DisplayName.String,
			"created_at":    post.CreatedAt.Format(time.RFC3339),
			"images":        images,
			"like_count":    post.LikeCount,
			"comment_count": post.CommentCount,
			"is_liked":      isLiked,
		},
	}

	// 返回结果
	c.JSON(http.StatusOK, response)
}

// PostOptionsResponse 帖子筛选选项响应
type PostOptionsResponse struct {
	Types      []OptionItem `json:"types"`
	Diseases   []OptionItem `json:"diseases"`
	Categories []OptionItem `json:"categories"`
}

// OptionItem 选项项
type OptionItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// UpdatePostRequest 更新帖子请求
type UpdatePostRequest struct {
	Title      *string  `json:"title"`
	Content    *string  `json:"content"`
	Images     []string `json:"images"`
	DiseaseID  *int64   `json:"disease_id"`
	CategoryID *int64   `json:"category_id"`
	Type       *string  `json:"type"`
}

// UpdateCommentRequest 更新评论请求
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// GetPostOptions 获取帖子筛选选项
func GetPostOptions(c *gin.Context) {
	// 帖子类型选项
	types := []OptionItem{
		{Label: "求助", Value: "help"},
		{Label: "经验", Value: "experience"},
		{Label: "情感", Value: "emotion"},
		{Label: "资讯", Value: "info"},
	}

	// 获取疾病分类选项
	diseaseQuery := "SELECT id, name FROM disease WHERE status = 1"
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
		var disease struct {
			ID   int64  `db:"id"`
			Name string `db:"name"`
		}
		if err := diseaseRows.Scan(&disease.ID, &disease.Name); err != nil {
			continue
		}
		diseases = append(diseases, OptionItem{
			Label: disease.Name,
			Value: strconv.FormatInt(disease.ID, 10),
		})
	}

	// 获取分类选项
	categoryQuery := "SELECT id, name FROM post_category WHERE status = 1"
	categoryRows, err := db.MySQL.Query(categoryQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询分类选项失败",
		})
		return
	}
	defer categoryRows.Close()

	var categories []OptionItem
	for categoryRows.Next() {
		var category struct {
			ID   int64  `db:"id"`
			Name string `db:"name"`
		}
		if err := categoryRows.Scan(&category.ID, &category.Name); err != nil {
			continue
		}
		categories = append(categories, OptionItem{
			Label: category.Name,
			Value: strconv.FormatInt(category.ID, 10),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": PostOptionsResponse{
			Types:      types,
			Diseases:   diseases,
			Categories: categories,
		},
	})
}

// UpdatePost 更新帖子
func UpdatePost(c *gin.Context) {
	postID := c.Param("id")
	id, err := strconv.ParseInt(postID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的帖子 ID",
		})
		return
	}

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 检查帖子是否存在且属于当前用户
	userID := int64(1001) // 实际应从 JWT 获取
	checkQuery := "SELECT id, user_id FROM post WHERE id = ? AND status = 1"
	var postUserID int64
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&postUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "帖子不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询帖子失败",
		})
		return
	}

	// 验证权限
	if postUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限修改该帖子",
		})
		return
	}

	// 构建动态更新语句
	updateFields := []string{}
	updateArgs := []interface{}{}

	if req.Title != nil {
		updateFields = append(updateFields, "title = ?")
		updateArgs = append(updateArgs, *req.Title)
	}
	if req.Content != nil {
		updateFields = append(updateFields, "content = ?")
		updateArgs = append(updateArgs, *req.Content)
	}
	if req.Images != nil {
		imagesJSON, _ := json.Marshal(req.Images)
		updateFields = append(updateFields, "images = ?")
		updateArgs = append(updateArgs, string(imagesJSON))
	}
	if req.DiseaseID != nil {
		updateFields = append(updateFields, "disease_id = ?")
		updateArgs = append(updateArgs, *req.DiseaseID)
	}
	if req.CategoryID != nil {
		updateFields = append(updateFields, "category_id = ?")
		updateArgs = append(updateArgs, *req.CategoryID)
	}
	if req.Type != nil {
		validTypes := map[string]bool{"help": true, "experience": true, "emotion": true, "info": true}
		if !validTypes[*req.Type] {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的帖子类型",
			})
			return
		}
		updateFields = append(updateFields, "type = ?")
		updateArgs = append(updateArgs, *req.Type)
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

	updateQuery := "UPDATE post SET " + strings.Join(updateFields, ", ") + " WHERE id = ?"
	_, err = db.MySQL.Exec(updateQuery, updateArgs...)
	if err != nil {
		fmt.Printf("Update post error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新帖子失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DeletePost 删除帖子（软删除）
func DeletePost(c *gin.Context) {
	postID := c.Param("id")
	id, err := strconv.ParseInt(postID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的帖子 ID",
		})
		return
	}

	// 检查帖子是否存在且属于当前用户
	userID := int64(1001) // 实际应从 JWT 获取
	checkQuery := "SELECT id, user_id FROM post WHERE id = ? AND status = 1"
	var postUserID int64
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&postUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "帖子不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询帖子失败",
		})
		return
	}

	// 验证权限
	if postUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限删除该帖子",
		})
		return
	}

	// 软删除：将 status 设为 0
	deleteQuery := "UPDATE post SET status = 0, updated_at = ? WHERE id = ?"
	_, err = db.MySQL.Exec(deleteQuery, time.Now(), id)
	if err != nil {
		fmt.Printf("Delete post error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除帖子失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// UpdateComment 更新评论
func UpdateComment(c *gin.Context) {
	commentID := c.Param("id")
	id, err := strconv.ParseInt(commentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的评论 ID",
		})
		return
	}

	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 检查评论是否存在且属于当前用户
	userID := int64(1001) // 实际应从 JWT 获取
	checkQuery := "SELECT id, user_id FROM comment WHERE id = ? AND status = 1"
	var commentUserID int64
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&commentUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "评论不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询评论失败",
		})
		return
	}

	// 验证权限
	if commentUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限修改该评论",
		})
		return
	}

	// 更新评论内容
	updateQuery := "UPDATE comment SET content = ?, updated_at = ? WHERE id = ?"
	_, err = db.MySQL.Exec(updateQuery, req.Content, time.Now(), id)
	if err != nil {
		fmt.Printf("Update comment error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新评论失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DeleteComment 删除评论（软删除）
func DeleteComment(c *gin.Context) {
	commentID := c.Param("id")
	id, err := strconv.ParseInt(commentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的评论 ID",
		})
		return
	}

	// 检查评论是否存在且属于当前用户
	userID := int64(1001) // 实际应从 JWT 获取
	checkQuery := "SELECT id, user_id, target_id FROM comment WHERE id = ? AND status = 1"
	var commentUserID, postID int64
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&commentUserID, &postID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "评论不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询评论失败",
		})
		return
	}

	// 验证权限
	if commentUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限删除该评论",
		})
		return
	}

	// 开启事务
	tx, err := db.MySQL.Begin()
	if err != nil {
		fmt.Printf("Begin transaction error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "开启事务失败",
		})
		return
	}
	defer tx.Rollback()

	// 软删除评论
	deleteQuery := "UPDATE comment SET status = 0, updated_at = ? WHERE id = ?"
	_, err = tx.Exec(deleteQuery, time.Now(), id)
	if err != nil {
		fmt.Printf("Delete comment error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除评论失败",
		})
		return
	}

	// 更新帖子的评论计数
	_, err = tx.Exec("UPDATE post SET comment_count = comment_count - 1 WHERE id = ? AND comment_count > 0", postID)
	if err != nil {
		fmt.Printf("Update comment count error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新评论计数失败",
		})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		fmt.Printf("Commit transaction error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}
