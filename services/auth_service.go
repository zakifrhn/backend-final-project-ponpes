package services

import (
	"backend-final-project-ponpes/models"
	"backend-final-project-ponpes/repositories"
	"backend-final-project-ponpes/utils"
	"time"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Login(loginReq models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(loginReq.Username)
	if err != nil {
		return nil, err
	}

	token, err := utils.GenerateJWT(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}
	err = s.userRepo.UpdateLastLogin(user.ID)
	if err != nil {
	}
	profile, err := s.userRepo.GetUserProfile(user.ID, user.Role)
	var nama string

	if err != nil {
		nama = user.Username
	} else {
		switch user.Role {
		case "santri":
			if santri, ok := profile.(struct {
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
			}); ok {
				nama = santri.NamaLengkap
			}
		case "ustad":
			if ustad, ok := profile.(struct {
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
				BidangKeahlian string `json:"bidang_kehahlian"`
				Jabatan        string `json:"jabatan"`
				StatusAktif    string `json:"status_aktif"`
			}); ok {
				nama = ustad.NamaLengkap
			}
		case "orang_tua":
			if orangTua, ok := profile.(struct {
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
			}); ok {
				nama = orangTua.NamaAyah
			}
		case "admin":
			nama = "Administrator"
		}
	}
	var lastLoginTime time.Time
	if user.LastLogin.Valid {
		lastLoginTime = user.LastLogin.Time
	} else {
		lastLoginTime = time.Time{}
	}

	response := &models.LoginResponse{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		Nama:      nama,
		LastLogin: lastLoginTime,
	}
	return response, nil
}
