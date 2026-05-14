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

// CommentTreeNode 评论树节点结构
type CommentTreeNode struct {
	ID             int64              `json:"id"`
	UserID         int64              `json:"user_id"`
	UserName       string             `json:"user_name"`
	UserAvatar     string             `json:"user_avatar"`
	TargetType     string             `json:"target_type"`
	TargetID       int64              `json:"target_id"`
	ParentID       int64              `json:"parent_id"`
	RootID         int64              `json:"root_id"`
	ReplyUserID    sql.NullInt64      `json:"reply_user_id"` // 可能为NULL
	ReplyUserName  string             `json:"reply_user_name"`
	Content        string             `json:"content"`
	LikeCount      int                `json:"like_count"`
	ReplyCount     int                `json:"reply_count"`
	Status         int                `json:"status"`
	CreatedAt      string             `json:"created_at"`
	Liked          bool               `json:"liked"`            // 实际项目中需根据当前登录用户判断
	HasMoreReplies bool               `json:"has_more_replies"` // 是否有更多未展示的回复
	Replies        []*CommentTreeNode `json:"replies,omitempty"`
}

// CommentReplyItem 单个回复项（用于 replies 接口）
type CommentReplyItem struct {
	ID            int64         `json:"id"`
	UserID        int64         `json:"user_id"`
	UserName      string        `json:"user_name"`
	UserAvatar    string        `json:"user_avatar"`
	ParentID      int64         `json:"parent_id"`
	RootID        int64         `json:"root_id"`
	ReplyUserID   sql.NullInt64 `json:"reply_user_id"`
	ReplyUserName string        `json:"reply_user_name"`
	Content       string        `json:"content"`
	LikeCount     int           `json:"like_count"`
	ReplyCount    int           `json:"reply_count"`
	Status        int           `json:"status"`
	CreatedAt     string        `json:"created_at"`
	Liked         bool          `json:"liked"`
}

// GetCommentTree 获取评论树
func GetCommentTree(c *gin.Context) {
	// 1. 解析参数
	targetType := c.Query("target_type")
	if targetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "target_type 不能为空"})
		return
	}

	targetIDStr := c.Query("target_id")
	if targetIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "target_id 不能为空"})
		return
	}
	targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的 target_id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	sort := c.DefaultQuery("sort", "latest") // latest or hot
	replyLimit, _ := strconv.Atoi(c.DefaultQuery("reply_limit", "3"))

	offset := (page - 1) * size

	// 2. 查询一级评论总数
	countQuery := "SELECT COUNT(*) FROM comment WHERE target_type = ? AND target_id = ? AND parent_id = 0 AND status = 1"
	var total int64
	err = db.MySQL.QueryRow(countQuery, targetType, targetID).Scan(&total)
	if err != nil {
		fmt.Printf("Count comments error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询评论总数失败"})
		return
	}

	// 3. 查询一级评论列表
	orderClause := "c.created_at DESC"
	if sort == "hot" {
		orderClause = "c.like_count DESC, c.created_at DESC"
	}

	rootCommentsQuery := `
		SELECT 
			c.id, c.user_id, u.display_name, u.avatar, 
			c.target_type, c.target_id, c.parent_id, c.root_id, 
			c.reply_user_id, c.content, c.like_count, c.reply_count, 
			c.status, c.created_at
		FROM comment c
		LEFT JOIN user u ON c.user_id = u.id
		WHERE c.target_type = ? AND c.target_id = ? AND c.parent_id = 0 AND c.status = 1
		ORDER BY ` + orderClause + `
		LIMIT ? OFFSET ?
	`
	rows, err := db.MySQL.Query(rootCommentsQuery, targetType, targetID, size, offset)
	if err != nil {
		fmt.Printf("Query root comments error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询评论列表失败"})
		return
	}
	defer rows.Close()

	var rootComments []*CommentTreeNode
	var rootIDs []int64

	for rows.Next() {
		var node CommentTreeNode
		var createdAt time.Time
		var avatar sql.NullString
		err := rows.Scan(
			&node.ID, &node.UserID, &node.UserName, &avatar,
			&node.TargetType, &node.TargetID, &node.ParentID, &node.RootID,
			&node.ReplyUserID, &node.Content, &node.LikeCount, &node.ReplyCount,
			&node.Status, &createdAt,
		)
		if err != nil {
			fmt.Printf("Scan root comment error: %v\n", err)
			continue
		}
		node.UserAvatar = avatar.String
		node.CreatedAt = createdAt.Format(time.RFC3339)
		// 初始化切片
		node.Replies = make([]*CommentTreeNode, 0)

		rootComments = append(rootComments, &node)
		rootIDs = append(rootIDs, node.ID)
	}

	// 4. 批量查询这些一级评论下的回复 (预加载)
	if len(rootIDs) > 0 {
		// 构造 IN 查询
		placeholders := strings.TrimRight(strings.Repeat("?,", len(rootIDs)), ",")
		repliesQuery := `
			SELECT 
				c.id, c.user_id, u.display_name, u.avatar, 
				c.parent_id, c.root_id, c.reply_user_id, ru.display_name as reply_user_name,
				c.content, c.like_count, c.reply_count, 
				c.status, c.created_at
			FROM comment c
			LEFT JOIN user u ON c.user_id = u.id
			LEFT JOIN user ru ON c.reply_user_id = ru.id
			WHERE c.root_id IN (` + placeholders + `) 
			  AND c.parent_id <> 0 
			  AND c.status = 1
			ORDER BY c.created_at ASC
		`

		// 准备参数
		args := make([]interface{}, len(rootIDs))
		for i, id := range rootIDs {
			args[i] = id
		}

		replyRows, err := db.MySQL.Query(repliesQuery, args...)
		if err != nil {
			fmt.Printf("Query replies error: %v\n", err)
			// 这里不直接返回错误，而是让主列表正常返回，只是没有回复数据
		} else {
			defer replyRows.Close()

			// 临时存储所有查出的回复，按 root_id 分组
			tempRepliesMap := make(map[int64][]*CommentTreeNode)

			for replyRows.Next() {
				var reply CommentTreeNode
				var createdAt time.Time
				var avatar sql.NullString
				var replyUserName sql.NullString

				err := replyRows.Scan(
					&reply.ID, &reply.UserID, &reply.UserName, &avatar,
					&reply.ParentID, &reply.RootID, &reply.ReplyUserID, &replyUserName,
					&reply.Content, &reply.LikeCount, &reply.ReplyCount,
					&reply.Status, &createdAt,
				)
				if err != nil {
					continue
				}

				reply.UserAvatar = avatar.String
				reply.ReplyUserName = replyUserName.String
				reply.CreatedAt = createdAt.Format(time.RFC3339)
				reply.Replies = make([]*CommentTreeNode, 0) // 二级回复通常不在这里展示，或者递归处理，这里简化为只展示一层或直接平铺

				tempRepliesMap[reply.RootID] = append(tempRepliesMap[reply.RootID], &reply)
			}

			// 将回复挂载到对应的一级评论下，并处理 reply_limit 和 has_more
			for _, rootComment := range rootComments {
				if replies, exists := tempRepliesMap[rootComment.ID]; exists {
					totalReplies := len(replies)
					// 截取前 replyLimit 条
					if totalReplies > replyLimit {
						rootComment.Replies = replies[:replyLimit]
						rootComment.HasMoreReplies = true
					} else {
						rootComment.Replies = replies
						rootComment.HasMoreReplies = false
					}
				}
			}
		}
	}

	// 5. 构造响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":  rootComments,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// GetCommentReplies 获取某根评论下的全部回复
func GetCommentReplies(c *gin.Context) {
	// 1. 解析参数
	rootIDStr := c.Query("root_id")
	if rootIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "root_id 不能为空"})
		return
	}
	rootID, err := strconv.ParseInt(rootIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的 root_id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	sort := c.DefaultQuery("sort", "earliest") // earliest or latest

	offset := (page - 1) * size

	// 2. 获取根评论信息 (用于返回 root_comment)
	rootCommentQuery := `
		SELECT 
			c.id, c.user_id, u.display_name, u.avatar, 
			c.content, c.like_count, c.reply_count, 
			c.created_at
		FROM comment c
		LEFT JOIN user u ON c.user_id = u.id
		WHERE c.id = ? AND c.status = 1
	`
	var rootComment CommentReplyItem
	var createdAt time.Time
	var avatar sql.NullString

	err = db.MySQL.QueryRow(rootCommentQuery, rootID).Scan(
		&rootComment.ID, &rootComment.UserID, &rootComment.UserName, &avatar,
		&rootComment.Content, &rootComment.LikeCount, &rootComment.ReplyCount,
		&createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "根评论不存在"})
		} else {
			fmt.Printf("Query root comment detail error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询根评论失败"})
		}
		return
	}
	rootComment.UserAvatar = avatar.String
	rootComment.CreatedAt = createdAt.Format(time.RFC3339)

	// 3. 统计回复总数 (不包括根评论自己)
	countQuery := "SELECT COUNT(*) FROM comment WHERE root_id = ? AND parent_id <> 0 AND status = 1"
	var total int64
	err = db.MySQL.QueryRow(countQuery, rootID).Scan(&total)
	if err != nil {
		fmt.Printf("Count replies error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询回复总数失败"})
		return
	}

	// 4. 查询回复列表
	orderClause := "c.created_at ASC" // 默认最早
	if sort == "latest" {
		orderClause = "c.created_at DESC"
	}

	repliesQuery := `
		SELECT 
			c.id, c.user_id, u.display_name, u.avatar, 
			c.parent_id, c.root_id, c.reply_user_id, ru.display_name as reply_user_name,
			c.content, c.like_count, c.reply_count, 
			c.status, c.created_at
		FROM comment c
		LEFT JOIN user u ON c.user_id = u.id
		LEFT JOIN user ru ON c.reply_user_id = ru.id
		WHERE c.root_id = ? AND c.parent_id <> 0 AND c.status = 1
		ORDER BY ` + orderClause + `
		LIMIT ? OFFSET ?
	`

	rows, err := db.MySQL.Query(repliesQuery, rootID, size, offset)
	if err != nil {
		fmt.Printf("Query replies list error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询回复列表失败"})
		return
	}
	defer rows.Close()

	var repliesList []*CommentReplyItem
	for rows.Next() {
		var item CommentReplyItem
		var createdAt time.Time
		var avatar sql.NullString
		var replyUserName sql.NullString

		err := rows.Scan(
			&item.ID, &item.UserID, &item.UserName, &avatar,
			&item.ParentID, &item.RootID, &item.ReplyUserID, &replyUserName,
			&item.Content, &item.LikeCount, &item.ReplyCount,
			&item.Status, &createdAt,
		)
		if err != nil {
			continue
		}

		item.UserAvatar = avatar.String
		item.ReplyUserName = replyUserName.String
		item.CreatedAt = createdAt.Format(time.RFC3339)

		repliesList = append(repliesList, &item)
	}

	// 5. 构造响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"root_comment": rootComment,
			"list":         repliesList,
			"total":        total,
			"page":         page,
			"size":         size,
		},
	})
}

// --- 结构体定义 ---

// PostResponse 帖子响应结构 (用于列表和详情)
type PostResponse struct {
	ID            int64    `json:"id"`
	UserID        int64    `json:"user_id"`
	DisplayName   string   `json:"display_name"`
	DiseaseID     *int64   `json:"disease_id,omitempty"`
	CategoryID    *int64   `json:"category_id,omitempty"`
	Type          string   `json:"type"`
	Title         string   `json:"title,omitempty"`
	Content       string   `json:"content,omitempty"` // 列表接口可考虑 omit 或截断
	Images        []string `json:"images"`
	ViewCount     int      `json:"view_count"`
	LikeCount     int      `json:"like_count"`
	CommentCount  int      `json:"comment_count"`
	FavoriteCount int      `json:"favorite_count"`
	IsTop         bool     `json:"is_top"`
	IsRecommend   bool     `json:"is_recommend"`
	Status        int      `json:"status"`
	RejectReason  string   `json:"reject_reason,omitempty"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at,omitempty"`
	IsLiked       bool     `json:"is_liked"`     // 当前用户是否点赞
	IsFavorited   bool     `json:"is_favorited"` // 当前用户是否收藏
}

// CreatePostRequest 创建帖子请求
type CreatePostRequest struct {
	UserID     int64    `json:"user_id" binding:"required"`
	DiseaseID  *int64   `json:"disease_id"`
	CategoryID *int64   `json:"category_id"`
	Type       string   `json:"type" binding:"required,oneof=help experience emotion info"`
	Title      *string  `json:"title"`
	Content    string   `json:"content" binding:"required"`
	Images     []string `json:"images"`
}

// UpdatePostRequest 更新帖子请求
type UpdatePostRequest struct {
	Title      *string   `json:"title"`
	Content    *string   `json:"content"`
	Images     *[]string `json:"images"` // 使用指针以便判断是否传参
	DiseaseID  *int64    `json:"disease_id"`
	CategoryID *int64    `json:"category_id"`
	Type       *string   `json:"type" binding:"omitempty,oneof=help experience emotion info"`

	// 以下字段通常仅管理员可更新，实际项目中应结合权限中间件判断
	IsTop        *int8   `json:"is_top"`
	IsRecommend  *int8   `json:"is_recommend"`
	Status       *int8   `json:"status"`
	RejectReason *string `json:"reject_reason"`
}

// --- Handler 实现 ---

// GetCommunityPosts 获取社区帖子列表
func GetCommunityPosts(c *gin.Context) {
	// 1. 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// 2. 获取筛选参数
	sort := c.DefaultQuery("sort", "latest")
	postType := c.Query("type")
	diseaseID := c.Query("disease_id")
	categoryID := c.Query("category_id")
	isTop := c.Query("is_top")             // 是否只看置顶
	isRecommend := c.Query("is_recommend") // 是否只看推荐

	// 3. 构建查询条件
	whereClause := "p.status = 1" // 默认只查正常状态的帖子
	args := []interface{}{}

	if postType != "" {
		whereClause += " AND p.type = ?"
		args = append(args, postType)
	}
	if diseaseID != "" {
		whereClause += " AND p.disease_id = ?"
		args = append(args, diseaseID)
	}
	if categoryID != "" {
		whereClause += " AND p.category_id = ?"
		args = append(args, categoryID)
	}
	if isTop == "1" {
		whereClause += " AND p.is_top = 1"
	}
	if isRecommend == "1" {
		whereClause += " AND p.is_recommend = 1"
	}

	// 4. 查询总数
	countQuery := "SELECT COUNT(*) FROM post p WHERE " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		fmt.Printf("Count query error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询总数失败"})
		return
	}

	// 5. 构建排序逻辑
	orderClause := "p.created_at DESC"
	switch sort {
	case "latest":
		orderClause = "p.created_at DESC"
	case "hot":
		orderClause = "p.like_count DESC, p.created_at DESC"
	case "top":
		orderClause = "p.is_top DESC, p.created_at DESC"
	}

	// 6. 查询帖子列表
	// 注意：列表接口通常不建议返回 content 全文，如果内容很长会影响性能。这里为了兼容原逻辑保留，生产环境建议截断或省略。
	listQuery := `
		SELECT 
			p.id, p.user_id, u.display_name, p.disease_id, p.category_id, p.type, p.title, p.content, p.images,
			p.view_count, p.like_count, p.comment_count, p.favorite_count,
			p.is_top, p.is_recommend, p.status, p.created_at
		FROM post p
		LEFT JOIN user u ON p.user_id = u.id
		WHERE ` + whereClause + `
		ORDER BY ` + orderClause + `
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		fmt.Printf("List query error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询帖子列表失败"})
		return
	}
	defer rows.Close()

	var records []PostResponse
	// 假设当前用户ID已从中件件获取，否则默认为0，不显示点赞/收藏状态
	currentUserID := int64(0) // TODO: 从 Context 获取真实用户ID

	for rows.Next() {
		var p PostResponse
		var imagesBytes []byte
		var createdAt time.Time
		var displayName sql.NullString
		var title sql.NullString

		// 【修改点】定义临时变量接收数据库的 TINYINT (int/int8)
		var isTopInt int
		var isRecommendInt int

		// 扫描数据
		err := rows.Scan(
			&p.ID, &p.UserID, &displayName, &p.DiseaseID, &p.CategoryID, &p.Type, &title, &p.Content, &imagesBytes,
			&p.ViewCount, &p.LikeCount, &p.CommentCount, &p.FavoriteCount,
			&p.IsTop, &p.IsRecommend, &p.Status, &createdAt,
		)
		if err != nil {
			fmt.Printf("Scan error: %v\n", err)
			continue
		}

		// 处理 Null 字段
		if displayName.Valid {
			p.DisplayName = displayName.String
		}
		if title.Valid {
			p.Title = title.String
		}
		p.CreatedAt = createdAt.Format(time.RFC3339)

		// 【修改点】将 int 转换为 bool
		p.IsTop = isTopInt == 1
		p.IsRecommend = isRecommendInt == 1

		// 处理 Images JSON
		if len(imagesBytes) > 0 {
			json.Unmarshal(imagesBytes, &p.Images)
		} else {
			p.Images = []string{}
		}

		// 查询当前用户是否点赞/收藏 (优化：可以在主查询中 LEFT JOIN 或者批量查询，这里简化为默认 false，实际需根据 currentUserID 查询)
		p.IsLiked = false
		p.IsFavorited = false
		if currentUserID > 0 {
			// 示例：实际生产中建议用 IN 查询批量获取状态，避免 N+1
			// var likeCount int
			// db.MySQL.QueryRow("SELECT COUNT(*) FROM post_like WHERE post_id=? AND user_id=?", p.ID, currentUserID).Scan(&likeCount)
			// p.IsLiked = likeCount > 0
		}

		records = append(records, p)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"records": records,
			"total":   total,
			"page":    page,
			"limit":   limit,
		},
	})
}

// GetPostDetail 获取帖子详情
func GetPostDetail(c *gin.Context) {
	postIDStr := c.Param("id")
	id, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的帖子 ID"})
		return
	}

	// 1. 原子增加浏览量 (异步执行或忽略错误，不影响主流程)
	go func() {
		db.MySQL.Exec("UPDATE post SET view_count = view_count + 1 WHERE id = ?", id)
	}()

	// 2. 查询详情
	query := `
		SELECT 
			p.id, p.user_id, u.display_name, p.disease_id, p.category_id, p.type, p.title, p.content, p.images,
			p.view_count, p.like_count, p.comment_count, p.favorite_count,
			p.is_top, p.is_recommend, p.status, p.reject_reason, p.created_at, p.updated_at
		FROM post p
		LEFT JOIN user u ON p.user_id = u.id
		WHERE p.id = ?
	`

	var p PostResponse
	var imagesBytes []byte
	var createdAt, updatedAt time.Time
	var displayName, title, rejectReason sql.NullString

	// 【修改点】定义临时变量接收数据库的 TINYINT
	var isTopInt int
	var isRecommendInt int

	err = db.MySQL.QueryRow(query, id).Scan(
		&p.ID, &p.UserID, &displayName, &p.DiseaseID, &p.CategoryID, &p.Type, &title, &p.Content, &imagesBytes,
		&p.ViewCount, &p.LikeCount, &p.CommentCount, &p.FavoriteCount,
		&p.IsTop, &p.IsRecommend, &p.Status, &rejectReason, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "帖子不存在"})
		} else {
			fmt.Printf("Query post detail error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		}
		return
	}

	// 填充字段
	if displayName.Valid {
		p.DisplayName = displayName.String
	}
	if title.Valid {
		p.Title = title.String
	}
	if rejectReason.Valid {
		p.RejectReason = rejectReason.String
	}
	p.CreatedAt = createdAt.Format(time.RFC3339)
	p.UpdatedAt = updatedAt.Format(time.RFC3339)

	// 【修改点】将 int 转换为 bool
	p.IsTop = isTopInt == 1
	p.IsRecommend = isRecommendInt == 1

	if len(imagesBytes) > 0 {
		json.Unmarshal(imagesBytes, &p.Images)
	} else {
		p.Images = []string{}
	}

	// 获取当前用户互动状态
	currentUserID := int64(0) // TODO: 从 Context 获取
	p.IsLiked = false
	p.IsFavorited = false

	if currentUserID > 0 {
		var likeCnt, favCnt int
		db.MySQL.QueryRow("SELECT COUNT(*) FROM post_like WHERE post_id=? AND user_id=?", id, currentUserID).Scan(&likeCnt)
		db.MySQL.QueryRow("SELECT COUNT(*) FROM post_favorite WHERE post_id=? AND user_id=?", id, currentUserID).Scan(&favCnt)
		p.IsLiked = likeCnt > 0
		p.IsFavorited = favCnt > 0
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    p,
	})
}

// CreatePost 创建帖子
func CreatePost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 处理图片 JSON
	imagesJSON, _ := json.Marshal(req.Images)
	if req.Images == nil {
		imagesJSON = []byte("[]")
	}

	// 插入数据
	// status 默认为 0 (审核中)，如果需要直接发布改为 1
	query := `
		INSERT INTO post (
			user_id, disease_id, category_id, type, title, content, images, 
			view_count, like_count, comment_count, favorite_count,
			is_top, is_recommend, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0, 0, 0, 0)
	`
	res, err := db.MySQL.Exec(query,
		req.UserID, req.DiseaseID, req.CategoryID, req.Type, req.Title, req.Content, string(imagesJSON),
	)
	if err != nil {
		fmt.Printf("Insert post error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建帖子失败"})
		return
	}

	postID, _ := res.LastInsertId()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    gin.H{"post_id": postID},
	})
}

// UpdatePost 更新帖子
func UpdatePost(c *gin.Context) {
	postIDStr := c.Param("id")
	id, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的帖子 ID"})
		return
	}

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 1. 权限检查：获取帖子所有者
	currentUserID := int64(1001) // TODO: 从 JWT 获取
	var postUserID int64
	var postStatus int
	err = db.MySQL.QueryRow("SELECT user_id, status FROM post WHERE id = ?", id).Scan(&postUserID, &postStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "帖子不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		}
		return
	}

	// 简单权限判断：如果是普通用户，只能改自己的帖子，且不能改状态/置顶等敏感字段
	// 如果是管理员，可以改所有字段。这里简化处理，假设 currentUserID == postUserID 为作者
	isAdmin := false // TODO: 判断是否为管理员
	if !isAdmin && currentUserID != postUserID {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权限修改"})
		return
	}

	// 2. 构建动态更新 SQL
	updateFields := []string{}
	updateArgs := []interface{}{}

	// 基础字段
	if req.Title != nil {
		updateFields = append(updateFields, "title = ?")
		updateArgs = append(updateArgs, *req.Title)
	}
	if req.Content != nil {
		updateFields = append(updateFields, "content = ?")
		updateArgs = append(updateArgs, *req.Content)
	}
	if req.Images != nil {
		imgJSON, _ := json.Marshal(*req.Images)
		updateFields = append(updateFields, "images = ?")
		updateArgs = append(updateArgs, string(imgJSON))
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
		updateFields = append(updateFields, "type = ?")
		updateArgs = append(updateArgs, *req.Type)
	}

	// 管理/敏感字段 (仅管理员或特定条件下允许)
	if isAdmin {
		if req.IsTop != nil {
			updateFields = append(updateFields, "is_top = ?")
			updateArgs = append(updateArgs, *req.IsTop)
		}
		if req.IsRecommend != nil {
			updateFields = append(updateFields, "is_recommend = ?")
			updateArgs = append(updateArgs, *req.IsRecommend)
		}
		if req.Status != nil {
			updateFields = append(updateFields, "status = ?")
			updateArgs = append(updateArgs, *req.Status)
		}
		if req.RejectReason != nil {
			updateFields = append(updateFields, "reject_reason = ?")
			updateArgs = append(updateArgs, *req.RejectReason)
		}
	} else {
		// 普通用户不允许修改后，如果之前被驳回，修改后状态应变回“审核中”(0) 或保持“正常”(1)?
		// 业务逻辑决定：通常修改后重新审核
		if postStatus == 2 { // 如果被驳回，修改后重置为审核中
			updateFields = append(updateFields, "status = ?")
			updateArgs = append(updateArgs, 0)
		}
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "未提供有效更新字段"})
		return
	}

	// 添加更新时间
	updateFields = append(updateFields, "updated_at = ?")
	updateArgs = append(updateArgs, time.Now())

	// 添加 Where 条件
	updateArgs = append(updateArgs, id)

	sqlStr := "UPDATE post SET " + strings.Join(updateFields, ", ") + " WHERE id = ?"
	_, err = db.MySQL.Exec(sqlStr, updateArgs...)
	if err != nil {
		fmt.Printf("Update post error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}
