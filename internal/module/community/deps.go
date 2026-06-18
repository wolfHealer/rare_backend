package community

import (
	"rare_backend/internal/module/community/repo"
	"rare_backend/internal/module/community/service"
)

var (
	postRepo    = repo.NewPostRepo()
	commentRepo = repo.NewCommentRepo()
	reportRepo  = repo.NewReportRepo()
	postSvc     = service.NewPostService(postRepo)
	commentSvc  = service.NewCommentService(postRepo, commentRepo)
	reportSvc   = service.NewReportService(postRepo, reportRepo)
)
