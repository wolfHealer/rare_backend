package charity

import (
	"errors"

	"rare_backend/internal/module/resource/charity/domain"
	"rare_backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func respondOK(c *gin.Context, data any) {
	response.OK(c, data)
}

func respondBadRequest(c *gin.Context, message string) {
	response.BadRequest(c, message)
}

func respondNotFound(c *gin.Context, message string) {
	response.NotFound(c, message)
}

func respondOKMessage(c *gin.Context, message string, data any) {
	response.OKMessage(c, message, data)
}

func respondPage(c *gin.Context, list any, total int64, page, pageSize int) {
	response.Page(c, list, total, page, pageSize)
}

func respondCreated(c *gin.Context, message string, data any) {
	response.OKMessage(c, message, data)
}

func respondProjectError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidProjectID):
		response.BadRequest(c, "无效的项目 ID")
	case errors.Is(err, domain.ErrProjectNotFound):
		response.NotFound(c, "项目不存在")
	case errors.Is(err, domain.ErrNoUpdateFields):
		response.BadRequest(c, "未提供更新字段")
	default:
		response.InternalError(c, "服务器错误")
	}
}

func respondProjectDetailError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrProjectNotFound) {
		response.NotFound(c, "项目不存在")
		return
	}
	response.InternalError(c, "查询项目详情失败")
}

func respondProjectListCountError(c *gin.Context) {
	response.InternalError(c, "查询总数失败")
}

func respondProjectListError(c *gin.Context) {
	response.InternalError(c, "查询列表失败")
}

func respondProjectCreateError(c *gin.Context, msg string) {
	response.InternalError(c, msg)
}

func respondProjectUpdateError(c *gin.Context, err error) {
	response.InternalError(c, "更新项目失败")
}

func respondProjectDeleteError(c *gin.Context, msg string) {
	response.InternalError(c, msg)
}

func respondProjectOptionsError(c *gin.Context) {
	response.InternalError(c, "查询项目选项失败")
}

func respondChannelError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidChannelID):
		response.BadRequest(c, "无效的渠道 ID")
	case errors.Is(err, domain.ErrChannelNotFound):
		response.NotFound(c, "渠道不存在")
	case errors.Is(err, domain.ErrNoUpdateFields):
		response.BadRequest(c, "未提供更新字段")
	default:
		response.InternalError(c, "服务器错误")
	}
}

func respondChannelListCountError(c *gin.Context) {
	response.InternalError(c, "查询总数失败")
}

func respondChannelListError(c *gin.Context) {
	response.InternalError(c, "查询渠道列表失败")
}

func respondChannelDetailError(c *gin.Context) {
	response.InternalError(c, "查询渠道详情失败")
}

func respondChannelCreateError(c *gin.Context) {
	response.InternalError(c, "创建渠道失败")
}

func respondChannelUpdateError(c *gin.Context) {
	response.InternalError(c, "更新渠道失败")
}

func respondChannelDeleteError(c *gin.Context) {
	response.InternalError(c, "删除渠道失败")
}

func respondCaseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidCaseID):
		response.BadRequest(c, "无效的案例 ID")
	case errors.Is(err, domain.ErrCaseNotFound):
		response.NotFound(c, "案例不存在")
	case errors.Is(err, domain.ErrCaseNotAudited):
		response.NotFound(c, "案例不存在或未审核")
	case errors.Is(err, domain.ErrCaseNoPDF):
		response.NotFound(c, "该案例暂无 PDF 版本")
	case errors.Is(err, domain.ErrNoUpdateFields):
		response.BadRequest(c, "未提供更新字段")
	case errors.Is(err, domain.ErrDiseaseIDRequired):
		response.BadRequest(c, "疾病ID不能为空")
	case errors.Is(err, domain.ErrProjectIDRequired):
		response.BadRequest(c, "关联项目ID不能为空")
	case errors.Is(err, domain.ErrCaseTitleRequired):
		response.BadRequest(c, "案例标题不能为空")
	case errors.Is(err, domain.ErrPatientDescRequired):
		response.BadRequest(c, "患者描述不能为空")
	case errors.Is(err, domain.ErrApplyCycleRequired):
		response.BadRequest(c, "申请周期不能为空")
	case errors.Is(err, domain.ErrActualReliefRequired):
		response.BadRequest(c, "实际救助金额不能为空")
	case errors.Is(err, domain.ErrExperienceRequired):
		response.BadRequest(c, "申请经验分享不能为空")
	case errors.Is(err, domain.ErrPitfallGuideRequired):
		response.BadRequest(c, "申请避坑要点不能为空")
	default:
		response.InternalError(c, "服务器错误")
	}
}

func respondCaseListCountError(c *gin.Context) {
	response.InternalError(c, "查询总数失败")
}

func respondCaseListError(c *gin.Context) {
	response.InternalError(c, "查询列表失败")
}

func respondCaseDetailError(c *gin.Context) {
	response.InternalError(c, "查询案例详情失败")
}

func respondCaseCreateError(c *gin.Context) {
	response.InternalError(c, "创建案例失败")
}

func respondCaseUpdateError(c *gin.Context) {
	response.InternalError(c, "更新案例失败")
}

func respondCaseDeleteError(c *gin.Context) {
	response.InternalError(c, "删除案例失败")
}

func respondCasePDFQueryError(c *gin.Context) {
	response.InternalError(c, "查询案例失败")
}

func respondCaseDiseaseOptionsError(c *gin.Context) {
	response.InternalError(c, "查询疾病选项失败")
}

func respondBindError(c *gin.Context, err error) {
	response.BadRequest(c, "参数错误")
}

func respondBindErrorSimple(c *gin.Context) {
	response.BadRequest(c, "参数错误")
}
