package auth

import (
	"rare_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// deactivateAccount 用户自助注销账号
func deactivateAccount(c *gin.Context) {
	userID, ok := middleware.MustGetUserID(c)
	if !ok {
		return
	}
	if err := userSvc.DeactivateAccount(userID); err != nil {
		respondServiceError(c, err)
		return
	}
	respondOKMessage(c, "账号已注销", nil)
}
