package orderapplication

import (
	"context"
	"strings"
	"testing"

	"writeandinviteco/inviteandco/card/carddomain"
	"writeandinviteco/inviteandco/order/orderdomain"
)

type stubCardRepo struct {
	card *carddomain.Card
	err  error
}

func (s *stubCardRepo) CreateCard(ctx context.Context, name string, description *string, priceFoilPKR int64, priceNofoilPKR int64, priceFoilNOK int64, priceNofoilNOK int64, insertPricePKR int64, insertPriceNOK int64, minOrder int32, includedInserts int32, image string, category string) (*carddomain.Card, error) {
	return nil, nil
}

func (s *stubCardRepo) UpdateCard(ctx context.Context, id int64, name string, description *string, priceFoilPKR int64, priceNofoilPKR int64, priceFoilNOK int64, priceNofoilNOK int64, insertPricePKR int64, insertPriceNOK int64, minOrder int32, includedInserts int32, image string, category string) error {
	return nil
}

func (s *stubCardRepo) DeleteCard(ctx context.Context, id int64) error {
	return nil
}

func (s *stubCardRepo) CreateCardImage(ctx context.Context, cardID int64, image string, sortOrder int32) (*carddomain.CardImage, error) {
	return nil, nil
}

func (s *stubCardRepo) DeleteCardImagesByCardID(ctx context.Context, cardID int64) error {
	return nil
}

func (s *stubCardRepo) GetAllCards(ctx context.Context) ([]*carddomain.Card, error) {
	return nil, nil
}

func (s *stubCardRepo) GetCardByID(ctx context.Context, id int64) (*carddomain.Card, error) {
	return s.card, s.err
}

func (s *stubCardRepo) GetCardsByCategory(ctx context.Context, category string) ([]*carddomain.Card, error) {
	return nil, nil
}

func (s *stubCardRepo) SearchCards(ctx context.Context, query string) ([]*carddomain.Card, error) {
	return nil, nil
}

func (s *stubCardRepo) GetCardImagesByCardID(ctx context.Context, cardID int64) ([]*carddomain.CardImage, error) {
	return nil, nil
}

func TestCalculatePricingIgnoresClientCurrency(t *testing.T) {
	t.Parallel()

	service := &Service{
		cardRepo: &stubCardRepo{
			card: &carddomain.Card{
				ID:              7,
				Name:            "Trusted Card",
				PriceFoilPKR:    500,
				PriceNofoilPKR:  450,
				PriceFoilNOK:    9000,
				PriceNofoilNOK:  8500,
				InsertPricePKR:  25,
				InsertPriceNOK:  400,
				MinOrder:        1,
				IncludedInserts: 0,
				Category:        "wedding-cards",
			},
		},
	}

	pricing, err := service.calculatePricing(context.Background(), 7, 2, "NOK", "foil", 3)
	if err != nil {
		t.Fatalf("calculatePricing returned error: %v", err)
	}

	if pricing.currency != "PKR" {
		t.Fatalf("expected trusted currency PKR, got %q", pricing.currency)
	}
	if pricing.basePrice != 500 {
		t.Fatalf("expected PKR foil price 500, got %d", pricing.basePrice)
	}
	if pricing.insertPrice != 25 {
		t.Fatalf("expected PKR insert price 25, got %d", pricing.insertPrice)
	}

	expectedTotal := int64((500 + (3 * 25)) * 2)
	if pricing.totalPrice != expectedTotal {
		t.Fatalf("expected trusted total %d, got %d", expectedTotal, pricing.totalPrice)
	}
}

func TestBuildPaymentAmountSummaryUsesFiftyPercentAdvance(t *testing.T) {
	t.Parallel()

	summary := BuildPaymentAmountSummary(24000, 0)

	if summary.TotalAmount != 24000 {
		t.Fatalf("expected total 24000, got %d", summary.TotalAmount)
	}
	if summary.AdvanceAmount != 12000 {
		t.Fatalf("expected advance 12000, got %d", summary.AdvanceAmount)
	}
	if summary.RemainingBalance != 12000 {
		t.Fatalf("expected remaining 12000, got %d", summary.RemainingBalance)
	}
}

func TestBuildPaymentAmountSummaryRoundsOddTotalUp(t *testing.T) {
	t.Parallel()

	summary := BuildPaymentAmountSummary(24001, 0)

	if summary.AdvanceAmount != 12001 {
		t.Fatalf("expected rounded-up advance 12001, got %d", summary.AdvanceAmount)
	}
	if summary.RemainingBalance != 12000 {
		t.Fatalf("expected remaining 12000, got %d", summary.RemainingBalance)
	}
}

func TestBuildCustomerOrderStatusEmailBodyForConfirmedOrderShowsAdvanceBreakdown(t *testing.T) {
	t.Parallel()

	body := buildCustomerOrderStatusEmailBody(orderEmailPayload{
		OrderID:       88,
		CustomerName:  "Aimen",
		ProductName:   "Floral Invite",
		Quantity:      100,
		TotalPrice:    32500,
		AdvanceAmount: 16250,
		Remaining:     16250,
		Currency:      "PKR",
		PaymentStatus: orderdomain.VerifiedPaymentStatus,
		Status:        orderdomain.ConfirmedOrderStatus,
	})

	for _, want := range []string{
		"Your order has been confirmed.",
		"Order Total: PKR 32500",
		"Advance Payment Received: PKR 16250",
		"Remaining Balance: PKR 16250",
		"Payment Status: Advance Payment Received / Final Payment Pending",
		"Our team will contact you on your provided WhatsApp number or email when your order is ready.",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected confirmed customer email to contain %q\nbody:\n%s", want, body)
		}
	}

	for _, unwanted := range []string{
		"\nTotal: PKR 32500\n",
		"Status: Confirmed",
		"full payment is complete",
	} {
		if strings.Contains(body, unwanted) {
			t.Fatalf("expected confirmed customer email to avoid %q\nbody:\n%s", unwanted, body)
		}
	}
}

func TestBuildAdminOrderStatusEmailBodyForConfirmedOrderShowsAdvanceBreakdown(t *testing.T) {
	t.Parallel()

	body := buildAdminOrderStatusEmailBody(orderEmailPayload{
		OrderID:       88,
		CustomerName:  "Aimen",
		CustomerPhone: "03001234567",
		CustomerEmail: "aimen@example.com",
		ProductName:   "Floral Invite",
		Quantity:      100,
		TotalPrice:    32500,
		AdvanceAmount: 16250,
		Remaining:     16250,
		Currency:      "PKR",
		PaymentStatus: orderdomain.VerifiedPaymentStatus,
		Status:        orderdomain.ConfirmedOrderStatus,
	})

	for _, want := range []string{
		"Order confirmed for processing.",
		"Phone/WhatsApp: 03001234567",
		"Email: aimen@example.com",
		"Order Total: PKR 32500",
		"Advance Payment Received: PKR 16250",
		"Remaining Balance: PKR 16250",
		"Payment Status: Advance Verified / Final Payment Pending",
		"The client has paid 50% advance only.",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected confirmed admin email to contain %q\nbody:\n%s", want, body)
		}
	}

	if strings.Contains(body, "Status: Confirmed") {
		t.Fatalf("expected confirmed admin email to avoid generic confirmed-only payment language\nbody:\n%s", body)
	}
}
