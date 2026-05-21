package domain

import "errors"

var (
	ErrInvalidInstitutionID = errors.New("invalid institution id")
	ErrInvalidPsychOrgID    = errors.New("invalid psych org id")
	ErrInvalidTrainingID    = errors.New("invalid training id")

	ErrInstitutionNotFound = errors.New("institution not found")
	ErrPsychOrgNotFound    = errors.New("psych org not found")
	ErrTrainingNotFound    = errors.New("training not found")

	ErrResourceNotFound = errors.New("resource not found")

	ErrCountList  = errors.New("count list failed")
	ErrQueryList  = errors.New("query list failed")
	ErrQueryDetail = errors.New("query detail failed")

	ErrTxBegin      = errors.New("tx begin failed")
	ErrTxCommit     = errors.New("tx commit failed")
	ErrRelFailed    = errors.New("disease relation failed")
	ErrRelCleanup   = errors.New("relation cleanup failed")
	ErrCreateFailed = errors.New("create failed")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
	ErrLastInsertID = errors.New("last insert id failed")
	ErrPicUrlsJSON  = errors.New("pic urls json failed")

	ErrProvinceOptions = errors.New("province options failed")
	ErrDiseaseOptions  = errors.New("disease options failed")
)

type CountListError struct {
	Cause error
}

func (e *CountListError) Error() string {
	return e.Cause.Error()
}

type QueryListError struct {
	Cause error
}

func (e *QueryListError) Error() string {
	return e.Cause.Error()
}
