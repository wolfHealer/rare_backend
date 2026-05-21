package charity

import (
	"strconv"

	"rare_backend/internal/module/resource/charity/domain"
	"rare_backend/internal/module/resource/charity/service"

	"github.com/gin-gonic/gin"
)

func parseChannelID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "无效的渠道 ID")
		return 0, false
	}
	return uint(id), true
}

func toCreateChannelInput(req CreateChannelRequest) domain.CreateChannelInput {
	return domain.CreateChannelInput{
		ChannelType:          req.ChannelType,
		Name:                 req.Name,
		ApplyCondition:       req.ApplyCondition,
		ResponseTime:         req.ResponseTime,
		ContactPhone:         req.ContactPhone,
		ContactUrl:           req.ContactUrl,
		HelpLetterTemplate:   req.HelpLetterTemplate,
		CrowdfundingTemplate: req.CrowdfundingTemplate,
		Sort:                 req.Sort,
		AuditStatus:          req.AuditStatus,
	}
}

func toUpdateChannelInput(req UpdateChannelRequest) domain.UpdateChannelInput {
	return domain.UpdateChannelInput{
		ChannelType:          req.ChannelType,
		Name:                 req.Name,
		ApplyCondition:       req.ApplyCondition,
		ResponseTime:         req.ResponseTime,
		ContactPhone:         req.ContactPhone,
		ContactUrl:           req.ContactUrl,
		HelpLetterTemplate:   req.HelpLetterTemplate,
		CrowdfundingTemplate: req.CrowdfundingTemplate,
		Sort:                 req.Sort,
		AuditStatus:          req.AuditStatus,
		RejectReason:         req.RejectReason,
	}
}

func GetChannels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	filter := service.BuildChannelListFilter(
		c.DefaultQuery("channelType", ""),
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("auditStatus", ""),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "10"),
	)

	result, err := channelSvc.List(filter)
	if err != nil {
		if err == domain.ErrCountList {
			respondChannelListCountError(c)
			return
		}
		respondChannelListError(c)
		return
	}

	respondPage(c, result.List, result.Total, page, pageSize)
}

func GetChannelDetail(c *gin.Context) {
	id, ok := parseChannelID(c)
	if !ok {
		return
	}

	result, err := channelSvc.GetByID(id)
	if err != nil {
		if err == domain.ErrChannelNotFound {
			respondNotFound(c, "渠道不存在")
			return
		}
		respondChannelDetailError(c)
		return
	}
	respondOK(c, result)
}

func CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	id, err := channelSvc.Create(toCreateChannelInput(req))
	if err != nil {
		respondChannelCreateError(c)
		return
	}

	respondCreated(c, "创建成功", gin.H{"id": id})
}

func UpdateChannel(c *gin.Context) {
	id, ok := parseChannelID(c)
	if !ok {
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	if err := channelSvc.Update(id, toUpdateChannelInput(req)); err != nil {
		if err == domain.ErrNoUpdateFields {
			respondBadRequest(c, "未提供更新字段")
			return
		}
		respondChannelUpdateError(c)
		return
	}

	respondOKMessage(c, "更新成功", nil)
}

func DeleteChannel(c *gin.Context) {
	id, ok := parseChannelID(c)
	if !ok {
		return
	}

	if err := channelSvc.Delete(id); err != nil {
		if err == domain.ErrChannelNotFound {
			respondNotFound(c, "渠道不存在")
			return
		}
		respondChannelDeleteError(c)
		return
	}

	respondOKMessage(c, "删除成功", nil)
}
