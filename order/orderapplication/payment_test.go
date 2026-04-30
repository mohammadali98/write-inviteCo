package orderapplication

import (
	"testing"

	"writeandinviteco/inviteandco/order/orderdomain"
)

func TestBuildPaymentAmountSummaryUsesProvidedExpectedAdvanceAmount(t *testing.T) {
	t.Parallel()

	summary := BuildPaymentAmountSummary(450, 225)

	if summary.AdvanceAmount != 225 {
		t.Fatalf("expected advance 225, got %d", summary.AdvanceAmount)
	}
	if summary.RemainingBalance != 225 {
		t.Fatalf("expected remaining 225, got %d", summary.RemainingBalance)
	}
}

func TestCanVerifySubmittedAdvance(t *testing.T) {
	t.Parallel()

	underpaid := int64(1)
	exact := int64(675)
	overpaid := int64(700)

	tests := []struct {
		name    string
		payment *orderdomain.OrderPayment
		want    bool
	}{
		{
			name:    "nil payment",
			payment: nil,
			want:    false,
		},
		{
			name: "missing submitted amount",
			payment: &orderdomain.OrderPayment{
				ExpectedAmount: 675,
			},
			want: false,
		},
		{
			name: "underpaid advance",
			payment: &orderdomain.OrderPayment{
				ExpectedAmount:  675,
				SubmittedAmount: &underpaid,
			},
			want: false,
		},
		{
			name: "exact advance",
			payment: &orderdomain.OrderPayment{
				ExpectedAmount:  675,
				SubmittedAmount: &exact,
			},
			want: true,
		},
		{
			name: "overpaid advance",
			payment: &orderdomain.OrderPayment{
				ExpectedAmount:  675,
				SubmittedAmount: &overpaid,
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := canVerifySubmittedAdvance(tc.payment)
			if got != tc.want {
				t.Fatalf("canVerifySubmittedAdvance() = %v, want %v", got, tc.want)
			}
		})
	}
}
