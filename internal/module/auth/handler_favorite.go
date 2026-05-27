package auth

import (
	"errors"
	"strconv"
	"strings"

	"rare_backend/internal/middleware"
	"rare_backend/internal/module/auth/domain"
	"rare_backend/internal/module/auth/repo"
	"rare_backend/internal/module/auth/service"
	"rare_backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

var favoriteSvc = service.NewFavoriteService(repo.NewFavoriteRepo())

type removeFavoriteByTargetRequest struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   int64  `json:"target_id" binding:"required"`
}

func listFavorites(c *gin.Context) {
	userID, ok := middleware.MustGetUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "0"))
	if pageSize <= 0 {
		pageSize, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	items, total, err := favoriteSvc.List(domain.FavoriteListFilter{
		UserID:     userID,
		TargetType: strings.TrimSpace(c.Query("target_type")),
		Keyword:    strings.TrimSpace(c.Query("keyword")),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		respondFavoriteError(c, err)
		return
	}
	if items == nil {
		items = []domain.FavoriteItem{}
	}
	respondPage(c, items, total, page, pageSize)
}

func removeFavoriteByID(c *gin.Context) {
	userID, ok := middleware.MustGetUserID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondBadRequest(c, "无效的收藏 ID")
		return
	}
	if err := favoriteSvc.RemoveByID(userID, id); err != nil {
		respondFavoriteError(c, err)
		return
	}
	respondOKMessage(c, "取消收藏成功", nil)
}

func removeFavoriteByTarget(c *gin.Context) {
	userID, ok := middleware.MustGetUserID(c)
	if !ok {
		return
	}
	var req removeFavoriteByTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	if err := favoriteSvc.RemoveByTarget(domain.RemoveFavoriteByTargetInput{
		UserID:     userID,
		TargetType: strings.TrimSpace(req.TargetType),
		TargetID:   req.TargetID,
	}); err != nil {
		respondFavoriteError(c, err)
		return
	}
	respondOKMessage(c, "取消收藏成功", nil)
}

func respondFavoriteError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrFavoriteNotFound):
		response.NotFound(c, "收藏记录不存在")
	case errors.Is(err, domain.ErrInvalidTargetType):
		response.BadRequest(c, "不支持的 target_type")
	default:
		respondInternalError(c, "服务器错误")
	}
}
