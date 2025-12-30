package utils

import (
	"regexp"
	"time"
)

func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func IsValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^[0-9]{10,15}$`)
	return phoneRegex.MatchString(phone)
}

func ValidatePaymentAmount(invoiceAmount, paymentAmount string) (bool, string) {
	if paymentAmount != invoiceAmount {
		return false, "Jumlah bayar harus sama dengan nominal tagihan"
	}
	return true, ""
}

func IsValidDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// Atau gunakan regex untuk validasi
func IsValidDateFormat(dateStr string) bool {
	if len(dateStr) != 10 {
		return false
	}

	// Pattern: YYYY-MM-DD
	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !re.MatchString(dateStr) {
		return false
	}

	// Cek apakah tanggal valid
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}
