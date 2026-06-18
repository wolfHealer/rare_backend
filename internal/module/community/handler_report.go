package community

import (
	"strconv"

	"rare_backend/internal/middleware"
	"rare_backend/internal/module/community/domain"

	"github.com/gin-gonic/gin"
)

// ReportPost 举报帖子
func ReportPost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || postID <= 0 {
		respondBadRequest(c, "无效的帖子 ID")
		return
	}

	userID, ok := middleware.MustGetUserID(c)
	if !ok {
		return
	}

	var req ReportPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}

	if err := reportSvc.ReportPost(domain.ReportPostInput{
		PostID:      postID,
		ReporterID:  userID,
		ReasonType:  req.Reason,
		Description: req.Description,
	}); err != nil {
		respondServiceError(c, err)
		return
	}

	respondOKMessage(c, "举报已提交", nil)
}
