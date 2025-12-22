package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID         int          `json:"id_user"`
	Username   string       `json:"username"`
	Password   string       `json:"password,omitempty"`
	Role       string       `json:"role"`
	IDSantri   *int         `json:"id_santri,omitempty"`
	IDUstad    *int         `json:"id_ustad,omitempty"`
	IDOrangTua *int         `json:"id_orang_tua,omitempty"`
	IsActive   bool         `json:"is_active"`
	LastLogin  sql.NullTime `json:"last_login"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  sql.NullTime `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Nama      string    `json:"nama"`
	LastLogin time.Time `json:"last_login"`
}

type UserDTO struct {
	ID         int    `json:"id_user"`
	Username   string `json:"username"`
	Password   string `json:"password,omitempty"`
	Role       string `json:"role"`
	IDSantri   *int   `json:"id_santri,omitempty"`
	IDUstad    *int   `json:"id_ustad,omitempty"`
	IDOrangTua *int   `json:"id_orang_tua,omitempty"`
	IsActive   bool   `json:"is_active"`
}
