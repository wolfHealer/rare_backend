package service

import (
	"encoding/json"

	"rare_backend/internal/module/community/domain"
	"rare_backend/internal/module/community/repo"
)

type PostService struct {
	repo *repo.PostRepo
}

func NewPostService(r *repo.PostRepo) *PostService {
	return &PostService{repo: r}
}

func (s *PostService) ToggleLike(postID, userID int64) (*domain.LikeResult, error) {
	ok, err := s.repo.ExistsActive(postID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrPostNotFound
	}
	isLiked, err := s.repo.IsLiked(postID, userID)
	if err != nil {
		return nil, err
	}
	if isLiked {
		if err := s.repo.Unlike(postID, userID); err != nil {
			return nil, err
		}
		isLiked = false
	} else {
		if err := s.repo.Like(postID, userID); err != nil {
			return nil, err
		}
		isLiked = true
	}
	likeCount, err := s.repo.RefreshLikeCount(postID)
	if err != nil {
		return nil, err
	}
	return &domain.LikeResult{IsLiked: isLiked, LikeCount: likeCount}, nil
}

func (s *PostService) ToggleFavorite(postID, userID int64) (*domain.FavoriteResult, error) {
	ok, err := s.repo.ExistsActive(postID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrPostNotFound
	}
	isFavorited, err := s.repo.IsFavorited(postID, userID)
	if err != nil {
		return nil, err
	}
	if isFavorited {
		if err := s.repo.Unfavorite(postID, userID); err != nil {
			return nil, err
		}
		isFavorited = false
	} else {
		if err := s.repo.Favorite(postID, userID); err != nil {
			return nil, err
		}
		isFavorited = true
	}
	favoriteCount, err := s.repo.RefreshFavoriteCount(postID)
	if err != nil {
		return nil, err
	}
	return &domain.FavoriteResult{IsFavorited: isFavorited, FavoriteCount: favoriteCount}, nil
}

func (s *PostService) Create(in domain.CreatePostInput) (int64, error) {
	return s.repo.Create(in)
}

func (s *PostService) Delete(postID, userID int64, isAdmin bool) error {
	ownerID, status, err := s.repo.GetOwnerAndStatus(postID)
	if err != nil {
		if s.repo.ErrNoRows(err) {
			return domain.ErrPostNotFound
		}
		return err
	}
	if status == 3 {
		return domain.ErrPostNotFound
	}
	if !isAdmin && ownerID != userID {
		return domain.ErrForbidden
	}
	return s.repo.SoftDelete(postID)
}

func (s *PostService) ListMyPosts(filter domain.MyPostsFilter) ([]domain.PostListItem, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 10
	}
	return s.repo.ListMyPosts(filter)
}

func (s *PostService) Update(postID, currentUserID int64, isAdmin bool, in domain.UpdatePostInput) error {
	ownerID, postStatus, err := s.repo.GetOwnerAndStatus(postID)
	if err != nil {
		if s.repo.ErrNoRows(err) {
			return domain.ErrPostNotFound
		}
		return err
	}
	if !isAdmin && ownerID != currentUserID {
		return domain.ErrForbidden
	}

	fields := map[string]interface{}{}
	if in.Title != nil {
		fields["title"] = *in.Title
	}
	if in.Content != nil {
		fields["content"] = *in.Content
	}
	if in.Images != nil {
		imgJSON, _ := json.Marshal(*in.Images)
		fields["images"] = string(imgJSON)
	}
	if in.DiseaseID != nil {
		fields["disease_id"] = *in.DiseaseID
	}
	if in.CategoryID != nil {
		fields["category_id"] = *in.CategoryID
	}
	if in.Type != nil {
		fields["type"] = *in.Type
	}
	if isAdmin {
		if in.IsTop != nil {
			fields["is_top"] = *in.IsTop
		}
		if in.IsRecommend != nil {
			fields["is_recommend"] = *in.IsRecommend
		}
		if in.Status != nil {
			fields["status"] = *in.Status
		}
		if in.RejectReason != nil {
			fields["reject_reason"] = *in.RejectReason
		}
	} else if postStatus == 2 {
		fields["status"] = 0
	}
	return s.repo.UpdateDynamic(postID, fields)
}
