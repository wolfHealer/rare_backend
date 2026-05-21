package medicare

import (
	"strconv"

	"rare_backend/internal/module/resource/medicare/domain"
	"rare_backend/internal/module/resource/medicare/service"

	"github.com/gin-gonic/gin"
)

func parsePolicyID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondInvalidPolicyID(c)
		return 0, false
	}
	return id, true
}

func GetPolicyList(c *gin.Context) {
	filter := service.BuildPolicyListFilter(
		c.DefaultQuery("province_code", ""),
		c.DefaultQuery("city_code", ""),
		c.DefaultQuery("disease_id", ""),
		c.DefaultQuery("need_latest_only", "1"),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("page_size", "10"),
	)

	result, err := policySvc.List(filter)
	if err != nil {
		if err == domain.ErrCountList {
			respondPolicyListCountError(c)
			return
		}
		respondPolicyListError(c)
		return
	}

	respondOK(c, toPolicyListResponse(result))
}

func GetPolicyDetail(c *gin.Context) {
	id, ok := parsePolicyID(c)
	if !ok {
		return
	}

	result, err := policySvc.GetByID(id)
	if err != nil {
		if err == domain.ErrPolicyNotFound {
			respondNotFound(c, "政策不存在")
			return
		}
		respondPolicyDetailError(c)
		return
	}

	respondOK(c, toPolicyDetailResponse(result))
}

func DownloadMaterials(c *gin.Context) {
	materials, err := policySvc.ListMaterials(c.DefaultQuery("type", ""))
	if err != nil {
		respondMaterialsError(c)
		return
	}

	respondOK(c, MaterialResponse{Materials: toMaterialItems(materials)})
}
