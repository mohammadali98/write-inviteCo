package orderdomain

import "time"

type PaymentMethod string

const (
	BankTransferPaymentMethod PaymentMethod = "bank_transfer"
)

type PaymentStatus string

const (
	PendingPaymentStatus              PaymentStatus = "pending_payment"
	AwaitingVerificationPaymentStatus PaymentStatus = "awaiting_verification"
	VerifiedPaymentStatus             PaymentStatus = "payment_verified"
	RejectedPaymentStatus             PaymentStatus = "payment_rejected"
)

type OrderPayment struct {
	ID                   int64
	OrderID              int64
	PaymentMethod        PaymentMethod
	PaymentStatus        PaymentStatus
	ExpectedAmount       int64
	SubmittedAmount      *int64
	SenderName           *string
	TransactionReference *string
	ProofFilePath        *string
	CustomerNote         *string
	SubmittedAt          *time.Time
	VerifiedAt           *time.Time
	RejectedAt           *time.Time
	AdminNote            *string
	CreatedAt            *time.Time
	UpdatedAt            *time.Time
}
