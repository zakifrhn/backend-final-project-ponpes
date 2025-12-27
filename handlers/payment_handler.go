package handlers

import (
	"backend-final-project-ponpes/models"
	"backend-final-project-ponpes/services"
	"backend-final-project-ponpes/utils"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService services.PaymentService
}

func NewPaymentHandler(paymentService services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePayment membuat pembayaran baru
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Request tidak valid: "+err.Error())
		return
	}

	// Get user info from context
	createdBy, exists := c.Get("username")
	if !exists {
		createdBy = "system"
	}
	req.CreatedBy = createdBy.(string)

	// Buat pembayaran
	paymentResp, err := h.paymentService.CreatePayment(req)
	if err != nil {
		errorMsg := err.Error()

		// Handle specific Midtrans errors
		if strings.Contains(errorMsg, "Autentikasi Midtrans gagal") ||
			strings.Contains(errorMsg, "Access denied") ||
			strings.Contains(errorMsg, "Unauthorized") {
			utils.SendError(c, http.StatusInternalServerError,
				"Error konfigurasi payment gateway. Silakan hubungi administrator.")
		} else if strings.Contains(errorMsg, "invoice sudah") {
			utils.SendError(c, http.StatusBadRequest, errorMsg)
		} else if strings.Contains(errorMsg, "jumlah bayar harus sama") {
			utils.SendError(c, http.StatusBadRequest, errorMsg)
		} else if strings.Contains(errorMsg, "tidak ditemukan") {
			utils.SendError(c, http.StatusNotFound, errorMsg)
		} else {
			utils.SendError(c, http.StatusInternalServerError,
				"Gagal membuat pembayaran: "+errorMsg)
		}
		return
	}

	utils.SendSuccess(c, "Pembayaran berhasil dibuat", paymentResp)
}

// CreatePaymentForInvoice membuat pembayaran untuk invoice tertentu
func (h *PaymentHandler) CreatePaymentForInvoice(c *gin.Context) {
	idStr := c.Param("id")
	idInvoice, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID invoice tidak valid")
		return
	}

	var req struct {
		MetodePembayaran string `json:"metode_pembayaran" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Request tidak valid")
		return
	}

	createdBy, exists := c.Get("username")
	if !exists {
		createdBy = "system"
	}

	paymentReq := models.CreatePaymentRequest{
		IDInvoice:        idInvoice,
		MetodePembayaran: req.MetodePembayaran,
		CreatedBy:        createdBy.(string),
	}

	paymentResp, err := h.paymentService.CreatePayment(paymentReq)
	if err != nil {
		if err.Error() == "invoice sudah LUNAS" {
			utils.SendError(c, http.StatusBadRequest, "Invoice sudah lunas")
		} else {
			utils.SendError(c, http.StatusInternalServerError, "Gagal membuat pembayaran")
		}
		return
	}

	utils.SendSuccess(c, "Pembayaran berhasil dibuat", paymentResp)
}

// HandlePaymentWebhook menerima webhook dari payment gateway (Midtrans)
func (h *PaymentHandler) HandlePaymentWebhook(c *gin.Context) {
	var webhookReq models.PaymentWebhookRequest

	if err := c.ShouldBindJSON(&webhookReq); err != nil {
		log.Printf("[WEBHOOK] Invalid webhook data: %v", err)
		utils.SendError(c, http.StatusBadRequest, "Webhook data tidak valid")
		return
	}

	log.Printf("[WEBHOOK] Processing webhook for order: %s", webhookReq.OrderID)

	// Proses webhook
	if err := h.paymentService.HandlePaymentWebhook(webhookReq); err != nil {
		log.Printf("[WEBHOOK] Error processing webhook: %v", err)
		utils.SendError(c, http.StatusInternalServerError, "Gagal memproses webhook")
		return
	}

	// Response untuk Midtrans
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Webhook berhasil diproses",
	})
}

// GetMyTransactions mendapatkan transaksi milik user (santri)
func (h *PaymentHandler) GetMyTransactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idSantri, err := strconv.Atoi(userID.(string))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID santri tidak valid")
		return
	}

	transactions, err := h.paymentService.GetTransactionHistory(idSantri)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal mengambil history transaksi")
		return
	}

	utils.SendSuccess(c, "History transaksi berhasil diambil", transactions)
}

// GetMyTransactionHistory alias untuk GetMyTransactions
func (h *PaymentHandler) GetMyTransactionHistory(c *gin.Context) {
	h.GetMyTransactions(c)
}

// GetAllTransactions untuk admin/ustad melihat semua transaksi
func (h *PaymentHandler) GetAllTransactions(c *gin.Context) {
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	midtransID := c.Query("midtrans_id")

	transactions, err := h.paymentService.GetAllTransactions(status, startDate, endDate)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal mengambil data transaksi")
		return
	}

	// Filter by Midtrans ID jika ada
	if midtransID != "" {
		filtered := []models.TransactionDetail{}
		for _, trans := range transactions {
			if trans.MidtransTransactionID != nil && *trans.MidtransTransactionID == midtransID {
				filtered = append(filtered, trans)
			}
		}
		transactions = filtered
	}

	utils.SendSuccess(c, "Data transaksi berhasil diambil", transactions)
}

// GetTransactionsBySantri untuk admin/ustad melihat transaksi santri tertentu
func (h *PaymentHandler) GetTransactionsBySantri(c *gin.Context) {
	idSantriStr := c.Param("id_santri")
	idSantri, err := strconv.Atoi(idSantriStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID santri tidak valid")
		return
	}

	transactions, err := h.paymentService.GetTransactionHistory(idSantri)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal mengambil data transaksi")
		return
	}

	utils.SendSuccess(c, "Data transaksi santri berhasil diambil", transactions)
}

// ValidatePayment validasi jumlah pembayaran
func (h *PaymentHandler) ValidatePayment(c *gin.Context) {
	idStr := c.Param("id_invoice")
	idInvoice, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID invoice tidak valid")
		return
	}

	amountStr := c.Query("amount")
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Jumlah tidak valid")
		return
	}

	isValid, err := h.paymentService.ValidatePaymentAmount(idInvoice, amountStr)
	if err != nil {
		if err.Error() == "jumlah bayar harus sama dengan nominal tagihan" {
			utils.SendError(c, http.StatusBadRequest, err.Error())
		} else {
			utils.SendError(c, http.StatusBadRequest, "Validasi gagal")
		}
		return
	}

	if !isValid {
		utils.SendError(c, http.StatusBadRequest, "Jumlah bayar tidak sesuai")
		return
	}

	utils.SendSuccess(c, "Jumlah bayar sesuai", gin.H{"valid": true})
}

// UpdateTransactionStatus mengupdate status transaksi
func (h *PaymentHandler) UpdateTransactionStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID transaksi tidak valid")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Request tidak valid")
		return
	}

	updatedBy, exists := c.Get("username")
	if !exists {
		updatedBy = "system"
	}

	if err := h.paymentService.UpdateTransactionStatus(id, req.Status, updatedBy.(string)); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal mengupdate status transaksi")
		return
	}

	utils.SendSuccess(c, "Status transaksi berhasil diupdate", nil)
}

// GetPaymentConfig mendapatkan konfigurasi payment gateway
func (h *PaymentHandler) GetPaymentConfig(c *gin.Context) {
	config, err := h.paymentService.GetPaymentConfig()
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal mengambil konfigurasi payment")
		return
	}

	utils.SendSuccess(c, "Konfigurasi payment berhasil diambil", config)
}

// UpdatePaymentConfig mengupdate konfigurasi payment gateway
func (h *PaymentHandler) UpdatePaymentConfig(c *gin.Context) {
	var req struct {
		MidtransServerKey string `json:"midtrans_server_key"`
		MidtransClientKey string `json:"midtrans_client_key"`
		IsProduction      bool   `json:"is_production"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Request tidak valid")
		return
	}

	updatedBy, exists := c.Get("username")
	if !exists {
		updatedBy = "system"
	}

	if err := h.paymentService.UpdatePaymentConfig(
		req.MidtransServerKey,
		req.MidtransClientKey,
		req.IsProduction,
		updatedBy.(string),
	); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal mengupdate konfigurasi payment")
		return
	}

	utils.SendSuccess(c, "Konfigurasi payment berhasil diupdate", nil)
}

// GetMyChildrenTransactions untuk orang tua melihat transaksi anak-anaknya
func (h *PaymentHandler) GetMyChildrenTransactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idOrangTua, err := strconv.Atoi(userID.(string))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID orang tua tidak valid")
		return
	}

	transactions, err := h.paymentService.GetTransactionsByOrangTua(idOrangTua)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal mengambil data transaksi anak")
		return
	}

	utils.SendSuccess(c, "Data transaksi anak berhasil diambil", transactions)
}
