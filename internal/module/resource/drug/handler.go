package drug

import (
	"database/sql"
	"fmt"
	"net/http"
	"os" // 新增
	"rare_backend/internal/pkg/db"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// DrugItem 药品项响应结构
type DrugItem struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	Indication       string `json:"indication"`
	Type             string `json:"type"`
	Insurance        bool   `json:"insurance"`
	Desc             string `json:"desc"`
	ManualURL        string `json:"manualUrl"`
	DosageForm       string `json:"dosageForm"`
	Spec             string `json:"spec"`
	RefPrice         string `json:"refPrice"`
	HasRelief        bool   `json:"hasRelief"`
	IsLaunched       bool   `json:"isLaunched"`
	NeedPrescription bool   `json:"needPrescription"`
}

// DrugListResponse 列表响应结构
type DrugListResponse struct {
	List     []DrugItem `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

// DrugOptionsResponse 筛选选项响应
type DrugOptionsResponse struct {
	Types      []OptionItem `json:"types"`
	Insurances []OptionItem `json:"insurances"`
}

// OptionItem 选项项
type OptionItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// ListDrugs 获取药品列表
func ListDrugs(c *gin.Context) {
	// 获取请求参数
	diseaseStr := c.DefaultQuery("disease", "0")
	keyword := c.DefaultQuery("keyword", "")
	typeFilter := c.DefaultQuery("type", "")
	insuranceStr := c.DefaultQuery("insurance", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	disease, _ := strconv.Atoi(diseaseStr)
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}

	if disease != 0 {
		whereClause += " AND disease_value = ?"
		args = append(args, disease)
	}
	if keyword != "" {
		whereClause += " AND (generic_name LIKE ? OR brand_name LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	if typeFilter != "" {
		whereClause += " AND drug_type = ?"
		args = append(args, typeFilter)
	}
	if insuranceStr != "" {
		isInsurance := 0
		if insuranceStr == "true" || insuranceStr == "1" {
			isInsurance = 1
		}
		whereClause += " AND is_insurance = ?"
		args = append(args, isInsurance)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM rare_drugs " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 查询列表
	listQuery := `
		SELECT id, generic_name, brand_name, indication, drug_type, is_insurance,
		       dosage_form, spec, ref_price, has_relief, is_launched, need_prescription,
		       manual_original, manual_popular
		FROM rare_drugs
		` + whereClause + `
		ORDER BY is_launched DESC, created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询列表失败",
		})
		return
	}
	defer rows.Close()

	var list []DrugItem
	for rows.Next() {
		var drug struct {
			ID               uint           `db:"id"`
			GenericName      string         `db:"generic_name"`
			BrandName        sql.NullString `db:"brand_name"`
			Indication       string         `db:"indication"`
			DrugType         string         `db:"drug_type"`
			IsInsurance      int8           `db:"is_insurance"`
			DosageForm       string         `db:"dosage_form"`
			Spec             string         `db:"spec"`
			RefPrice         sql.NullString `db:"ref_price"`
			HasRelief        int8           `db:"has_relief"`
			IsLaunched       int8           `db:"is_launched"`
			NeedPrescription int8           `db:"need_prescription"`
			ManualOriginal   string         `db:"manual_original"`
			ManualPopular    string         `db:"manual_popular"`
		}
		if err := rows.Scan(
			&drug.ID, &drug.GenericName, &drug.BrandName, &drug.Indication,
			&drug.DrugType, &drug.IsInsurance, &drug.DosageForm, &drug.Spec,
			&drug.RefPrice, &drug.HasRelief, &drug.IsLaunched, &drug.NeedPrescription,
			&drug.ManualOriginal, &drug.ManualPopular,
		); err != nil {
			continue
		}

		// 药品名称优先使用商品名，否则使用通用名
		name := drug.GenericName
		if drug.BrandName.Valid && drug.BrandName.String != "" {
			name = drug.BrandName.String
		}

		// 转换类型枚举
		drugType := convertDrugType(drug.DrugType)

		list = append(list, DrugItem{
			ID:               drug.ID,
			Name:             name,
			Indication:       drug.Indication,
			Type:             drugType,
			Insurance:        drug.IsInsurance == 1,
			Desc:             drug.DosageForm + " | " + drug.Spec,
			ManualURL:        drug.ManualOriginal,
			DosageForm:       drug.DosageForm,
			Spec:             drug.Spec,
			RefPrice:         drug.RefPrice.String,
			HasRelief:        drug.HasRelief == 1,
			IsLaunched:       drug.IsLaunched == 1,
			NeedPrescription: drug.NeedPrescription == 1,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": DrugListResponse{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// GetDrugDetail 获取药品详情
func GetDrugDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的药品 ID",
		})
		return
	}

	query := `
		SELECT id, generic_name, brand_name, indication, drug_type, is_insurance,
		       dosage_form, spec, ref_price, has_relief, is_launched, need_prescription,
		       manual_original, manual_popular, disease_value, created_at
		FROM rare_drugs
		WHERE id = ? AND is_audit = 1
	`

	var drug struct {
		ID               uint           `db:"id"`
		GenericName      string         `db:"generic_name"`
		BrandName        sql.NullString `db:"brand_name"`
		Indication       string         `db:"indication"`
		DrugType         string         `db:"drug_type"`
		IsInsurance      int8           `db:"is_insurance"`
		DosageForm       string         `db:"dosage_form"`
		Spec             string         `db:"spec"`
		RefPrice         sql.NullString `db:"ref_price"`
		HasRelief        int8           `db:"has_relief"`
		IsLaunched       int8           `db:"is_launched"`
		NeedPrescription int8           `db:"need_prescription"`
		ManualOriginal   string         `db:"manual_original"`
		ManualPopular    string         `db:"manual_popular"`
		DiseaseValue     int            `db:"disease_value"`
		CreatedAt        string         `db:"created_at"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&drug.ID, &drug.GenericName, &drug.BrandName, &drug.Indication,
		&drug.DrugType, &drug.IsInsurance, &drug.DosageForm, &drug.Spec,
		&drug.RefPrice, &drug.HasRelief, &drug.IsLaunched, &drug.NeedPrescription,
		&drug.ManualOriginal, &drug.ManualPopular, &drug.DiseaseValue, &drug.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "药品不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询药品详情失败",
		})
		return
	}

	// 药品名称优先使用商品名
	name := drug.GenericName
	if drug.BrandName.Valid && drug.BrandName.String != "" {
		name = drug.BrandName.String
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": DrugItem{
			ID:               drug.ID,
			Name:             name,
			Indication:       drug.Indication,
			Type:             convertDrugType(drug.DrugType),
			Insurance:        drug.IsInsurance == 1,
			Desc:             drug.DosageForm + " | " + drug.Spec,
			ManualURL:        drug.ManualOriginal,
			DosageForm:       drug.DosageForm,
			Spec:             drug.Spec,
			RefPrice:         drug.RefPrice.String,
			HasRelief:        drug.HasRelief == 1,
			IsLaunched:       drug.IsLaunched == 1,
			NeedPrescription: drug.NeedPrescription == 1,
		},
	})
}

// DownloadManual 下载药品说明书
func DownloadManual(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的药品 ID",
		})
		return
	}

	query := `SELECT manual_original, manual_popular FROM rare_drugs WHERE id = ? AND is_audit = 1`

	var manualOriginal, manualPopular string
	err = db.MySQL.QueryRow(query, id).Scan(&manualOriginal, &manualPopular)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "药品不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询说明书失败",
		})
		return
	}

	// 返回说明书下载地址
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"original": manualOriginal,
			"popular":  manualPopular,
			"url":      manualOriginal, // 默认返回官方版
		},
	})
}

// ExportDrugs 导出药品名录 Excel
func ExportDrugs(c *gin.Context) {
	// 获取筛选参数
	diseaseStr := c.DefaultQuery("disease", "0")
	keyword := c.DefaultQuery("keyword", "")
	typeFilter := c.DefaultQuery("type", "")
	insuranceStr := c.DefaultQuery("insurance", "")

	disease, _ := strconv.Atoi(diseaseStr)

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}

	if disease != 0 {
		whereClause += " AND disease_value = ?"
		args = append(args, disease)
	}
	if keyword != "" {
		whereClause += " AND (generic_name LIKE ? OR brand_name LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	if typeFilter != "" {
		whereClause += " AND drug_type = ?"
		args = append(args, typeFilter)
	}
	if insuranceStr != "" {
		isInsurance := 0
		if insuranceStr == "true" || insuranceStr == "1" {
			isInsurance = 1
		}
		whereClause += " AND is_insurance = ?"
		args = append(args, isInsurance)
	}

	// 查询药品数据
	query := `
		SELECT generic_name, brand_name, indication, drug_type, is_insurance,
		       dosage_form, spec, ref_price, has_relief, is_launched
		FROM rare_drugs
		` + whereClause + `
		ORDER BY is_launched DESC, created_at DESC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询药品数据失败",
		})
		return
	}
	defer rows.Close()

	// 创建 Excel 文件
	excel := excelize.NewFile()
	sheetName := "药品名录"
	index, err := excel.NewSheet(sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建 Excel 工作表失败",
		})
		return
	}
	excel.SetActiveSheet(index)
	excel.DeleteSheet("Sheet1")

	// 设置表头
	headers := []string{"通用名", "商品名", "适应症", "类型", "医保", "剂型", "规格", "参考价格", "赠药援助", "国内上市"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		excel.SetCellValue(sheetName, cell, header)
	}

	// 填充数据
	rowNum := 2
	for rows.Next() {
		var drug struct {
			GenericName string         `db:"generic_name"`
			BrandName   sql.NullString `db:"brand_name"`
			Indication  string         `db:"indication"`
			DrugType    string         `db:"drug_type"`
			IsInsurance int8           `db:"is_insurance"`
			DosageForm  string         `db:"dosage_form"`
			Spec        string         `db:"spec"`
			RefPrice    sql.NullString `db:"ref_price"`
			HasRelief   int8           `db:"has_relief"`
			IsLaunched  int8           `db:"is_launched"`
		}
		if err := rows.Scan(
			&drug.GenericName, &drug.BrandName, &drug.Indication,
			&drug.DrugType, &drug.IsInsurance, &drug.DosageForm, &drug.Spec,
			&drug.RefPrice, &drug.HasRelief, &drug.IsLaunched,
		); err != nil {
			continue
		}

		data := []interface{}{
			drug.GenericName,
			drug.BrandName.String,
			truncateString(drug.Indication, 50),
			convertDrugType(drug.DrugType),
			mapYesNo(drug.IsInsurance),
			drug.DosageForm,
			drug.Spec,
			drug.RefPrice.String,
			mapYesNo(drug.HasRelief),
			mapYesNo(drug.IsLaunched),
		}

		for i, value := range data {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowNum)
			excel.SetCellValue(sheetName, cell, value)
		}
		rowNum++
	}

	// 设置列宽
	for i := 1; i <= len(headers); i++ {
		col, _ := excelize.ColumnNumberToName(i)
		excel.SetColWidth(sheetName, col, col, 20)
	}

	// 设置响应头
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("药品名录_%s.xlsx", timestamp)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// 输出文件
	if err := excel.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成 Excel 文件失败",
		})
		return
	}
}

// GetDrugOptions 获取筛选选项
func GetDrugOptions(c *gin.Context) {
	// 药品类型选项
	types := []OptionItem{
		{Label: "进口药", Value: "进口"},
		{Label: "国产药", Value: "国产"},
		{Label: "仿制药", Value: "仿制药"},
	}

	// 医保选项
	insurances := []OptionItem{
		{Label: "医保", Value: "true"},
		{Label: "非医保", Value: "false"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": DrugOptionsResponse{
			Types:      types,
			Insurances: insurances,
		},
	})
}

// convertDrugType 转换药品类型枚举
func convertDrugType(drugType string) string {
	typeMap := map[string]string{
		"进口":  "imported",
		"国产":  "domestic",
		"仿制药": "generic",
	}
	if val, ok := typeMap[drugType]; ok {
		return val
	}
	return drugType
}

// mapYesNo 布尔值转中文
func mapYesNo(val int8) string {
	if val == 1 {
		return "是"
	}
	return "否"
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ChannelItem 渠道项响应结构
type ChannelItem struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Address       string `json:"address"`
	Desc          string `json:"desc"`
	Region        string `json:"region"`
	Contact       string `json:"contact"`
	IsInsurance   bool   `json:"isInsurance"`
	DeliveryScope string `json:"deliveryScope"`
	DeliveryCycle string `json:"deliveryCycle"`
}

// ChannelOptionsResponse 渠道筛选选项响应
type ChannelOptionsResponse struct {
	Regions    []OptionItem `json:"regions"`
	Deliveries []OptionItem `json:"deliveries"`
}

// ChannelContactResponse 渠道联系方式响应
type ChannelContactResponse struct {
	Phone  string `json:"phone"`
	Wechat string `json:"wechat"`
	Email  string `json:"email"`
}

// GetChannelOptions 获取渠道筛选选项
func GetChannelOptions(c *gin.Context) {
	// 获取所有不重复的地区
	regionQuery := "SELECT DISTINCT region FROM drug_channels WHERE is_audit = 1 AND region != ''"
	regionRows, err := db.MySQL.Query(regionQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询地区选项失败",
		})
		return
	}
	defer regionRows.Close()

	var regions []OptionItem
	for regionRows.Next() {
		var region string
		if err := regionRows.Scan(&region); err != nil {
			continue
		}
		regions = append(regions, OptionItem{
			Label: region,
			Value: region,
		})
	}

	// 获取配送方式选项（从 delivery_scope 提取或固定配置）
	deliveries := []OptionItem{
		{Label: "自提", Value: "pickup"},
		{Label: "快递", Value: "delivery"},
		{Label: "同城配送", Value: "local"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": ChannelOptionsResponse{
			Regions:    regions,
			Deliveries: deliveries,
		},
	})
}

// GetChannelList 获取渠道列表
func GetChannelList(c *gin.Context) {
	// 获取请求参数
	region := c.DefaultQuery("region", "")
	delivery := c.DefaultQuery("delivery", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}

	if region != "" {
		whereClause += " AND region = ?"
		args = append(args, region)
	}

	// 根据配送方式筛选（匹配 delivery_scope）
	if delivery != "" {
		switch delivery {
		case "pickup":
			whereClause += " AND delivery_scope LIKE ?"
			args = append(args, "%自提%")
		case "delivery":
			whereClause += " AND delivery_scope LIKE ?"
			args = append(args, "%快递%")
		case "local":
			whereClause += " AND delivery_scope LIKE ?"
			args = append(args, "%同城%")
		}
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM drug_channels " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 查询列表
	listQuery := `
		SELECT id, name, channel_type, address, region, contact, 
		       delivery_scope, delivery_cycle, is_insurance_settle
		FROM drug_channels
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询列表失败",
		})
		return
	}
	defer rows.Close()

	var list []ChannelItem
	for rows.Next() {
		var channel struct {
			ID            uint           `db:"id"`
			Name          string         `db:"name"`
			ChannelType   string         `db:"channel_type"`
			Address       sql.NullString `db:"address"`
			Region        string         `db:"region"`
			Contact       string         `db:"contact"`
			DeliveryScope string         `db:"delivery_scope"`
			DeliveryCycle string         `db:"delivery_cycle"`
			IsInsurance   int8           `db:"is_insurance_settle"`
		}
		if err := rows.Scan(
			&channel.ID, &channel.Name, &channel.ChannelType, &channel.Address,
			&channel.Region, &channel.Contact, &channel.DeliveryScope,
			&channel.DeliveryCycle, &channel.IsInsurance,
		); err != nil {
			continue
		}

		// 构建描述信息
		desc := channel.DeliveryScope
		if channel.DeliveryCycle != "" {
			desc += " | " + channel.DeliveryCycle
		}

		list = append(list, ChannelItem{
			ID:            channel.ID,
			Name:          channel.Name,
			Type:          channel.ChannelType,
			Address:       channel.Address.String,
			Desc:          desc,
			Region:        channel.Region,
			Contact:       channel.Contact,
			IsInsurance:   channel.IsInsurance == 1,
			DeliveryScope: channel.DeliveryScope,
			DeliveryCycle: channel.DeliveryCycle,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// ContactChannel 获取联系方式
func ContactChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道 ID",
		})
		return
	}

	// 使用匿名结构体解析请求体
	var req struct {
		ContactType string `json:"contactType"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 查询渠道信息
	query := `SELECT contact, name FROM drug_channels WHERE id = ? AND is_audit = 1`
	var contact, name string
	err = db.MySQL.QueryRow(query, id).Scan(&contact, &name)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "渠道不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询联系方式失败",
		})
		return
	}

	// 根据 contactType 解析联系方式
	contactData := ChannelContactResponse{
		Phone:  contact,
		Wechat: "",
		Email:  "",
	}

	// 如果 contact 是邮箱格式
	if len(contact) > 5 && contact[len(contact)-4:] == ".com" {
		contactData.Email = contact
	} else if len(contact) >= 11 {
		contactData.Phone = contact
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    contactData,
	})
}

// FeedbackChannel 提交反馈评价
func FeedbackChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道 ID",
		})
		return
	}

	// 使用匿名结构体解析请求体
	var req struct {
		Rating  int    `json:"rating"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 验证评分范围
	if req.Rating < 1 || req.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "评分必须在 1-5 之间",
		})
		return
	}

	// 验证渠道是否存在
	checkQuery := "SELECT id FROM drug_channels WHERE id = ? AND is_audit = 1"
	var channelID uint
	err = db.MySQL.QueryRow(checkQuery, id).Scan(&channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "渠道不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "验证渠道失败",
		})
		return
	}

	// 插入反馈记录
	insertQuery := `
		INSERT INTO drug_channel_feedback (channel_id, rating, content, created_at) 
		VALUES (?, ?, ?, ?)
	`
	_, err = db.MySQL.Exec(insertQuery, id, req.Rating, req.Content, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交反馈失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// GetChannelDetail 获取渠道详情及联系方式
func GetChannelDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道 ID",
		})
		return
	}

	// 查询渠道信息
	query := `
		SELECT id, name, channel_type, region, address, contact, 
		       delivery_scope, delivery_cycle, is_insurance_settle, is_audit
		FROM drug_channels
		WHERE id = ? AND is_audit = 1
	`

	var channel struct {
		ID            uint           `db:"id"`
		Name          string         `db:"name"`
		ChannelType   string         `db:"channel_type"`
		Region        string         `db:"region"`
		Address       sql.NullString `db:"address"`
		Contact       string         `db:"contact"`
		DeliveryScope string         `db:"delivery_scope"`
		DeliveryCycle string         `db:"delivery_cycle"`
		IsInsurance   int8           `db:"is_insurance_settle"`
		IsAudit       int8           `db:"is_audit"`
	}

	err = db.MySQL.QueryRow(query, id).Scan(
		&channel.ID, &channel.Name, &channel.ChannelType, &channel.Region,
		&channel.Address, &channel.Contact, &channel.DeliveryScope,
		&channel.DeliveryCycle, &channel.IsInsurance, &channel.IsAudit,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "渠道不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询渠道详情失败",
		})
		return
	}

	// 构建描述信息
	descParts := []string{}
	if channel.DeliveryScope != "" {
		descParts = append(descParts, channel.DeliveryScope)
	}
	if channel.IsInsurance == 1 {
		descParts = append(descParts, "支持医保")
	}
	if channel.DeliveryCycle != "" {
		descParts = append(descParts, channel.DeliveryCycle)
	}
	desc := ""
	if len(descParts) > 0 {
		desc = strings.Join(descParts, "，")
	}

	// 解析联系方式
	phone := ""
	wechat := ""
	email := ""

	// 根据 contact 字段格式解析联系方式
	if len(channel.Contact) > 5 && channel.Contact[len(channel.Contact)-4:] == ".com" {
		email = channel.Contact
	} else if len(channel.Contact) >= 11 {
		// 假设 contact 可能包含多个联系方式，用逗号分隔
		parts := strings.Split(channel.Contact, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if len(part) >= 11 && part[0] == '1' {
				phone = part
			} else if strings.HasPrefix(part, "wx") || strings.HasPrefix(part, "wechat") {
				wechat = part
			}
		}
		if phone == "" && len(channel.Contact) >= 11 {
			phone = channel.Contact
		}
	}

	// 构建配送方式
	delivery := channel.DeliveryScope
	if strings.Contains(channel.DeliveryScope, "自提") {
		delivery = "自提"
	} else if strings.Contains(channel.DeliveryScope, "快递") {
		delivery = "快递"
	} else if strings.Contains(channel.DeliveryScope, "同城") {
		delivery = "同城配送"
	}

	// 直接使用 gin.H 构建返回数据
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":       channel.ID,
			"name":     channel.Name,
			"type":     channel.ChannelType,
			"region":   channel.Region,
			"delivery": delivery,
			"address":  channel.Address.String,
			"desc":     desc,
			"phone":    phone,
			"wechat":   wechat,
			"email":    email,
			"verified": channel.IsAudit == 1,
		},
	})
}

// GetDonationList 获取赠药项目列表
// GetDonationList 获取赠药援助项目列表
func GetDonationList(c *gin.Context) {
	// 获取请求参数
	diseaseStr := c.DefaultQuery("disease", "0")
	drugIDStr := c.DefaultQuery("drugId", "0")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	disease, _ := strconv.Atoi(diseaseStr)
	drugID, _ := strconv.Atoi(drugIDStr)
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}

	if disease != 0 {
		whereClause += " AND disease_value = ?"
		args = append(args, disease)
	}

	if drugID != 0 {
		whereClause += " AND drug_id = ?"
		args = append(args, drugID)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM drug_relief_projects " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 查询列表
	listQuery := `
		SELECT id, drug_id, disease_value, name, organizer, apply_condition,
		       relief_cycle, drug_dosage, apply_form, apply_guide, material_list, progress_query
		FROM drug_relief_projects
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询列表失败",
		})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		var project struct {
			ID             uint           `db:"id"`
			DrugID         uint           `db:"drug_id"`
			DiseaseValue   int            `db:"disease_value"`
			Name           string         `db:"name"`
			Organizer      string         `db:"organizer"`
			ApplyCondition string         `db:"apply_condition"`
			ReliefCycle    string         `db:"relief_cycle"`
			DrugDosage     string         `db:"drug_dosage"`
			ApplyForm      sql.NullString `db:"apply_form"`
			ApplyGuide     sql.NullString `db:"apply_guide"`
			MaterialList   sql.NullString `db:"material_list"`
			ProgressQuery  sql.NullString `db:"progress_query"`
		}
		if err := rows.Scan(
			&project.ID, &project.DrugID, &project.DiseaseValue, &project.Name,
			&project.Organizer, &project.ApplyCondition, &project.ReliefCycle,
			&project.DrugDosage, &project.ApplyForm, &project.ApplyGuide,
			&project.MaterialList, &project.ProgressQuery,
		); err != nil {
			continue
		}

		list = append(list, gin.H{
			"id":            project.ID,
			"drugId":        project.DrugID,
			"diseaseValue":  project.DiseaseValue,
			"name":          project.Name,
			"organizer":     project.Organizer,
			"condition":     project.ApplyCondition,
			"period":        project.ReliefCycle,
			"dosage":        project.DrugDosage,
			"applyForm":     project.ApplyForm.String,
			"applyGuide":    project.ApplyGuide.String,
			"materialList":  project.MaterialList.String,
			"progressQuery": project.ProgressQuery.String,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// ApplyDonation 提交赠药援助申请
func ApplyDonation(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目 ID",
		})
		return
	}

	// 使用匿名结构体解析请求体
	var req struct {
		UserID         uint   `json:"userId"`
		PatientName    string `json:"patientName"`
		PatientIdCard  string `json:"patientIdCard"`
		DiagnosisProof string `json:"diagnosisProof"`
		IncomeProof    string `json:"incomeProof"`
		ContactPhone   string `json:"contactPhone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 验证必填字段
	if req.PatientName == "" || req.PatientIdCard == "" || req.ContactPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请填写完整信息",
		})
		return
	}

	// 验证身份证号格式
	if len(req.PatientIdCard) != 18 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "身份证号格式错误",
		})
		return
	}

	// 验证赠药项目是否存在
	checkQuery := "SELECT id, name FROM drug_relief_projects WHERE id = ? AND is_audit = 1"
	var projectName string
	err = db.MySQL.QueryRow(checkQuery, projectID).Scan(&projectID, &projectName)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "赠药项目不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "验证项目失败",
		})
		return
	}

	// 生成申请编号
	applicationID := fmt.Sprintf("RELIEF%s%03d", time.Now().Format("20060102"), projectID)

	// 插入申请记录
	insertQuery := `
		INSERT INTO drug_relief_applications 
		(application_id, project_id, user_id, patient_name, patient_id_card, 
		 diagnosis_proof, income_proof, contact_phone, status, status_text, submit_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending', '待审核', ?)
	`
	_, err = db.MySQL.Exec(insertQuery,
		applicationID, projectID, req.UserID, req.PatientName, req.PatientIdCard,
		req.DiagnosisProof, req.IncomeProof, req.ContactPhone, time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交申请失败",
		})
		return
	}

	// 插入进度日志
	logQuery := `
		INSERT INTO drug_relief_logs (application_id, status, desc, created_at)
		VALUES (?, 'submitted', '申请已提交', ?)
	`
	_, err = db.MySQL.Exec(logQuery, applicationID, time.Now())
	if err != nil {
		// 日志插入失败不影响主流程
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "申请提交成功",
		"data": gin.H{
			"applicationId": applicationID,
			"status":        "pending",
		},
	})
}

// GetDonationProgress 查询申请进度
// GetDonationProgress 查询申请进度
func GetDonationProgress(c *gin.Context) {
	idStr := c.Param("id")
	applicationID := idStr

	// 查询申请信息
	query := `
		SELECT a.application_id, a.project_id, p.name, a.status, a.status_text,
		       a.submit_time, a.update_time
		FROM drug_relief_applications a
		JOIN drug_relief_projects p ON a.project_id = p.id
		WHERE a.application_id = ?
	`

	var progress struct {
		ApplicationID string         `db:"application_id"`
		ProjectID     uint           `db:"project_id"`
		ProjectName   string         `db:"name"`
		Status        string         `db:"status"`
		StatusText    sql.NullString `db:"status_text"`
		SubmitTime    time.Time      `db:"submit_time"`
		UpdateTime    time.Time      `db:"update_time"`
	}

	err := db.MySQL.QueryRow(query, applicationID).Scan(
		&progress.ApplicationID, &progress.ProjectID, &progress.ProjectName,
		&progress.Status, &progress.StatusText, &progress.SubmitTime, &progress.UpdateTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "申请记录不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询进度失败",
		})
		return
	}

	// 查询进度日志
	logQuery := `
		SELECT status, desc, created_at
		FROM drug_relief_logs
		WHERE application_id = ?
		ORDER BY created_at ASC
	`

	logRows, err := db.MySQL.Query(logQuery, applicationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询日志失败",
		})
		return
	}
	defer logRows.Close()

	var logs []gin.H
	for logRows.Next() {
		var log struct {
			Status    string    `db:"status"`
			Desc      string    `db:"desc"`
			CreatedAt time.Time `db:"created_at"`
		}
		if err := logRows.Scan(&log.Status, &log.Desc, &log.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, gin.H{
			"time":   log.CreatedAt.Format("2006-01-02 15:04:05"),
			"status": log.Status,
			"desc":   log.Desc,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"applicationId": progress.ApplicationID,
			"projectId":     progress.ProjectID,
			"projectName":   progress.ProjectName,
			"status":        progress.Status,
			"statusText":    progress.StatusText.String,
			"submitTime":    progress.SubmitTime.Format("2006-01-02 15:04:05"),
			"updateTime":    progress.UpdateTime.Format("2006-01-02 15:04:05"),
			"logs":          logs,
		},
	})
}

// DownloadDonationGuide 下载赠药指南
func DownloadDonationGuide(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目 ID",
		})
		return
	}

	// 查询指南文件路径
	query := `SELECT apply_guide, name FROM drug_relief_projects WHERE id = ? AND is_audit = 1`
	var guideURL, name string
	err = db.MySQL.QueryRow(query, projectID).Scan(&guideURL, &name)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "赠药项目不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询指南失败",
		})
		return
	}

	if guideURL == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "指南文件不存在",
		})
		return
	}

	// 如果 guide_url 是本地文件路径
	if strings.HasPrefix(guideURL, "/") || strings.HasPrefix(guideURL, "./") {
		filePath := guideURL
		if !strings.HasPrefix(filePath, "/") {
			filePath = "./" + filePath
		}

		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "指南文件不存在",
			})
			return
		}

		// 设置响应头
		filename := fmt.Sprintf("%s_申请指南.pdf", name)
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

		// 输出文件
		c.File(filePath)
		return
	}

	// 如果 guide_url 是远程 URL，重定向到该 URL
	c.Redirect(http.StatusTemporaryRedirect, guideURL)
}

// GetToolList 获取工具列表
// GetToolList 获取用药管理工具列表
func GetToolList(c *gin.Context) {
	// 获取请求参数
	diseaseStr := c.DefaultQuery("disease", "0")
	toolType := c.DefaultQuery("toolType", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	disease, _ := strconv.Atoi(diseaseStr)
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := "WHERE is_audit = 1"
	args := []interface{}{}

	if disease != 0 {
		whereClause += " AND disease_value = ?"
		args = append(args, disease)
	}

	if toolType != "" {
		whereClause += " AND tool_type = ?"
		args = append(args, toolType)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM drug_manage_tools " + whereClause
	var total int64
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询总数失败",
		})
		return
	}

	// 查询列表
	listQuery := `
		SELECT id, disease_value, tool_type, name, record_template_excel, 
		       record_template_word, store_guide_pdf, content_intro, updated_at
		FROM drug_manage_tools
		` + whereClause + `
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询列表失败",
		})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		var tool struct {
			ID                  uint           `db:"id"`
			DiseaseValue        int            `db:"disease_value"`
			ToolType            string         `db:"tool_type"`
			Name                string         `db:"name"`
			RecordTemplateExcel sql.NullString `db:"record_template_excel"`
			RecordTemplateWord  sql.NullString `db:"record_template_word"`
			StoreGuidePDF       sql.NullString `db:"store_guide_pdf"`
			ContentIntro        string         `db:"content_intro"`
			UpdatedAt           time.Time      `db:"updated_at"`
		}
		if err := rows.Scan(
			&tool.ID, &tool.DiseaseValue, &tool.ToolType, &tool.Name,
			&tool.RecordTemplateExcel, &tool.RecordTemplateWord, &tool.StoreGuidePDF,
			&tool.ContentIntro, &tool.UpdatedAt,
		); err != nil {
			continue
		}

		// 构建文件列表
		files := []gin.H{}
		if tool.RecordTemplateExcel.Valid && tool.RecordTemplateExcel.String != "" {
			files = append(files, gin.H{
				"fileType":    "excel",
				"title":       tool.Name + "_用药记录模板",
				"downloadUrl": "/api/resource/drug/tool/download/" + strconv.FormatUint(uint64(tool.ID), 10) + "?type=excel",
			})
		}
		if tool.RecordTemplateWord.Valid && tool.RecordTemplateWord.String != "" {
			files = append(files, gin.H{
				"fileType":    "word",
				"title":       tool.Name + "_用药记录模板",
				"downloadUrl": "/api/resource/drug/tool/download/" + strconv.FormatUint(uint64(tool.ID), 10) + "?type=word",
			})
		}
		if tool.StoreGuidePDF.Valid && tool.StoreGuidePDF.String != "" {
			files = append(files, gin.H{
				"fileType":    "pdf",
				"title":       tool.Name + "_储存指南",
				"downloadUrl": "/api/resource/drug/tool/download/" + strconv.FormatUint(uint64(tool.ID), 10) + "?type=pdf",
			})
		}

		list = append(list, gin.H{
			"id":           tool.ID,
			"title":        tool.Name,
			"description":  tool.ContentIntro,
			"toolType":     tool.ToolType,
			"diseaseValue": tool.DiseaseValue,
			"files":        files,
			"updatedAt":    tool.UpdatedAt.Format("2006-01-02"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// DownloadTool 下载工具文件
// DownloadTool 下载工具文件
func DownloadTool(c *gin.Context) {
	idStr := c.Param("id")
	toolID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的工具 ID",
		})
		return
	}

	// 获取文件类型参数
	fileType := c.DefaultQuery("type", "")
	if fileType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请指定文件类型（type=excel/word/pdf）",
		})
		return
	}

	// 查询文件信息
	query := `SELECT name, record_template_excel, record_template_word, store_guide_pdf 
			  FROM drug_manage_tools WHERE id = ? AND is_audit = 1`

	var toolName string
	var recordTemplateExcel, recordTemplateWord, storeGuidePDF sql.NullString
	err = db.MySQL.QueryRow(query, toolID).Scan(
		&toolName, &recordTemplateExcel, &recordTemplateWord, &storeGuidePDF,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "工具文件不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询文件失败",
		})
		return
	}

	// 根据文件类型获取对应路径
	var filePath string
	var contentType string
	var fileExtension string

	switch fileType {
	case "excel":
		if !recordTemplateExcel.Valid || recordTemplateExcel.String == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Excel 模板文件不存在",
			})
			return
		}
		filePath = recordTemplateExcel.String
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		fileExtension = "xlsx"
	case "word":
		if !recordTemplateWord.Valid || recordTemplateWord.String == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Word 模板文件不存在",
			})
			return
		}
		filePath = recordTemplateWord.String
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		fileExtension = "docx"
	case "pdf":
		if !storeGuidePDF.Valid || storeGuidePDF.String == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "PDF 指南文件不存在",
			})
			return
		}
		filePath = storeGuidePDF.String
		contentType = "application/pdf"
		fileExtension = "pdf"
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的文件类型",
		})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件不存在",
		})
		return
	}

	// 设置响应头
	filename := fmt.Sprintf("%s.%s", toolName, fileExtension)
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// 输出文件
	c.File(filePath)
}
