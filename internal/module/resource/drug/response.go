package drug

import (
	"errors"
	"fmt"

	"rare_backend/internal/module/resource/drug/domain"
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

func respondInternalError(c *gin.Context, message string) {
	response.InternalError(c, message)
}

func respondPage(c *gin.Context, list any, total int64, page, pageSize int) {
	response.Page(c, list, total, page, pageSize)
}

func respondBindError(c *gin.Context, err error) {
	response.BadRequest(c, "参数错误")
}

func respondBindErrorSimple(c *gin.Context) {
	response.BadRequest(c, "参数错误")
}

func respondDrugError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidDrugID):
		response.BadRequest(c, "无效的药品 ID")
	case errors.Is(err, domain.ErrDrugNotFound):
		response.NotFound(c, "药品不存在")
	case errors.Is(err, domain.ErrGenericNameRequired):
		response.BadRequest(c, "通用名不能为空")
	case errors.Is(err, domain.ErrDrugTypeRequired):
		response.BadRequest(c, "药品类型不能为空")
	case errors.Is(err, domain.ErrDrugRelFailed):
		response.InternalError(c, "新增药品成功，但疾病关联失败")
	default:
		response.InternalError(c, "服务器错误")
	}
}

func respondDrugCreateError(c *gin.Context, msg string) {
	response.InternalError(c, msg)
}

func respondDrugUpdateError(c *gin.Context, err error) {
	response.InternalError(c, "更新药品主表失败")
}

func respondDrugListCountError(c *gin.Context, err error) {
	response.InternalError(c, "查询总数失败")
}

func respondDrugListError(c *gin.Context, err error) {
	response.InternalError(c, "查询列表失败")
}

func respondDrugDetailError(c *gin.Context, err error) {
	response.InternalError(c, "查询药品详情失败")
}

func respondDrugDeleteRelError(c *gin.Context, msg string) {
	response.InternalError(c, msg)
}

func respondChannelError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidChannelID):
		response.BadRequest(c, "无效的渠道 ID")
	case errors.Is(err, domain.ErrChannelNotFound):
		response.NotFound(c, "渠道不存在")
	case errors.Is(err, domain.ErrChannelNameRequired):
		response.BadRequest(c, "渠道名称不能为空")
	case errors.Is(err, domain.ErrRegionCodeRequired):
		response.BadRequest(c, "省份和城市 Code 不能为空")
	case errors.Is(err, domain.ErrContactRequired):
		response.BadRequest(c, "至少提供一种联系方式")
	case errors.Is(err, domain.ErrNoUpdateFields):
		response.BadRequest(c, "未提供更新字段")
	default:
		response.InternalError(c, "服务器错误")
	}
}

func respondChannelListCountError(c *gin.Context, err error) {
	response.InternalError(c, "查询总数失败")
}

func respondChannelListError(c *gin.Context, err error) {
	response.InternalError(c, "查询列表失败")
}

func respondChannelDetailError(c *gin.Context, err error) {
	response.InternalError(c, "查询渠道详情失败")
}

func respondChannelCreateError(c *gin.Context) {
	response.InternalError(c, "新增渠道失败")
}

func respondChannelUpdateError(c *gin.Context) {
	response.InternalError(c, "更新渠道失败")
}

func respondChannelDeleteError(c *gin.Context) {
	response.InternalError(c, "删除渠道失败")
}

func respondChannelContactError(c *gin.Context) {
	response.InternalError(c, "查询联系方式失败")
}

func respondDonationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidDonationID):
		response.BadRequest(c, "无效的项目 ID")
	case errors.Is(err, domain.ErrDonationNotFound):
		response.NotFound(c, "赠药项目不存在")
	case errors.Is(err, domain.ErrDonationNameRequired):
		response.BadRequest(c, "项目名称不能为空")
	case errors.Is(err, domain.ErrOrganizerRequired):
		response.BadRequest(c, "主办方不能为空")
	case errors.Is(err, domain.ErrDrugIDRequired):
		response.BadRequest(c, "关联药品 ID 不能为空")
	case errors.Is(err, domain.ErrDrugNotApproved):
		response.BadRequest(c, "关联药品不存在")
	case errors.Is(err, domain.ErrNoUpdateFields):
		response.BadRequest(c, "未提供更新字段")
	case errors.Is(err, domain.ErrApplyInfoIncomplete):
		response.BadRequest(c, "请填写完整信息")
	case errors.Is(err, domain.ErrInvalidIDCard):
		response.BadRequest(c, "身份证号格式错误")
	case errors.Is(err, domain.ErrGuideNotFound):
		response.NotFound(c, "指南文件不存在")
	default:
		response.InternalError(c, "服务器错误")
	}
}

func respondDonationListCountError(c *gin.Context, err error) {
	response.InternalError(c, "查询总数失败")
}

func respondDonationListError(c *gin.Context) {
	response.InternalError(c, "查询列表失败")
}

func respondDonationDetailError(c *gin.Context, err error) {
	response.InternalError(c, "查询项目详情失败")
}

func respondDonationApplySuccess(c *gin.Context, data any) {
	response.OKMessage(c, "申请提交成功", data)
}

func respondDonationApplyError(c *gin.Context) {
	response.InternalError(c, "提交申请失败")
}

func respondDonationProgressError(c *gin.Context) {
	response.InternalError(c, "查询进度失败")
}

func respondDonationProgressLogError(c *gin.Context) {
	response.InternalError(c, "查询日志失败")
}

func respondDonationOptionsDiseaseError(c *gin.Context) {
	response.InternalError(c, "查询疾病选项失败")
}

func respondDonationOptionsDrugError(c *gin.Context) {
	response.InternalError(c, "查询药品选项失败")
}

func respondExportError(c *gin.Context, msg string) {
	response.InternalError(c, msg)
}

func respondExcelWriteError(c *gin.Context) {
	response.InternalError(c, "生成 Excel 文件失败")
}

func setExcelHeaders(c *gin.Context, filename string) {
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
}

func respondGuideDownloadError(c *gin.Context) {
	response.InternalError(c, "查询指南失败")
}
