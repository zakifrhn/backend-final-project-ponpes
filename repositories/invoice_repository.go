package repositories

import (
	"backend-final-project-ponpes/models"
	"database/sql"
	"fmt"
	"time"
)

type InvoiceRepository interface {
	GetInvoiceByID(id int) (*models.InvoiceDetail, error)
	GetInvoicesBySantri(idSantri int) ([]models.InvoiceDetail, error)
	GetAllInvoices(status, month, year string) ([]models.InvoiceDetail, error)
	CreateInvoice(invoice *models.Invoice) (int, error)
	UpdateInvoice(invoice *models.Invoice) error
	UpdateInvoiceStatus(id int, status, updatedBy string) error
	DeleteInvoice(id int, deletedBy string) error
	GetInvoicesByOrangTua(idOrangTua int) ([]models.InvoiceDetail, error)
	CheckInvoiceExists(id int) (bool, error)
}

type invoiceRepository struct {
	DB *sql.DB
}

func NewInvoiceRepository(db *sql.DB) InvoiceRepository {
	return &invoiceRepository{DB: db}
}

func (r *invoiceRepository) GetInvoiceByID(id int) (*models.InvoiceDetail, error) {
	query := `SELECT 
		i.id_invoice, i.deskripsi, i.deadline_tagihan, 
		i.id_santri, i.id_orang_tua, i.nominal_tagihan, 
		i.status, i.created_date, i.created_by,
		s.nis, s.nama_lengkap, s.no_telepon, s.email,
		o.nama_ayah, o.nama_ibu, o.no_telepon_ayah, o.no_telepon_ibu
	FROM tr_invoice_spp i
	INNER JOIN md_biodata_santri s ON i.id_santri = s.id_santri
	LEFT JOIN md_biodata_orang_tua o ON i.id_orang_tua = o.id_orang_tua
	WHERE i.id_invoice = $1 AND i.deleted_date IS NULL`

	var invoice models.InvoiceDetail
	var idOrangTua sql.NullInt64
	var createdBy, noTelepon, email, namaAyah, namaIbu, teleponAyah, teleponIbu sql.NullString

	err := r.DB.QueryRow(query, id).Scan(
		&invoice.IDInvoice, &invoice.Deskripsi, &invoice.DeadlineTagihan,
		&invoice.IDSantri, &idOrangTua, &invoice.NominalTagihan,
		&invoice.Status, &invoice.CreatedDate, &createdBy,
		&invoice.NIS, &invoice.NamaSantri, &noTelepon, &email,
		&namaAyah, &namaIbu, &teleponAyah, &teleponIbu,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice tidak ditemukan")
		}
		return nil, err
	}

	if idOrangTua.Valid {
		idOrangTuaInt := int(idOrangTua.Int64)
		invoice.IDOrangTua = &idOrangTuaInt
	}
	if createdBy.Valid {
		invoice.CreatedBy = &createdBy.String
	}
	if noTelepon.Valid {
		invoice.TeleponSantri = &noTelepon.String
	}
	if email.Valid {
		invoice.EmailSantri = &email.String
	}
	if namaAyah.Valid {
		invoice.NamaAyah = &namaAyah.String
	}
	if namaIbu.Valid {
		invoice.NamaIbu = &namaIbu.String
	}
	if teleponAyah.Valid {
		invoice.TeleponAyah = &teleponAyah.String
	}
	if teleponIbu.Valid {
		invoice.TeleponIbu = &teleponIbu.String
	}

	return &invoice, nil
}

func (r *invoiceRepository) GetInvoicesBySantri(idSantri int) ([]models.InvoiceDetail, error) {
	query := `SELECT 
		i.id_invoice, i.deskripsi, i.deadline_tagihan, 
		i.id_santri, i.nominal_tagihan, i.status, i.created_date,
		s.nis, s.nama_lengkap
	FROM tr_invoice_spp i
	INNER JOIN md_biodata_santri s ON i.id_santri = s.id_santri
	WHERE i.id_santri = $1 AND i.deleted_date IS NULL
	ORDER BY i.deadline_tagihan DESC`

	rows, err := r.DB.Query(query, idSantri)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []models.InvoiceDetail
	for rows.Next() {
		var invoice models.InvoiceDetail
		err := rows.Scan(
			&invoice.IDInvoice, &invoice.Deskripsi, &invoice.DeadlineTagihan,
			&invoice.IDSantri, &invoice.NominalTagihan, &invoice.Status,
			&invoice.CreatedDate, &invoice.NIS, &invoice.NamaSantri,
		)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *invoiceRepository) GetAllInvoices(status, month, year string) ([]models.InvoiceDetail, error) {
	query := `SELECT 
		i.id_invoice, i.deskripsi, i.deadline_tagihan, 
		i.id_santri, i.nominal_tagihan, i.status, i.created_date,
		s.nis, s.nama_lengkap
	FROM tr_invoice_spp i
	INNER JOIN md_biodata_santri s ON i.id_santri = s.id_santri
	WHERE i.deleted_date IS NULL`

	args := []interface{}{}
	argIndex := 1

	if status != "" {
		query += fmt.Sprintf(" AND i.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if month != "" {
		query += fmt.Sprintf(" AND EXTRACT(MONTH FROM i.deadline_tagihan) = $%d", argIndex)
		args = append(args, month)
		argIndex++
	}

	if year != "" {
		query += fmt.Sprintf(" AND EXTRACT(YEAR FROM i.deadline_tagihan) = $%d", argIndex)
		args = append(args, year)
		argIndex++
	}

	query += " ORDER BY i.deadline_tagihan DESC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []models.InvoiceDetail
	for rows.Next() {
		var invoice models.InvoiceDetail
		err := rows.Scan(
			&invoice.IDInvoice, &invoice.Deskripsi, &invoice.DeadlineTagihan,
			&invoice.IDSantri, &invoice.NominalTagihan, &invoice.Status,
			&invoice.CreatedDate, &invoice.NIS, &invoice.NamaSantri,
		)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *invoiceRepository) CreateInvoice(invoice *models.Invoice) (int, error) {
	query := `INSERT INTO tr_invoice_spp 
		(deskripsi, deadline_tagihan, id_santri, id_orang_tua, 
		 nominal_tagihan, status, created_by, created_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id_invoice`

	var id int
	err := r.DB.QueryRow(query,
		invoice.Deskripsi, invoice.DeadlineTagihan, invoice.IDSantri,
		invoice.IDOrangTua, invoice.NominalTagihan, invoice.Status,
		invoice.CreatedBy, time.Now(),
	).Scan(&id)

	return id, err
}

func (r *invoiceRepository) UpdateInvoice(invoice *models.Invoice) error {
	query := `UPDATE tr_invoice_spp 
		SET deskripsi = $1, deadline_tagihan = $2, nominal_tagihan = $3,
			status = $4, updated_by = $5, updated_date = $6
		WHERE id_invoice = $7 AND deleted_date IS NULL`

	_, err := r.DB.Exec(query,
		invoice.Deskripsi, invoice.DeadlineTagihan, invoice.NominalTagihan,
		invoice.Status, invoice.UpdatedBy, time.Now(), invoice.IDInvoice,
	)

	return err
}

func (r *invoiceRepository) UpdateInvoiceStatus(id int, status, updatedBy string) error {
	query := `UPDATE tr_invoice_spp 
		SET status = $1, updated_by = $2, updated_date = $3
		WHERE id_invoice = $4 AND deleted_date IS NULL`

	_, err := r.DB.Exec(query, status, updatedBy, time.Now(), id)
	return err
}

func (r *invoiceRepository) DeleteInvoice(id int, deletedBy string) error {
	query := `UPDATE tr_invoice_spp 
		SET deleted_by = $1, deleted_date = $2
		WHERE id_invoice = $3`

	_, err := r.DB.Exec(query, deletedBy, time.Now(), id)
	return err
}

func (r *invoiceRepository) GetInvoicesByOrangTua(idOrangTua int) ([]models.InvoiceDetail, error) {
	query := `SELECT 
		i.id_invoice, i.deskripsi, i.deadline_tagihan, 
		i.id_santri, i.nominal_tagihan, i.status, i.created_date,
		s.nis, s.nama_lengkap, s.no_telepon
	FROM tr_invoice_spp i
	INNER JOIN md_biodata_santri s ON i.id_santri = s.id_santri
	INNER JOIN md_biodata_orang_tua o ON i.id_santri = o.id_santri
	WHERE o.id_orang_tua = $1 AND i.deleted_date IS NULL
	ORDER BY i.deadline_tagihan DESC`

	rows, err := r.DB.Query(query, idOrangTua)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []models.InvoiceDetail
	for rows.Next() {
		var invoice models.InvoiceDetail
		var noTelepon sql.NullString
		err := rows.Scan(
			&invoice.IDInvoice, &invoice.Deskripsi, &invoice.DeadlineTagihan,
			&invoice.IDSantri, &invoice.NominalTagihan, &invoice.Status,
			&invoice.CreatedDate, &invoice.NIS, &invoice.NamaSantri, &noTelepon,
		)
		if err != nil {
			return nil, err
		}
		if noTelepon.Valid {
			invoice.TeleponSantri = &noTelepon.String
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *invoiceRepository) CheckInvoiceExists(id int) (bool, error) {
	query := `SELECT COUNT(*) FROM tr_invoice_spp 
		WHERE id_invoice = $1 AND deleted_date IS NULL`

	var count int
	err := r.DB.QueryRow(query, id).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
