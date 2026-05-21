package drug

import (
	"strconv"

	"rare_backend/internal/module/resource/drug/domain"
	"rare_backend/internal/module/resource/drug/service"

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
		DrugID:            req.DrugID,
		Name:              req.Name,
		ChannelType:       req.ChannelType,
		ProvinceCode:      req.ProvinceCode,
		CityCode:          req.CityCode,
		DistrictCode:      req.DistrictCode,
		Address:           req.Address,
		ContactPhone:      req.ContactPhone,
		ContactURL:        req.ContactURL,
		DeliveryScope:     req.DeliveryScope,
		DeliveryCycle:     req.DeliveryCycle,
		IsInsuranceSettle: req.IsInsuranceSettle,
		Qualification:     req.Qualification,
	}
}

func toUpdateChannelInput(req UpdateChannelRequest) domain.UpdateChannelInput {
	return domain.UpdateChannelInput{
		Name:              req.Name,
		ChannelType:       req.ChannelType,
		ProvinceCode:      req.ProvinceCode,
		CityCode:          req.CityCode,
		DistrictCode:      req.DistrictCode,
		Address:           req.Address,
		ContactPhone:      req.ContactPhone,
		ContactURL:        req.ContactURL,
		DeliveryScope:     req.DeliveryScope,
		DeliveryCycle:     req.DeliveryCycle,
		IsInsuranceSettle: req.IsInsuranceSettle,
		Qualification:     req.Qualification,
		DrugID:            req.DrugID,
	}
}

func GetChannelList(c *gin.Context) {
	filter := service.BuildChannelListFilter(
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("provinceCode", ""),
		c.DefaultQuery("cityCode", ""),
		c.DefaultQuery("districtCode", ""),
		c.DefaultQuery("channelType", ""),
		c.DefaultQuery("deliveryScope", ""),
		c.DefaultQuery("auditStatus", ""),
		c.DefaultQuery("isInsuranceSettle", ""),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "10"),
	)

	result, err := channelSvc.List(filter)
	if err != nil {
		if err == domain.ErrCountList {
			respondChannelListCountError(c, err)
			return
		}
		respondChannelListError(c, err)
		return
	}
	respondPage(c, result.List, result.Total, result.Page, result.PageSize)
}

func GetChannelDetail(c *gin.Context) {
	id, ok := parseChannelID(c)
	if !ok {
		return
	}

	result, err := channelSvc.GetByID(id)
	if err != nil {
		respondChannelError(c, err)
		return
	}
	respondOK(c, result)
}

func CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindErrorSimple(c)
		return
	}
	if req.Name == "" {
		respondBadRequest(c, "渠道名称不能为空")
		return
	}
	if req.ProvinceCode == "" || req.CityCode == "" {
		respondBadRequest(c, "省份和城市 Code 不能为空")
		return
	}
	if req.ContactPhone == "" && req.ContactURL == "" {
		respondBadRequest(c, "至少提供一种联系方式")
		return
	}

	id, err := channelSvc.Create(toCreateChannelInput(req))
	if err != nil {
		respondChannelCreateError(c)
		return
	}
	respondOK(c, gin.H{"id": id})
}

func UpdateChannel(c *gin.Context) {
	id, ok := parseChannelID(c)
	if !ok {
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindErrorSimple(c)
		return
	}

	if err := channelSvc.Update(id, toUpdateChannelInput(req)); err != nil {
		respondChannelError(c, err)
		return
	}
	respondOK(c, nil)
}

func DeleteChannel(c *gin.Context) {
	id, ok := parseChannelID(c)
	if !ok {
		return
	}

	if err := channelSvc.Delete(id); err != nil {
		respondChannelError(c, err)
		return
	}
	respondOK(c, nil)
}

func ContactChannel(c *gin.Context) {
	id, ok := parseChannelID(c)
	if !ok {
		return
	}

	var req ContactChannelRequest
	_ = c.ShouldBindJSON(&req)

	result, err := channelSvc.Contact(id, domain.ChannelContactInput{ContactType: req.ContactType})
	if err != nil {
		respondChannelError(c, err)
		return
	}
	respondOK(c, result)
}
