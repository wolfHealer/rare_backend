package community

import (
	"strconv"
	"strings"

	"rare_backend/internal/middleware"
	"rare_backend/internal/module/community/domain"

	"github.com/gin-gonic/gin"
)

// ListMyPosts 我的发布列表
func ListMyPosts(c *gin.Context) {
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

	items, total, err := postSvc.ListMyPosts(domain.MyPostsFilter{
		UserID:   userID,
		Status:   strings.TrimSpace(c.Query("status")),
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}

	records := make([]PostResponse, 0, len(items))
	postIDs := make([]int64, 0, len(items))
	for _, item := range items {
		records = append(records, toPostResponse(item))
		postIDs = append(postIDs, item.ID)
	}
	if len(postIDs) > 0 {
		likedSet, _ := postRepo.BatchLikedPostIDs(userID, postIDs)
		favSet, _ := postRepo.BatchFavoritedPostIDs(userID, postIDs)
		for i := range records {
			records[i].IsLiked = likedSet[records[i].ID]
			records[i].IsFavorited = favSet[records[i].ID]
		}
	}
	respondPage(c, records, total, page, pageSize)
}

// DeleteMyPost 删除我的发布（仅本人）
func DeleteMyPost(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondBadRequest(c, "无效的帖子 ID")
		return
	}
	userID, ok := middleware.MustGetUserID(c)
	if !ok {
		return
	}
	if err := postSvc.Delete(id, userID, false); err != nil {
		respondServiceError(c, err)
		return
	}
	respondOKMessage(c, "删除成功", nil)
}

func toPostResponse(item domain.PostListItem) PostResponse {
	p := PostResponse{
		ID:            item.ID,
		UserID:        item.UserID,
		DisplayName:   item.DisplayName,
		DiseaseID:     item.DiseaseID,
		CategoryID:    item.CategoryID,
		Type:          item.Type,
		Title:         item.Title,
		Content:       item.Content,
		Images:        item.Images,
		ViewCount:     item.ViewCount,
		LikeCount:     item.LikeCount,
		CommentCount:  item.CommentCount,
		FavoriteCount: item.FavoriteCount,
		IsTop:         item.IsTop,
		IsRecommend:   item.IsRecommend,
		Status:        item.Status,
		RejectReason:  item.RejectReason,
		CreatedAt:     item.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if !item.UpdatedAt.IsZero() {
		p.UpdatedAt = item.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	return p
}
