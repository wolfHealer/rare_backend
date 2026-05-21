package rehab

import "rare_backend/internal/module/resource/rehab/domain"

type CreateInstitutionRequest struct {
	Name          string `json:"name" binding:"required"`
	ProvinceCode  string `json:"provinceCode" binding:"required"`
	CityCode      string `json:"cityCode" binding:"required"`
	DistrictCode  string `json:"districtCode"`
	ProvinceName  string `json:"provinceName"`
	CityName      string `json:"cityName"`
	DistrictName  string `json:"districtName"`
	Qualification string `json:"qualification"`
	RehabProjects string `json:"rehabProjects" binding:"required"`
	FeeStandard   string `json:"feeStandard" binding:"required"`
	ContactPhone  string `json:"contactPhone"`
	ContactUrl    string `json:"contactUrl"`
	Address       string `json:"address" binding:"required"`
	DiseaseIds    []int  `json:"diseaseIds"`
	AuditStatus   int    `json:"auditStatus"`
	RejectReason  string `json:"rejectReason"`
}

type UpdateInstitutionRequest struct {
	Name          string  `json:"name"`
	ProvinceCode  *string `json:"provinceCode"`
	CityCode      *string `json:"cityCode"`
	DistrictCode  *string `json:"districtCode"`
	Qualification string  `json:"qualification"`
	RehabProjects string  `json:"rehabProjects"`
	FeeStandard   string  `json:"feeStandard"`
	ContactPhone  string  `json:"contactPhone"`
	ContactUrl    string  `json:"contactUrl"`
	Address       string  `json:"address"`
	DiseaseIds    []int   `json:"diseaseIds"`
	Sort          int     `json:"sort"`
	AuditStatus   *int    `json:"auditStatus"`
	RejectReason  *string `json:"rejectReason"`
	ProvinceName  *string `json:"provinceName"`
	CityName      *string `json:"cityName"`
	DistrictName  *string `json:"districtName"`
}

type CreatePsychologicalOrgRequest struct {
	Name         string `json:"name" binding:"required"`
	ProvinceCode string `json:"provinceCode" binding:"required"`
	CityCode     string `json:"cityCode" binding:"required"`
	DistrictCode string `json:"districtCode"`
	Address      string `json:"address"`
	ContactPhone string `json:"contactPhone"`
	ContactUrl   string `json:"contactUrl"`
	IsFree       bool   `json:"isFree"`
	ConsultWay   string `json:"consultWay"`
	ContentIntro string `json:"contentIntro" binding:"required"`
	DiseaseIds   []int  `json:"diseaseIds"`
	Sort         int    `json:"sort"`
}

type UpdatePsychologicalOrgRequest struct {
	Name         string  `json:"name"`
	ProvinceCode *string `json:"provinceCode"`
	CityCode     *string `json:"cityCode"`
	DistrictCode *string `json:"districtCode"`
	Address      string  `json:"address"`
	ContactPhone string  `json:"contactPhone"`
	ContactUrl   string  `json:"contactUrl"`
	IsFree       *bool   `json:"isFree"`
	ConsultWay   string  `json:"consultWay"`
	ContentIntro string  `json:"contentIntro"`
	DiseaseIds   []int   `json:"diseaseIds"`
	AuditStatus  int     `json:"auditStatus"`
}

type CreateTrainingRequest struct {
	Title           string   `json:"title" binding:"required"`
	TrainContent    string   `json:"trainContent" binding:"required"`
	RehabStage      string   `json:"rehabStage" binding:"required"`
	TrainPurpose    string   `json:"trainPurpose"`
	ForbiddenAction string   `json:"forbiddenAction"`
	PicUrls         []string `json:"picUrls"`
	GuidePDF        string   `json:"guidePdf"`
	GuideWord       string   `json:"guideWord"`
	Sort            int      `json:"sort"`
	AuditStatus     *int     `json:"auditStatus"`
	RejectReason    *string  `json:"rejectReason"`
	DiseaseIds      []int    `json:"diseaseIds"`
}

type UpdateTrainingRequest struct {
	Title           string   `json:"title"`
	TrainContent    string   `json:"trainContent"`
	RehabStage      string   `json:"rehabStage"`
	TrainPurpose    string   `json:"trainPurpose"`
	ForbiddenAction string   `json:"forbiddenAction"`
	PicUrls         []string `json:"picUrls"`
	GuidePDF        string   `json:"guidePdf"`
	GuideWord       string   `json:"guideWord"`
	Sort            int      `json:"sort"`
	AuditStatus     *int     `json:"auditStatus"`
	RejectReason    *string  `json:"rejectReason"`
	DiseaseIds      []int    `json:"diseaseIds"`
}

func toCreateInstitutionInput(req CreateInstitutionRequest) domain.CreateInstitutionInput {
	return domain.CreateInstitutionInput{
		Name:          req.Name,
		ProvinceCode:  req.ProvinceCode,
		CityCode:      req.CityCode,
		DistrictCode:  req.DistrictCode,
		ProvinceName:  req.ProvinceName,
		CityName:      req.CityName,
		DistrictName:  req.DistrictName,
		Qualification: req.Qualification,
		RehabProjects: req.RehabProjects,
		FeeStandard:   req.FeeStandard,
		ContactPhone:  req.ContactPhone,
		ContactUrl:    req.ContactUrl,
		Address:       req.Address,
		DiseaseIDs:    req.DiseaseIds,
		AuditStatus:   req.AuditStatus,
		RejectReason:  req.RejectReason,
	}
}

func toUpdateInstitutionInput(req UpdateInstitutionRequest) domain.UpdateInstitutionInput {
	return domain.UpdateInstitutionInput{
		Name:          req.Name,
		ProvinceCode:  req.ProvinceCode,
		CityCode:      req.CityCode,
		DistrictCode:  req.DistrictCode,
		ProvinceName:  req.ProvinceName,
		CityName:      req.CityName,
		DistrictName:  req.DistrictName,
		Qualification: req.Qualification,
		RehabProjects: req.RehabProjects,
		FeeStandard:   req.FeeStandard,
		ContactPhone:  req.ContactPhone,
		ContactUrl:    req.ContactUrl,
		Address:       req.Address,
		DiseaseIDs:    req.DiseaseIds,
		Sort:          req.Sort,
		AuditStatus:   req.AuditStatus,
		RejectReason:  req.RejectReason,
		HasDiseaseIDs: req.DiseaseIds != nil,
	}
}

func toCreatePsychOrgInput(req CreatePsychologicalOrgRequest) domain.CreatePsychOrgInput {
	return domain.CreatePsychOrgInput{
		Name:         req.Name,
		ProvinceCode: req.ProvinceCode,
		CityCode:     req.CityCode,
		DistrictCode: req.DistrictCode,
		Address:      req.Address,
		ContactPhone: req.ContactPhone,
		ContactUrl:   req.ContactUrl,
		IsFree:       req.IsFree,
		ConsultWay:   req.ConsultWay,
		ContentIntro: req.ContentIntro,
		DiseaseIDs:   req.DiseaseIds,
	}
}

func toUpdatePsychOrgInput(req UpdatePsychologicalOrgRequest) domain.UpdatePsychOrgInput {
	return domain.UpdatePsychOrgInput{
		Name:          req.Name,
		ProvinceCode:  req.ProvinceCode,
		CityCode:      req.CityCode,
		DistrictCode:  req.DistrictCode,
		Address:       req.Address,
		ContactPhone:  req.ContactPhone,
		ContactUrl:    req.ContactUrl,
		IsFree:        req.IsFree,
		ConsultWay:    req.ConsultWay,
		ContentIntro:  req.ContentIntro,
		DiseaseIDs:    req.DiseaseIds,
		AuditStatus:   req.AuditStatus,
		HasDiseaseIDs: req.DiseaseIds != nil,
	}
}

func toCreateTrainingInput(req CreateTrainingRequest) domain.CreateTrainingInput {
	return domain.CreateTrainingInput{
		Title:           req.Title,
		TrainContent:    req.TrainContent,
		RehabStage:      req.RehabStage,
		TrainPurpose:    req.TrainPurpose,
		ForbiddenAction: req.ForbiddenAction,
		PicUrls:         req.PicUrls,
		GuidePDF:        req.GuidePDF,
		GuideWord:       req.GuideWord,
		Sort:            req.Sort,
		AuditStatus:     req.AuditStatus,
		RejectReason:    req.RejectReason,
		DiseaseIDs:      req.DiseaseIds,
	}
}

func toUpdateTrainingInput(req UpdateTrainingRequest) domain.UpdateTrainingInput {
	return domain.UpdateTrainingInput{
		Title:           req.Title,
		TrainContent:    req.TrainContent,
		RehabStage:      req.RehabStage,
		TrainPurpose:    req.TrainPurpose,
		ForbiddenAction: req.ForbiddenAction,
		PicUrls:         req.PicUrls,
		HasPicUrls:      req.PicUrls != nil,
		GuidePDF:        req.GuidePDF,
		GuideWord:       req.GuideWord,
		Sort:            req.Sort,
		AuditStatus:     req.AuditStatus,
		RejectReason:    req.RejectReason,
		DiseaseIDs:      req.DiseaseIds,
		HasDiseaseIDs:   req.DiseaseIds != nil,
	}
}
