package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"

	"rare_backend/internal/module/resource/drug/domain"
	"rare_backend/internal/module/resource/drug/repo"
)

type DrugService struct {
	repo *repo.DrugRepo
}

func NewDrugService(r *repo.DrugRepo) *DrugService {
	return &DrugService{repo: r}
}

func BuildDrugListFilter(keyword, drugType, isInsurance, hasRelief, auditStatusStr, pageStr, pageSizeStr string) domain.DrugListFilter {
	filter := domain.DrugListFilter{
		Keyword:     keyword,
		DrugType:    drugType,
		IsInsurance: isInsurance,
		HasRelief:   hasRelief,
		Page:        1,
		PageSize:    10,
	}
	if page, err := strconv.Atoi(pageStr); err == nil && page >= 1 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
		if pageSize >= 1 && pageSize <= 100 {
			filter.PageSize = pageSize
		}
	}
	if auditStatusStr != "" {
		if auditStatus, err := strconv.Atoi(auditStatusStr); err == nil {
			filter.AuditStatus = auditStatus
			filter.AuditStatusExplicit = true
		}
	}
	return filter
}

func BuildDrugExportFilter(diseaseStr, keyword, typeFilter, insuranceStr string) domain.DrugExportFilter {
	diseaseID, _ := strconv.Atoi(diseaseStr)
	return domain.DrugExportFilter{
		DiseaseID:  diseaseID,
		Keyword:    keyword,
		TypeFilter: typeFilter,
		Insurance:  insuranceStr,
	}
}

func (s *DrugService) List(filter domain.DrugListFilter) (*domain.DrugListResult, error) {
	total, err := s.repo.CountList(filter)
	if err != nil {
		return nil, domain.ErrCountList
	}

	rows, err := s.repo.List(filter)
	if err != nil {
		return nil, domain.ErrQueryList
	}

	drugIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		drugIDs = append(drugIDs, uint64(row.ID))
	}
	relMap, _ := s.repo.ListDiseaseIDsByDrugIDs(drugIDs)

	list := make([]domain.DrugListItem, 0, len(rows))
	for _, row := range rows {
		diseaseIds := relMap[uint64(row.ID)]
		if diseaseIds == nil {
			diseaseIds = []uint64{}
		}
		list = append(list, domain.DrugListItem{
			ID:               row.ID,
			GenericName:      row.GenericName,
			BrandName:        nullString(row.BrandName),
			Indication:       row.Indication,
			DrugType:         row.DrugType,
			IsInsurance:      row.IsInsurance == 1,
			DosageForm:       row.DosageForm,
			Spec:             row.Spec,
			RefPrice:         nullString(row.RefPrice),
			HasRelief:        row.HasRelief == 1,
			IsLaunched:       row.IsLaunched == 1,
			NeedPrescription: row.NeedPrescription == 1,
			ManualOriginal:   nullString(row.ManualOriginal),
			ManualPopular:    nullString(row.ManualPopular),
			AuditStatus:      row.AuditStatus,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
			DiseaseIds:       diseaseIds,
		})
	}
	if list == nil {
		list = []domain.DrugListItem{}
	}

	return &domain.DrugListResult{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *DrugService) GetByID(id uint) (*domain.DrugDetailResponse, error) {
	row, err := s.repo.GetByID(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrDrugNotFound
		}
		return nil, err
	}

	diseases, err := s.repo.ListDiseasesByDrugID(id)
	if err != nil {
		diseases = nil
	}
	if diseases == nil {
		diseases = []domain.DiseaseSimple{}
	}

	return &domain.DrugDetailResponse{
		ID:               row.ID,
		GenericName:      row.GenericName,
		BrandName:        nullString(row.BrandName),
		Indication:       row.Indication,
		DrugType:         row.DrugType,
		IsInsurance:      row.IsInsurance == 1,
		DosageForm:       row.DosageForm,
		Spec:             row.Spec,
		RefPrice:         nullString(row.RefPrice),
		HasRelief:        row.HasRelief == 1,
		IsLaunched:       row.IsLaunched == 1,
		NeedPrescription: row.NeedPrescription == 1,
		ManualOriginal:   nullString(row.ManualOriginal),
		ManualPopular:    nullString(row.ManualPopular),
		AuditStatus:      row.AuditStatus,
		CreatedAt:        row.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        row.UpdatedAt.Format(time.RFC3339),
		Diseases:         diseases,
	}, nil
}

func (s *DrugService) Create(input domain.CreateDrugInput) (int64, error) {
	id, err := s.repo.Create(input)
	if err != nil {
		return 0, err
	}
	if len(input.DiseaseIds) > 0 {
		if err := s.repo.InsertDiseaseRels(id, input.DiseaseIds); err != nil {
			return id, domain.ErrDrugRelFailed
		}
	}
	return id, nil
}

func (s *DrugService) Update(id uint, input domain.UpdateDrugInput) error {
	exists, err := s.repo.Exists(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return domain.ErrDrugNotFound
		}
		return err
	}
	if !exists {
		return domain.ErrDrugNotFound
	}

	fields := []string{}
	args := []interface{}{}

	if input.GenericName != nil {
		fields = append(fields, "generic_name = ?")
		args = append(args, *input.GenericName)
	}
	if input.BrandName != nil {
		fields = append(fields, "brand_name = ?")
		args = append(args, *input.BrandName)
	}
	if input.Indication != nil {
		fields = append(fields, "indication = ?")
		args = append(args, *input.Indication)
	}
	if input.DrugType != nil {
		fields = append(fields, "drug_type = ?")
		args = append(args, *input.DrugType)
	}
	if input.IsInsurance != nil {
		fields = append(fields, "is_insurance = ?")
		args = append(args, boolToInt(*input.IsInsurance))
	}
	if input.DosageForm != nil {
		fields = append(fields, "dosage_form = ?")
		args = append(args, *input.DosageForm)
	}
	if input.Spec != nil {
		fields = append(fields, "spec = ?")
		args = append(args, *input.Spec)
	}
	if input.RefPrice != nil {
		fields = append(fields, "ref_price = ?")
		args = append(args, fmt.Sprintf("%.2f", *input.RefPrice))
	}
	if input.HasRelief != nil {
		fields = append(fields, "has_relief = ?")
		args = append(args, boolToInt(*input.HasRelief))
	}
	if input.IsLaunched != nil {
		fields = append(fields, "is_launched = ?")
		args = append(args, boolToInt(*input.IsLaunched))
	}
	if input.NeedPrescription != nil {
		fields = append(fields, "need_prescription = ?")
		args = append(args, boolToInt(*input.NeedPrescription))
	}
	if input.ManualOriginal != nil {
		fields = append(fields, "manual_original = ?")
		args = append(args, *input.ManualOriginal)
	}
	if input.ManualPopular != nil {
		fields = append(fields, "manual_popular = ?")
		args = append(args, *input.ManualPopular)
	}
	if input.AuditStatus != nil {
		fields = append(fields, "audit_status = ?")
		args = append(args, *input.AuditStatus)
	}
	if input.RejectReason != nil {
		fields = append(fields, "reject_reason = ?")
		args = append(args, *input.RejectReason)
	}

	if len(fields) > 0 {
		if err := s.repo.Update(id, fields, args); err != nil {
			return err
		}
	}

	if input.HasDiseaseIds {
		if err := s.repo.DeleteDiseaseRels(id); err != nil {
			return err
		}
		if len(input.DiseaseIds) > 0 {
			if err := s.repo.InsertDiseaseRels(int64(id), input.DiseaseIds); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *DrugService) Delete(id uint) error {
	exists, err := s.repo.Exists(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return domain.ErrDrugNotFound
		}
		return err
	}
	if !exists {
		return domain.ErrDrugNotFound
	}
	if err := s.repo.DeleteDiseaseRels(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *DrugService) GetManual(id uint) (*domain.ManualResponse, error) {
	original, popular, err := s.repo.GetManual(id)
	if err != nil {
		if repo.IsNoRows(err) {
			return nil, domain.ErrDrugNotFound
		}
		return nil, err
	}
	return &domain.ManualResponse{
		Original: original,
		Popular:  popular,
		URL:      original,
	}, nil
}

func (s *DrugService) ExportExcel(filter domain.DrugExportFilter) (*excelize.File, string, error) {
	rows, err := s.repo.ListForExport(filter)
	if err != nil {
		return nil, "", err
	}

	excel := excelize.NewFile()
	sheetName := "药品名录"
	index, err := excel.NewSheet(sheetName)
	if err != nil {
		return nil, "", err
	}
	excel.SetActiveSheet(index)
	excel.DeleteSheet("Sheet1")

	headers := []string{"通用名", "商品名", "适应症", "类型", "医保", "剂型", "规格", "参考价格", "赠药援助", "国内上市"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		excel.SetCellValue(sheetName, cell, header)
	}

	rowNum := 2
	for _, drug := range rows {
		data := []interface{}{
			drug.GenericName,
			drug.BrandName,
			truncateString(drug.Indication, 50),
			convertDrugType(drug.DrugType),
			mapYesNo(drug.IsInsurance),
			drug.DosageForm,
			drug.Spec,
			drug.RefPrice,
			mapYesNo(drug.HasRelief),
			mapYesNo(drug.IsLaunched),
		}
		for i, value := range data {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowNum)
			excel.SetCellValue(sheetName, cell, value)
		}
		rowNum++
	}

	for i := 1; i <= len(headers); i++ {
		col, _ := excelize.ColumnNumberToName(i)
		excel.SetColWidth(sheetName, col, col, 20)
	}

	filename := fmt.Sprintf("药品名录_%s.xlsx", time.Now().Format("20060102150405"))
	return excel, filename, nil
}

func (s *DrugService) Options() *domain.DrugOptionsResult {
	return &domain.DrugOptionsResult{
		Types: []domain.OptionItem{
			{Label: "进口药", Value: "origin_import"},
			{Label: "国产药", Value: "origin_domestic"},
			{Label: "仿制药", Value: "generic"},
			{Label: "其他", Value: "other"},
		},
		Insurances: []domain.OptionItem{
			{Label: "医保", Value: "true"},
			{Label: "非医保", Value: "false"},
		},
	}
}
