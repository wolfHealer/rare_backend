package auth

import (
	"database/sql"
	"net/http"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/hash"
	"rare_backend/internal/pkg/jwt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func login(c *gin.Context) {
	// 获取请求参数
	var req struct {
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required"` // 接收明文密码
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 查询用户是否存在
	var user struct {
		ID           int64          `db:"id"`
		PasswordHash string         `db:"password_hash"`
		LoginCount   int            `db:"login_count"`
		Nickname     sql.NullString `db:"display_name"`
		Avatar       sql.NullString `db:"avatar"`
	}

	query := "SELECT id, password_hash, login_count, display_name, avatar FROM user WHERE phone = ?"
	err := db.MySQL.QueryRow(query, req.Phone).Scan(&user.ID, &user.PasswordHash, &user.LoginCount, &user.Nickname, &user.Avatar)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "手机号或密码错误",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "数据库查询失败",
			})
		}
		return
	}

	// 验证密码（使用 bcrypt 比对）
	if !hash.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "手机号或密码错误",
		})
		return
	}

	// 更新登录信息
	updateQuery := "UPDATE user SET last_login_at = NOW(), login_count = ? WHERE id = ?"
	_, err = db.MySQL.Exec(updateQuery, user.LoginCount+1, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新登录信息失败",
		})
		return
	}

	// 生成 JWT Token
	token, err := jwt.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成 Token 失败",
		})
		return
	}

	// 构造返回数据
	response := gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"user_id":    user.ID,
			"nickname":   user.Nickname.String,
			"phone":      req.Phone,
			"avatar":     user.Avatar.String,
			"token":      token,
			"expires_in": int(jwt.TokenExpireDuration.Seconds()),
		},
	}

	// 返回结果
	c.JSON(http.StatusOK, response)
}

func register(c *gin.Context) {
	// 获取请求参数
	var req struct {
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required"` // 接收明文密码
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 检查手机号是否已存在
	var count int
	checkQuery := "SELECT COUNT(*) FROM user WHERE phone = ?"
	err := db.MySQL.QueryRow(checkQuery, req.Phone).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库查询失败",
		})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "手机号已被注册",
		})
		return
	}

	// 密码哈希加密
	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "密码加密失败",
		})
		return
	}

	// 插入新用户
	insertQuery := `
		INSERT INTO user (phone, password_hash, display_name, role, status)
		VALUES (?, ?, CONCAT('用户', LPAD(FLOOR(RAND() * 10000), 4, '0')), 1, 1)
	`
	_, err = db.MySQL.Exec(insertQuery, req.Phone, hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "注册失败",
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "注册成功",
	})
}

// UserItem 用户列表/详情响应结构
type UserItem struct {
	ID          int64  `json:"id"`
	Phone       string `json:"phone"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Role        int    `json:"role"`
	Status      int    `json:"status"`
	LastLoginAt string `json:"lastLoginAt"`
	LoginCount  int    `json:"loginCount"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// GetUserList 获取用户列表（支持分页和关键词搜索）
// GetUserList 获取用户列表（支持分页、关键词搜索和角色筛选）
func GetUserList(c *gin.Context) {
	// 1. 获取请求参数
	keyword := c.DefaultQuery("keyword", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	statusStr := c.DefaultQuery("status", "") // 可选：按状态筛选
	roleStr := c.DefaultQuery("role", "")     // 新增：按角色筛选

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 2. 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// 关键词搜索：匹配手机号或展示名
	if keyword != "" {
		whereClause += " AND (phone LIKE ? OR display_name LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态筛选
	if statusStr != "" {
		status, err := strconv.Atoi(statusStr)
		if err == nil {
			whereClause += " AND status = ?"
			args = append(args, status)
		}
	}

	// 新增：角色筛选
	if roleStr != "" {
		role, err := strconv.Atoi(roleStr)
		if err == nil {
			// 可选：验证角色合法性，例如只允许 1, 2, 9
			// if role != 1 && role != 2 && role != 9 {
			//     c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色类型"})
			//     return
			// }
			whereClause += " AND role = ?"
			args = append(args, role)
		}
	}

	// 3. 查询总数
	countQuery := "SELECT COUNT(*) FROM user " + whereClause
	var total int64
	err = db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询用户总数失败",
		})
		return
	}

	// 4. 查询列表
	listQuery := `
		SELECT id, phone, display_name, avatar, role, status, last_login_at, login_count, created_at, updated_at
		FROM user
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	// 注意：这里追加 LIMIT 和 OFFSET 的参数，必须放在最后
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询用户列表失败",
		})
		return
	}
	defer rows.Close()

	var list []UserItem
	for rows.Next() {
		var u struct {
			ID          int64          `db:"id"`
			Phone       string         `db:"phone"`
			DisplayName sql.NullString `db:"display_name"`
			Avatar      sql.NullString `db:"avatar"`
			Role        int8           `db:"role"`
			Status      int8           `db:"status"`
			LastLoginAt sql.NullTime   `db:"last_login_at"`
			LoginCount  int            `db:"login_count"`
			CreatedAt   time.Time      `db:"created_at"`
			UpdatedAt   time.Time      `db:"updated_at"`
		}

		if err := rows.Scan(
			&u.ID, &u.Phone, &u.DisplayName, &u.Avatar,
			&u.Role, &u.Status, &u.LastLoginAt, &u.LoginCount,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			continue
		}

		// 格式化时间
		lastLoginAtStr := ""
		if u.LastLoginAt.Valid {
			lastLoginAtStr = u.LastLoginAt.Time.Format("2006-01-02 15:04:05")
		}

		list = append(list, UserItem{
			ID:          u.ID,
			Phone:       u.Phone,
			DisplayName: u.DisplayName.String,
			Avatar:      u.Avatar.String,
			Role:        int(u.Role),
			Status:      int(u.Status),
			LastLoginAt: lastLoginAtStr,
			LoginCount:  u.LoginCount,
			CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   u.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
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

// GetUserByID 获取单个用户详情
func GetUserByID(c *gin.Context) {
	// 1. 获取路径参数 ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户 ID",
		})
		return
	}

	// 2. 查询数据库
	query := `
		SELECT id, phone, display_name, avatar, role, status, last_login_at, login_count, created_at, updated_at
		FROM user
		WHERE id = ?
	`

	var u struct {
		ID          int64          `db:"id"`
		Phone       string         `db:"phone"`
		DisplayName sql.NullString `db:"display_name"`
		Avatar      sql.NullString `db:"avatar"`
		Role        int8           `db:"role"`
		Status      int8           `db:"status"`
		LastLoginAt sql.NullTime   `db:"last_login_at"`
		LoginCount  int            `db:"login_count"`
		CreatedAt   time.Time      `db:"created_at"`
		UpdatedAt   time.Time      `db:"updated_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&u.ID, &u.Phone, &u.DisplayName, &u.Avatar,
		&u.Role, &u.Status, &u.LastLoginAt, &u.LoginCount,
		&u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询用户详情失败",
			})
		}
		return
	}

	// 3. 格式化时间
	lastLoginAtStr := ""
	if u.LastLoginAt.Valid {
		lastLoginAtStr = u.LastLoginAt.Time.Format("2006-01-02 15:04:05")
	}

	// 4. 构造返回数据
	userItem := UserItem{
		ID:          u.ID,
		Phone:       u.Phone,
		DisplayName: u.DisplayName.String,
		Avatar:      u.Avatar.String,
		Role:        int(u.Role),
		Status:      int(u.Status),
		LastLoginAt: lastLoginAtStr,
		LoginCount:  u.LoginCount,
		CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   u.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    userItem,
	})
}

// UpdateUserRequest 更新用户请求结构
// 使用指针类型，以便区分“未传递该字段”和“传递了空值”
type UpdateUserRequest struct {
	DisplayName *string `json:"displayName"` // 展示名
	Avatar      *string `json:"avatar"`      // 头像URL
	Role        *int    `json:"role"`        // 角色: 1普通 2专家 9管理员
	Status      *int    `json:"status"`      // 状态: 1正常 0禁用
}

// UpdateUser 更新用户信息
func UpdateUser(c *gin.Context) {
	// 1. 获取路径参数 ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户 ID",
		})
		return
	}

	// 2. 绑定请求参数
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 3. 检查用户是否存在
	checkQuery := "SELECT id FROM user WHERE id = ?"
	var exists int64
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询用户失败",
			})
		}
		return
	}

	// 4. 构建动态更新语句
	updateFields := []string{}
	updateArgs := []interface{}{}

	if req.DisplayName != nil {
		updateFields = append(updateFields, "display_name = ?")
		updateArgs = append(updateArgs, *req.DisplayName)
	}
	if req.Avatar != nil {
		updateFields = append(updateFields, "avatar = ?")
		updateArgs = append(updateArgs, *req.Avatar)
	}
	if req.Role != nil {
		// 可选：验证角色合法性
		if *req.Role != 1 && *req.Role != 2 && *req.Role != 9 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的角色类型",
			})
			return
		}
		updateFields = append(updateFields, "role = ?")
		updateArgs = append(updateArgs, *req.Role)
	}
	if req.Status != nil {
		// 可选：验证状态合法性
		if *req.Status != 0 && *req.Status != 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的状态值",
			})
			return
		}
		updateFields = append(updateFields, "status = ?")
		updateArgs = append(updateArgs, *req.Status)
	}

	// 如果没有提供可更新字段
	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "未提供需要更新的字段",
		})
		return
	}

	// 添加自动更新时间和 ID
	updateFields = append(updateFields, "updated_at = ?")
	updateArgs = append(updateArgs, time.Now())
	updateArgs = append(updateArgs, id)

	// 拼接 SQL
	updateQuery := "UPDATE user SET " + strings.Join(updateFields, ", ") + " WHERE id = ?"

	_, err = db.MySQL.Exec(updateQuery, updateArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DeleteUser 删除用户（软删除：将 status 设为 0）
func DeleteUser(c *gin.Context) {
	// 1. 获取路径参数 ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户 ID",
		})
		return
	}

	// 2. 检查用户是否存在且未删除
	checkQuery := "SELECT id FROM user WHERE id = ? AND status = 1"
	var exists int64
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在或已禁用",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询用户失败",
			})
		}
		return
	}

	// 3. 执行软删除：更新 status 为 0
	deleteQuery := "UPDATE user SET status = 0, updated_at = ? WHERE id = ?"
	_, err = db.MySQL.Exec(deleteQuery, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// UpdateUserRoleRequest 更新用户角色请求结构
type UpdateUserRoleRequest struct {
	Role int `json:"role" binding:"required,min=1,max=9"` // 绑定验证：必填，范围1-9
}

// UpdateUserRole 更新用户角色
func UpdateUserRole(c *gin.Context) {
	// 1. 获取路径参数 ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户 ID",
		})
		return
	}

	// 2. 绑定请求参数
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 3. 验证角色合法性 (根据业务逻辑定义: 1普通 2专家 9管理员)
	allowedRoles := map[int]bool{1: true, 2: true, 9: true}
	if !allowedRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的角色类型",
		})
		return
	}

	// 4. 检查用户是否存在
	checkQuery := "SELECT id FROM user WHERE id = ?"
	var exists int64
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询用户失败",
			})
		}
		return
	}

	// 5. 执行更新
	updateQuery := "UPDATE user SET role = ?, updated_at = ? WHERE id = ?"
	_, err = db.MySQL.Exec(updateQuery, req.Role, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新角色失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// ResetPasswordRequest 重置密码请求结构
type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required,min=6"` // 新密码，至少6位
}

// ResetPassword 重置用户密码
func ResetPassword(c *gin.Context) {
	// 1. 获取路径参数 ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户 ID",
		})
		return
	}

	// 2. 绑定请求参数
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 3. 检查用户是否存在
	checkQuery := "SELECT id FROM user WHERE id = ?"
	var exists int64
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询用户失败",
			})
		}
		return
	}

	// 4. 密码哈希加密
	hashedPassword, err := hash.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "密码加密失败",
		})
		return
	}

	// 5. 执行更新
	updateQuery := "UPDATE user SET password_hash = ?, updated_at = ? WHERE id = ?"
	_, err = db.MySQL.Exec(updateQuery, hashedPassword, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "重置密码失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码重置成功",
	})
}

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	Phone       string `json:"phone" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"` // 初始密码
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Role        int    `json:"role" binding:"required,oneof=1 2 9"` // 1普通 2专家 9管理员
	Status      int    `json:"status" binding:"required,oneof=0 1"` // 0禁用 1正常
}

// CreateUser 创建新用户（管理员后台添加）
func CreateUser(c *gin.Context) {
	// 1. 绑定请求参数
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 2. 检查手机号是否已存在
	var count int
	checkQuery := "SELECT COUNT(*) FROM user WHERE phone = ?"
	err := db.MySQL.QueryRow(checkQuery, req.Phone).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库查询失败",
		})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "手机号已被注册",
		})
		return
	}

	// 3. 密码哈希加密
	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "密码加密失败",
		})
		return
	}

	// 4. 插入新用户
	// 注意：如果 displayName 为空，可以设置一个默认值，或者直接存入空字符串
	displayName := req.DisplayName
	if displayName == "" {
		displayName = "用户" + req.Phone[len(req.Phone)-4:] // 简单默认名
	}

	insertQuery := `
		INSERT INTO user (phone, password_hash, display_name, avatar, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := db.MySQL.Exec(insertQuery, req.Phone, hashedPassword, displayName, req.Avatar, req.Role, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建用户失败",
		})
		return
	}

	// 获取新插入用户的 ID
	newID, err := result.LastInsertId()
	if err != nil {
		// 即使获取ID失败，用户可能已经创建成功，这里可以根据业务需求决定如何处理
		// 通常记录日志即可，不影响返回成功
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data": gin.H{
			"id": newID,
		},
	})
}
