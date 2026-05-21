package service

import (
	"time"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/module/knowledge/repo"
)

type TagService struct {
	repo *repo.TagRepo
}

func NewTagService(r *repo.TagRepo) *TagService {
	return &TagService{repo: r}
}

func (s *TagService) Create(in domain.CreateTagInput) (int64, error) {
	if in.Name == "" {
		return 0, domain.ErrNameRequired
	}
	if in.Code == "" {
		return 0, domain.ErrCodeRequired
	}
	exists, err := s.repo.CodeExistsActive(in.Code)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, domain.ErrCodeExists
	}
	return s.repo.Create(in)
}

func (s *TagService) Update(id uint, in domain.UpdateTagInput) error {
	ok, err := s.repo.Exists(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotFound
	}

	setParts := []string{}
	args := []interface{}{}

	if in.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *in.Name)
	}
	if in.Code != nil {
		exists, err := s.repo.CodeExistsActiveExcept(*in.Code, id)
		if err != nil {
			return err
		}
		if exists {
			return domain.ErrCodeExists
		}
		setParts = append(setParts, "code = ?")
		args = append(args, *in.Code)
	}
	if in.SortOrder != nil {
		setParts = append(setParts, "sort_order = ?")
		args = append(args, *in.SortOrder)
	}
	if in.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *in.Status)
	}

	if len(setParts) == 0 {
		return domain.ErrNoUpdateFields
	}

	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, id)

	return s.repo.UpdateDynamic(id, setParts, args)
}

func (s *TagService) Delete(id uint) error {
	ok, err := s.repo.ExistsActive(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotFound
	}
	return s.repo.SoftDelete(id)
}

func (s *TagService) List() ([]domain.TagItem, error) {
	return s.repo.ListActive()
}
