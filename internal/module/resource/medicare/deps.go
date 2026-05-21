package medicare

import (
	"rare_backend/internal/module/resource/medicare/repo"
	"rare_backend/internal/module/resource/medicare/service"
)

var (
	policyRepo = repo.NewPolicyRepo()
	policySvc  = service.NewPolicyService(policyRepo)
)
