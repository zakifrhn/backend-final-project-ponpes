package utils

import (
	"regexp"
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
