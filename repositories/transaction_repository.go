package repositories

import (
	"backend-final-project-ponpes/models"
	"database/sql"
	"fmt"
	"time"
)

type TransactionRepository interface {
	CreateTransaction(trans *models.Transaction) (int, error)
	GetTransactionByKode(kode string) (*models.Transaction, error)
	GetTransactionByID(id int) (*models.Transaction, error)
	UpdateTransactionStatus(kode, status, updatedBy string) error
	UpdateTransactionFromWebhook(transactionID, status string, amount string, midtransID, updatedBy string) error
	GetTransactionHistory(idSantri int) ([]models.TransactionDetail, error)
	GetAllTransactions(status, startDate, endDate string) ([]models.TransactionDetail, error)
	GetTransactionsByOrangTua(idOrangTua int) ([]models.TransactionDetail, error)
}

type transactionRepository struct {
	DB *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{DB: db}
}

func (r *transactionRepository) CreateTransaction(trans *models.Transaction) (int, error) {
	query := `INSERT INTO tr_transaksi_spp 
		(id_invoice, nominal_tagihan, jumlah_bayar, status, 
		 metode_pembayaran, kode_pembayaran, payment_url,
		 created_by, created_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id_transaksi`

	var id int
	err := r.DB.QueryRow(query,
		trans.IDInvoice, trans.NominalTagihan, trans.JumlahBayar,
		trans.Status, trans.MetodePembayaran, trans.KodePembayaran,
		trans.PaymentURL, trans.CreatedBy, time.Now(),
	).Scan(&id)

	return id, err
}

func (r *transactionRepository) GetTransactionByKode(kode string) (*models.Transaction, error) {
	query := `SELECT id_transaksi, id_invoice, nominal_tagihan, jumlah_bayar,
			 transaction_date, status, metode_pembayaran, kode_pembayaran,
			 payment_url, midtrans_transaction_id, created_date, created_by
		FROM tr_transaksi_spp 
		WHERE kode_pembayaran = $1 AND deleted_date IS NULL`

	trans := &models.Transaction{}
	err := r.DB.QueryRow(query, kode).Scan(
		&trans.IDTransaksi, &trans.IDInvoice, &trans.NominalTagihan,
		&trans.JumlahBayar, &trans.TransactionDate, &trans.Status,
		&trans.MetodePembayaran, &trans.KodePembayaran, &trans.PaymentURL,
		&trans.MidtransTransactionID, &trans.CreatedDate, &trans.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaksi tidak ditemukan")
		}
		return nil, err
	}

	return trans, nil
}

func (r *transactionRepository) GetTransactionByID(id int) (*models.Transaction, error) {
	query := `SELECT id_transaksi, id_invoice, nominal_tagihan, jumlah_bayar,
			 transaction_date, status, metode_pembayaran, kode_pembayaran,
			 payment_url, midtrans_transaction_id, created_date, created_by
		FROM tr_transaksi_spp 
		WHERE id_transaksi = $1 AND deleted_date IS NULL`

	trans := &models.Transaction{}
	err := r.DB.QueryRow(query, id).Scan(
		&trans.IDTransaksi, &trans.IDInvoice, &trans.NominalTagihan,
		&trans.JumlahBayar, &trans.TransactionDate, &trans.Status,
		&trans.MetodePembayaran, &trans.KodePembayaran, &trans.PaymentURL,
		&trans.MidtransTransactionID, &trans.CreatedDate, &trans.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaksi tidak ditemukan")
		}
		return nil, err
	}

	return trans, nil
}

func (r *transactionRepository) UpdateTransactionStatus(kode, status, updatedBy string) error {
	query := `UPDATE tr_transaksi_spp 
		SET status = $1, updated_by = $2, updated_date = $3
		WHERE kode_pembayaran = $4 AND deleted_date IS NULL`

	_, err := r.DB.Exec(query, status, updatedBy, time.Now(), kode)
	return err
}

func (r *transactionRepository) UpdateTransactionFromWebhook(transactionID, status string, amount string, midtransID, updatedBy string) error {
	query := `UPDATE tr_transaksi_spp 
		SET status = $1, jumlah_bayar = $2, transaction_date = $3,
		    midtrans_transaction_id = $4, updated_by = $5, updated_date = $6
		WHERE kode_pembayaran = $7 AND deleted_date IS NULL`

	_, err := r.DB.Exec(query, status, amount, time.Now(),
		midtransID, updatedBy, time.Now(), transactionID)
	return err
}

func (r *transactionRepository) GetTransactionHistory(idSantri int) ([]models.TransactionDetail, error) {
	query := `SELECT 
		t.id_transaksi, t.id_invoice, t.nominal_tagihan, t.jumlah_bayar,
		t.transaction_date, t.status, t.metode_pembayaran, t.kode_pembayaran,
		t.midtrans_transaction_id, t.created_date, i.deskripsi, s.nis, s.nama_lengkap
	FROM tr_transaksi_spp t
	INNER JOIN tr_invoice_spp i ON t.id_invoice = i.id_invoice
	INNER JOIN md_biodata_santri s ON i.id_santri = s.id_santri
	WHERE i.id_santri = $1 AND t.deleted_date IS NULL
	ORDER BY t.transaction_date DESC`

	rows, err := r.DB.Query(query, idSantri)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.TransactionDetail
	for rows.Next() {
		var trans models.TransactionDetail
		var metodePembayaran, midtransTransactionID sql.NullString
		err := rows.Scan(
			&trans.IDTransaksi, &trans.IDInvoice, &trans.NominalTagihan,
			&trans.JumlahBayar, &trans.TransactionDate, &trans.Status,
			&metodePembayaran, &trans.KodePembayaran, &midtransTransactionID,
			&trans.CreatedDate, &trans.Deskripsi, &trans.NIS, &trans.NamaSantri,
		)
		if err != nil {
			return nil, err
		}
		if metodePembayaran.Valid {
			trans.MetodePembayaran = &metodePembayaran.String
		}
		if midtransTransactionID.Valid {
			trans.MidtransTransactionID = &midtransTransactionID.String
		}
		transactions = append(transactions, trans)
	}

	return transactions, nil
}

func (r *transactionRepository) GetAllTransactions(status, startDate, endDate string) ([]models.TransactionDetail, error) {
	query := `SELECT 
		t.id_transaksi, t.id_invoice, t.nominal_tagihan, t.jumlah_bayar,
		t.transaction_date, t.status, t.metode_pembayaran, t.kode_pembayaran,
		t.midtrans_transaction_id, t.created_date, i.deskripsi, s.nis, s.nama_lengkap
	FROM tr_transaksi_spp t
	INNER JOIN tr_invoice_spp i ON t.id_invoice = i.id_invoice
	INNER JOIN md_biodata_santri s ON i.id_santri = s.id_santri
	WHERE t.deleted_date IS NULL`

	args := []interface{}{}
	argIndex := 1

	if status != "" {
		query += fmt.Sprintf(" AND t.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if startDate != "" {
		query += fmt.Sprintf(" AND t.transaction_date >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}

	if endDate != "" {
		query += fmt.Sprintf(" AND t.transaction_date <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	query += " ORDER BY t.transaction_date DESC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.TransactionDetail
	for rows.Next() {
		var trans models.TransactionDetail
		var metodePembayaran, midtransTransactionID sql.NullString
		err := rows.Scan(
			&trans.IDTransaksi, &trans.IDInvoice, &trans.NominalTagihan,
			&trans.JumlahBayar, &trans.TransactionDate, &trans.Status,
			&metodePembayaran, &trans.KodePembayaran, &midtransTransactionID,
			&trans.CreatedDate, &trans.Deskripsi, &trans.NIS, &trans.NamaSantri,
		)
		if err != nil {
			return nil, err
		}
		if metodePembayaran.Valid {
			trans.MetodePembayaran = &metodePembayaran.String
		}
		if midtransTransactionID.Valid {
			trans.MidtransTransactionID = &midtransTransactionID.String
		}
		transactions = append(transactions, trans)
	}

	return transactions, nil
}

func (r *transactionRepository) GetTransactionsByOrangTua(idOrangTua int) ([]models.TransactionDetail, error) {
	query := `SELECT 
		t.id_transaksi, t.id_invoice, t.nominal_tagihan, t.jumlah_bayar,
		t.transaction_date, t.status, t.metode_pembayaran, t.kode_pembayaran,
		t.midtrans_transaction_id, t.created_date, i.deskripsi, s.nis, s.nama_lengkap
	FROM tr_transaksi_spp t
	INNER JOIN tr_invoice_spp i ON t.id_invoice = i.id_invoice
	INNER JOIN md_biodata_santri s ON i.id_santri = s.id_santri
	INNER JOIN md_biodata_orang_tua o ON i.id_santri = o.id_santri
	WHERE o.id_orang_tua = $1 AND t.deleted_date IS NULL
	ORDER BY t.transaction_date DESC`

	rows, err := r.DB.Query(query, idOrangTua)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.TransactionDetail
	for rows.Next() {
		var trans models.TransactionDetail
		var metodePembayaran, midtransTransactionID sql.NullString
		err := rows.Scan(
			&trans.IDTransaksi, &trans.IDInvoice, &trans.NominalTagihan,
			&trans.JumlahBayar, &trans.TransactionDate, &trans.Status,
			&metodePembayaran, &trans.KodePembayaran, &midtransTransactionID,
			&trans.CreatedDate, &trans.Deskripsi, &trans.NIS, &trans.NamaSantri,
		)
		if err != nil {
			return nil, err
		}
		if metodePembayaran.Valid {
			trans.MetodePembayaran = &metodePembayaran.String
		}
		if midtransTransactionID.Valid {
			trans.MidtransTransactionID = &midtransTransactionID.String
		}
		transactions = append(transactions, trans)
	}

	return transactions, nil
}
