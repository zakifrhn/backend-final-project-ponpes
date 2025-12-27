package repositories

import (
	"backend-final-project-ponpes/models"
	"database/sql"
	"fmt"
)

type SantriRepository interface {
	GetSantriByID(id int) (*models.Santri, error)
	GetSantriByOrangTua(idOrangTua int) ([]models.Santri, error)
	GetOrangTuaBySantri(idSantri int) (*models.OrangTua, error)
}

type santriRepository struct {
	DB *sql.DB
}

func NewSantriRepository(db *sql.DB) SantriRepository {
	return &santriRepository{DB: db}
}

func (r *santriRepository) GetSantriByID(id int) (*models.Santri, error) {
	var santri models.Santri
	var email, noTelepon sql.NullString

	query := `SELECT id_santri, nama_lengkap, email, no_telepon, status_aktif
              FROM md_biodata_santri 
              WHERE id_santri = $1`

	err := r.DB.QueryRow(query, id).Scan(
		&santri.IDSantri,
		&santri.NamaLengkap,
		&email,
		&noTelepon,
		&santri.StatusAktif,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("santri dengan ID %d tidak ditemukan", id)
		}
		return nil, fmt.Errorf("gagal mengambil data santri: %v", err)
	}

	// Convert NullString to string pointer
	if email.Valid {
		santri.Email = &email.String
	}

	if noTelepon.Valid {
		santri.NoTelepon = &noTelepon.String
	}

	return &santri, nil
}

func (r *santriRepository) GetSantriByOrangTua(idOrangTua int) ([]models.Santri, error) {
	query := `SELECT 
		s.id_santri, s.nis, s.nama_lengkap, s.nama_panggilan, s.tempat_lahir,
		s.tanggal_lahir, s.jenis_kelamin, s.no_telepon, s.email, s.status_aktif
	FROM md_biodata_santri s
	INNER JOIN md_biodata_orang_tua o ON s.id_santri = o.id_santri
	WHERE o.id_orang_tua = $1 AND s.deleted_at IS NULL`

	rows, err := r.DB.Query(query, idOrangTua)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var santriList []models.Santri
	for rows.Next() {
		var santri models.Santri
		var (
			namaPanggilan, tempatLahir, jenisKelamin, noTelepon, email sql.NullString
			tanggalLahir                                               sql.NullTime
		)

		err := rows.Scan(
			&santri.IDSantri, &santri.NIS, &santri.NamaLengkap,
			&namaPanggilan, &tempatLahir, &tanggalLahir,
			&jenisKelamin, &noTelepon, &email, &santri.StatusAktif,
		)
		if err != nil {
			return nil, err
		}

		if namaPanggilan.Valid {
			santri.NamaPanggilan = &namaPanggilan.String
		}
		if tempatLahir.Valid {
			santri.TempatLahir = &tempatLahir.String
		}
		if tanggalLahir.Valid {
			santri.TanggalLahir = &tanggalLahir.Time
		}
		if jenisKelamin.Valid {
			santri.JenisKelamin = &jenisKelamin.String
		}
		if noTelepon.Valid {
			santri.NoTelepon = &noTelepon.String
		}
		if email.Valid {
			santri.Email = &email.String
		}

		santriList = append(santriList, santri)
	}

	return santriList, nil
}

func (r *santriRepository) GetOrangTuaBySantri(idSantri int) (*models.OrangTua, error) {
	query := `SELECT 
		id_orang_tua, id_santri, nama_ayah, pekerjaan_ayah, pendidikan_ayah,
		no_telepon_ayah, nama_ibu, pekerjaan_ibu, pendidikan_ibu, no_telepon_ibu
	FROM md_biodata_orang_tua 
	WHERE id_santri = $1 AND deleted_at IS NULL`

	var orangTua models.OrangTua
	var (
		idSantriDB sql.NullInt64
		namaAyah, pekerjaanAyah, pendidikanAyah, noTeleponAyah,
		namaIbu, pekerjaanIbu, pendidikanIbu, noTeleponIbu sql.NullString
	)

	err := r.DB.QueryRow(query, idSantri).Scan(
		&orangTua.IDOrangTua, &idSantriDB, &namaAyah, &pekerjaanAyah,
		&pendidikanAyah, &noTeleponAyah, &namaIbu, &pekerjaanIbu,
		&pendidikanIbu, &noTeleponIbu,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Orang tua not found, return nil without error
		}
		return nil, err
	}

	// Set nullable fields
	if idSantriDB.Valid {
		id := int(idSantriDB.Int64)
		orangTua.IDSantri = &id
	}
	if namaAyah.Valid {
		orangTua.NamaAyah = &namaAyah.String
	}
	if pekerjaanAyah.Valid {
		orangTua.PekerjaanAyah = &pekerjaanAyah.String
	}
	if pendidikanAyah.Valid {
		orangTua.PendidikanAyah = &pendidikanAyah.String
	}
	if noTeleponAyah.Valid {
		orangTua.NoTeleponAyah = &noTeleponAyah.String
	}
	if namaIbu.Valid {
		orangTua.NamaIbu = &namaIbu.String
	}
	if pekerjaanIbu.Valid {
		orangTua.PekerjaanIbu = &pekerjaanIbu.String
	}
	if pendidikanIbu.Valid {
		orangTua.PendidikanIbu = &pendidikanIbu.String
	}
	if noTeleponIbu.Valid {
		orangTua.NoTeleponIbu = &noTeleponIbu.String
	}

	return &orangTua, nil
}
