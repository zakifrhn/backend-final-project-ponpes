package models

import "time"

type Transaction struct {
	IDTransaksi           int        `db:"id_transaksi" json:"id_transaksi"`
	IDInvoice             int        `db:"id_invoice" json:"id_invoice"`
	NominalTagihan        float64    `db:"nominal_tagihan" json:"nominal_tagihan"`
	JumlahBayar           float64    `db:"jumlah_bayar" json:"jumlah_bayar"`
	TransactionDate       time.Time  `db:"transaction_date" json:"transaction_date"`
	Status                string     `db:"status" json:"status"`
	BuktiBayarPath        *string    `db:"bukti_bayar_path" json:"bukti_bayar_path,omitempty"`
	MetodePembayaran      *string    `db:"metode_pembayaran" json:"metode_pembayaran,omitempty"`
	KodePembayaran        string     `db:"kode_pembayaran" json:"kode_pembayaran"`
	PaymentURL            *string    `db:"payment_url" json:"payment_url,omitempty"`
	MidtransTransactionID *string    `db:"midtrans_transaction_id" json:"midtrans_transaction_id,omitempty"`
	CreatedDate           time.Time  `db:"created_date" json:"created_date"`
	CreatedBy             *string    `db:"created_by" json:"created_by,omitempty"`
	UpdatedDate           time.Time  `db:"updated_date" json:"updated_date"`
	UpdatedBy             *string    `db:"updated_by" json:"updated_by,omitempty"`
	DeletedBy             *string    `db:"deleted_by" json:"deleted_by,omitempty"`
	DeletedDate           *time.Time `db:"deleted_date" json:"deleted_date,omitempty"`
}

type TransactionDetail struct {
	Transaction
	Deskripsi  string `db:"deskripsi" json:"deskripsi"`
	NIS        string `db:"nis" json:"nis"`
	NamaSantri string `db:"nama_lengkap" json:"nama_santri"`
}
