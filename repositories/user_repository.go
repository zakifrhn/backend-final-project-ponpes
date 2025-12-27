package repositories

import (
	"backend-final-project-ponpes/models"
	"backend-final-project-ponpes/utils"
	"database/sql"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	query := `
        SELECT id_user, username, password, role, id_santri, id_ustad, id_orang_tua, 
               is_active, last_login, created_at, updated_at
        FROM users 
        WHERE username = $1 AND is_active = true
    `

	var user models.User
	var idSantri, idUstad, idOrangTua sql.NullInt64

	err := r.DB.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role,
		&idSantri, &idUstad, &idOrangTua,
		&user.IsActive, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if idSantri.Valid {
		santriID := int(idSantri.Int64)
		user.IDSantri = &santriID
	}
	if idUstad.Valid {
		ustadID := int(idUstad.Int64)
		user.IDUstad = &ustadID
	}
	if idOrangTua.Valid {
		orangTuaID := int(idOrangTua.Int64)
		user.IDOrangTua = &orangTuaID
	}

	return &user, nil
}

func (r *UserRepository) FindByUserId(id int) (*models.User, error) {
	query := `
        SELECT id_user, username, password, role, id_santri, id_ustad, id_orang_tua, 
               is_active, last_login, created_at, updated_at
        FROM users 
        WHERE id_user = $1 AND is_active = true
    `

	var user models.User
	var idSantri, idUstad, idOrangTua sql.NullInt64

	err := r.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Role,
		&idSantri,
		&idUstad,
		&idOrangTua,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if idSantri.Valid {
		v := int(idSantri.Int64)
		user.IDSantri = &v
	}
	if idUstad.Valid {
		v := int(idUstad.Int64)
		user.IDUstad = &v
	}
	if idOrangTua.Valid {
		v := int(idOrangTua.Int64)
		user.IDOrangTua = &v
	}

	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(userID int) error {
	query := `UPDATE users SET last_login = NOW() WHERE id_user = $1`
	_, err := r.DB.Exec(query, userID)
	return err
}

func (r *UserRepository) GetUserProfile(userID int, role string) (interface{}, error) {
	switch role {
	case "santri":
		return r.getSantriProfile(userID)
	case "ustad":
		return r.getUstadProfile(userID)
	case "orang_tua":
		return r.getOrangTuaProfile(userID)
	case "admin":
		return r.getAdminProfile(userID)
	default:
		return nil, sql.ErrNoRows
	}
}

func (r *UserRepository) getSantriProfile(userID int) (interface{}, error) {
	query := `
        SELECT s.id_santri, s.nis, s.nama_lengkap, s.nama_panggilan, s.tempat_lahir, 
               s.tanggal_lahir, s.jenis_kelamin, s.alamat_lengkap, s.no_telepon, s.email,
               s.foto_path, s.status_aktif, s.tanggal_masuk
        FROM users u
        JOIN md_biodata_santri s ON u.id_santri = s.id_santri
        WHERE u.id_user = $1
    `

	var santri struct {
		ID            int    `json:"id_santri"`
		NIS           string `json:"nis"`
		NamaLengkap   string `json:"nama_lengkap"`
		NamaPanggilan string `json:"nama_panggilan"`
		TempatLahir   string `json:"tempat_lahir"`
		TanggalLahir  string `json:"tanggal_lahir"`
		JenisKelamin  string `json:"jenis_kelamin"`
		Alamat        string `json:"alamat_lengkap"`
		NoTelepon     string `json:"no_telepon"`
		Email         string `json:"email"`
		FotoPath      string `json:"foto_path"`
		StatusAktif   string `json:"status_aktif"`
		TanggalMasuk  string `json:"tanggal_masuk"`
	}

	err := r.DB.QueryRow(query, userID).Scan(
		&santri.ID, &santri.NIS, &santri.NamaLengkap, &santri.NamaPanggilan,
		&santri.TempatLahir, &santri.TanggalLahir, &santri.JenisKelamin,
		&santri.Alamat, &santri.NoTelepon, &santri.Email, &santri.FotoPath,
		&santri.StatusAktif, &santri.TanggalMasuk,
	)

	return santri, err
}

func (r *UserRepository) getUstadProfile(userID int) (interface{}, error) {
	query := `
        SELECT u.id_ustad, u.nip, u.nama_lengkap, u.tempat_lahir, u.tanggal_lahir,
               u.jenis_kelamin, u.alamat_lengkap, u.no_telepon, u.email, u.foto_path,
               u.bidang_keahlian, u.jabatan, u.status_aktif
        FROM users us
        JOIN md_biodata_ustad u ON us.id_ustad = u.id_ustad
        WHERE us.id_user = $1
    `

	var ustad struct {
		ID             int    `json:"id_ustad"`
		NIP            string `json:"nip"`
		NamaLengkap    string `json:"nama_lengkap"`
		TempatLahir    string `json:"tempat_lahir"`
		TanggalLahir   string `json:"tanggal_lahir"`
		JenisKelamin   string `json:"jenis_kelamin"`
		Alamat         string `json:"alamat_lengkap"`
		NoTelepon      string `json:"no_telepon"`
		Email          string `json:"email"`
		FotoPath       string `json:"foto_path"`
		BidangKeahlian string `json:"bidang_keahlian"`
		Jabatan        string `json:"jabatan"`
		StatusAktif    string `json:"status_aktif"`
	}

	err := r.DB.QueryRow(query, userID).Scan(
		&ustad.ID, &ustad.NIP, &ustad.NamaLengkap, &ustad.TempatLahir,
		&ustad.TanggalLahir, &ustad.JenisKelamin, &ustad.Alamat, &ustad.NoTelepon,
		&ustad.Email, &ustad.FotoPath, &ustad.BidangKeahlian, &ustad.Jabatan,
		&ustad.StatusAktif,
	)

	return ustad, err
}

func (r *UserRepository) getOrangTuaProfile(userID int) (interface{}, error) {
	query := `
        SELECT ot.id_orang_tua, ot.nama_ayah, ot.pekerjaan_ayah, ot.no_telepon_ayah,
               ot.nama_ibu, ot.pekerjaan_ibu, ot.no_telepon_ibu, ot.alamat_orang_tua,
               s.nama_lengkap as nama_santri, s.nis
        FROM users us
        JOIN md_biodata_orang_tua ot ON us.id_orang_tua = ot.id_orang_tua
        JOIN md_biodata_santri s ON ot.id_santri = s.id_santri
        WHERE us.id_user = $1
    `

	var orangTua struct {
		ID            int    `json:"id_orang_tua"`
		NamaAyah      string `json:"nama_ayah"`
		PekerjaanAyah string `json:"pekerjaan_ayah"`
		TeleponAyah   string `json:"no_telepon_ayah"`
		NamaIbu       string `json:"nama_ibu"`
		PekerjaanIbu  string `json:"pekerjaan_ibu"`
		TeleponIbu    string `json:"no_telepon_ibu"`
		Alamat        string `json:"alamat_orang_tua"`
		NamaSantri    string `json:"nama_santri"`
		NISSantri     string `json:"nis_santri"`
	}

	err := r.DB.QueryRow(query, userID).Scan(
		&orangTua.ID, &orangTua.NamaAyah, &orangTua.PekerjaanAyah, &orangTua.TeleponAyah,
		&orangTua.NamaIbu, &orangTua.PekerjaanIbu, &orangTua.TeleponIbu, &orangTua.Alamat,
		&orangTua.NamaSantri, &orangTua.NISSantri,
	)

	return orangTua, err
}

func (r *UserRepository) getAdminProfile(userID int) (interface{}, error) {
	query := `SELECT id_user, username, role, created_at FROM users WHERE id_user = $1`

	var admin struct {
		ID        int    `json:"id_user"`
		Username  string `json:"username"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
	}

	err := r.DB.QueryRow(query, userID).Scan(
		&admin.ID, &admin.Username, &admin.Role, &admin.CreatedAt,
	)

	return admin, err
}

func (r *UserRepository) CreateUser(users *models.UserDTO) (string, error) {
	// Implementation for creating a new user in the database

	query := `
		INSERT INTO users (username, password, role, id_santri, id_ustad, id_orang_tua, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		returning id_user
	`

	var newID int

	hashPassword, errors := utils.HashPassword(users.Password)
	if errors != nil {
		return "", errors
	}

	users.Password = hashPassword

	err := r.DB.QueryRow(
		query,
		users.Username,
		users.Password,
		users.Role,
		users.IDSantri,
		users.IDUstad,
		users.IDOrangTua,
		users.IsActive,
	).Scan(&newID)

	if err != nil {
		return "", err
	}

	users.ID = newID

	return "user created successfully", nil

}

func (r *UserRepository) UpdateUser(users *models.UserDTO) (string, error) {
	// Mulai transaksi
	tx, err := r.DB.Begin()
	if err != nil {
		return "", err
	}

	// Jika terjadi panic atau error → otomatis rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := `
        UPDATE users
        SET
            username      = COALESCE(NULLIF($1, ''), username),
            password      = COALESCE(NULLIF($2, ''), password),
            role          = COALESCE(NULLIF($3, ''), role),
            id_santri     = COALESCE($4, id_santri),
            id_ustad      = COALESCE($5, id_ustad),
            id_orang_tua  = COALESCE($6, id_orang_tua),
            is_active     = COALESCE($7, is_active),
            updated_at    = NOW()
        WHERE id_user = $8
        RETURNING id_user
    `

	var updatedID int

	hashPassword, errors := utils.HashPassword(users.Password)
	if errors != nil {
		return "", errors
	}

	users.Password = hashPassword

	// Eksekusi query pakai tx.QueryRow
	err = tx.QueryRow(
		query,
		users.Username,
		users.Password,
		users.Role,
		users.IDSantri,
		users.IDUstad,
		users.IDOrangTua,
		users.IsActive,
		users.ID,
	).Scan(&updatedID)

	if err != nil {
		tx.Rollback()
		return "", err
	}

	// Commit kalau semua aman
	if err := tx.Commit(); err != nil {
		return "", err
	}

	return "user updated successfully", nil
}
