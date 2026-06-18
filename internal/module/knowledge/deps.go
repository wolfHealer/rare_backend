package knowledge

import (
	"rare_backend/internal/module/knowledge/repo"
	"rare_backend/internal/module/knowledge/service"
)

var (
	categoryRepo  = repo.NewCategoryRepo()
	tagRepo       = repo.NewTagRepo()
	diseaseRepo   = repo.NewDiseaseRepo()
	articleRepo   = repo.NewArticleRepo()
	articleTagRepo = repo.NewArticleTagRepo()

	categorySvc   = service.NewCategoryService(categoryRepo)
	tagSvc        = service.NewTagService(tagRepo)
	diseaseSvc    = service.NewDiseaseService(diseaseRepo, categoryRepo)
	articleSvc    = service.NewArticleService(articleRepo)
	articleTagSvc = service.NewArticleTagService(articleTagRepo)
)
