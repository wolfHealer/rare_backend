package drug

import (
	"rare_backend/internal/module/resource/drug/repo"
	"rare_backend/internal/module/resource/drug/service"
)

var (
	drugRepo     = repo.NewDrugRepo()
	channelRepo  = repo.NewChannelRepo()
	donationRepo = repo.NewDonationRepo()

	drugSvc     = service.NewDrugService(drugRepo)
	channelSvc  = service.NewChannelService(channelRepo)
	donationSvc = service.NewDonationService(donationRepo)
)
