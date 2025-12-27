package services

import (
	"backend-final-project-ponpes/models"
	"backend-final-project-ponpes/repositories"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type PaymentService interface {
	CreatePayment(req models.CreatePaymentRequest) (*models.PaymentGatewayResponse, error)
	HandlePaymentWebhook(webhookReq models.PaymentWebhookRequest) error
	GetTransactionHistory(idSantri int) ([]models.TransactionDetail, error)
	GetAllTransactions(status, startDate, endDate string) ([]models.TransactionDetail, error)
	GetTransactionsByOrangTua(idOrangTua int) ([]models.TransactionDetail, error)
	ValidatePaymentAmount(invoiceID int, jumlahBayar string) (bool, error)
	UpdateTransactionStatus(id int, status, updatedBy string) error
	GetPaymentConfig() (map[string]interface{}, error)
	UpdatePaymentConfig(serverKey, clientKey string, isProduction bool, updatedBy string) error
	VerifyMidtransSignature(signatureKey, orderID, statusCode, grossAmount string) bool
}

type paymentService struct {
	db           *sql.DB
	invoiceRepo  repositories.InvoiceRepository
	transRepo    repositories.TransactionRepository
	santriRepo   repositories.SantriRepository
	serverKey    string
	clientKey    string
	isProduction bool
}

func NewPaymentService(
	db *sql.DB,
	invoiceRepo repositories.InvoiceRepository,
	transRepo repositories.TransactionRepository,
	santriRepo repositories.SantriRepository,
	serverKey, clientKey string,
	isProduction bool,
) PaymentService {

	// Jika serverKey atau clientKey kosong, log warning
	if serverKey == "" || clientKey == "" {
		log.Printf("[PAYMENT] WARNING: Server Key or Client Key is empty. Midtrans transactions will fail.")
	}

	log.Printf("[PAYMENT] Initializing REAL Midtrans Payment Service")
	log.Printf("[PAYMENT] Server Key: %s...", safeSubstring(serverKey, 20))
	log.Printf("[PAYMENT] Environment: %s", map[bool]string{true: "PRODUCTION", false: "SANDBOX"}[isProduction])

	return &paymentService{
		db:           db,
		invoiceRepo:  invoiceRepo,
		transRepo:    transRepo,
		santriRepo:   santriRepo,
		serverKey:    serverKey,
		clientKey:    clientKey,
		isProduction: isProduction,
	}
}

func safeSubstring(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func (s *paymentService) CreatePayment(req models.CreatePaymentRequest) (*models.PaymentGatewayResponse, error) {
	log.Printf("[PAYMENT] Starting payment creation for invoice %d", req.IDInvoice)

	// 1. Validasi invoice
	invoice, err := s.invoiceRepo.GetInvoiceByID(req.IDInvoice)
	if err != nil {
		log.Printf("[PAYMENT] Error getting invoice: %v", err)
		return nil, fmt.Errorf("invoice tidak ditemukan: %v", err)
	}

	log.Printf("[PAYMENT] Invoice found: ID=%d, Status=%s, Amount=%.2f",
		invoice.IDInvoice, invoice.Status, invoice.NominalTagihan)

	// 2. Validasi status invoice
	if invoice.Status != "BELUM_LUNAS" {
		errMsg := fmt.Sprintf("invoice sudah %s", invoice.Status)
		log.Printf("[PAYMENT] Invalid invoice status: %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 3. Generate kode pembayaran unik
	kodePembayaran := fmt.Sprintf("SPP-%d-%s", invoice.IDInvoice,
		time.Now().Format("20060102150405"))

	log.Printf("[PAYMENT] Generated payment code: %s", kodePembayaran)

	// 4. Mulai transaksi database
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("[PAYMENT] Failed to start DB transaction: %v", err)
		return nil, fmt.Errorf("gagal memulai transaksi: %v", err)
	}
	defer tx.Rollback()

	// 5. Buat data transaksi
	transaction := &models.Transaction{
		IDInvoice:        invoice.IDInvoice,
		NominalTagihan:   invoice.NominalTagihan,
		JumlahBayar:      invoice.NominalTagihan,
		Status:           "MENUNGGU",
		MetodePembayaran: &req.MetodePembayaran,
		KodePembayaran:   kodePembayaran,
		CreatedBy:        &req.CreatedBy,
	}

	// 6. Simpan transaksi ke database
	transID, err := s.transRepo.CreateTransaction(transaction)
	if err != nil {
		log.Printf("[PAYMENT] Failed to save transaction: %v", err)
		return nil, fmt.Errorf("gagal menyimpan transaksi: %v", err)
	}

	log.Printf("[PAYMENT] Transaction saved with ID: %d", transID)

	// 7. Buat pembayaran di Midtrans
	log.Println("[PAYMENT] Creating Midtrans transaction...")
	paymentURL, err := s.createPaymentGatewayTransaction(transaction, invoice)
	if err != nil {
		log.Printf("[PAYMENT] Midtrans error: %v", err)

		// Hapus transaksi yang sudah dibuat karena Midtrans gagal
		deleteQuery := `DELETE FROM tr_transaksi_spp WHERE id_transaksi = $1`
		if _, delErr := tx.Exec(deleteQuery, transID); delErr != nil {
			log.Printf("[PAYMENT] Failed to delete transaction: %v", delErr)
		}

		// Return error yang lebih informatif
		if strings.Contains(err.Error(), "Access denied") || strings.Contains(err.Error(), "Unauthorized") {
			return nil, fmt.Errorf("kesalahan autentikasi Midtrans. Periksa Server Key Anda")
		}
		return nil, fmt.Errorf("gagal membuat pembayaran di gateway: %v", err)
	}

	// 8. Update transaksi dengan payment URL
	updateQuery := `UPDATE tr_transaksi_spp SET payment_url = $1 WHERE id_transaksi = $2`
	_, err = tx.Exec(updateQuery, paymentURL, transID)
	if err != nil {
		log.Printf("[PAYMENT] Failed to update payment URL: %v", err)
		return nil, fmt.Errorf("gagal update payment URL: %v", err)
	}

	// 9. Commit transaksi
	if err = tx.Commit(); err != nil {
		log.Printf("[PAYMENT] Failed to commit transaction: %v", err)
		return nil, fmt.Errorf("gagal commit transaksi: %v", err)
	}

	log.Printf("[PAYMENT] Payment created successfully. URL: %s", paymentURL)

	// 10. Response
	response := &models.PaymentGatewayResponse{
		Status:         true,
		Message:        "Pembayaran berhasil dibuat",
		PaymentURL:     paymentURL,
		KodePembayaran: kodePembayaran,
		OrderID:        kodePembayaran,
	}

	return response, nil
}

func (s *paymentService) createPaymentGatewayTransaction(trans *models.Transaction, invoice *models.InvoiceDetail) (string, error) {
	log.Printf("[MIDTRANS] Creating transaction for order: %s", trans.KodePembayaran)

	// Validasi server key
	if s.serverKey == "" {
		return "", fmt.Errorf("server key tidak ditemukan")
	}

	if strings.Contains(s.serverKey, "SB-Mid-server-0M3ckIvV22lFcF1Gl25R-LXW") || !strings.HasPrefix(s.serverKey, "SB-Mid-server-") {
		log.Printf("[MIDTRANS] WARNING: Server key mungkin tidak valid: %s...", s.serverKey[:20])
	}

	// Get santri data
	santri, err := s.santriRepo.GetSantriByID(invoice.IDSantri)
	if err != nil {
		log.Printf("[MIDTRANS] Error getting santri: %v", err)
		return "", fmt.Errorf("gagal mengambil data santri: %v", err)
	}

	if santri == nil {
		log.Printf("[MIDTRANS] Santri not found for ID: %d", invoice.IDSantri)
		return "", fmt.Errorf("data santri tidak ditemukan")
	}

	// Prepare customer details
	customerEmail := "santri@ponpes-alaziziyah.com"
	if santri.Email != nil && *santri.Email != "" {
		customerEmail = *santri.Email
	}

	customerPhone := "081234567890"
	if santri.NoTelepon != nil && *santri.NoTelepon != "" {
		customerPhone = *santri.NoTelepon
	}

	// API URL berdasarkan environment
	baseURL := "https://api.sandbox.midtrans.com"
	if s.isProduction {
		baseURL = "https://api.midtrans.com"
	}

	log.Printf("[MIDTRANS] Using base URL: %s", baseURL)

	// Prepare payload
	payload := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     trans.KodePembayaran,
			"gross_amount": invoice.NominalTagihan,
		},
		"customer_details": map[string]interface{}{
			"first_name": santri.NamaLengkap,
			"email":      customerEmail,
			"phone":      customerPhone,
		},
		"item_details": []map[string]interface{}{
			{
				"id":       fmt.Sprintf("INV-%d", invoice.IDInvoice),
				"price":    invoice.NominalTagihan,
				"quantity": 1,
				"name":     invoice.Deskripsi,
			},
		},
		"callbacks": map[string]interface{}{
			"finish":  "https://your-frontend-url.com/payment/finish",
			"error":   "https://your-frontend-url.com/payment/error",
			"pending": "https://your-frontend-url.com/payment/pending",
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[MIDTRANS] JSON marshal error: %v", err)
		return "", err
	}

	log.Printf("[MIDTRANS] Request payload: %s", string(jsonData))

	// Create HTTP request
	req, err := http.NewRequest("POST", baseURL+"/snap/v1/transactions",
		strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("[MIDTRANS] HTTP request creation error: %v", err)
		return "", err
	}

	// Set headers dengan benar
	auth := base64.StdEncoding.EncodeToString([]byte(s.serverKey + ":"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+auth)

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[MIDTRANS] HTTP request error: %v", err)
		return "", fmt.Errorf("gagal terhubung ke Midtrans: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[MIDTRANS] Error reading response body: %v", err)
		return "", err
	}

	log.Printf("[MIDTRANS] Response status: %d", resp.StatusCode)
	log.Printf("[MIDTRANS] Response body: %s", string(body))

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[MIDTRANS] JSON unmarshal error: %v", err)
		return "", fmt.Errorf("gagal memproses respons Midtrans: %v", err)
	}

	// Check response
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		errorMessage := "Autentikasi Midtrans gagal. Periksa Server Key Anda"
		if errorField, ok := result["error_messages"].([]interface{}); ok && len(errorField) > 0 {
			errorMessage = fmt.Sprintf("Midtrans error: %v", errorField[0])
		}
		log.Printf("[MIDTRANS] Authentication Error: %s", errorMessage)
		return "", fmt.Errorf(errorMessage)
	}

	if resp.StatusCode != 201 {
		errorMessage := fmt.Sprintf("Midtrans API error (status: %d)", resp.StatusCode)
		if errorField, ok := result["error_messages"].([]interface{}); ok && len(errorField) > 0 {
			errorMessage = fmt.Sprintf("Midtrans error: %v", errorField[0])
		} else if msg, ok := result["message"].(string); ok {
			errorMessage = msg
		}
		log.Printf("[MIDTRANS] API Error: %s", errorMessage)
		return "", fmt.Errorf(errorMessage)
	}

	// Get payment URL
	paymentURL, ok := result["redirect_url"].(string)
	if !ok {
		log.Printf("[MIDTRANS] No redirect_url in response: %v", result)
		return "", fmt.Errorf("redirect_url tidak ditemukan dalam response")
	}

	log.Printf("[MIDTRANS] Payment URL generated successfully")
	return paymentURL, nil
}

func (s *paymentService) VerifyMidtransSignature(signatureKey, orderID, statusCode, grossAmount string) bool {
	// Generate signature untuk verifikasi
	input := orderID + statusCode + grossAmount + s.serverKey
	expectedSignature := base64.StdEncoding.EncodeToString([]byte(input))

	// Bandingkan signature
	return signatureKey == expectedSignature
}

func (s *paymentService) HandlePaymentWebhook(webhookReq models.PaymentWebhookRequest) error {
	log.Printf("[WEBHOOK] Received webhook: %+v", webhookReq)

	// 1. Verifikasi signature (opsional tapi direkomendasikan)
	if !s.VerifyMidtransSignature(webhookReq.SignatureKey, webhookReq.OrderID,
		webhookReq.StatusCode, fmt.Sprintf("%.0f", webhookReq.GrossAmount)) {
		log.Printf("[WEBHOOK] WARNING: Signature verification failed")
		// Anda bisa memilih untuk tetap melanjutkan atau return error
	}

	// 2. Cari transaksi berdasarkan order_id
	transaksi, err := s.transRepo.GetTransactionByKode(webhookReq.OrderID)
	if err != nil {
		log.Printf("[WEBHOOK] Error getting transaction with kode %s: %v", webhookReq.OrderID, err)
		return fmt.Errorf("transaksi tidak ditemukan: %v", err)
	}
	log.Printf("[WEBHOOK] Found transaction: %+v", transaksi)

	// 3. Update status transaksi
	var transStatus string
	var invoiceStatus string

	switch webhookReq.TransactionStatus {
	case "capture", "settlement":
		if webhookReq.FraudStatus == "accept" {
			transStatus = "BERHASIL"
			invoiceStatus = "LUNAS"
		} else if webhookReq.FraudStatus == "challenge" {
			transStatus = "DIPROSES"
			invoiceStatus = "BELUM_LUNAS"
		}
	case "pending":
		transStatus = "MENUNGGU"
		invoiceStatus = "BELUM_LUNAS"
	case "deny", "cancel", "expire":
		transStatus = "GAGAL"
		invoiceStatus = "BELUM_LUNAS"
	case "refund", "partial_refund":
		transStatus = "DIREFUND"
		invoiceStatus = "BELUM_LUNAS"
	default:
		transStatus = "MENUNGGU"
		invoiceStatus = "BELUM_LUNAS"
	}

	log.Printf("[WEBHOOK] Transaction Status: %s, Fraud Status: %s, mapping to: %s (trans), %s (invoice)",
		webhookReq.TransactionStatus, webhookReq.FraudStatus, transStatus, invoiceStatus)

	// 4. Update transaksi dengan status dan jumlah bayar dari Midtrans
	if err := s.transRepo.UpdateTransactionFromWebhook(
		webhookReq.OrderID,
		transStatus,
		webhookReq.GrossAmount,
		webhookReq.TransactionID,
		"system_webhook",
	); err != nil {
		log.Printf("[WEBHOOK] Error updating transaction: %v", err)
		return fmt.Errorf("gagal update status transaksi: %v", err)
	}
	log.Printf("[WEBHOOK] Transaction updated with status %s and Midtrans ID %s",
		transStatus, webhookReq.TransactionID)

	// 5. Update status invoice jika pembayaran berhasil
	if transStatus == "BERHASIL" {
		if err := s.invoiceRepo.UpdateInvoiceStatus(
			transaksi.IDInvoice,
			invoiceStatus,
			"system_webhook",
		); err != nil {
			log.Printf("[WEBHOOK] Error updating invoice status: %v", err)
			return fmt.Errorf("gagal update status invoice: %v", err)
		}
		log.Printf("[WEBHOOK] Invoice status updated to %s", invoiceStatus)

		// Update transaction date dengan waktu dari Midtrans jika tersedia
		if webhookReq.TransactionTime != "" {
			transactionTime, err := time.Parse("2006-01-02 15:04:05", webhookReq.TransactionTime)
			if err == nil {
				updateTimeQuery := `UPDATE tr_transaksi_spp 
					SET transaction_date = $1 
					WHERE kode_pembayaran = $2`
				_, err = s.db.Exec(updateTimeQuery, transactionTime, webhookReq.OrderID)
				if err != nil {
					log.Printf("[WEBHOOK] Error updating transaction time: %v", err)
				}
			}
		}
	}

	// 6. Log untuk monitoring
	log.Printf("[WEBHOOK] Webhook processed successfully for order %s", webhookReq.OrderID)

	return nil
}

func (s *paymentService) GetTransactionHistory(idSantri int) ([]models.TransactionDetail, error) {
	return s.transRepo.GetTransactionHistory(idSantri)
}

func (s *paymentService) GetAllTransactions(status, startDate, endDate string) ([]models.TransactionDetail, error) {
	return s.transRepo.GetAllTransactions(status, startDate, endDate)
}

func (s *paymentService) GetTransactionsByOrangTua(idOrangTua int) ([]models.TransactionDetail, error) {
	return s.transRepo.GetTransactionsByOrangTua(idOrangTua)
}

func (s *paymentService) ValidatePaymentAmount(invoiceID int, jumlahBayar string) (bool, error) {
	invoice, err := s.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		return false, err
	}

	if jumlahBayar != invoice.NominalTagihan {
		return false, fmt.Errorf("jumlah bayar harus sama dengan nominal tagihan")
	}

	return true, nil
}

func (s *paymentService) UpdateTransactionStatus(id int, status, updatedBy string) error {
	trans, err := s.transRepo.GetTransactionByID(id)
	if err != nil {
		return err
	}

	validStatus := map[string]bool{
		"MENUNGGU":   true,
		"BERHASIL":   true,
		"GAGAL":      true,
		"KADALUARSA": true,
		"DIBATALKAN": true,
		"DIPROSES":   true,
		"DIREFUND":   true,
	}

	if !validStatus[status] {
		return fmt.Errorf("status tidak valid")
	}

	return s.transRepo.UpdateTransactionStatus(trans.KodePembayaran, status, updatedBy)
}

func (s *paymentService) GetPaymentConfig() (map[string]interface{}, error) {
	config := map[string]interface{}{
		"midtrans_server_key": s.serverKey,
		"midtrans_client_key": s.clientKey,
		"is_production":       s.isProduction,
		"webhook_url":         "https://enormous-especially-sawfish.ngrok-free.app/api/v1/payment/webhook",
	}
	return config, nil
}

func (s *paymentService) UpdatePaymentConfig(serverKey, clientKey string, isProduction bool, updatedBy string) error {
	// Simpan config ke database atau environment
	// Untuk sementara, kita update di memory saja
	s.serverKey = serverKey
	s.clientKey = clientKey
	s.isProduction = isProduction
	return nil
}
