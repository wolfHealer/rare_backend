package service

import (
	"time"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/module/knowledge/repo"
)

type CategoryService struct {
	repo *repo.CategoryRepo
}

func NewCategoryService(r *repo.CategoryRepo) *CategoryService {
	return &CategoryService{repo: r}
}

func (s *CategoryService) Create(in domain.CreateCategoryInput) (int64, error) {
	if in.Name == "" {
		return 0, domain.ErrNameRequired
	}
	if in.Code == "" {
		return 0, domain.ErrCodeRequired
	}
	if in.Level < 1 || in.Level > 3 {
		return 0, domain.ErrInvalidLevel
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

func (s *CategoryService) Update(id uint, in domain.UpdateCategoryInput) error {
	ok, err := s.repo.Exists(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotFound
	}

	setParts := []string{}
	args := []interface{}{}

	if in.ParentID != nil {
		setParts = append(setParts, "parent_id = ?")
		args = append(args, *in.ParentID)
	}
	if in.Level != nil {
		if *in.Level < 1 || *in.Level > 3 {
			return domain.ErrInvalidLevel
		}
		setParts = append(setParts, "level = ?")
		args = append(args, *in.Level)
	}
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
	if in.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *in.Description)
	}
	if in.IconURL != nil {
		setParts = append(setParts, "icon_url = ?")
		args = append(args, *in.IconURL)
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

func (s *CategoryService) Delete(id uint) error {
	ok, err := s.repo.ExistsActive(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotFound
	}
	return s.repo.SoftDelete(id)
}

func (s *CategoryService) List(filter domain.CategoryListFilter) ([]domain.CategoryItem, error) {
	return s.repo.List(filter)
}

func (s *CategoryService) Tree(filter domain.CategoryTreeFilter) []domain.CategoryTreeNode {
	all, err := s.repo.ListFlatForTree(filter)
	if err != nil {
		return []domain.CategoryTreeNode{}
	}
	return buildCategoryTree(all)
}

func buildCategoryTree(categories []domain.CategoryTreeNode) []domain.CategoryTreeNode {
	nodeMap := make(map[uint]*domain.CategoryTreeNode)
	var rootNodes []domain.CategoryTreeNode

	for i := range categories {
		nodeMap[categories[i].ID] = &categories[i]
	}

	for i := range categories {
		node := &categories[i]
		if node.ParentID == 0 {
			rootNodes = append(rootNodes, *node)
		} else {
			if parent, exists := nodeMap[node.ParentID]; exists {
				parent.Children = append(parent.Children, *node)
			} else {
				rootNodes = append(rootNodes, *node)
			}
		}
	}
	return rootNodes
}
