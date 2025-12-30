package models

import (
	"time"
)

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
	IDInvoice       int       `json:"id_invoice"`
	Deskripsi       string    `json:"deskripsi"`
	DeadlineTagihan time.Time `json:"deadline_tagihan"`
	IDSantri        int       `json:"id_santri"`
	IDOrangTua      *int      `json:"id_orang_tua,omitempty"`
	NominalTagihan  string    `json:"nominal_tagihan"`
	Status          string    `json:"status"`
	CreatedDate     time.Time `json:"created_date"`
	UpdatedDate     time.Time `json:"updated_date"`
	CreatedBy       *string   `json:"created_by,omitempty"`

	// Data santri
	NIS           string  `json:"nis"`
	NamaSantri    string  `json:"nama_santri"`
	TeleponSantri *string `json:"telepon_santri,omitempty"`
	EmailSantri   *string `json:"email_santri,omitempty"`

	// Data orang tua (opsional)
	NamaAyah    *string `json:"nama_ayah,omitempty"`
	NamaIbu     *string `json:"nama_ibu,omitempty"`
	TeleponAyah *string `json:"telepon_ayah,omitempty"`
	TeleponIbu  *string `json:"telepon_ibu,omitempty"`
}

type CreateInvoiceRequest struct {
	Deskripsi       string `json:"deskripsi" binding:"required"`
	DeadlineTagihan string `json:"deadline_tagihan" binding:"required"`
	IDSantri        int    `json:"id_santri" binding:"required"`
	NominalTagihan  string `json:"nominal_tagihan" binding:"required"`
}

type GetAllInvoicesRequest struct {
	StartDate string `json:"start_date" example:"2024-01-01"`
	EndDate   string `json:"end_date" example:"2024-12-31"`
}
