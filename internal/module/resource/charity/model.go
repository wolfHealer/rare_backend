package charity

type CreateProjectRequest struct {
	Name            string  `json:"name" binding:"required"`
	ApplyProcess    string  `json:"applyProcess" binding:"required"`
	ApplyCondition  string  `json:"applyCondition"`
	ReliefType      string  `json:"reliefType" binding:"required"`
	DiseaseIDs      []int   `json:"diseaseIds"`
	ReliefStandard  string  `json:"reliefStandard"`
	ApplyDifficulty string  `json:"applyDifficulty"`
	ApplyDeadline   *string `json:"applyDeadline"`
	ContactPhone    string  `json:"contactPhone"`
	ContactURL      string  `json:"contactUrl"`
	ApplyForm       string  `json:"applyForm"`
	ApplyGuide      string  `json:"applyGuide"`
	MaterialList    string  `json:"materialList"`
	Organizer       string  `json:"organizer"`
	Sort            int     `json:"sort"`
}

type UpdateProjectRequest struct {
	Name            string  `json:"name"`
	ApplyProcess    string  `json:"applyProcess"`
	ApplyCondition  string  `json:"applyCondition"`
	ReliefType      string  `json:"reliefType"`
	Type            string  `json:"type"`
	DiseaseIDs      []int   `json:"diseaseIds"`
	ReliefStandard  string  `json:"reliefStandard"`
	ApplyDifficulty string  `json:"applyDifficulty"`
	ApplyDeadline   *string `json:"applyDeadline"`
	ContactPhone    string  `json:"contactPhone"`
	ContactURL      string  `json:"contactUrl"`
	ApplyForm       string  `json:"applyForm"`
	ApplyGuide      string  `json:"applyGuide"`
	MaterialList    string  `json:"materialList"`
	Organizer       string  `json:"organizer"`
	Sort            *int    `json:"sort"`
	AuditStatus     *int    `json:"auditStatus"`
	RejectReason    *string `json:"rejectReason"`
}

type CreateChannelRequest struct {
	ChannelType          string `json:"channelType" binding:"required"`
	Name                 string `json:"name" binding:"required"`
	ApplyCondition       string `json:"applyCondition" binding:"required"`
	ResponseTime         string `json:"responseTime" binding:"required"`
	ContactPhone         string `json:"contactPhone"`
	ContactUrl           string `json:"contactUrl"`
	HelpLetterTemplate   string `json:"helpLetterTemplate"`
	CrowdfundingTemplate string `json:"crowdfundingTemplate"`
	Sort                 int    `json:"sort"`
	AuditStatus          *int   `json:"auditStatus"`
}

type UpdateChannelRequest struct {
	ChannelType          string  `json:"channelType"`
	Name                 string  `json:"name"`
	ApplyCondition       string  `json:"applyCondition"`
	ResponseTime         string  `json:"responseTime"`
	ContactPhone         string  `json:"contactPhone"`
	ContactUrl           string  `json:"contactUrl"`
	HelpLetterTemplate   string  `json:"helpLetterTemplate"`
	CrowdfundingTemplate string  `json:"crowdfundingTemplate"`
	Sort                 *int    `json:"sort"`
	AuditStatus          *int    `json:"auditStatus"`
	RejectReason         *string `json:"rejectReason"`
}

type CreateCaseRequest struct {
	DiseaseID        int64  `json:"diseaseId" binding:"required"`
	ProjectID        uint   `json:"projectId" binding:"required"`
	CaseTitle        string `json:"caseTitle" binding:"required"`
	PatientDesc      string `json:"patientDesc" binding:"required"`
	ApplyCycle       string `json:"applyCycle" binding:"required"`
	ActualRelief     string `json:"actualRelief" binding:"required"`
	Experience       string `json:"experience" binding:"required"`
	PitfallGuide     string `json:"pitfallGuide" binding:"required"`
	CasePdf          string `json:"casePdf"`
	MaterialTemplate string `json:"materialTemplate"`
	AuditStatus      *int   `json:"auditStatus"`
	RejectReason     string `json:"rejectReason"`
}

type UpdateCaseRequest struct {
	Title            string `json:"title"`
	PatientDesc      string `json:"patientDesc"`
	DiseaseID        *int64 `json:"diseaseId"`
	ActualRelief     string `json:"actualRelief"`
	Experience       string `json:"experience"`
	PitfallGuide     string `json:"pitfallGuide"`
	ApplyCycle       string `json:"applyCycle"`
	CasePdf          string `json:"casePdf"`
	MaterialTemplate string `json:"materialTemplate"`
	ProjectID        *uint  `json:"projectId"`
	AuditStatus      *int   `json:"auditStatus"`
}
