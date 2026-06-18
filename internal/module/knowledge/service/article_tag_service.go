package service

import (
	"database/sql"
	"strings"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/module/knowledge/repo"
)

type ArticleTagService struct {
	repo *repo.ArticleTagRepo
}

func NewArticleTagService(r *repo.ArticleTagRepo) *ArticleTagService {
	return &ArticleTagService{repo: r}
}

func (s *ArticleTagService) List(filter domain.ArticleTagListFilter) ([]domain.ArticleTagItem, error) {
	return s.repo.List(filter)
}

func (s *ArticleTagService) GetByID(id uint) (*domain.ArticleTagItem, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *ArticleTagService) Create(in domain.CreateArticleTagInput) (int64, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return 0, domain.ErrNameRequired
	}
	tagType := strings.TrimSpace(in.Type)
	if tagType == "" {
		tagType = "general"
	}
	return s.repo.Create(domain.CreateArticleTagInput{Name: name, Type: tagType})
}

func (s *ArticleTagService) Update(id uint, in domain.UpdateArticleTagInput) error {
	ok, err := s.repo.Exists(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotFound
	}

	update := domain.UpdateArticleTagInput{}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return domain.ErrNameRequired
		}
		update.Name = &name
	}
	if in.Type != nil {
		tagType := strings.TrimSpace(*in.Type)
		if tagType == "" {
			tagType = "general"
		}
		update.Type = &tagType
	}
	return s.repo.Update(id, update)
}

func (s *ArticleTagService) Delete(id uint) error {
	ok, err := s.repo.Exists(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotFound
	}
	err = s.repo.Delete(id)
	if err == sql.ErrNoRows {
		return domain.ErrNotFound
	}
	return err
}
