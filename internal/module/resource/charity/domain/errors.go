package domain

import "errors"

var (
	ErrInvalidProjectID = errors.New("invalid project id")
	ErrInvalidChannelID = errors.New("invalid channel id")
	ErrInvalidCaseID    = errors.New("invalid case id")

	ErrProjectNotFound = errors.New("project not found")
	ErrChannelNotFound = errors.New("channel not found")
	ErrCaseNotFound    = errors.New("case not found")

	ErrCaseNotAudited = errors.New("case not audited")
	ErrCaseNoPDF      = errors.New("case no pdf")

	ErrNoUpdateFields = errors.New("no update fields")

	ErrCountList = errors.New("count list failed")
	ErrQueryList = errors.New("query list failed")

	ErrDiseaseIDRequired    = errors.New("disease id required")
	ErrProjectIDRequired    = errors.New("project id required")
	ErrCaseTitleRequired    = errors.New("case title required")
	ErrPatientDescRequired  = errors.New("patient desc required")
	ErrApplyCycleRequired   = errors.New("apply cycle required")
	ErrActualReliefRequired = errors.New("actual relief required")
	ErrExperienceRequired   = errors.New("experience required")
	ErrPitfallGuideRequired = errors.New("pitfall guide required")

	ErrDiseaseOptions = errors.New("disease options failed")
)
