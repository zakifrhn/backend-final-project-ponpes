package services

import (
	"backend-final-project-ponpes/models"
	"backend-final-project-ponpes/repositories"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(reqUser models.UserDTO) (string, error) {
	// Cek username apakah sudah ada
	_, err := s.userRepo.FindByUsername(reqUser.Username)

	if err == nil {
		// Username sudah ada
		return "", fmt.Errorf("username already exists")
	}

	// Error lain selain ErrNoRows
	if err != sql.ErrNoRows {
		return "", err
	}

	// Username belum ada → create user
	_, err = s.userRepo.CreateUser(&reqUser)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("user %s created successfully", reqUser.Username), nil
}

func (s *UserService) UpdatedUser(reqUser models.UserDTO) (string, error) {

	log.Printf("[SERVICE] Starting update for user ID: %d", reqUser.ID)

	// Cek apakah user exists
	existingUser, err := s.userRepo.FindByUserId(reqUser.ID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[SERVICE] User ID %d not found in database", reqUser.ID)
			return "", fmt.Errorf("user not found")
		}
		log.Printf("[SERVICE] Error while checking user: %v", err)
		return "", err
	}

	log.Printf("[SERVICE] User found: %+v", existingUser)

	// Eksekusi update
	msg, err := s.userRepo.UpdateUser(&reqUser)
	if err != nil {
		log.Printf("[SERVICE] Failed updating user ID %d: %v", reqUser.ID, err)
		return "", err
	}

	log.Printf("[SERVICE] Update success: %s", msg)

	return fmt.Sprintf("user id %d updated successfully", reqUser.ID), nil
}
