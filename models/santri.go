package models

import "time"

// Santri model sesuai dengan tabel md_biodata_santri
type Santri struct {
	IDSantri      int        `db:"id_santri" json:"id_santri"`
	NIS           string     `db:"nis" json:"nis"`
	NamaLengkap   string     `db:"nama_lengkap" json:"nama_lengkap"`
	NamaPanggilan *string    `db:"nama_panggilan" json:"nama_panggilan,omitempty"`
	TempatLahir   *string    `db:"tempat_lahir" json:"tempat_lahir,omitempty"`
	TanggalLahir  *time.Time `db:"tanggal_lahir" json:"tanggal_lahir,omitempty"`
	JenisKelamin  *string    `db:"jenis_kelamin" json:"jenis_kelamin,omitempty"`
	AlamatLengkap *string    `db:"alamat_lengkap" json:"alamat_lengkap,omitempty"`
	Provinsi      *string    `db:"provinsi" json:"provinsi,omitempty"`
	Kabupaten     *string    `db:"kabupaten" json:"kabupaten,omitempty"`
	Kecamatan     *string    `db:"kecamatan" json:"kecamatan,omitempty"`
	Desa          *string    `db:"desa" json:"desa,omitempty"`
	NoTelepon     *string    `db:"no_telepon" json:"no_telepon,omitempty"`
	Email         *string    `db:"email" json:"email,omitempty"`
	FotoPath      *string    `db:"foto_path" json:"foto_path,omitempty"`
	StatusAktif   string     `db:"status_aktif" json:"status_aktif"`
	TanggalMasuk  *time.Time `db:"tanggal_masuk" json:"tanggal_masuk,omitempty"`
	TanggalKeluar *time.Time `db:"tanggal_keluar" json:"tanggal_keluar,omitempty"`
}

// OrangTua model sesuai dengan tabel md_biodata_orang_tua
type OrangTua struct {
	IDOrangTua     int     `db:"id_orang_tua" json:"id_orang_tua"`
	IDSantri       *int    `db:"id_santri" json:"id_santri,omitempty"`
	NamaAyah       *string `db:"nama_ayah" json:"nama_ayah,omitempty"`
	PekerjaanAyah  *string `db:"pekerjaan_ayah" json:"pekerjaan_ayah,omitempty"`
	PendidikanAyah *string `db:"pendidikan_ayah" json:"pendidikan_ayah,omitempty"`
	NoTeleponAyah  *string `db:"no_telepon_ayah" json:"no_telepon_ayah,omitempty"`
	NamaIbu        *string `db:"nama_ibu" json:"nama_ibu,omitempty"`
	PekerjaanIbu   *string `db:"pekerjaan_ibu" json:"pekerjaan_ibu,omitempty"`
	PendidikanIbu  *string `db:"pendidikan_ibu" json:"pendidikan_ibu,omitempty"`
	NoTeleponIbu   *string `db:"no_telepon_ibu" json:"no_telepon_ibu,omitempty"`
}
