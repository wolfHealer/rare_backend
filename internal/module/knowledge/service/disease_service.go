package service

import (
	"encoding/json"
	"time"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/module/knowledge/repo"
)

type DiseaseService struct {
	diseaseRepo  *repo.DiseaseRepo
	categoryRepo *repo.CategoryRepo
}

func NewDiseaseService(dr *repo.DiseaseRepo, cr *repo.CategoryRepo) *DiseaseService {
	return &DiseaseService{diseaseRepo: dr, categoryRepo: cr}
}

func (s *DiseaseService) Create(in domain.CreateDiseaseInput) (int64, error) {
	if in.Name == "" {
		return 0, domain.ErrNameRequired
	}
	if in.Status != 0 && in.Status != 1 {
		in.Status = 1
	}

	imagesJSON := "[]"
	if len(in.Images) > 0 {
		bytes, err := json.Marshal(in.Images)
		if err != nil {
			return 0, err
		}
		imagesJSON = string(bytes)
	}
	return s.diseaseRepo.Create(in, imagesJSON)
}

func (s *DiseaseService) GetByID(id uint) (*domain.DiseaseItem, error) {
	item, err := s.diseaseRepo.GetActiveByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *DiseaseService) Update(id uint, in domain.UpdateDiseaseInput) error {
	ok, err := s.diseaseRepo.Exists(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotFound
	}

	now := time.Now()
	setParts := []string{}
	args := []interface{}{}

	if in.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *in.Name)
	}
	if in.Alias != nil {
		setParts = append(setParts, "alias = ?")
		args = append(args, *in.Alias)
	}
	if in.Introduction != nil {
		setParts = append(setParts, "introduction = ?")
		args = append(args, *in.Introduction)
	}
	if in.Symptoms != nil {
		setParts = append(setParts, "symptoms = ?")
		args = append(args, *in.Symptoms)
	}
	if in.Images != nil {
		imagesJSON := "[]"
		if len(*in.Images) > 0 {
			bytes, err := json.Marshal(*in.Images)
			if err != nil {
				return err
			}
			imagesJSON = string(bytes)
		}
		setParts = append(setParts, "images = ?")
		args = append(args, imagesJSON)
	}
	if in.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *in.Status)
	}

	hasMainUpdate := len(setParts) > 0
	if hasMainUpdate {
		setParts = append(setParts, "updated_at = ?")
		args = append(args, now)
		args = append(args, id)
	}

	if !hasMainUpdate && in.CategoryIDs == nil && in.PrimaryCategoryID == nil && in.TagIDs == nil {
		return domain.ErrNoUpdateFields
	}

	return s.diseaseRepo.Update(id, setParts, args, in)
}

func (s *DiseaseService) Delete(id uint) error {
	err := s.diseaseRepo.Delete(id)
	if repo.IsNoRows(err) {
		return domain.ErrNotFound
	}
	return err
}

func (s *DiseaseService) List(filter domain.DiseaseListFilter) (*domain.DiseaseListResult, error) {
	return s.diseaseRepo.List(filter)
}

func (s *DiseaseService) ListByCategory(filter domain.DiseasesByCategoryFilter) (*domain.DiseaseListResult, error) {
	ok, err := s.categoryRepo.ExistsActive(filter.CategoryID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrNotFound
	}
	return s.diseaseRepo.ListByCategory(filter)
}

func (s *DiseaseService) Search(keyword string, page, pageSize int) (*domain.SearchDiseasesResult, error) {
	if keyword == "" {
		return nil, domain.ErrKeywordRequired
	}
	return s.diseaseRepo.Search(keyword, page, pageSize)
}

func (s *DiseaseService) Options(keyword string) ([]domain.DiseaseOptionItem, error) {
	return s.diseaseRepo.Options(keyword)
}
