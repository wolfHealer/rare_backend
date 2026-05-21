package rehab

import (
	"rare_backend/internal/module/resource/rehab/repo"
	"rare_backend/internal/module/resource/rehab/service"
)

var (
	institutionRepo   = repo.NewInstitutionRepo()
	optionsRepo       = repo.NewOptionsRepo()
	psychologicalRepo = repo.NewPsychologicalRepo()
	trainingRepo      = repo.NewTrainingRepo()

	institutionSvc   = service.NewInstitutionService(institutionRepo, optionsRepo)
	psychologicalSvc = service.NewPsychologicalService(psychologicalRepo)
	trainingSvc      = service.NewTrainingService(trainingRepo)
)
