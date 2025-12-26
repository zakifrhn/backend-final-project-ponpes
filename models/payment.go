package models

type CreatePaymentRequest struct {
	IDInvoice        int    `json:"id_invoice" binding:"required"`
	MetodePembayaran string `json:"metode_pembayaran" binding:"required"`
	CreatedBy        string `json:"created_by" binding:"required"`
}

type PaymentGatewayResponse struct {
	Status         bool   `json:"status"`
	Message        string `json:"message"`
	PaymentURL     string `json:"payment_url"`
	KodePembayaran string `json:"kode_pembayaran"`
	OrderID        string `json:"order_id"`
}

type PaymentWebhookRequest struct {
	OrderID           string  `json:"order_id"`
	TransactionID     string  `json:"transaction_id"`
	TransactionStatus string  `json:"transaction_status"`
	GrossAmount       float64 `json:"gross_amount"`
	PaymentType       string  `json:"payment_type"`
	TransactionTime   string  `json:"transaction_time"`
	FraudStatus       string  `json:"fraud_status"`
	StatusMessage     string  `json:"status_message"`
	StatusCode        string  `json:"status_code"`
	SignatureKey      string  `json:"signature_key"`
	MerchantID        string  `json:"merchant_id"`
}
