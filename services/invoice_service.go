package services

import (
	"backend-final-project-ponpes/models"
	"backend-final-project-ponpes/repositories"
	"database/sql"
	"fmt"
	"time"
)

type InvoiceService interface {
	GetInvoiceByID(id int) (*models.InvoiceDetail, error)
	GetInvoicesBySantri(idSantri int) ([]models.InvoiceDetail, error)
	GetAllInvoices(status, month, year string) ([]models.InvoiceDetail, error)
	CreateInvoice(req models.CreateInvoiceRequest, createdBy string) (int, error)
	UpdateInvoice(id int, req map[string]interface{}, updatedBy string) error
	UpdateInvoiceStatus(id int, status, updatedBy string) error
	DeleteInvoice(id int, deletedBy string) error
	GetInvoicesByOrangTua(idOrangTua int) ([]models.InvoiceDetail, error)
	GenerateMonthlyInvoices(createdBy string) (int, error)
	CheckOverdueInvoices() error
}

type invoiceService struct {
	invoiceRepo repositories.InvoiceRepository
	santriRepo  repositories.SantriRepository
	db          *sql.DB
}

func NewInvoiceService(invoiceRepo repositories.InvoiceRepository, santriRepo repositories.SantriRepository, db *sql.DB) InvoiceService {
	return &invoiceService{
		invoiceRepo: invoiceRepo,
		santriRepo:  santriRepo,
		db:          db,
	}
}

func (s *invoiceService) GetInvoiceByID(id int) (*models.InvoiceDetail, error) {
	return s.invoiceRepo.GetInvoiceByID(id)
}

func (s *invoiceService) GetInvoicesBySantri(idSantri int) ([]models.InvoiceDetail, error) {
	return s.invoiceRepo.GetInvoicesBySantri(idSantri)
}

func (s *invoiceService) GetAllInvoices(status, month, year string) ([]models.InvoiceDetail, error) {
	return s.invoiceRepo.GetAllInvoices(status, month, year)
}

func (s *invoiceService) CreateInvoice(req models.CreateInvoiceRequest, createdBy string) (int, error) {
	// Parse deadline
	deadline, err := time.Parse("2006-01-02", req.DeadlineTagihan)
	if err != nil {
		return 0, fmt.Errorf("format tanggal tidak valid: %v", err)
	}

	// Validasi nominal
	if req.NominalTagihan <= 0 {
		return 0, fmt.Errorf("nominal tagihan harus lebih dari 0")
	}

	// Validasi santri exists
	invoice := &models.Invoice{
		Deskripsi:       req.Deskripsi,
		DeadlineTagihan: deadline,
		IDSantri:        req.IDSantri,
		NominalTagihan:  req.NominalTagihan,
		Status:          "BELUM_LUNAS",
		CreatedBy:       &createdBy,
	}

	return s.invoiceRepo.CreateInvoice(invoice)
}

func (s *invoiceService) UpdateInvoice(id int, req map[string]interface{}, updatedBy string) error {
	// Cek invoice exists
	exists, err := s.invoiceRepo.CheckInvoiceExists(id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("invoice tidak ditemukan")
	}

	// Get current invoice
	currentInvoice, err := s.invoiceRepo.GetInvoiceByID(id)
	if err != nil {
		return err
	}

	// Update fields
	invoice := &models.Invoice{
		IDInvoice:   id,
		UpdatedBy:   &updatedBy,
		UpdatedDate: time.Now(),
	}

	// Update deskripsi jika ada
	if deskripsi, ok := req["deskripsi"].(string); ok && deskripsi != "" {
		invoice.Deskripsi = deskripsi
	} else {
		invoice.Deskripsi = currentInvoice.Deskripsi
	}

	// Update deadline jika ada
	if deadlineStr, ok := req["deadline_tagihan"].(string); ok && deadlineStr != "" {
		deadline, err := time.Parse("2006-01-02", deadlineStr)
		if err != nil {
			return fmt.Errorf("format tanggal tidak valid: %v", err)
		}
		invoice.DeadlineTagihan = deadline
	} else {
		invoice.DeadlineTagihan = currentInvoice.DeadlineTagihan
	}

	// Update nominal jika ada
	if nominal, ok := req["nominal_tagihan"].(float64); ok && nominal > 0 {
		invoice.NominalTagihan = nominal
	} else {
		invoice.NominalTagihan = currentInvoice.NominalTagihan
	}

	// Update status jika ada
	if status, ok := req["status"].(string); ok && status != "" {
		validStatus := map[string]bool{
			"BELUM_LUNAS": true,
			"LUNAS":       true,
			"JATUH_TEMPO": true,
			"DIBATALKAN":  true,
		}
		if !validStatus[status] {
			return fmt.Errorf("status tidak valid")
		}
		invoice.Status = status
	} else {
		invoice.Status = currentInvoice.Status
	}

	return s.invoiceRepo.UpdateInvoice(invoice)
}

func (s *invoiceService) UpdateInvoiceStatus(id int, status, updatedBy string) error {
	validStatus := map[string]bool{
		"BELUM_LUNAS": true,
		"LUNAS":       true,
		"JATUH_TEMPO": true,
		"DIBATALKAN":  true,
	}
	if !validStatus[status] {
		return fmt.Errorf("status tidak valid")
	}

	return s.invoiceRepo.UpdateInvoiceStatus(id, status, updatedBy)
}

func (s *invoiceService) DeleteInvoice(id int, deletedBy string) error {
	return s.invoiceRepo.DeleteInvoice(id, deletedBy)
}

func (s *invoiceService) GetInvoicesByOrangTua(idOrangTua int) ([]models.InvoiceDetail, error) {
	return s.invoiceRepo.GetInvoicesByOrangTua(idOrangTua)
}

func (s *invoiceService) GenerateMonthlyInvoices(createdBy string) (int, error) {
	// Ambil bulan dan tahun sekarang
	now := time.Now()
	currentMonth := int(now.Month())
	currentYear := now.Year()

	// Cek apakah sudah ada invoice untuk bulan ini
	checkQuery := `SELECT COUNT(*) FROM tr_invoice_spp 
		WHERE EXTRACT(MONTH FROM deadline_tagihan) = $1 
		AND EXTRACT(YEAR FROM deadline_tagihan) = $2 
		AND deleted_date IS NULL`

	var existingCount int
	err := s.db.QueryRow(checkQuery, currentMonth, currentYear).Scan(&existingCount)
	if err != nil {
		return 0, err
	}

	if existingCount > 0 {
		return 0, fmt.Errorf("invoice untuk bulan %d %d sudah ada", currentMonth, currentYear)
	}

	// Ambil semua santri aktif
	query := `SELECT s.id_santri, s.nis, s.nama_lengkap, o.id_orang_tua
		FROM md_biodata_santri s
		LEFT JOIN md_biodata_orang_tua o ON s.id_santri = o.id_santri
		WHERE s.status_aktif = 'Aktif' AND s.deleted_at IS NULL`

	rows, err := s.db.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	monthName := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}

	// Deadline: akhir bulan
	deadline := time.Date(currentYear, now.Month()+1, 0, 23, 59, 59, 0, now.Location())

	for rows.Next() {
		var idSantri int
		var nis, namaSantri string
		var idOrangTua sql.NullInt64

		if err := rows.Scan(&idSantri, &nis, &namaSantri, &idOrangTua); err != nil {
			continue
		}

		// Buat deskripsi
		deskripsi := fmt.Sprintf("SPP %s %d", monthName[currentMonth], currentYear)
		nominal := 500000.0 // Default, bisa disesuaikan

		invoice := &models.Invoice{
			Deskripsi:       deskripsi,
			DeadlineTagihan: deadline,
			IDSantri:        idSantri,
			NominalTagihan:  nominal,
			Status:          "BELUM_LUNAS",
			CreatedBy:       &createdBy,
		}

		if idOrangTua.Valid {
			idOrtu := int(idOrangTua.Int64)
			invoice.IDOrangTua = &idOrtu
		}

		_, err := s.invoiceRepo.CreateInvoice(invoice)
		if err != nil {
			// Log error tapi lanjut ke santri berikutnya
			fmt.Printf("Gagal membuat invoice untuk santri %s: %v\n", nis, err)
			continue
		}

		count++
	}

	return count, nil
}

func (s *invoiceService) CheckOverdueInvoices() error {
	query := `UPDATE tr_invoice_spp 
		SET status = 'JATUH_TEMPO', updated_by = 'system', updated_date = $1
		WHERE status = 'BELUM_LUNAS' 
		AND deadline_tagihan < CURRENT_DATE
		AND deleted_date IS NULL`

	_, err := s.db.Exec(query, time.Now())
	return err
}
