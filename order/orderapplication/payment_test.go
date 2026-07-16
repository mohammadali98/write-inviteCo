package orderapplication

import (
	"context"
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

func TestTrustedAdvanceAmountUsesPaymentExpectedAmount(t *testing.T) {
	t.Parallel()

	service := &Service{
		orderRepo: stubPaymentOrderRepo{
			order: &orderdomain.Order{ID: 42, TotalPrice: 99999},
		},
	}

	got, err := service.trustedAdvanceAmount(context.Background(), 42, &orderdomain.OrderPayment{
		ExpectedAmount: 16500,
	})
	if err != nil {
		t.Fatalf("trustedAdvanceAmount returned error: %v", err)
	}
	if got != 16500 {
		t.Fatalf("expected stored trusted advance 16500, got %d", got)
	}
}

func TestTrustedAdvanceAmountFallsBackToServerCalculatedAdvance(t *testing.T) {
	t.Parallel()

	service := &Service{
		orderRepo: stubPaymentOrderRepo{
			order: &orderdomain.Order{ID: 42, TotalPrice: 33001},
		},
	}

	got, err := service.trustedAdvanceAmount(context.Background(), 42, nil)
	if err != nil {
		t.Fatalf("trustedAdvanceAmount returned error: %v", err)
	}
	if got != 16501 {
		t.Fatalf("expected ceil 50%% trusted advance 16501, got %d", got)
	}
}

type stubPaymentOrderRepo struct {
	order          *orderdomain.Order
	capturedFilter *orderdomain.AdminOrderFilter
}

func (s stubPaymentOrderRepo) CreateOrder(ctx context.Context, customerID int64, cardID int64, quantity int64, totalPrice int64, status orderdomain.OrderStatus, currency string) (*orderdomain.Order, error) {
	return nil, nil
}

func (s stubPaymentOrderRepo) CreateOrderDetail(ctx context.Context, orderID int64, side string, extraInsertsPerCard int64, bidBoxTopLabel *string, bidBoxCoupleName *string, bidBoxEventDate *string, bidBoxDetails *string, brideName *string, groomName *string, brideFatherName *string, groomFatherName *string, mehndiDate *string, mehndiDay *string, mehndiTimeType *string, mehndiTime *string, mehndiDinnerTime *string, mehndiVenueName *string, mehndiVenueAddress *string, baraatDate *string, baraatDay *string, baraatTimeType *string, baraatTime *string, baraatDinnerTime *string, baraatArrivalTime *string, rukhsatiTime *string, baraatSehrabandiTime *string, baraatVenueName *string, baraatVenueAddress *string, nikkahDate *string, nikkahDay *string, nikkahTimeType *string, nikkahTime *string, nikkahDinnerTime *string, nikkahVenueName *string, nikkahVenueAddress *string, walimaDate *string, walimaDay *string, walimaTimeType *string, walimaTime *string, walimaDinnerTime *string, receptionTime *string, walimaVenueName *string, walimaVenueAddress *string, rsvpName string, rsvpPhone string, notes *string) (*orderdomain.OrderDetail, error) {
	return nil, nil
}

func (s stubPaymentOrderRepo) UpdateOrderStatus(ctx context.Context, id int64, status orderdomain.OrderStatus) error {
	return nil
}

func (s stubPaymentOrderRepo) GetOrderByID(ctx context.Context, id int64) (*orderdomain.Order, error) {
	return s.order, nil
}

func (s stubPaymentOrderRepo) GetOrderByPublicToken(ctx context.Context, token string) (*orderdomain.Order, error) {
	return s.order, nil
}

func (s stubPaymentOrderRepo) GetOrdersByCustomerID(ctx context.Context, customerID int64) ([]*orderdomain.Order, error) {
	return nil, nil
}

func (s stubPaymentOrderRepo) GetAdminOrders(ctx context.Context, filter orderdomain.AdminOrderFilter) ([]*orderdomain.AdminOrder, error) {
	if s.capturedFilter != nil {
		*s.capturedFilter = filter
	}
	return nil, nil
}

func (s stubPaymentOrderRepo) GetOrderDetailByOrderID(ctx context.Context, orderID int64) (*orderdomain.OrderDetail, error) {
	return nil, nil
}

func (s stubPaymentOrderRepo) GetOrderPaymentByOrderID(ctx context.Context, orderID int64) (*orderdomain.OrderPayment, error) {
	return nil, nil
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
