package rehab

import (
	"encoding/json"
	"errors"
	"strconv"

	"rare_backend/internal/module/resource/rehab/domain"
	"rare_backend/internal/module/resource/rehab/service"

	"github.com/gin-gonic/gin"
)

func parseTrainingID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondInvalidTrainingID(c)
		return 0, false
	}
	return id, true
}

func GetTrainingList(c *gin.Context) {
	filter := service.BuildTrainingListFilter(
		c.DefaultQuery("diseaseId", ""),
		c.DefaultQuery("rehabStage", ""),
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("auditStatus", ""),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "10"),
	)

	result, err := trainingSvc.List(filter)
	if err != nil {
		var countErr *domain.CountListError
		if errors.As(err, &countErr) {
			respondTrainingListCountError(c, countErr.Cause)
			return
		}
		var queryErr *domain.QueryListError
		if errors.As(err, &queryErr) {
			respondTrainingListError(c, queryErr.Cause)
			return
		}
		respondTrainingListCountError(c, err)
		return
	}

	var listItems []TrainingListItem
	for _, item := range result.List {
		listItems = append(listItems, TrainingListItem{
			ID:              item.ID,
			RehabStage:      item.RehabStage,
			Title:           item.Title,
			TrainPurpose:    item.TrainPurpose,
			TrainContent:    item.TrainContent,
			ForbiddenAction: item.ForbiddenAction,
			PicUrls:         item.PicUrls,
			GuidePdf:        item.GuidePdf,
			GuideWord:       item.GuideWord,
			AuditStatus:     item.AuditStatus,
			RejectReason:    item.RejectReason,
			Sort:            item.Sort,
			DiseaseIds:      item.DiseaseIds,
			Diseases:        toTrainingDiseaseItems(item.Diseases),
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
		})
	}

	respondPage(c, listItems, result.Total, filter.Page, filter.PageSize)
}

func GetTrainingDetail(c *gin.Context) {
	id, ok := parseTrainingID(c)
	if !ok {
		return
	}

	result, err := trainingSvc.GetByID(id)
	if err != nil {
		handleTrainingDetailError(c, err)
		return
	}

	respondOK(c, TrainingDetailDataResponse{
		ID:              result.ID,
		RehabStage:      result.RehabStage,
		Title:           result.Title,
		TrainPurpose:    result.TrainPurpose,
		TrainContent:    result.TrainContent,
		ForbiddenAction: result.ForbiddenAction,
		PicUrls:         result.PicUrls,
		GuidePdf:        result.GuidePdf,
		GuideWord:       result.GuideWord,
		AuditStatus:     result.AuditStatus,
		RejectReason:    result.RejectReason,
		Sort:            result.Sort,
		DiseaseIds:      result.DiseaseIds,
		Diseases:        toTrainingDiseaseItems(result.Diseases),
		CreatedAt:       result.CreatedAt,
		UpdatedAt:       result.UpdatedAt,
	})
}

func GetTrainingResource(c *gin.Context) {
	id, ok := parseTrainingID(c)
	if !ok {
		return
	}

	resourceType := c.DefaultQuery("type", "pdf")

	result, err := trainingSvc.GetResource(id, resourceType)
	if err != nil {
		handleTrainingResourceError(c, err)
		return
	}

	respondOK(c, ResourceResponse{
		DownloadUrl: result.DownloadUrl,
		PreviewUrl:  result.PreviewUrl,
		FileName:    result.FileName,
		FileSize:    result.FileSize,
	})
}

func CreateTraining(c *gin.Context) {
	var req CreateTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	if len(req.PicUrls) > 0 {
		if _, err := json.Marshal(req.PicUrls); err != nil {
			respondTrainingPicUrlsError(c)
			return
		}
	}

	id, err := trainingSvc.Create(toCreateTrainingInput(req))
	if err != nil {
		respondTrainingCreateError(c, err)
		return
	}

	respondOKMessage(c, "创建成功", gin.H{"id": id})
}

func UpdateTraining(c *gin.Context) {
	id, ok := parseTrainingID(c)
	if !ok {
		return
	}

	var req UpdateTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	if req.PicUrls != nil {
		if _, err := json.Marshal(req.PicUrls); err != nil {
			respondTrainingPicUrlsError(c)
			return
		}
	}

	if err := trainingSvc.Update(id, toUpdateTrainingInput(req)); err != nil {
		respondTrainingUpdateError(c, err)
		return
	}

	respondOKMessage(c, "更新成功", nil)
}

func DeleteTraining(c *gin.Context) {
	id, ok := parseTrainingID(c)
	if !ok {
		return
	}

	if err := trainingSvc.Delete(id); err != nil {
		handleTrainingDeleteError(c, err)
		return
	}

	respondOKMessage(c, "删除成功", nil)
}
