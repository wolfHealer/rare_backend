package service

import (
	"strconv"
	"strings"
	"time"
)

// DoctorItem 医生信息项
type DoctorItem struct {
	Name      string
	Title     string
	Specialty string
}

// CounselorItem 心理咨询师信息项
type CounselorItem struct {
	Name      string
	Title     string
	Specialty string
}

func convertRegionToCode(region string) string {
	regionMap := map[string]string{
		"北京": "bj",
		"上海": "sh",
		"广东": "gd",
		"广州": "gz",
		"深圳": "sz",
		"浙江": "zj",
		"江苏": "js",
		"四川": "sc",
		"湖北": "hb",
		"山东": "sd",
		"河南": "hn",
		"福建": "fj",
		"湖南": "hun",
		"安徽": "ah",
		"辽宁": "ln",
		"陕西": "sx",
		"重庆": "cq",
		"天津": "tj",
	}
	if code, ok := regionMap[region]; ok {
		return code
	}
	return "other"
}

func convertInstitutionType(name string) (string, string) {
	if strings.Contains(name, "医院") {
		return "hospital", "康复医院"
	}
	if strings.Contains(name, "中心") {
		return "center", "康复中心"
	}
	if strings.Contains(name, "诊所") {
		return "clinic", "康复诊所"
	}
	if strings.Contains(name, "社区") {
		return "community", "社区康复站"
	}
	return "other", "其他机构"
}

func parseServices(projects string) []string {
	if projects == "" {
		return []string{"康复指导", "康复训练", "护理培训"}
	}
	services := strings.Split(projects, "，")
	if len(services) == 0 {
		services = strings.Split(projects, ",")
	}
	if len(services) == 0 {
		return []string{"康复指导", "康复训练", "护理培训"}
	}
	return services
}

func getInstitutionDoctors(instID uint) []DoctorItem {
	doctorsMap := map[uint][]DoctorItem{
		1: {
			{Name: "张医生", Title: "主任医师", Specialty: "神经康复"},
			{Name: "李医生", Title: "副主任医师", Specialty: "肢体康复"},
		},
		2: {
			{Name: "王医生", Title: "主任医师", Specialty: "儿童康复"},
		},
	}
	if doctors, ok := doctorsMap[instID]; ok {
		return doctors
	}
	return []DoctorItem{
		{Name: "赵医生", Title: "主治医师", Specialty: "康复指导"},
	}
}

func convertPsychologicalOrgType(name, consultWay string) (string, string) {
	if strings.Contains(name, "热线") {
		return "hotline", "心理热线"
	}
	if strings.Contains(name, "中心") {
		return "center", "心理中心"
	}
	if strings.Contains(name, "医院") {
		return "hospital", "心理医院"
	}
	if consultWay == "线上" {
		return "online", "在线咨询"
	}
	return "center", "心理中心"
}

func parsePsychologicalServices(name, consultWay string) []string {
	if strings.Contains(name, "热线") {
		return []string{"心理疏导", "危机干预", "情绪支持"}
	}
	if strings.Contains(name, "中心") {
		return []string{"心理咨询", "心理治疗", "团体辅导", "心理测评"}
	}
	if strings.Contains(name, "医院") {
		return []string{"心理诊断", "心理治疗", "药物治疗", "康复指导"}
	}
	return []string{"心理咨询", "心理支持"}
}

func getServiceTime(orgType string) string {
	if orgType == "hotline" {
		return "24 小时"
	}
	return "周一至周日 08:00-17:00"
}

func getPsychologicalCounselors(orgID uint) []CounselorItem {
	counselorsMap := map[uint][]CounselorItem{
		1: {
			{Name: "李老师", Title: "资深心理咨询师", Specialty: "危机干预、创伤治疗"},
			{Name: "张老师", Title: "心理咨询师", Specialty: "情绪管理、压力疏导"},
		},
		2: {
			{Name: "王医生", Title: "心理治疗师", Specialty: "心理治疗、认知行为疗法"},
			{Name: "赵医生", Title: "主任医师", Specialty: "精神障碍诊断与治疗"},
		},
		3: {
			{Name: "陈老师", Title: "心理咨询师", Specialty: "生命教育、心理支持"},
		},
	}
	if counselors, ok := counselorsMap[orgID]; ok {
		return counselors
	}
	return []CounselorItem{
		{Name: "刘老师", Title: "心理咨询师", Specialty: "心理咨询、情绪疏导"},
	}
}

func convertStageToValue(stage string) string {
	valueMap := map[string]string{
		"早期": "early",
		"中期": "mid",
		"晚期": "late",
	}
	if val, ok := valueMap[stage]; ok {
		return val
	}
	return stage
}

func convertStageToType(stage string) string {
	typeMap := map[string]string{
		"early":       "基础训练",
		"middle":      "强化训练",
		"late":        "维持训练",
		"stable":      "维持训练",
		"progressive": "强化训练",
	}
	if val, ok := typeMap[stage]; ok {
		return val
	}
	return "康复训练"
}

func getManualUpdateTime() string {
	return time.Now().Format("2006-01-02T15:04:05Z")
}

func getFileSize(url string) string {
	if strings.Contains(url, ".pdf") {
		return "1.5MB"
	}
	if strings.Contains(url, ".doc") {
		return "800KB"
	}
	return "未知"
}

func parsePicUrls(picUrlsStr string) []string {
	var picUrls []string
	if picUrlsStr == "" {
		return []string{}
	}
	picUrlsStr = strings.TrimPrefix(picUrlsStr, "[")
	picUrlsStr = strings.TrimSuffix(picUrlsStr, "]")
	if picUrlsStr != "" {
		rawUrls := strings.Split(picUrlsStr, ",")
		for _, u := range rawUrls {
			u = strings.TrimSpace(u)
			u = strings.Trim(u, "\"")
			if u != "" {
				picUrls = append(picUrls, u)
			}
		}
	}
	if picUrls == nil {
		picUrls = []string{}
	}
	return picUrls
}

func calcRating(id uint) float64 {
	return 4.5 + float64(id%10)/10
}

func parsePageParams(pageStr, pageSizeStr string, defaultPageSize int) (int, int) {
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = defaultPageSize
	}
	return page, pageSize
}

func resolveAuditStatus(auditStatusStr string) string {
	if auditStatusStr != "" {
		if auditStatus, err := strconv.Atoi(auditStatusStr); err == nil {
			return strconv.Itoa(auditStatus)
		}
	}
	return ""
}

func resolveDiseaseID(diseaseStr string) string {
	if diseaseStr != "" {
		diseaseID, err := strconv.Atoi(diseaseStr)
		if err == nil && diseaseID > 0 {
			return strconv.Itoa(diseaseID)
		}
	}
	return ""
}

func resolvePsychDiseaseID(diseaseStr string) string {
	if diseaseStr != "" {
		diseaseID, _ := strconv.Atoi(diseaseStr)
		if diseaseID > 0 {
			return strconv.Itoa(diseaseID)
		}
	}
	return ""
}

func nullStringPtr(valid bool, s string) *string {
	if valid {
		return &s
	}
	return nil
}

func institutionRating(id uint) float64 {
	return calcRating(id)
}
