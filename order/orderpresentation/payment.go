package orderpresentation

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"writeandinviteco/inviteandco/order/orderapplication"
	"writeandinviteco/inviteandco/order/orderdomain"
	"writeandinviteco/inviteandco/webui"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

const (
	maxPaymentProofFileSize = 10 << 20
	paymentProofPathPrefix  = "/static/payment-proofs/"
)

var (
	errInvalidPaymentProofFile = errors.New("invalid payment proof file")
	allowedPaymentProofTypes   = map[string]string{
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
	PublicPath string
	SavedPath  string
}

func (h *OrderHandler) BankTransferPage(c *gin.Context) {
	orderID, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
		return
	}

	payload, err := h.service.GetOrderStatusDetail(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, orderapplication.ErrInvalidInput) {
			webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that number.")
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
	})
}

func (h *OrderHandler) SubmitPaymentProof(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please refresh the page and try uploading your payment proof again.")
		return
	}

	orderID, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
		return
	}

	payload, err := h.service.GetOrderStatusDetail(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, orderapplication.ErrInvalidInput) {
			webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that number.")
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

	submittedAmount, err := parsePositiveInt64(c.PostForm("submitted_amount"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Amount", "Please enter the advance amount you transferred as a valid number.")
		return
	}

	proof, err := h.savePaymentProof(c, "payment_proof")
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Proof Upload", "Upload a JPG, JPEG, PNG, WEBP, or PDF file up to 10 MB.")
		return
	}

	oldProofPath := ""
	if payload.Payment != nil && payload.Payment.ProofFilePath != nil {
		oldProofPath = *payload.Payment.ProofFilePath
	}

	err = h.service.SubmitBankTransferProof(c.Request.Context(), orderID, orderapplication.PaymentProofInput{
		SubmittedAmount:      submittedAmount,
		SenderName:           c.PostForm("sender_name"),
		TransactionReference: c.PostForm("transaction_reference"),
		CustomerNote:         c.PostForm("payment_note"),
		ProofFilePath:        proof.PublicPath,
	})
	if err != nil {
		_ = os.Remove(proof.SavedPath)
		switch {
		case errors.Is(err, orderapplication.ErrInvalidInput):
			webui.RenderError(c, http.StatusBadRequest, "Invalid Payment Details", "Please review the submitted payment proof details and try again.")
		case errors.Is(err, orderapplication.ErrPaymentActionNotAllowed):
			webui.RenderError(c, http.StatusBadRequest, "Upload Not Allowed", "This order is not currently accepting a new payment proof upload.")
		case errors.Is(err, pgx.ErrNoRows):
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that number.")
		default:
			log.Println("PAYMENT PROOF SUBMISSION ERROR:", err)
			webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not save your payment proof right now.")
		}
		return
	}

	if oldProofPath != "" && oldProofPath != proof.PublicPath {
		removeUploadedPaymentProof(h.paymentProofDir, oldProofPath)
	}

	c.Redirect(http.StatusSeeOther, "/order/"+strconv.FormatInt(orderID, 10)+"/payment?proof_submitted=1")
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

func canUploadPaymentProof(payment *orderdomain.OrderPayment) bool {
	if payment == nil {
		return false
	}
	return payment.PaymentStatus == orderdomain.PendingPaymentStatus ||
		payment.PaymentStatus == orderdomain.RejectedPaymentStatus
}

func (h *OrderHandler) savePaymentProof(c *gin.Context, field string) (*savedPaymentProof, error) {
	fileHeader, err := c.FormFile(field)
	if err != nil {
		return nil, errInvalidPaymentProofFile
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxPaymentProofFileSize {
		return nil, errInvalidPaymentProofFile
	}

	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileHeader.Filename)))
	expectedContentType, ok := allowedPaymentProofTypes[ext]
	if !ok {
		return nil, errInvalidPaymentProofFile
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, errInvalidPaymentProofFile
	}
	defer file.Close()

	buffer := make([]byte, 512)
	readBytes, readErr := io.ReadFull(file, buffer)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) {
		return nil, errInvalidPaymentProofFile
	}
	if readBytes == 0 {
		return nil, errInvalidPaymentProofFile
	}

	detectedContentType := http.DetectContentType(buffer[:readBytes])
	canonicalExt, ok := detectedPaymentProofTypes[detectedContentType]
	if !ok || detectedContentType != expectedContentType {
		return nil, errInvalidPaymentProofFile
	}

	if err := os.MkdirAll(h.paymentProofDir, 0o755); err != nil {
		return nil, err
	}

	filename, err := newPaymentProofFilename(canonicalExt)
	if err != nil {
		return nil, err
	}

	savePath := filepath.Join(h.paymentProofDir, filename)
	if err := c.SaveUploadedFile(fileHeader, savePath); err != nil {
		return nil, err
	}

	return &savedPaymentProof{
		PublicPath: paymentProofPathPrefix + filename,
		SavedPath:  savePath,
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

func removeUploadedPaymentProof(uploadDir string, publicPath string) {
	savePath, ok := paymentProofFilePath(uploadDir, publicPath)
	if !ok {
		return
	}
	_ = os.Remove(savePath)
}

func paymentProofFilePath(uploadDir string, publicPath string) (string, bool) {
	path := strings.TrimSpace(publicPath)
	if !strings.HasPrefix(path, paymentProofPathPrefix) {
		return "", false
	}

	relPath := strings.TrimPrefix(path, paymentProofPathPrefix)
	if relPath == "" || strings.Contains(relPath, "..") {
		return "", false
	}

	return filepath.Join(uploadDir, filepath.FromSlash(relPath)), true
}
