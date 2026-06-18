package community

import (
	"strconv"

	"rare_backend/internal/middleware"
	"rare_backend/internal/module/community/domain"

	"github.com/gin-gonic/gin"
)

// ListPostReports 管理端：举报列表
// GET /api/community/post-reports?page=1&pageSize=10&status=0&reasonType=spam&keyword=
func ListPostReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filter := domain.ReportListFilter{
		Page:       page,
		PageSize:   pageSize,
		ReasonType: c.Query("reasonType"),
		Keyword:    c.Query("keyword"),
	}
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			filter.Status = &v
		}
	}

	items, total, err := reportSvc.ListReports(filter)
	if err != nil {
		respondInternalError(c, "查询举报列表失败")
		return
	}
	if items == nil {
		items = []domain.ReportItem{}
	}
	respondPage(c, items, total, page, pageSize)
}

// GetPostReport 管理端：举报详情
// GET /api/community/post-reports/:id
func GetPostReport(c *gin.Context) {
	reportID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || reportID <= 0 {
		respondBadRequest(c, "无效的举报 ID")
		return
	}

	item, err := reportSvc.GetReport(reportID)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	respondOK(c, item)
}

// HandlePostReport 管理端：处理举报
// PUT /api/community/post-reports/:id/handle
func HandlePostReport(c *gin.Context) {
	reportID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || reportID <= 0 {
		respondBadRequest(c, "无效的举报 ID")
		return
	}
	handlerID, ok := middleware.MustGetUserID(c)
	if !ok {
		return
	}

	var req HandleReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}

	if err := reportSvc.HandleReport(domain.HandleReportInput{
		ReportID:   reportID,
		HandlerID:  handlerID,
		Status:     req.Status,
		HandleNote: req.HandleNote,
		RemovePost: req.RemovePost,
	}); err != nil {
		respondServiceError(c, err)
		return
	}
	respondOKMessage(c, "处理成功", nil)
}
