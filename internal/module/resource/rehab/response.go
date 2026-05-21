package rehab

import (
	"errors"

	"rare_backend/internal/module/resource/rehab/domain"
	"rare_backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type ResourceResponse struct {
	DownloadUrl string `json:"downloadUrl"`
	PreviewUrl  string `json:"previewUrl"`
	FileName    string `json:"fileName"`
	FileSize    string `json:"fileSize"`
}

type OptionItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

type DoctorItem struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Specialty string `json:"specialty"`
}

type InstitutionItem struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	ProvinceCode  string   `json:"provinceCode"`
	CityCode      string   `json:"cityCode"`
	DistrictCode  string   `json:"districtCode"`
	ProvinceName  string   `json:"provinceName"`
	CityName      string   `json:"cityName"`
	DistrictName  string   `json:"districtName"`
	Address       string   `json:"address"`
	ContactPhone  string   `json:"contactPhone"`
	ContactUrl    string   `json:"contactUrl"`
	Qualification string   `json:"qualification"`
	RehabProjects string   `json:"rehabProjects"`
	FeeStandard   string   `json:"feeStandard"`
	DiseaseIds    []uint64 `json:"diseaseIds"`
	AuditStatus   int8     `json:"auditStatus"`
	Rating        float64  `json:"rating"`
	Status        string   `json:"status"`
	UpdateAt      string   `json:"updatedAt"`
}

type InstitutionListResponse struct {
	List     []InstitutionItem `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

type InstitutionDiseaseItem struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type RegionItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

type RegionResponse struct {
	Regions []RegionItem `json:"regions"`
}

type CounselorItem struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Specialty string `json:"specialty"`
}

type PsychologicalOrgItem struct {
	ID           uint                     `json:"id"`
	Name         string                   `json:"name"`
	ProvinceCode string                   `json:"provinceCode"`
	CityCode     string                   `json:"cityCode"`
	DistrictCode string                   `json:"districtCode"`
	ProvinceName string                   `json:"provinceName"`
	CityName     string                   `json:"cityName"`
	DistrictName string                   `json:"districtName"`
	Address      string                   `json:"address"`
	ContactPhone string                   `json:"contactPhone"`
	ContactUrl   string                   `json:"contactUrl"`
	IsFree       bool                     `json:"isFree"`
	ConsultWay   string                   `json:"consultWay"`
	ContentIntro string                   `json:"contentIntro"`
	AuditStatus  int8                     `json:"auditStatus"`
	RejectReason *string                  `json:"rejectReason"`
	Type         string                   `json:"type"`
	TypeName     string                   `json:"typeName"`
	Region       string                   `json:"region"`
	RegionCode   string                   `json:"regionCode"`
	ServiceTime  string                   `json:"serviceTime"`
	Description  string                   `json:"description"`
	Services     []string                 `json:"services"`
	Rating       float64                  `json:"rating"`
	CoverUrl     string                   `json:"coverUrl"`
	Status       string                   `json:"status"`
	DiseaseIds   []uint64                 `json:"diseaseIds"`
	Diseases     []InstitutionDiseaseItem `json:"diseases"`
	CreatedAt    string                   `json:"createdAt"`
	UpdatedAt    string                   `json:"updatedAt"`
}

type PsychologicalOrgListResponse struct {
	List     []PsychologicalOrgItem `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

type PsychologicalOrgDetailResponse struct {
	ID           uint                     `json:"id"`
	Name         string                   `json:"name"`
	ProvinceCode string                   `json:"provinceCode"`
	CityCode     string                   `json:"cityCode"`
	DistrictCode string                   `json:"districtCode"`
	ProvinceName string                   `json:"provinceName"`
	CityName     string                   `json:"cityName"`
	DistrictName string                   `json:"districtName"`
	Address      string                   `json:"address"`
	ContactPhone string                   `json:"contactPhone"`
	ContactUrl   string                   `json:"contactUrl"`
	IsFree       bool                     `json:"isFree"`
	ConsultWay   string                   `json:"consultWay"`
	ContentIntro string                   `json:"contentIntro"`
	AuditStatus  int8                     `json:"auditStatus"`
	RejectReason *string                  `json:"rejectReason"`
	Type         string                   `json:"type"`
	TypeName     string                   `json:"typeName"`
	Region       string                   `json:"region"`
	RegionCode   string                   `json:"regionCode"`
	ServiceTime  string                   `json:"serviceTime"`
	Description  string                   `json:"description"`
	Services     []string                 `json:"services"`
	Rating       float64                  `json:"rating"`
	CoverUrl     string                   `json:"coverUrl"`
	Status       string                   `json:"status"`
	DiseaseIds   []uint64                 `json:"diseaseIds"`
	Diseases     []InstitutionDiseaseItem `json:"diseases"`
	DiseaseCount int                      `json:"diseaseCount"`
	Images       []string                 `json:"images"`
	Counselors   []CounselorItem          `json:"counselors"`
	CreatedAt    string                   `json:"createdAt"`
	UpdatedAt    string                   `json:"updatedAt"`
}

type TargetItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

type TargetResponse struct {
	Targets []TargetItem `json:"targets"`
}

type OrgTypeItem struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

type OrgTypeResponse struct {
	Types []OrgTypeItem `json:"types"`
}

type InstitutionOptionsResponse struct {
	Regions  []RegionItem  `json:"regions"`
	Types    []OrgTypeItem `json:"types"`
	Diseases []OptionItem  `json:"diseases"`
}

type TrainingDiseaseItem struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type TrainingListItem struct {
	ID              uint                  `json:"id"`
	RehabStage      string                `json:"rehabStage"`
	Title           string                `json:"title"`
	TrainPurpose    string                `json:"trainPurpose"`
	TrainContent    string                `json:"trainContent"`
	ForbiddenAction string                `json:"forbiddenAction"`
	PicUrls         []string              `json:"picUrls"`
	GuidePdf        string                `json:"guidePdf"`
	GuideWord       string                `json:"guideWord"`
	AuditStatus     int8                  `json:"auditStatus"`
	RejectReason    string                `json:"rejectReason"`
	Sort            int                   `json:"sort"`
	DiseaseIds      []uint64              `json:"diseaseIds"`
	Diseases        []TrainingDiseaseItem `json:"diseases"`
	CreatedAt       string                `json:"createdAt"`
	UpdatedAt       string                `json:"updatedAt"`
}

type TrainingDetailDataResponse struct {
	ID              uint                  `json:"id"`
	RehabStage      string                `json:"rehabStage"`
	Title           string                `json:"title"`
	TrainPurpose    string                `json:"trainPurpose"`
	TrainContent    string                `json:"trainContent"`
	ForbiddenAction string                `json:"forbiddenAction"`
	PicUrls         []string              `json:"picUrls"`
	GuidePdf        string                `json:"guidePdf"`
	GuideWord       string                `json:"guideWord"`
	AuditStatus     int8                  `json:"auditStatus"`
	RejectReason    *string               `json:"rejectReason"`
	Sort            int                   `json:"sort"`
	DiseaseIds      []uint64              `json:"diseaseIds"`
	Diseases        []TrainingDiseaseItem `json:"diseases"`
	CreatedAt       string                `json:"createdAt"`
	UpdatedAt       string                `json:"updatedAt"`
}

func respondOK(c *gin.Context, data any) {
	response.OK(c, data)
}

func respondOKMessage(c *gin.Context, message string, data any) {
	response.OKMessage(c, message, data)
}

func respondPage(c *gin.Context, list any, total int64, page, pageSize int) {
	response.Page(c, list, total, page, pageSize)
}

func respondBindError(c *gin.Context, err error) {
	response.BadRequest(c, "参数错误")
}

func respondInvalidInstitutionID(c *gin.Context) {
	response.BadRequest(c, "无效的机构 ID")
}

func respondInvalidTrainingID(c *gin.Context) {
	response.BadRequest(c, "无效的训练 ID")
}

func respondInstitutionNotFound(c *gin.Context) {
	response.NotFound(c, "机构不存在")
}

func respondPsychOrgNotFound(c *gin.Context) {
	response.NotFound(c, "机构不存在")
}

func respondTrainingNotFound(c *gin.Context) {
	response.NotFound(c, "训练指南不存在")
}

func respondResourceNotFound(c *gin.Context) {
	response.NotFound(c, "资源文件不存在")
}

func respondInstitutionListCountError(c *gin.Context, err error) {
	response.InternalError(c, "查询总数失败")
}

func respondInstitutionListError(c *gin.Context, err error) {
	response.InternalError(c, "查询列表失败")
}

func respondInstitutionDetailError(c *gin.Context, err error) {
	response.InternalError(c, "查询机构详情失败")
}

func respondProvinceOptionsError(c *gin.Context) {
	response.InternalError(c, "查询省份选项失败")
}

func respondDiseaseOptionsError(c *gin.Context) {
	response.InternalError(c, "查询疾病选项失败")
}

func respondInstitutionTxBeginError(c *gin.Context) {
	response.InternalError(c, "数据库事务启动失败")
}

func respondInstitutionCreateError(c *gin.Context, err error) {
	response.InternalError(c, "创建康复机构失败")
}

func respondInstitutionLastInsertError(c *gin.Context) {
	response.InternalError(c, "获取新增ID失败")
}

func respondInstitutionRelError(c *gin.Context) {
	response.InternalError(c, "关联疾病失败")
}

func respondInstitutionCommitError(c *gin.Context) {
	response.InternalError(c, "事务提交失败")
}

func respondInstitutionBindParseError(c *gin.Context, err error) {
	response.BadRequest(c, "参数解析错误")
}

func respondInstitutionUpdateError(c *gin.Context, err error) {
	response.InternalError(c, "更新机构信息失败")
}

func respondInstitutionRelCleanupError(c *gin.Context) {
	response.InternalError(c, "清理旧疾病关联失败")
}

func respondInstitutionRelCreateError(c *gin.Context) {
	response.InternalError(c, "创建新疾病关联失败")
}

func respondInstitutionDeleteRelError(c *gin.Context) {
	response.InternalError(c, "删除关联数据失败")
}

func respondInstitutionDeleteError(c *gin.Context) {
	response.InternalError(c, "删除机构失败")
}

func respondPsychOrgListCountError(c *gin.Context, err error) {
	response.InternalError(c, "查询总数失败")
}

func respondPsychOrgListError(c *gin.Context, err error) {
	response.InternalError(c, "查询列表失败")
}

func respondPsychOrgDetailError(c *gin.Context, err error) {
	response.InternalError(c, "查询机构详情失败")
}

func respondPsychOrgBindError(c *gin.Context) {
	response.BadRequest(c, "参数错误")
}

func respondPsychOrgCreateError(c *gin.Context) {
	response.InternalError(c, "创建心理咨询机构失败")
}

func respondPsychOrgUpdateError(c *gin.Context) {
	response.InternalError(c, "更新机构信息失败")
}

func respondPsychOrgRelCleanupError(c *gin.Context) {
	response.InternalError(c, "清理旧疾病关联失败")
}

func respondPsychOrgRelCreateError(c *gin.Context) {
	response.InternalError(c, "创建新疾病关联失败")
}

func respondTrainingListCountError(c *gin.Context, err error) {
	response.InternalError(c, "查询总数失败")
}

func respondTrainingListError(c *gin.Context, err error) {
	response.InternalError(c, "查询列表失败")
}

func respondTrainingDetailError(c *gin.Context, err error) {
	response.InternalError(c, "查询训练详情失败")
}

func respondTrainingPicUrlsError(c *gin.Context) {
	response.InternalError(c, "图片URL序列化失败")
}

func respondTrainingCreateError(c *gin.Context, err error) {
	response.InternalError(c, "创建训练指南失败")
}

func respondTrainingUpdateError(c *gin.Context, err error) {
	response.InternalError(c, "更新训练指南失败")
}

func respondTrainingDeleteRelError(c *gin.Context, err error) {
	response.InternalError(c, "删除关联疾病数据失败")
}

func respondTrainingDeleteError(c *gin.Context, err error) {
	response.InternalError(c, "删除训练指南失败")
}

func respondTrainingResourceQueryError(c *gin.Context) {
	response.InternalError(c, "查询资源文件失败")
}

func toInstitutionDiseaseItems(items []domain.DiseaseItem) []InstitutionDiseaseItem {
	result := make([]InstitutionDiseaseItem, len(items))
	for i, d := range items {
		result[i] = InstitutionDiseaseItem{ID: d.ID, Name: d.Name, Alias: d.Alias}
	}
	return result
}

func toTrainingDiseaseItems(items []domain.DiseaseItem) []TrainingDiseaseItem {
	result := make([]TrainingDiseaseItem, len(items))
	for i, d := range items {
		result[i] = TrainingDiseaseItem{ID: d.ID, Name: d.Name, Alias: d.Alias}
	}
	return result
}

func handleInstitutionDetailError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrInstitutionNotFound) {
		respondInstitutionNotFound(c)
		return
	}
	respondInstitutionDetailError(c, err)
}

func handlePsychOrgDetailError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrPsychOrgNotFound) {
		respondPsychOrgNotFound(c)
		return
	}
	respondPsychOrgDetailError(c, err)
}

func handleTrainingDetailError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrTrainingNotFound) {
		respondTrainingNotFound(c)
		return
	}
	respondTrainingDetailError(c, err)
}

func handleTrainingResourceError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrTrainingNotFound) {
		respondTrainingNotFound(c)
		return
	}
	if errors.Is(err, domain.ErrResourceNotFound) {
		respondResourceNotFound(c)
		return
	}
	respondTrainingResourceQueryError(c)
}

func handleTrainingDeleteError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrTrainingNotFound) {
		respondTrainingNotFound(c)
		return
	}
	respondTrainingDeleteError(c, err)
}

func handleInstitutionOptionsError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrProvinceOptions) {
		respondProvinceOptionsError(c)
		return
	}
	if errors.Is(err, domain.ErrDiseaseOptions) {
		respondDiseaseOptionsError(c)
		return
	}
	respondProvinceOptionsError(c)
}
