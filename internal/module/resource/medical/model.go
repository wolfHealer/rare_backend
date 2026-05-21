package medical

import "rare_backend/internal/module/resource/medical/domain"

type CreateHospitalRequest struct {
	Name          string   `json:"name" binding:"required"`
	ProvinceCode  string   `json:"provinceCode" binding:"required"`
	CityCode      string   `json:"cityCode" binding:"required"`
	DistrictCode  string   `json:"districtCode"`
	ProvinceName  string   `json:"provinceName"`
	CityName      string   `json:"cityName"`
	DistrictName  string   `json:"districtName"`
	Level         string   `json:"level" binding:"required"`
	IsRareNetwork int8     `json:"isRareNetwork"`
	TreatScope    string   `json:"treatScope"`
	Address       string   `json:"address" binding:"required"`
	Phone         string   `json:"phone" binding:"required"`
	HospitalURL   string   `json:"hospitalUrl"`
	DiseaseIDs    []uint64 `json:"diseaseIds"`
}

type UpdateHospitalRequest struct {
	Name          string   `json:"name"`
	ProvinceCode  *string  `json:"provinceCode"`
	CityCode      *string  `json:"cityCode"`
	DistrictCode  *string  `json:"districtCode"`
	ProvinceName  *string  `json:"provinceName"`
	CityName      *string  `json:"cityName"`
	DistrictName  *string  `json:"districtName"`
	Level         string   `json:"level"`
	IsRareNetwork *int8    `json:"isRareNetwork"`
	TreatScope    *string  `json:"treatScope"`
	Address       string   `json:"address"`
	Phone         string   `json:"phone"`
	HospitalURL   *string  `json:"hospitalUrl"`
	AuditStatus   *int8    `json:"auditStatus"`
	RejectReason  *string  `json:"rejectReason"`
	DiseaseIDs    []uint64 `json:"diseaseIds"`
}

type CreateDoctorRequest struct {
	Name       string   `json:"name" binding:"required"`
	Title      string   `json:"title" binding:"required"`
	Department string   `json:"department" binding:"required"`
	GoodAt     string   `json:"goodAt"`
	ClinicTime string   `json:"clinicTime"`
	Contact    string   `json:"contact"`
	HospitalID uint64   `json:"hospitalId" binding:"required"`
	DiseaseIDs []uint64 `json:"diseaseIds"`
	AuditStatus *int8   `json:"auditStatus"`
	Score       *float64 `json:"score"`
	CommentNum  *int     `json:"commentNum"`
}

type UpdateDoctorRequest struct {
	Name         string   `json:"name"`
	Title        string   `json:"title"`
	Department   string   `json:"department"`
	GoodAt       *string  `json:"goodAt"`
	ClinicTime   *string  `json:"clinicTime"`
	Contact      *string  `json:"contact"`
	HospitalID   *uint64  `json:"hospitalId"`
	AuditStatus  *int8    `json:"auditStatus"`
	RejectReason *string  `json:"rejectReason"`
	DiseaseIDs   []uint64 `json:"diseaseIds"`
}

type TemplatesRequest struct {
	Excel   string `json:"excel"`
	Word    string `json:"word"`
	Compare string `json:"compare"`
}

type CreateExaminationRequest struct {
	ExamName          string            `json:"examName" binding:"required"`
	ExamType          string            `json:"examType" binding:"required"`
	ExamPurpose       string            `json:"examPurpose" binding:"required"`
	ReferenceValue    string            `json:"referenceValue"`
	AbnormalInterpret string            `json:"abnormalInterpret"`
	SampleNotes       string            `json:"sampleNotes"`
	Institution       string            `json:"institution"`
	Templates         *TemplatesRequest `json:"templates"`
	Sort              int               `json:"sort"`
	DiseaseIDs        []uint64          `json:"diseaseIds"`
	AuditStatus       *int8             `json:"auditStatus"`
}

type UpdateExaminationRequest struct {
	ExamName          *string           `json:"examName"`
	ExamType          *string           `json:"examType"`
	ExamPurpose       *string           `json:"examPurpose"`
	ReferenceValue    *string           `json:"referenceValue"`
	AbnormalInterpret *string           `json:"abnormalInterpret"`
	SampleNotes       *string           `json:"sampleNotes"`
	Institution       *string           `json:"institution"`
	Templates         *TemplatesRequest `json:"templates"`
	Sort              *int              `json:"sort"`
	AuditStatus       *int8             `json:"auditStatus"`
	RejectReason      *string           `json:"rejectReason"`
	DiseaseIDs        []uint64          `json:"diseaseIds"`
}

func toCreateHospitalInput(req CreateHospitalRequest) domain.CreateHospitalInput {
	return domain.CreateHospitalInput{
		Name:          req.Name,
		ProvinceCode:  req.ProvinceCode,
		CityCode:      req.CityCode,
		DistrictCode:  req.DistrictCode,
		ProvinceName:  req.ProvinceName,
		CityName:      req.CityName,
		DistrictName:  req.DistrictName,
		Level:         req.Level,
		IsRareNetwork: req.IsRareNetwork,
		TreatScope:    req.TreatScope,
		Address:       req.Address,
		Phone:         req.Phone,
		HospitalURL:   req.HospitalURL,
		DiseaseIDs:    req.DiseaseIDs,
	}
}

func toUpdateHospitalInput(req UpdateHospitalRequest) domain.UpdateHospitalInput {
	return domain.UpdateHospitalInput{
		Name:          req.Name,
		ProvinceCode:  req.ProvinceCode,
		CityCode:      req.CityCode,
		DistrictCode:  req.DistrictCode,
		ProvinceName:  req.ProvinceName,
		CityName:      req.CityName,
		DistrictName:  req.DistrictName,
		Level:         req.Level,
		IsRareNetwork: req.IsRareNetwork,
		TreatScope:    req.TreatScope,
		Address:       req.Address,
		Phone:         req.Phone,
		HospitalURL:   req.HospitalURL,
		AuditStatus:   req.AuditStatus,
		RejectReason:  req.RejectReason,
		DiseaseIDs:    req.DiseaseIDs,
		HasDiseaseIDs: req.DiseaseIDs != nil,
	}
}

func toCreateDoctorInput(req CreateDoctorRequest) domain.CreateDoctorInput {
	return domain.CreateDoctorInput{
		Name:        req.Name,
		Title:       req.Title,
		Department:  req.Department,
		GoodAt:      req.GoodAt,
		ClinicTime:  req.ClinicTime,
		Contact:     req.Contact,
		HospitalID:  req.HospitalID,
		DiseaseIDs:  req.DiseaseIDs,
		AuditStatus: req.AuditStatus,
		Score:       req.Score,
		CommentNum:  req.CommentNum,
	}
}

func toUpdateDoctorInput(req UpdateDoctorRequest) domain.UpdateDoctorInput {
	return domain.UpdateDoctorInput{
		Name:          req.Name,
		Title:         req.Title,
		Department:    req.Department,
		GoodAt:        req.GoodAt,
		ClinicTime:    req.ClinicTime,
		Contact:       req.Contact,
		HospitalID:    req.HospitalID,
		AuditStatus:   req.AuditStatus,
		RejectReason:  req.RejectReason,
		DiseaseIDs:    req.DiseaseIDs,
		HasDiseaseIDs: req.DiseaseIDs != nil,
	}
}

func toExaminationTemplates(req *TemplatesRequest) *domain.ExaminationTemplates {
	if req == nil {
		return nil
	}
	return &domain.ExaminationTemplates{
		Excel:   req.Excel,
		Word:    req.Word,
		Compare: req.Compare,
	}
}

func toCreateExaminationInput(req CreateExaminationRequest) domain.CreateExaminationInput {
	return domain.CreateExaminationInput{
		ExamName:          req.ExamName,
		ExamType:          req.ExamType,
		ExamPurpose:       req.ExamPurpose,
		ReferenceValue:    req.ReferenceValue,
		AbnormalInterpret: req.AbnormalInterpret,
		SampleNotes:       req.SampleNotes,
		Institution:       req.Institution,
		Templates:         toExaminationTemplates(req.Templates),
		Sort:              req.Sort,
		DiseaseIDs:        req.DiseaseIDs,
		AuditStatus:       req.AuditStatus,
	}
}

func toUpdateExaminationInput(req UpdateExaminationRequest) domain.UpdateExaminationInput {
	return domain.UpdateExaminationInput{
		ExamName:          req.ExamName,
		ExamType:          req.ExamType,
		ExamPurpose:       req.ExamPurpose,
		ReferenceValue:    req.ReferenceValue,
		AbnormalInterpret: req.AbnormalInterpret,
		SampleNotes:       req.SampleNotes,
		Institution:       req.Institution,
		Templates:         toExaminationTemplates(req.Templates),
		Sort:              req.Sort,
		AuditStatus:       req.AuditStatus,
		RejectReason:      req.RejectReason,
		DiseaseIDs:        req.DiseaseIDs,
		HasDiseaseIDs:     req.DiseaseIDs != nil,
	}
}
