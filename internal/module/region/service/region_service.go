package service

import (
	"rare_backend/internal/module/region/domain"
	"rare_backend/internal/module/region/repo"
)

type RegionService struct {
	repo *repo.RegionRepo
}

func NewRegionService(r *repo.RegionRepo) *RegionService {
	return &RegionService{repo: r}
}

func (s *RegionService) Create(in domain.CreateRegionInput) (int64, error) {
	exists, err := s.repo.CodeExists(in.Code)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, domain.ErrCodeExists
	}
	if in.ParentCode != "" {
		ok, err := s.repo.ParentCodeExists(in.ParentCode)
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, domain.ErrParentNotFound
		}
	}
	if in.IsEnabled != 0 && in.IsEnabled != 1 {
		in.IsEnabled = 1
	}
	return s.repo.Create(in)
}

func (s *RegionService) Update(id uint, in domain.UpdateRegionInput) error {
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
	if in.FullName != nil {
		setParts = append(setParts, "full_name = ?")
		args = append(args, *in.FullName)
	}
	if in.ParentCode != nil {
		if *in.ParentCode != "" {
			ok, err := s.repo.ParentCodeExists(*in.ParentCode)
			if err != nil {
				return err
			}
			if !ok {
				return domain.ErrParentNotFound
			}
		}
		setParts = append(setParts, "parent_code = ?")
		args = append(args, *in.ParentCode)
	}
	if in.Level != nil {
		if *in.Level < 1 || *in.Level > 3 {
			return domain.ErrInvalidLevel
		}
		setParts = append(setParts, "level = ?")
		args = append(args, *in.Level)
	}
	if in.Sort != nil {
		setParts = append(setParts, "sort = ?")
		args = append(args, *in.Sort)
	}
	if in.IsEnabled != nil {
		setParts = append(setParts, "is_enabled = ?")
		args = append(args, *in.IsEnabled)
	}

	return s.repo.Update(id, setParts, args)
}

func (s *RegionService) Delete(id uint) error {
	code, err := s.repo.GetCodeByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return domain.ErrNotFound
		}
		return err
	}
	count, err := s.repo.CountChildren(code)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrHasChildren
	}
	return s.repo.Delete(id)
}

func (s *RegionService) List(filter domain.RegionListFilter) (*domain.RegionListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 1000 {
		filter.PageSize = 100
	}
	return s.repo.List(filter)
}

func (s *RegionService) GetByID(id uint) (*domain.RegionItem, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *RegionService) GetTree(isEnabled int) ([]*domain.RegionTreeNode, error) {
	flatList, err := s.repo.ListFlat(isEnabled, 0)
	if err != nil {
		return nil, err
	}
	return buildRegionTree(flatList), nil
}

func (s *RegionService) GetProvinceCityTree(isEnabled int) ([]*domain.RegionTreeNode, error) {
	flatList, err := s.repo.ListFlat(isEnabled, 2)
	if err != nil {
		return nil, err
	}
	return buildRegionTree(flatList), nil
}

func buildRegionTree(flatList []domain.FlatRegion) []*domain.RegionTreeNode {
	nodeMap := make(map[string]*domain.RegionTreeNode)
	var rootNodes []*domain.RegionTreeNode

	for _, r := range flatList {
		node := &domain.RegionTreeNode{
			Code:     r.Code,
			Name:     r.Name,
			Level:    r.Level,
			Children: []*domain.RegionTreeNode{},
		}
		nodeMap[r.Code] = node
	}

	for _, r := range flatList {
		node := nodeMap[r.Code]
		if r.ParentCode == "" || r.ParentCode == "0" {
			rootNodes = append(rootNodes, node)
		} else if parent, exists := nodeMap[r.ParentCode]; exists {
			parent.Children = append(parent.Children, node)
		} else {
			rootNodes = append(rootNodes, node)
		}
	}
	return rootNodes
}
