package domain

import "errors"

var (
	ErrInvalidHospitalID   = errors.New("invalid hospital id")
	ErrInvalidDoctorID     = errors.New("invalid doctor id")
	ErrInvalidExaminationID = errors.New("invalid examination id")

	ErrHospitalNotFound    = errors.New("hospital not found")
	ErrDoctorNotFound      = errors.New("doctor not found")
	ErrExaminationNotFound = errors.New("examination not found")

	ErrHospitalHasDoctors  = errors.New("hospital has active doctors")
	ErrHospitalNotApproved = errors.New("hospital not approved")

	ErrCountList = errors.New("count list failed")
	ErrQueryList = errors.New("query list failed")

	ErrTxBegin      = errors.New("tx begin failed")
	ErrTxCommit     = errors.New("tx commit failed")
	ErrRelFailed    = errors.New("disease relation failed")
	ErrRelCleanup   = errors.New("relation cleanup failed")
	ErrCreateFailed = errors.New("create failed")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

type HospitalHasDoctorsError struct {
	Count int64
}

func (e *HospitalHasDoctorsError) Error() string {
	return ErrHospitalHasDoctors.Error()
}
