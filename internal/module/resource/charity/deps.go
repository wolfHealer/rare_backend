package charity

import (
	"rare_backend/internal/module/resource/charity/repo"
	"rare_backend/internal/module/resource/charity/service"
)

var (
	projectRepo = repo.NewProjectRepo()
	channelRepo = repo.NewChannelRepo()
	caseRepo    = repo.NewCaseRepo()

	projectSvc = service.NewProjectService(projectRepo)
	channelSvc = service.NewChannelService(channelRepo)
	caseSvc    = service.NewCaseService(caseRepo)
)
