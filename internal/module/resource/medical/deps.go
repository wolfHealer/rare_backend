package medical

import (
	"rare_backend/internal/module/resource/medical/repo"
	"rare_backend/internal/module/resource/medical/service"
)

var (
	diseaseRelRepo = repo.NewDiseaseRelRepo()
	hospitalRepo   = repo.NewHospitalRepo(diseaseRelRepo)
	doctorRepo     = repo.NewDoctorRepo(diseaseRelRepo, hospitalRepo)
	examinationRepo = repo.NewExaminationRepo(diseaseRelRepo)

	hospitalSvc    = service.NewHospitalService(hospitalRepo, diseaseRelRepo)
	doctorSvc      = service.NewDoctorService(doctorRepo, diseaseRelRepo)
	examinationSvc = service.NewExaminationService(examinationRepo, diseaseRelRepo)
)
