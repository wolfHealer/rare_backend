package service

import (
	"time"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/module/knowledge/repo"
)

type ArticleService struct {
	repo *repo.ArticleRepo
}

func NewArticleService(r *repo.ArticleRepo) *ArticleService {
	return &ArticleService{repo: r}
}

func (s *ArticleService) Create(in domain.CreateArticleInput) (int64, error) {
	return s.repo.Create(in)
}

func (s *ArticleService) GetByID(id uint) (*domain.ArticleDetailResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *ArticleService) Update(id uint, in domain.UpdateArticleInput) error {
	setParts := []string{}
	args := []interface{}{}

	if in.Title != nil {
		setParts = append(setParts, "title=?")
		args = append(args, *in.Title)
	}
	if in.Summary != nil {
		setParts = append(setParts, "summary=?")
		args = append(args, *in.Summary)
	}
	if in.CoverImage != nil {
		setParts = append(setParts, "cover_image=?")
		args = append(args, *in.CoverImage)
	}
	if in.ContentType != nil {
		setParts = append(setParts, "content_type=?")
		args = append(args, *in.ContentType)
	}
	if in.AuthorID != nil {
		setParts = append(setParts, "author_id=?")
		args = append(args, *in.AuthorID)
	}
	if in.SourceName != nil {
		setParts = append(setParts, "source_name=?")
		args = append(args, *in.SourceName)
	}
	if in.SourceURL != nil {
		setParts = append(setParts, "source_url=?")
		args = append(args, *in.SourceURL)
	}
	if in.Status != nil {
		setParts = append(setParts, "status=?")
		args = append(args, *in.Status)
	}
	if in.PublishTime != nil {
		if *in.PublishTime == "" {
			setParts = append(setParts, "publish_time=NULL")
		} else {
			setParts = append(setParts, "publish_time=?")
			args = append(args, *in.PublishTime)
		}
	}
	if in.IsTop != nil {
		setParts = append(setParts, "is_top=?")
		args = append(args, *in.IsTop)
	}
	if in.IsRecommend != nil {
		setParts = append(setParts, "is_recommend=?")
		args = append(args, *in.IsRecommend)
	}
	if in.SeoTitle != nil {
		setParts = append(setParts, "seo_title=?")
		args = append(args, *in.SeoTitle)
	}
	if in.SeoKeywords != nil {
		setParts = append(setParts, "seo_keywords=?")
		args = append(args, *in.SeoKeywords)
	}
	if in.SeoDescription != nil {
		setParts = append(setParts, "seo_description=?")
		args = append(args, *in.SeoDescription)
	}

	if len(setParts) > 0 {
		setParts = append(setParts, "updated_at=?")
		args = append(args, time.Now())
		args = append(args, id)
	}

	return s.repo.Update(id, setParts, args, in)
}

func (s *ArticleService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *ArticleService) List(filter domain.ArticleListFilter) (*domain.ArticleListResult, error) {
	return s.repo.List(filter)
}
