package repo

import (
	"database/sql"
	"time"

	"rare_backend/internal/module/resource/drug/domain"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/search"
)

type ChannelRepo struct{}

func NewChannelRepo() *ChannelRepo {
	return &ChannelRepo{}
}

type channelDetailRow struct {
	ID              uint
	Name            string
	ChannelType     string
	ProvinceCode    string
	CityCode        string
	DistrictCode    sql.NullString
	Address         sql.NullString
	ContactPhone    sql.NullString
	ContactURL      sql.NullString
	DeliveryScope   string
	DeliveryCycle   string
	IsInsurance     int8
	AuditStatus     int8
	Qualification   sql.NullString
	ProvinceName    sql.NullString
	CityName        sql.NullString
	DistrictName    sql.NullString
}

type channelDrugRow struct {
	ID          uint
	GenericName string
	BrandName   sql.NullString
}

func (r *ChannelRepo) buildListWhere(filter domain.ChannelListFilter) (string, []interface{}) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.AuditStatusExplicit {
		whereClause += " AND c.audit_status = ?"
		args = append(args, filter.AuditStatus)
	} else {
		whereClause += " AND c.audit_status = 1"
	}

	if filter.IsInsuranceSettle != "" {
		isInsuranceSettle := 0
		if filter.IsInsuranceSettle == "true" || filter.IsInsuranceSettle == "1" {
			isInsuranceSettle = 1
		}
		whereClause += " AND c.is_insurance_settle = ?"
		args = append(args, isInsuranceSettle)
	}
	if filter.Keyword != "" {
		if clause, arg, ok := search.MatchClause("c.name", filter.Keyword); ok {
			whereClause += clause
			args = append(args, arg)
		}
	}
	if filter.ProvinceCode != "" {
		whereClause += " AND c.province_code = ?"
		args = append(args, filter.ProvinceCode)
	}
	if filter.CityCode != "" {
		whereClause += " AND c.city_code = ?"
		args = append(args, filter.CityCode)
	}
	if filter.DistrictCode != "" {
		whereClause += " AND c.district_code = ?"
		args = append(args, filter.DistrictCode)
	}
	if filter.ChannelType != "" {
		whereClause += " AND c.channel_type = ?"
		args = append(args, filter.ChannelType)
	}
	if filter.DeliveryScope != "" {
		switch filter.DeliveryScope {
		case "pickup":
			whereClause += " AND c.delivery_scope LIKE ?"
			args = append(args, "%仅门店自提%")
		case "delivery":
			whereClause += " AND c.delivery_scope LIKE ?"
			args = append(args, "%全国%")
		case "local":
			whereClause += " AND c.delivery_scope LIKE ?"
			args = append(args, "%同城%")
		}
	}
	return whereClause, args
}

func (r *ChannelRepo) CountList(filter domain.ChannelListFilter) (int64, error) {
	whereClause, args := r.buildListWhere(filter)
	var total int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) FROM drug_channel c "+whereClause, args...).Scan(&total)
	return total, err
}

func (r *ChannelRepo) List(filter domain.ChannelListFilter) ([]map[string]interface{}, error) {
	whereClause, args := r.buildListWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT 
			c.id, c.name, c.channel_type, c.province_code, c.city_code, c.district_code, 
			c.address, c.contact_phone, c.contact_url,
			c.delivery_scope, c.delivery_cycle, c.is_insurance_settle, c.qualification, 
			c.created_at, c.updated_at,
			c.province_name, c.city_name, c.district_name,
			c.audit_status
		FROM drug_channel c
		` + whereClause + `
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var (
			id                uint
			name              string
			channelType       string
			provinceCode      string
			cityCode          string
			districtCode      sql.NullString
			address           sql.NullString
			contactPhone      sql.NullString
			contactUrl        sql.NullString
			deliveryScope     string
			deliveryCycle     string
			isInsuranceSettle int8
			qualification     sql.NullString
			createdAt         time.Time
			updatedAt         time.Time
			provinceName      sql.NullString
			cityName          sql.NullString
			districtName      sql.NullString
			auditStatus       int8
		)
		if err := rows.Scan(
			&id, &name, &channelType, &provinceCode, &cityCode, &districtCode,
			&address, &contactPhone, &contactUrl,
			&deliveryScope, &deliveryCycle, &isInsuranceSettle,
			&qualification, &createdAt, &updatedAt,
			&provinceName, &cityName, &districtName,
			&auditStatus,
		); err != nil {
			continue
		}

		item := map[string]interface{}{
			"id":                id,
			"name":              name,
			"channelType":       channelType,
			"provinceCode":      provinceCode,
			"cityCode":          cityCode,
			"districtCode":      districtCode.String,
			"address":           address.String,
			"contactPhone":      contactPhone.String,
			"contactUrl":        contactUrl.String,
			"deliveryScope":     deliveryScope,
			"deliveryCycle":     deliveryCycle,
			"isInsuranceSettle": isInsuranceSettle == 1,
			"qualification":     qualification.String,
			"createdAt":         createdAt.Format(time.RFC3339),
			"updatedAt":         updatedAt.Format(time.RFC3339),
			"provinceName":      provinceName.String,
			"cityName":          cityName.String,
			"districtName":      districtName.String,
			"auditStatus":       auditStatus,
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *ChannelRepo) GetApprovedByID(id uint) (*channelDetailRow, error) {
	query := `
		SELECT 
			id, name, channel_type, province_code, city_code, district_code, address, contact_phone, contact_url,
			delivery_scope, delivery_cycle, is_insurance_settle, audit_status, qualification,
			province_name, city_name, district_name
		FROM drug_channel
		WHERE id = ? AND audit_status = 1
	`
	var row channelDetailRow
	err := db.MySQL.QueryRow(query, id).Scan(
		&row.ID, &row.Name, &row.ChannelType, &row.ProvinceCode, &row.CityCode, &row.DistrictCode,
		&row.Address, &row.ContactPhone, &row.ContactURL,
		&row.DeliveryScope, &row.DeliveryCycle, &row.IsInsurance, &row.AuditStatus, &row.Qualification,
		&row.ProvinceName, &row.CityName, &row.DistrictName,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ChannelRepo) ListDrugsByChannelID(channelID uint) ([]channelDrugRow, error) {
	drugQuery := `
		SELECT d.id, d.generic_name, d.brand_name
		FROM drug_channel_drug_rel rel
		JOIN rare_drug d ON rel.drug_id = d.id
		WHERE rel.channel_id = ?
	`
	rows, err := db.MySQL.Query(drugQuery, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drugs []channelDrugRow
	for rows.Next() {
		var d channelDrugRow
		if err := rows.Scan(&d.ID, &d.GenericName, &d.BrandName); err != nil {
			continue
		}
		drugs = append(drugs, d)
	}
	return drugs, rows.Err()
}

func (r *ChannelRepo) Exists(id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM drug_channel WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *ChannelRepo) ApprovedExists(id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM drug_channel WHERE id = ? AND audit_status = 1", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *ChannelRepo) Create(input domain.CreateChannelInput) (int64, error) {
	insertQuery := `
		INSERT INTO drug_channel 
		(drug_id, name, channel_type, province_code, city_code, district_code, address, contact_phone, contact_url,
		 delivery_scope, delivery_cycle, is_insurance_settle, qualification, audit_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
	`
	now := time.Now()
	result, err := db.MySQL.Exec(insertQuery,
		input.DrugID, input.Name, input.ChannelType, input.ProvinceCode, input.CityCode, input.DistrictCode, input.Address,
		input.ContactPhone, input.ContactURL,
		input.DeliveryScope, input.DeliveryCycle, boolToInt(input.IsInsuranceSettle),
		input.Qualification, now, now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *ChannelRepo) Update(id uint, fields []string, args []interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	fields = append(fields, "updated_at = ?")
	args = append(args, time.Now(), id)
	updateQuery := "UPDATE drug_channel SET " + joinUpdateFields(fields) + " WHERE id = ?"
	_, err := db.MySQL.Exec(updateQuery, args...)
	return err
}

func (r *ChannelRepo) SoftDelete(id uint) error {
	_, err := db.MySQL.Exec("UPDATE drug_channel SET audit_status = 2, updated_at = ? WHERE id = ?", time.Now(), id)
	return err
}

func (r *ChannelRepo) GetContact(id uint) (phone, url, name string, err error) {
	query := `SELECT contact_phone, contact_url, name FROM drug_channel WHERE id = ? AND audit_status = 1`
	err = db.MySQL.QueryRow(query, id).Scan(&phone, &url, &name)
	return
}
