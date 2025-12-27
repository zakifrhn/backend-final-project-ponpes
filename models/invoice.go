package models

import "time"

type Invoice struct {
	IDInvoice       int        `db:"id_invoice" json:"id_invoice"`
	Deskripsi       string     `db:"deskripsi" json:"deskripsi"`
	DeadlineTagihan time.Time  `db:"deadline_tagihan" json:"deadline_tagihan"`
	IDSantri        int        `db:"id_santri" json:"id_santri"`
	IDOrangTua      *int       `db:"id_orang_tua" json:"id_orang_tua,omitempty"`
	NominalTagihan  string     `db:"nominal_tagihan" json:"nominal_tagihan"`
	Status          string     `db:"status" json:"status"`
	CreatedDate     time.Time  `db:"created_date" json:"created_date"`
	CreatedBy       *string    `db:"created_by" json:"created_by,omitempty"`
	UpdatedDate     time.Time  `db:"updated_date" json:"updated_date"`
	UpdatedBy       *string    `db:"updated_by" json:"updated_by,omitempty"`
	DeletedBy       *string    `db:"deleted_by" json:"deleted_by,omitempty"`
	DeletedDate     *time.Time `db:"deleted_date" json:"deleted_date,omitempty"`
}

type InvoiceDetail struct {
	Invoice
	NIS           string  `db:"nis" json:"nis"`
	NamaSantri    string  `db:"nama_lengkap" json:"nama_santri"`
	TeleponSantri *string `db:"no_telepon" json:"telepon_santri,omitempty"`
	EmailSantri   *string `db:"email" json:"email_santri,omitempty"`
	NamaAyah      *string `db:"nama_ayah" json:"nama_ayah,omitempty"`
	NamaIbu       *string `db:"nama_ibu" json:"nama_ibu,omitempty"`
	TeleponAyah   *string `db:"no_telepon_ayah" json:"telepon_ayah,omitempty"`
	TeleponIbu    *string `db:"no_telepon_ibu" json:"telepon_ibu,omitempty"`
}

type CreateInvoiceRequest struct {
	Deskripsi       string `json:"deskripsi" binding:"required"`
	DeadlineTagihan string `json:"deadline_tagihan" binding:"required"`
	IDSantri        int    `json:"id_santri" binding:"required"`
	NominalTagihan  string `json:"nominal_tagihan" binding:"required"`
}
