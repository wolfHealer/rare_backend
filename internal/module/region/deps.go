package region

import (
	"rare_backend/internal/module/region/repo"
	"rare_backend/internal/module/region/service"
)

var regionSvc = service.NewRegionService(repo.NewRegionRepo())
