package service

import (
	"database/sql"
	"strings"
	"unicode/utf8"

	"rare_backend/internal/module/auth/domain"
	"rare_backend/internal/module/auth/repo"
)

type FavoriteService struct {
	repo *repo.FavoriteRepo
}

func NewFavoriteService(r *repo.FavoriteRepo) *FavoriteService {
	return &FavoriteService{repo: r}
}

func (s *FavoriteService) List(filter domain.FavoriteListFilter) ([]domain.FavoriteItem, int64, error) {
	targetType := strings.TrimSpace(filter.TargetType)
	switch targetType {
	case "", domain.TargetTypePost:
		items, total, err := s.repo.ListPostFavorites(filter)
		if err != nil {
			return nil, 0, err
		}
		for i := range items {
			items[i].Summary = truncateRunes(items[i].Summary, 120)
			items[i].CreatedAtS = items[i].CreatedAt.Format("2006-01-02 15:04:05")
		}
		return items, total, nil
	case domain.TargetTypeArticle, domain.TargetTypeResource:
		// 暂无对应收藏表，返回空列表
		return []domain.FavoriteItem{}, 0, nil
	default:
		return nil, 0, domain.ErrInvalidTargetType
	}
}

func (s *FavoriteService) RemoveByID(userID, favoriteID int64) error {
	postID, err := s.repo.DeletePostFavoriteByID(userID, favoriteID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrFavoriteNotFound
		}
		return err
	}
	return s.repo.RefreshPostFavoriteCount(postID)
}

func (s *FavoriteService) RemoveByTarget(in domain.RemoveFavoriteByTargetInput) error {
	switch in.TargetType {
	case domain.TargetTypePost:
		if err := s.repo.DeletePostFavoriteByTarget(in.UserID, in.TargetID); err != nil {
			if err == sql.ErrNoRows {
				return domain.ErrFavoriteNotFound
			}
			return err
		}
		return s.repo.RefreshPostFavoriteCount(in.TargetID)
	case domain.TargetTypeArticle, domain.TargetTypeResource:
		return domain.ErrInvalidTargetType
	default:
		return domain.ErrInvalidTargetType
	}
}

func truncateRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	if s == "" || max <= 0 {
		return s
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "..."
}
