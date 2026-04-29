package orderapplication

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"writeandinviteco/inviteandco/order/orderdomain"
	orderwriter "writeandinviteco/inviteandco/order/orderinfrastructure/postgres/writer"
)

var (
	ErrPaymentActionNotAllowed     = errors.New("payment action not allowed")
	ErrPaymentVerificationRequired = errors.New("payment verification required")
)

const (
	maxSubmittedAmount            = 1_000_000_000_000
	maxTransactionReferenceLength = 150
	maxCustomerPaymentNoteLength  = 1000
	maxAdminPaymentNoteLength     = 1000
	paymentProofStaticPrefix      = "/static/payment-proofs/"
)

type BankTransferDetails struct {
	AccountTitle  string
	BankName      string
	BranchName    string
	AccountNumber string
	IBAN          string
}

var bankTransferDetails = BankTransferDetails{
	AccountTitle:  "AIMEN FATIMA SOHAIL",
	BankName:      "Meezan Bank",
	BranchName:    "Karim Block, Lahore",
	AccountNumber: "02840104731135",
	IBAN:          "PK56MEZN0002840104731135",
}

type PaymentProofInput struct {
	SubmittedAmount      int64
	SenderName           string
	TransactionReference string
	CustomerNote         string
	ProofFilePath        string
}

type PaymentAmountSummary struct {
	TotalAmount      int64
	AdvanceAmount    int64
	RemainingBalance int64
}

func BankTransferInstructions() BankTransferDetails {
	return bankTransferDetails
}

func BuildPaymentAmountSummary(totalAmount int64, expectedAdvanceAmount int64) PaymentAmountSummary {
	total := totalAmount
	if total < 0 {
		total = 0
	}

	advance := expectedAdvanceAmount
	if advance <= 0 {
		advance = total / 2
		if total%2 != 0 {
			advance++
		}
	}
	if advance < 0 {
		advance = 0
	}
	if advance > total {
		advance = total
	}

	return PaymentAmountSummary{
		TotalAmount:      total,
		AdvanceAmount:    advance,
		RemainingBalance: total - advance,
	}
}

func (s *Service) SubmitBankTransferProof(ctx context.Context, orderID int64, input PaymentProofInput) error {
	if orderID <= 0 {
		return ErrInvalidInput
	}

	input = sanitizePaymentProofInput(input)
	if err := validatePaymentProofInput(input); err != nil {
		return err
	}

	payment, err := s.orderRepo.GetOrderPaymentByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	switch payment.PaymentStatus {
	case orderdomain.PendingPaymentStatus, orderdomain.RejectedPaymentStatus:
	default:
		return ErrPaymentActionNotAllowed
	}

	status := string(orderdomain.AwaitingVerificationPaymentStatus)
	_, err = s.orderWriter.SubmitOrderPaymentProof(ctx, orderwriter.SubmitOrderPaymentProofParams{
		OrderID:              orderID,
		PaymentStatus:        status,
		SubmittedAmount:      &input.SubmittedAmount,
		SenderName:           &input.SenderName,
		TransactionReference: &input.TransactionReference,
		ProofFilePath:        &input.ProofFilePath,
		CustomerNote:         nullableString(input.CustomerNote),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) AdminProcessPayment(ctx context.Context, orderID int64, action string, adminNote string) error {
	if orderID <= 0 {
		return ErrInvalidInput
	}

	action = strings.ToLower(strings.TrimSpace(action))
	if action != "verify" && action != "reject" && action != "request_reupload" {
		return ErrInvalidInput
	}

	adminNote = sanitizeMultiline(adminNote)
	if action == "request_reupload" && adminNote == "" {
		adminNote = "Please upload your advance payment proof again."
	}
	if utf8.RuneCountInString(adminNote) > maxAdminPaymentNoteLength {
		return ErrInvalidInput
	}

	payment, err := s.orderRepo.GetOrderPaymentByOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	if payment.PaymentStatus != orderdomain.AwaitingVerificationPaymentStatus {
		return ErrPaymentActionNotAllowed
	}

	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	writer := s.orderWriter.WithTx(tx)

	switch action {
	case "verify":
		if payment.SubmittedAmount == nil ||
			payment.TransactionReference == nil ||
			payment.ProofFilePath == nil ||
			strings.TrimSpace(*payment.ProofFilePath) == "" {
			return ErrPaymentActionNotAllowed
		}

		if _, err := writer.VerifyOrderPayment(ctx, orderwriter.VerifyOrderPaymentParams{
			OrderID:   orderID,
			AdminNote: nullableString(adminNote),
		}); err != nil {
			return err
		}

		if order.Status != orderdomain.CompletedOrderStatus {
			status := string(orderdomain.ConfirmedOrderStatus)
			if err := writer.UpdateOrderStatus(ctx, orderwriter.UpdateOrderStatusParams{
				ID:     orderID,
				Status: &status,
			}); err != nil {
				return err
			}
			order.Status = orderdomain.ConfirmedOrderStatus
		}

	case "reject", "request_reupload":
		if _, err := writer.RejectOrderPayment(ctx, orderwriter.RejectOrderPaymentParams{
			OrderID:   orderID,
			AdminNote: nullableString(adminNote),
		}); err != nil {
			return err
		}

	default:
		return ErrInvalidInput
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if action == "verify" {
		s.sendOrderStatusEmailAsync(order, order.Status)
	}

	return nil
}

func sanitizePaymentProofInput(input PaymentProofInput) PaymentProofInput {
	input.SenderName = sanitizeSingleLine(input.SenderName)
	input.TransactionReference = sanitizeSingleLine(input.TransactionReference)
	input.CustomerNote = sanitizeMultiline(input.CustomerNote)
	input.ProofFilePath = strings.TrimSpace(input.ProofFilePath)
	return input
}

func validatePaymentProofInput(input PaymentProofInput) error {
	if input.SubmittedAmount <= 0 || input.SubmittedAmount > maxSubmittedAmount {
		return ErrInvalidInput
	}
	if input.SenderName == "" || input.TransactionReference == "" || input.ProofFilePath == "" {
		return ErrInvalidInput
	}
	if utf8.RuneCountInString(input.SenderName) > maxCustomerNameLength ||
		utf8.RuneCountInString(input.TransactionReference) > maxTransactionReferenceLength ||
		utf8.RuneCountInString(input.CustomerNote) > maxCustomerPaymentNoteLength {
		return ErrInvalidInput
	}
	if !containsLetterOrDigit(input.SenderName) || !containsLetterOrDigit(input.TransactionReference) {
		return ErrInvalidInput
	}
	if !strings.HasPrefix(input.ProofFilePath, paymentProofStaticPrefix) || strings.Contains(input.ProofFilePath, "..") {
		return ErrInvalidInput
	}
	return nil
}

func paymentStatusLogValue(payment *orderdomain.OrderPayment) string {
	if payment == nil {
		return ""
	}
	return string(payment.PaymentStatus)
}

func PaymentStatusDisplay(status orderdomain.PaymentStatus) string {
	switch status {
	case orderdomain.PendingPaymentStatus:
		return "Advance Payment Pending"
	case orderdomain.AwaitingVerificationPaymentStatus:
		return "Advance Payment Awaiting Verification"
	case orderdomain.VerifiedPaymentStatus:
		return "Advance Payment Verified"
	case orderdomain.RejectedPaymentStatus:
		return "Advance Payment Rejected"
	default:
		return "Not Recorded"
	}
}

func PaymentStatusMessage(payment *orderdomain.OrderPayment, totalAmount int64) string {
	expectedAdvanceAmount := int64(0)
	if payment != nil {
		expectedAdvanceAmount = payment.ExpectedAmount
	}
	summary := BuildPaymentAmountSummary(totalAmount, expectedAdvanceAmount)

	if payment == nil {
		return fmt.Sprintf("Please transfer the 50%% advance payment of PKR %d to confirm your order. The remaining balance is PKR %d.", summary.AdvanceAmount, summary.RemainingBalance)
	}

	switch payment.PaymentStatus {
	case orderdomain.PendingPaymentStatus:
		return fmt.Sprintf("Please transfer the 50%% advance payment of PKR %d to confirm your order. The remaining balance is PKR %d.", summary.AdvanceAmount, summary.RemainingBalance)
	case orderdomain.AwaitingVerificationPaymentStatus:
		return fmt.Sprintf("Your advance payment proof is awaiting verification. The remaining balance of PKR %d will be collected after your order is completed and before delivery.", summary.RemainingBalance)
	case orderdomain.VerifiedPaymentStatus:
		if summary.RemainingBalance > 0 {
			return fmt.Sprintf("Your advance payment has been verified. The remaining balance of PKR %d will be collected after your order is completed and before delivery.", summary.RemainingBalance)
		}
		return "Your advance payment has been verified."
	case orderdomain.RejectedPaymentStatus:
		return fmt.Sprintf("Your advance payment proof was rejected. Please review the admin note below, if provided, and resubmit your proof. The remaining balance is PKR %d.", summary.RemainingBalance)
	default:
		return "Payment status will be updated here."
	}
}
