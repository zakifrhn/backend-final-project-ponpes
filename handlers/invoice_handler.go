package handlers

import (
	"backend-final-project-ponpes/models"
	"backend-final-project-ponpes/services"
	"backend-final-project-ponpes/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct {
	invoiceService services.InvoiceService
}

func NewInvoiceHandler(invoiceService services.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceService: invoiceService,
	}
}

// GetMyInvoices untuk santri melihat invoice mereka sendiri
func (h *InvoiceHandler) GetMyInvoices(c *gin.Context) {
	// Ambil user ID dari context (dari middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Convert userID ke int (asumsi user_id adalah ID santri)
	idSantri, err := strconv.Atoi(userID.(string))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID santri tidak valid")
		return
	}

	invoices, err := h.invoiceService.GetInvoicesBySantri(idSantri)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal mengambil data invoice")
		return
	}

	utils.SendSuccess(c, "Data invoice berhasil diambil", invoices)
}

// GetInvoiceDetail mengambil detail invoice
func (h *InvoiceHandler) GetInvoiceDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID invoice tidak valid")
		return
	}

	invoice, err := h.invoiceService.GetInvoiceByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			utils.SendError(c, http.StatusNotFound, "Invoice tidak ditemukan")
		} else {
			utils.SendError(c, http.StatusInternalServerError, "Gagal mengambil detail invoice")
		}
		return
	}

	utils.SendSuccess(c, "Detail invoice berhasil diambil", invoice)
}

// GetAllInvoices untuk admin/ustad melihat semua invoice
func (h *InvoiceHandler) GetAllInvoices(c *gin.Context) {
	var req models.GetAllInvoicesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	if req.StartDate != "" && !utils.IsValidDate(req.StartDate) {
		utils.SendError(c, http.StatusBadRequest, "Format start_date harus YYYY-MM-DD")
		return
	}

	if req.EndDate != "" && !utils.IsValidDate(req.EndDate) {
		utils.SendError(c, http.StatusBadRequest, "Format end_date harus YYYY-MM-DD")
		return
	}

	invoices, err := h.invoiceService.GetAllInvoices(req.StartDate, req.EndDate)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal mengambil data invoice")
		return
	}

	if len(invoices) == 0 {
		utils.SendError(c, http.StatusNotFound, "Data invoice tidak ditemukan")
		return
	}

	utils.SendSuccess(c, "Data invoice berhasil diambil", h.formatInvoices(invoices))
}

func (h *InvoiceHandler) formatInvoices(invoices []models.InvoiceDetail) []map[string]interface{} {
	response := make([]map[string]interface{}, 0, len(invoices))

	for _, inv := range invoices {
		item := map[string]interface{}{
			"id_invoice":       inv.IDInvoice,
			"deskripsi":        inv.Deskripsi,
			"deadline_tagihan": inv.DeadlineTagihan.Format("2006-01-02"),
			"id_santri":        inv.IDSantri,
			"nominal_tagihan":  inv.NominalTagihan,
			"status":           inv.Status,
			"created_date":     inv.CreatedDate.Format("2006-01-02"),
			"nis":              inv.NIS,
			"nama_santri":      inv.NamaSantri,
		}

		if !inv.UpdatedDate.IsZero() {
			item["updated_date"] = inv.UpdatedDate.Format("2006-01-02")
		}

		response = append(response, item)
	}

	return response
}

// GetInvoicesBySantri untuk admin/ustad melihat invoice santri tertentu
func (h *InvoiceHandler) GetInvoicesBySantri(c *gin.Context) {
	idSantriStr := c.Param("id_santri")
	idSantri, err := strconv.Atoi(idSantriStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID santri tidak valid")
		return
	}

	invoices, err := h.invoiceService.GetInvoicesBySantri(idSantri)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal mengambil data invoice")
		return
	}

	utils.SendSuccess(c, "Data invoice berhasil diambil", invoices)
}

// CreateInvoice untuk admin/ustad membuat invoice baru
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var req models.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Request tidak valid")
		return
	}

	// Get creator from context
	createdBy, exists := c.Get("username")
	if !exists {
		createdBy = "system"
	}

	invoiceID, err := h.invoiceService.CreateInvoice(req, createdBy.(string))
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			utils.SendError(c, http.StatusNotFound, "Santri tidak ditemukan")
		} else if strings.Contains(err.Error(), "tidak aktif") {
			utils.SendError(c, http.StatusBadRequest, "Santri tidak aktif")
		} else {
			utils.SendError(c, http.StatusInternalServerError, "Gagal membuat invoice")
		}
		return
	}

	utils.SendSuccess(c, "Invoice berhasil dibuat", gin.H{"id_invoice": invoiceID})
}

// UpdateInvoice untuk mengupdate invoice
func (h *InvoiceHandler) UpdateInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID invoice tidak valid")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Request tidak valid")
		return
	}

	updatedBy, exists := c.Get("username")
	if !exists {
		updatedBy = "system"
	}

	if err := h.invoiceService.UpdateInvoice(id, req, updatedBy.(string)); err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			utils.SendError(c, http.StatusNotFound, "Invoice tidak ditemukan")
		} else if strings.Contains(err.Error(), "tidak valid") {
			utils.SendError(c, http.StatusBadRequest, err.Error())
		} else {
			utils.SendError(c, http.StatusInternalServerError, "Gagal mengupdate invoice")
		}
		return
	}

	utils.SendSuccess(c, "Invoice berhasil diupdate", nil)
}

// UpdateInvoiceStatus khusus untuk update status
func (h *InvoiceHandler) UpdateInvoiceStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID invoice tidak valid")
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

	if err := h.invoiceService.UpdateInvoiceStatus(id, req.Status, updatedBy.(string)); err != nil {
		if strings.Contains(err.Error(), "tidak valid") {
			utils.SendError(c, http.StatusBadRequest, err.Error())
		} else {
			utils.SendError(c, http.StatusInternalServerError, "Gagal mengupdate status invoice")
		}
		return
	}

	utils.SendSuccess(c, "Status invoice berhasil diupdate", nil)
}

// DeleteInvoice soft delete invoice
func (h *InvoiceHandler) DeleteInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "ID invoice tidak valid")
		return
	}

	deletedBy, exists := c.Get("username")
	if !exists {
		deletedBy = "system"
	}

	if err := h.invoiceService.DeleteInvoice(id, deletedBy.(string)); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal menghapus invoice")
		return
	}

	utils.SendSuccess(c, "Invoice berhasil dihapus", nil)
}

// GenerateMonthlyInvoices untuk generate invoice bulanan otomatis
func (h *InvoiceHandler) GenerateMonthlyInvoices(c *gin.Context) {
	createdBy, exists := c.Get("username")
	if !exists {
		createdBy = "system"
	}

	count, err := h.invoiceService.GenerateMonthlyInvoices(createdBy.(string))
	if err != nil {
		if strings.Contains(err.Error(), "sudah ada") {
			utils.SendError(c, http.StatusBadRequest, err.Error())
		} else {
			utils.SendError(c, http.StatusInternalServerError, "Gagal generate invoice bulanan")
		}
		return
	}

	utils.SendSuccess(c, "Invoice bulanan berhasil digenerate",
		gin.H{"total_invoices": count})
}

// GetMyChildrenInvoices untuk orang tua melihat invoice anak-anaknya
func (h *InvoiceHandler) GetMyChildrenInvoices(c *gin.Context) {
	// Ambil user ID dari context (asumsi user_id adalah ID orang tua)
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

	invoices, err := h.invoiceService.GetInvoicesByOrangTua(idOrangTua)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Gagal mengambil data invoice anak")
		return
	}

	utils.SendSuccess(c, "Data invoice anak berhasil diambil", invoices)
}
