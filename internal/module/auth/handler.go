package auth

import (
	"database/sql"
	"net/http"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/hash"
	"rare_backend/internal/pkg/jwt"

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
