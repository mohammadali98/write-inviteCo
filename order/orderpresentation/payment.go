package orderpresentation

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"writeandinviteco/inviteandco/order/orderapplication"
	"writeandinviteco/inviteandco/order/orderdomain"
	"writeandinviteco/inviteandco/webui"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

const (
	maxPaymentProofFileSize = 10 << 20
)

var (
	errInvalidPaymentProofFile        = errors.New("invalid payment proof file")
	errPaymentProofStorageUnavailable = errors.New("payment proof storage not configured")
	allowedPaymentProofTypes          = map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
		".pdf":  "application/pdf",
	}
	detectedPaymentProofTypes = map[string]string{
		"image/jpeg":      ".jpg",
		"image/png":       ".png",
		"image/webp":      ".webp",
		"application/pdf": ".pdf",
	}
)

type savedPaymentProof struct {
	SecureURL string
	PublicID  string
}

func (h *OrderHandler) BankTransferPage(c *gin.Context) {
	token, err := parsePublicToken(c.Param("token"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please use a valid order tracking link.")
		return
	}

	payload, err := h.service.GetOrderStatusDetailByToken(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, orderapplication.ErrInvalidInput) {
			webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please use a valid order tracking link.")
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that tracking link.")
			return
		}

		log.Println("BANK TRANSFER PAGE ERROR:", err)
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the payment instructions right now.")
		return
	}

	c.HTML(http.StatusOK, "order-payment.html", gin.H{
		"order":                payload.Order,
		"customer":             payload.Customer,
		"payment":              payload.Payment,
		"bankDetails":          orderapplication.BankTransferInstructions(),
		"csrfToken":            webui.EnsureCSRFToken(c),
		"proofSubmitted":       c.Query("proof_submitted") == "1",
		"canUploadProof":       canUploadPaymentProof(payload.Payment),
		"paymentMessage":       orderapplication.PaymentStatusMessage(payload.Payment, payload.Order.TotalPrice),
		"amountSummary":        paymentAmountSummary(payload.Order, payload.Payment),
		"paymentStatusDisplay": paymentStatusDisplay(payload.Payment),
		"originalAmount":       payload.OriginalAmount,
		"insertAmount":         payload.InsertAmount,
		"discountAmount":       payload.DiscountAmount,
	})
}

func (h *OrderHandler) SubmitPaymentProof(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please refresh the page and try uploading your payment proof again.")
		return
	}

	token, err := parsePublicToken(c.Param("token"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please use a valid order tracking link.")
		return
	}

	payload, err := h.service.GetOrderStatusDetailByToken(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, orderapplication.ErrInvalidInput) {
			webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please use a valid order tracking link.")
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that tracking link.")
			return
		}

		log.Println("PAYMENT PROOF LOAD ERROR:", err)
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the payment record right now.")
		return
	}

	if !canUploadPaymentProof(payload.Payment) {
		webui.RenderError(c, http.StatusBadRequest, "Upload Not Allowed", "This order is not currently accepting a new payment proof upload.")
		return
	}

	proof, err := h.savePaymentProof(c, "payment_proof", payload.Order.ID)
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Proof Upload", "Upload a JPG, JPEG, PNG, WEBP, or PDF file up to 10 MB.")
		return
	}

	oldProofPath := ""
	if payload.Payment != nil && payload.Payment.ProofFilePath != nil {
		oldProofPath = *payload.Payment.ProofFilePath
	}

	err = h.service.SubmitBankTransferProof(c.Request.Context(), payload.Order.ID, orderapplication.PaymentProofInput{
		SenderName:           c.PostForm("sender_name"),
		TransactionReference: c.PostForm("transaction_reference"),
		CustomerNote:         c.PostForm("payment_note"),
		ProofFilePath:        proof.SecureURL,
	})
	if err != nil {
		h.proofUploader.destroy(proof.PublicID)
		switch {
		case errors.Is(err, orderapplication.ErrInvalidInput):
			webui.RenderError(c, http.StatusBadRequest, "Invalid Payment Details", "Please review the submitted payment proof details and try again.")
		case errors.Is(err, orderapplication.ErrPaymentActionNotAllowed):
			webui.RenderError(c, http.StatusBadRequest, "Upload Not Allowed", "This order is not currently accepting a new payment proof upload.")
		case errors.Is(err, pgx.ErrNoRows):
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that tracking link.")
		default:
			log.Println("PAYMENT PROOF SUBMISSION ERROR:", err)
			webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not save your payment proof right now.")
		}
		return
	}

	if oldProofPath != "" && oldProofPath != proof.SecureURL && !isCloudinaryProofURL(oldProofPath) {
		removeUploadedPaymentProof(h.paymentProofDir, oldProofPath)
	}

	c.Redirect(http.StatusSeeOther, "/order/"+token+"/payment?proof_submitted=1")
}

func (h *OrderHandler) AdminPaymentAction(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please refresh the page and try the payment action again.")
		return
	}

	orderID, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
		return
	}

	action := strings.ToLower(strings.TrimSpace(c.PostForm("action")))

	if err := h.service.AdminProcessPayment(c.Request.Context(), orderID, action, c.PostForm("admin_note")); err != nil {
		switch {
		case errors.Is(err, orderapplication.ErrInvalidInput):
			c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "payment_error=invalid_action"))
		case errors.Is(err, orderapplication.ErrPaymentAmountTooLow):
			c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "payment_error=amount_too_low"))
		case errors.Is(err, orderapplication.ErrPaymentActionNotAllowed):
			c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "payment_error=action_not_allowed"))
		case errors.Is(err, pgx.ErrNoRows):
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that number.")
		default:
			log.Println("ADMIN PAYMENT ACTION ERROR:", err)
			webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not update the payment right now.")
		}
		return
	}

	switch action {
	case "verify":
		c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "payment_notice=payment_verified"))
	case "request_reupload":
		c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "payment_notice=payment_reupload_requested"))
	default:
		c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "payment_notice=payment_rejected"))
	}
}

func (h *OrderHandler) SubmitFinalPaymentProof(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please refresh the page and try uploading your final payment proof again.")
		return
	}

	token, err := parsePublicToken(c.Param("token"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please use a valid order tracking link.")
		return
	}

	payload, err := h.service.GetOrderStatusDetailByToken(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, orderapplication.ErrInvalidInput) {
			webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please use a valid order tracking link.")
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that tracking link.")
			return
		}

		log.Println("FINAL PAYMENT PROOF LOAD ERROR:", err)
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the payment record right now.")
		return
	}

	if !canUploadFinalPaymentProof(payload.Order) {
		webui.RenderError(c, http.StatusBadRequest, "Upload Not Allowed", "This order is not currently accepting a final payment proof upload.")
		return
	}

	proof, err := h.saveFinalPaymentProof(c, "final_payment_proof", payload.Order.ID)
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Proof Upload", "Upload a JPG, JPEG, PNG, WEBP, or PDF file up to 10 MB.")
		return
	}

	err = h.service.SubmitFinalPaymentProof(c.Request.Context(), payload.Order.ID, orderapplication.FinalPaymentProofInput{
		SenderName:    c.PostForm("sender_name"),
		ProofFilePath: proof.SecureURL,
	})
	if err != nil {
		h.proofUploader.destroy(proof.PublicID)
		switch {
		case errors.Is(err, orderapplication.ErrInvalidInput):
			webui.RenderError(c, http.StatusBadRequest, "Invalid Payment Details", "Please review the submitted payment proof details and try again.")
		case errors.Is(err, orderapplication.ErrPaymentActionNotAllowed):
			webui.RenderError(c, http.StatusBadRequest, "Upload Not Allowed", "This order is not currently accepting a final payment proof upload.")
		case errors.Is(err, pgx.ErrNoRows):
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that tracking link.")
		default:
			log.Println("FINAL PAYMENT PROOF SUBMISSION ERROR:", err)
			webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not save your payment proof right now.")
		}
		return
	}

	c.Redirect(http.StatusSeeOther, "/order/"+token+"/final-payment?proof_submitted=1")
}

func (h *OrderHandler) AdminFinalPaymentAction(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please refresh the page and try the payment action again.")
		return
	}

	orderID, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
		return
	}

	action := strings.ToLower(strings.TrimSpace(c.PostForm("action")))

	if err := h.service.AdminProcessFinalPayment(c.Request.Context(), orderID, action, c.PostForm("admin_note")); err != nil {
		switch {
		case errors.Is(err, orderapplication.ErrInvalidInput):
			c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "final_payment_error=invalid_action"))
		case errors.Is(err, orderapplication.ErrPaymentActionNotAllowed):
			c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "final_payment_error=action_not_allowed"))
		case errors.Is(err, pgx.ErrNoRows):
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that number.")
		default:
			log.Println("ADMIN FINAL PAYMENT ACTION ERROR:", err)
			webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not update the payment right now.")
		}
		return
	}

	switch action {
	case "verify":
		c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "final_payment_notice=payment_verified"))
	case "request_reupload":
		c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "final_payment_notice=payment_reupload_requested"))
	default:
		c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "final_payment_notice=payment_rejected"))
	}
}

func (h *OrderHandler) AdminServeFinalPaymentProof(c *gin.Context) {
	orderID, err := parsePositiveInt64(c.Param("orderID"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
		return
	}

	payload, err := h.service.GetAdminOrderDetail(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that number.")
			return
		}
		log.Println("ADMIN FINAL PAYMENT PROOF LOAD ERROR:", err)
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the payment record right now.")
		return
	}

	if payload.Order == nil || payload.Order.FinalPaymentProofURL == nil || strings.TrimSpace(*payload.Order.FinalPaymentProofURL) == "" {
		webui.RenderError(c, http.StatusNotFound, "No Proof Uploaded", "No final payment proof has been uploaded for this order.")
		return
	}

	proofPath := strings.TrimSpace(*payload.Order.FinalPaymentProofURL)

	if !isCloudinaryProofURL(proofPath) {
		webui.RenderError(c, http.StatusNotFound, "No Proof Uploaded", "No final payment proof has been uploaded for this order.")
		return
	}

	c.Redirect(http.StatusFound, proofPath)
}

func (h *OrderHandler) AdminServePaymentProof(c *gin.Context) {
	orderID, err := parsePositiveInt64(c.Param("orderID"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
		return
	}

	payload, err := h.service.GetAdminOrderDetail(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that number.")
			return
		}
		log.Println("ADMIN PAYMENT PROOF LOAD ERROR:", err)
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the payment record right now.")
		return
	}

	if payload.Payment == nil || payload.Payment.ProofFilePath == nil || strings.TrimSpace(*payload.Payment.ProofFilePath) == "" {
		webui.RenderError(c, http.StatusNotFound, "No Proof Uploaded", "No payment proof has been uploaded for this order.")
		return
	}

	proofPath := strings.TrimSpace(*payload.Payment.ProofFilePath)

	if isCloudinaryProofURL(proofPath) {
		c.Redirect(http.StatusFound, proofPath)
		return
	}

	// Legacy local proof from before the Cloudinary migration (Railway's disk
	// is ephemeral, so no new proofs are saved here — this path only serves
	// pre-migration test orders that were never migrated).
	filePath, ok := paymentProofFilePath(h.paymentProofDir, proofPath)
	if !ok {
		webui.RenderError(c, http.StatusNotFound, "No Proof Uploaded", "No payment proof has been uploaded for this order.")
		return
	}

	if _, err := os.Stat(filePath); err != nil {
		webui.RenderError(c, http.StatusNotFound, "No Proof Uploaded", "No payment proof has been uploaded for this order.")
		return
	}

	c.File(filePath)
}

func isCloudinaryProofURL(raw string) bool {
	return strings.HasPrefix(strings.TrimSpace(raw), "https://res.cloudinary.com/")
}

func canUploadPaymentProof(payment *orderdomain.OrderPayment) bool {
	if payment == nil {
		return false
	}
	return payment.PaymentStatus == orderdomain.PendingPaymentStatus ||
		payment.PaymentStatus == orderdomain.RejectedPaymentStatus
}

func (h *OrderHandler) savePaymentProof(c *gin.Context, field string, orderID int64) (*savedPaymentProof, error) {
	fileHeader, err := c.FormFile(field)
	if err != nil {
		return nil, errInvalidPaymentProofFile
	}

	canonicalExt, err := webui.ValidateUploadedFile(fileHeader, maxPaymentProofFileSize, allowedPaymentProofTypes, detectedPaymentProofTypes)
	if err != nil {
		return nil, errInvalidPaymentProofFile
	}

	if !h.proofUploader.configured() {
		return nil, errPaymentProofStorageUnavailable
	}

	filename, err := newPaymentProofFilename(canonicalExt)
	if err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	folder := fmt.Sprintf("payment-proofs/%d", orderID)
	secureURL, publicID, err := h.proofUploader.upload(file, filename, folder)
	if err != nil {
		return nil, err
	}

	return &savedPaymentProof{
		SecureURL: secureURL,
		PublicID:  publicID,
	}, nil
}

func (h *OrderHandler) saveFinalPaymentProof(c *gin.Context, field string, orderID int64) (*savedPaymentProof, error) {
	fileHeader, err := c.FormFile(field)
	if err != nil {
		return nil, errInvalidPaymentProofFile
	}

	canonicalExt, err := webui.ValidateUploadedFile(fileHeader, maxPaymentProofFileSize, allowedPaymentProofTypes, detectedPaymentProofTypes)
	if err != nil {
		return nil, errInvalidPaymentProofFile
	}

	if !h.proofUploader.configured() {
		return nil, errPaymentProofStorageUnavailable
	}

	filename, err := newPaymentProofFilename(canonicalExt)
	if err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	folder := fmt.Sprintf("final-payment-proofs/%d", orderID)
	secureURL, publicID, err := h.proofUploader.upload(file, filename, folder)
	if err != nil {
		return nil, err
	}

	return &savedPaymentProof{
		SecureURL: secureURL,
		PublicID:  publicID,
	}, nil
}

func newPaymentProofFilename(ext string) (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	buffer[6] = (buffer[6] & 0x0f) | 0x40
	buffer[8] = (buffer[8] & 0x3f) | 0x80

	encoded := hex.EncodeToString(buffer)
	uuid := encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
	return uuid + strings.ToLower(ext), nil
}

func removeUploadedPaymentProof(uploadDir string, filename string) {
	savePath, ok := paymentProofFilePath(uploadDir, filename)
	if !ok {
		return
	}
	_ = os.Remove(savePath)
}

func paymentProofFilePath(uploadDir string, filename string) (string, bool) {
	name := strings.TrimSpace(filename)
	if name == "" || strings.Contains(name, "..") || strings.ContainsAny(name, "/\\") {
		return "", false
	}

	return filepath.Join(uploadDir, name), true
}
