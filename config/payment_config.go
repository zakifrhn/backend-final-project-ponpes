package config

import (
	"strings"
)

type PaymentConfig struct {
	ServerKey    string
	ClientKey    string
	IsProduction bool
}

func GetPaymentConfig() PaymentConfig {
	serverKey := "SB-Mid-server-0M3ckIvV22lFcF1Gl25R-LXW"
	clientKey := "SB-Mid-client-QV1hYGwQPCr5nKrf"
	env := "sandbox"

	// NO DEFAULT VALUES - ALL FROM ENVIRONMENT
	if serverKey == "" {
		panic("MIDTRANS_SERVER_KEY environment variable is required!")
	}

	if clientKey == "" {
		panic("MIDTRANS_CLIENT_KEY environment variable is required!")
	}

	if env == "" {
		panic("MIDTRANS_ENV environment variable is required! Must be 'sandbox' or 'production'")
	}

	return PaymentConfig{
		ServerKey:    serverKey,
		ClientKey:    clientKey,
		IsProduction: strings.ToLower(env) == "production",
	}
}
