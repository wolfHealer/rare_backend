package domain

import "errors"

var (
	ErrInvalidDrugID     = errors.New("invalid drug id")
	ErrInvalidChannelID  = errors.New("invalid channel id")
	ErrInvalidDonationID = errors.New("invalid donation id")

	ErrDrugNotFound     = errors.New("drug not found")
	ErrChannelNotFound  = errors.New("channel not found")
	ErrDonationNotFound = errors.New("donation not found")

	ErrDrugRelFailed       = errors.New("drug disease relation failed")
	ErrNoUpdateFields      = errors.New("no update fields")
	ErrDrugNotApproved     = errors.New("drug not approved")
	ErrDonationNotApproved = errors.New("donation not approved")

	ErrGenericNameRequired = errors.New("generic name required")
	ErrDrugTypeRequired    = errors.New("drug type required")
	ErrChannelNameRequired = errors.New("channel name required")
	ErrRegionCodeRequired  = errors.New("region code required")
	ErrContactRequired     = errors.New("contact required")
	ErrDonationNameRequired = errors.New("donation name required")
	ErrOrganizerRequired   = errors.New("organizer required")
	ErrDrugIDRequired      = errors.New("drug id required")

	ErrInvalidIDCard       = errors.New("invalid id card")
	ErrApplyInfoIncomplete = errors.New("apply info incomplete")
	ErrGuideNotFound       = errors.New("guide not found")

	ErrCountList = errors.New("count list failed")
	ErrQueryList = errors.New("query list failed")
)
