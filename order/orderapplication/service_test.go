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

func TestCalculatePricingTreatsRequestedInsertsAsAdditionalExtras(t *testing.T) {
	t.Parallel()

	service := &Service{
		cardRepo: &stubCardRepo{
			card: &carddomain.Card{
				ID:              7,
				Name:            "Trusted Card",
				PriceFoilPKR:    500,
				PriceNofoilPKR:  450,
				InsertPricePKR:  25,
				MinOrder:        1,
				IncludedInserts: 2,
				Category:        "wedding-cards",
			},
		},
	}

	pricing, err := service.calculatePricing(context.Background(), 7, 2, "PKR", "foil", 1)
	if err != nil {
		t.Fatalf("calculatePricing returned error: %v", err)
	}

	if pricing.includedInserts != 2 {
		t.Fatalf("expected included inserts 2, got %d", pricing.includedInserts)
	}
	if pricing.extraInserts != 1 {
		t.Fatalf("expected selected extra inserts 1, got %d", pricing.extraInserts)
	}

	expectedTotal := int64((500 + 25) * 2)
	if pricing.totalPrice != expectedTotal {
		t.Fatalf("expected total with one extra insert per card %d, got %d", expectedTotal, pricing.totalPrice)
	}
}

func TestCalculatePricingUsesSelectedExtraInsertsAsChargeableExtras(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		requestedInserts int64
		wantExtraInserts int64
		wantExtraCost    int64
		wantPerCardTotal int64
		wantOrderTotal   int64
	}{
		{
			name:             "no extra inserts",
			requestedInserts: 0,
			wantExtraInserts: 0,
			wantExtraCost:    0,
			wantPerCardTotal: 580,
			wantOrderTotal:   29000,
		},
		{
			name:             "one extra insert beyond included inserts",
			requestedInserts: 1,
			wantExtraInserts: 1,
			wantExtraCost:    80,
			wantPerCardTotal: 660,
			wantOrderTotal:   33000,
		},
		{
			name:             "two extra inserts beyond included inserts",
			requestedInserts: 2,
			wantExtraInserts: 2,
			wantExtraCost:    160,
			wantPerCardTotal: 740,
			wantOrderTotal:   37000,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := &Service{
				cardRepo: &stubCardRepo{
					card: &carddomain.Card{
						ID:              5,
						Name:            "Sage & Gold Nikkah Suite",
						PriceFoilPKR:    580,
						PriceNofoilPKR:  580,
						InsertPricePKR:  80,
						MinOrder:        50,
						IncludedInserts: 2,
						Category:        "wedding-cards",
					},
				},
			}

			pricing, err := service.calculatePricing(context.Background(), 5, 50, "PKR", "foil", tc.requestedInserts)
			if err != nil {
				t.Fatalf("calculatePricing returned error: %v", err)
			}

			if pricing.includedInserts != 2 {
				t.Fatalf("expected included inserts 2, got %d", pricing.includedInserts)
			}
			if pricing.extraInserts != tc.wantExtraInserts {
				t.Fatalf("expected extra inserts %d, got %d", tc.wantExtraInserts, pricing.extraInserts)
			}
			if pricing.extraInsertCost != tc.wantExtraCost {
				t.Fatalf("expected extra insert cost %d, got %d", tc.wantExtraCost, pricing.extraInsertCost)
			}
			if pricing.perCardPrice != tc.wantPerCardTotal {
				t.Fatalf("expected per-card total %d, got %d", tc.wantPerCardTotal, pricing.perCardPrice)
			}
			if pricing.totalPrice != tc.wantOrderTotal {
				t.Fatalf("expected order total %d, got %d", tc.wantOrderTotal, pricing.totalPrice)
			}
		})
	}
}

func TestCalculatePricingBulkDiscount(t *testing.T) {
	t.Parallel()

	newService := func() *Service {
		return &Service{
			cardRepo: &stubCardRepo{
				card: &carddomain.Card{
					ID:              9,
					Name:            "Bulk Discount Card",
					PriceFoilPKR:    380,
					PriceNofoilPKR:  380,
					InsertPricePKR:  50,
					MinOrder:        1,
					IncludedInserts: 0,
					Category:        "wedding-cards",
				},
			},
		}
	}

	tests := []struct {
		name             string
		quantity         int64
		requestedInserts int64
		wantDiscount     bool
		wantCardSubtotal int64
		wantInsertSub    int64
		wantTotal        int64
	}{
		{
			name:             "quantity 70 no discount",
			quantity:         70,
			requestedInserts: 0,
			wantDiscount:     false,
			wantCardSubtotal: 380 * 70,
			wantInsertSub:    0,
			wantTotal:        380 * 70,
		},
		{
			name:             "quantity 71 discount applies",
			quantity:         71,
			requestedInserts: 0,
			wantDiscount:     true,
			wantCardSubtotal: 380 * 71,
			wantInsertSub:    0,
			wantTotal:        int64(float64(380*71) * 0.85),
		},
		{
			name:             "quantity 100 with extra inserts, inserts not discounted",
			quantity:         100,
			requestedInserts: 3,
			wantDiscount:     true,
			wantCardSubtotal: 380 * 100,
			wantInsertSub:    3 * 50 * 100,
			wantTotal:        int64(float64(380*100)*0.85) + 3*50*100,
		},
		{
			name:             "quantity 1 no discount sanity check",
			quantity:         1,
			requestedInserts: 0,
			wantDiscount:     false,
			wantCardSubtotal: 380,
			wantInsertSub:    0,
			wantTotal:        380,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := newService()
			pricing, err := service.calculatePricing(context.Background(), 9, tc.quantity, "PKR", "foil", tc.requestedInserts)
			if err != nil {
				t.Fatalf("calculatePricing returned error: %v", err)
			}

			if pricing.discountApplied != tc.wantDiscount {
				t.Fatalf("expected discountApplied=%v, got %v", tc.wantDiscount, pricing.discountApplied)
			}
			if pricing.cardSubtotal != tc.wantCardSubtotal {
				t.Fatalf("expected card subtotal %d, got %d", tc.wantCardSubtotal, pricing.cardSubtotal)
			}
			if pricing.insertSubtotal != tc.wantInsertSub {
				t.Fatalf("expected insert subtotal %d, got %d", tc.wantInsertSub, pricing.insertSubtotal)
			}
			if pricing.totalPrice != tc.wantTotal {
				t.Fatalf("expected total price %d, got %d", tc.wantTotal, pricing.totalPrice)
			}
		})
	}
}

func TestPrepareOrderReviewAllowsMultipleRSVPContacts(t *testing.T) {
	t.Parallel()

	service := newReviewTestService()

	review, err := service.PrepareOrderReview(context.Background(), PlaceOrderInput{
		CardID:           7,
		Quantity:         10,
		FoilOption:       "foil",
		RequestedInserts: 0,
		Name:             "Aimen",
		Email:            "aimen@example.com",
		Phone:            "03001234567",
		Address:          "123 Karim Block",
		City:             "Lahore",
		PostalCode:       "54000",
		Side:             "bride",
		RsvpName:         "Ali\nSara",
		RsvpPhone:        "0300 1234567\n0300 7654321",
	})
	if err != nil {
		t.Fatalf("PrepareOrderReview returned error: %v", err)
	}

	if review.Input.RsvpName != "Ali\nSara" {
		t.Fatalf("expected RSVP names preserved, got %q", review.Input.RsvpName)
	}
	if review.Input.RsvpPhone != "03001234567\n03007654321" {
		t.Fatalf("expected RSVP phones normalized and preserved, got %q", review.Input.RsvpPhone)
	}
}

func TestPrepareOrderReviewTreatsRSVPPhoneAsOptional(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		rsvpName  string
		rsvpPhone string
		wantName  string
		wantPhone string
	}{
		{
			name:      "blank RSVP phone allowed",
			rsvpPhone: " \n\t ",
			wantPhone: "",
		},
		{
			name:      "local RSVP phone allowed",
			rsvpPhone: "03084549268",
			wantPhone: "03084549268",
		},
		{
			name:      "international RSVP phone allowed",
			rsvpPhone: "+923084549268",
			wantPhone: "+923084549268",
		},
		{
			name:      "923 RSVP phone normalized",
			rsvpPhone: "923084549268",
			wantPhone: "+923084549268",
		},
		{
			name:      "leading 3 RSVP phone normalized",
			rsvpPhone: "3084549268",
			wantPhone: "03084549268",
		},
		{
			name:      "hyphen and space RSVP phones normalized",
			rsvpPhone: "0300-1234567\n0300 7654321",
			wantPhone: "03001234567\n03007654321",
		},
		{
			name:      "blank rows ignored",
			rsvpName:  "Ali\n\nSara",
			rsvpPhone: "\n03084549268\n\n+923084549268\n",
			wantName:  "Ali\nSara",
			wantPhone: "03084549268\n+923084549268",
		},
		{
			name:      "more names than phones allowed",
			rsvpName:  "Ali\nSara\nZara",
			rsvpPhone: "03084549268\n\n",
			wantName:  "Ali\nSara\nZara",
			wantPhone: "03084549268",
		},
		{
			name:      "phones without names allowed",
			rsvpPhone: "03084549268\n3084549268",
			wantPhone: "03084549268\n03084549268",
		},
		{
			name:      "optional invalid RSVP phone skipped",
			rsvpName:  "Ali\nSara",
			rsvpPhone: "not-a-phone\n0300 1234567",
			wantName:  "Ali\nSara",
			wantPhone: "03001234567",
		},
		{
			name:      "repeated names and phones preserved",
			rsvpName:  "Ali\nAli\nSara",
			rsvpPhone: "03084549268\n03084549268\n+923084549268",
			wantName:  "Ali\nAli\nSara",
			wantPhone: "03084549268\n03084549268\n+923084549268",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			review, err := newReviewTestService().PrepareOrderReview(context.Background(), baseReviewInput(PlaceOrderInput{
				RsvpName:  tc.rsvpName,
				RsvpPhone: tc.rsvpPhone,
			}))
			if err != nil {
				t.Fatalf("PrepareOrderReview returned error: %v", err)
			}

			if review.Input.RsvpName != tc.wantName {
				t.Fatalf("expected RSVP names %q, got %q", tc.wantName, review.Input.RsvpName)
			}
			if review.Input.RsvpPhone != tc.wantPhone {
				t.Fatalf("expected RSVP phones %q, got %q", tc.wantPhone, review.Input.RsvpPhone)
			}
		})
	}
}

func TestPrepareOrderReviewNormalizesCommonPakistanPhoneFormats(t *testing.T) {
	t.Parallel()

	service := newReviewTestService()

	review, err := service.PrepareOrderReview(context.Background(), PlaceOrderInput{
		CardID:           7,
		Quantity:         10,
		FoilOption:       "foil",
		RequestedInserts: 0,
		Name:             "Aimen",
		Email:            "aimen@example.com",
		Phone:            "923001234567",
		Address:          "123 Karim Block",
		City:             "Lahore",
		PostalCode:       "54000",
		Side:             "bride",
		RsvpPhone:        "3007654321",
	})
	if err != nil {
		t.Fatalf("PrepareOrderReview returned error: %v", err)
	}

	if review.Input.Phone != "+923001234567" {
		t.Fatalf("expected customer phone normalized, got %q", review.Input.Phone)
	}
	if review.Input.RsvpPhone != "03007654321" {
		t.Fatalf("expected RSVP phone normalized, got %q", review.Input.RsvpPhone)
	}
}

func newReviewTestService() *Service {
	return &Service{
		cardRepo: &stubCardRepo{
			card: &carddomain.Card{
				ID:              7,
				Name:            "Trusted Card",
				PriceFoilPKR:    500,
				PriceNofoilPKR:  450,
				InsertPricePKR:  25,
				MinOrder:        1,
				IncludedInserts: 0,
				Category:        "wedding-cards",
			},
		},
	}
}

func baseReviewInput(override PlaceOrderInput) PlaceOrderInput {
	input := PlaceOrderInput{
		CardID:           7,
		Quantity:         10,
		FoilOption:       "foil",
		RequestedInserts: 0,
		Name:             "Aimen",
		Email:            "aimen@example.com",
		Phone:            "03001234567",
		Address:          "123 Karim Block",
		City:             "Lahore",
		PostalCode:       "54000",
		Side:             "bride",
	}
	if override.CardID != 0 {
		input.CardID = override.CardID
	}
	if override.Quantity != 0 {
		input.Quantity = override.Quantity
	}
	if override.FoilOption != "" {
		input.FoilOption = override.FoilOption
	}
	input.RequestedInserts = override.RequestedInserts
	if override.Name != "" {
		input.Name = override.Name
	}
	if override.Email != "" {
		input.Email = override.Email
	}
	if override.Phone != "" {
		input.Phone = override.Phone
	}
	if override.Address != "" {
		input.Address = override.Address
	}
	if override.City != "" {
		input.City = override.City
	}
	if override.PostalCode != "" {
		input.PostalCode = override.PostalCode
	}
	if override.Side != "" {
		input.Side = override.Side
	}
	input.RsvpName = override.RsvpName
	input.RsvpPhone = override.RsvpPhone
	return input
}

func TestPrepareOrderReviewAcceptsExactBrowserCustomerValues(t *testing.T) {
	t.Parallel()

	service := &Service{
		cardRepo: &stubCardRepo{
			card: &carddomain.Card{
				ID:              5,
				Name:            "Sage & Gold Nikkah Suite",
				PriceFoilPKR:    580,
				PriceNofoilPKR:  580,
				InsertPricePKR:  80,
				MinOrder:        50,
				IncludedInserts: 2,
				Category:        "wedding-cards",
			},
		},
	}

	review, err := service.PrepareOrderReview(context.Background(), PlaceOrderInput{
		CardID:           5,
		Quantity:         50,
		FoilOption:       "foil",
		RequestedInserts: 1,
		Name:             "muhammad ali",
		Email:            "meetali098@gmail.com",
		Phone:            "03084549268",
		Address:          "househouse",
		City:             "Lahore",
		PostalCode:       "2448",
		Side:             "bride",
	})
	if err != nil {
		t.Fatalf("PrepareOrderReview returned error: %v", err)
	}

	if review.Input.Phone != "03084549268" {
		t.Fatalf("expected customer phone preserved, got %q", review.Input.Phone)
	}
	if review.Input.PostalCode != "2448" {
		t.Fatalf("expected short local postal code preserved, got %q", review.Input.PostalCode)
	}
	if review.Summary.ExtraInserts != 1 {
		t.Fatalf("expected extra inserts to remain 1, got %d", review.Summary.ExtraInserts)
	}
	if review.Summary.ExtraInsertCost != 80 {
		t.Fatalf("expected extra insert cost 80, got %d", review.Summary.ExtraInsertCost)
	}
	if review.Summary.PerCardTotal != 660 {
		t.Fatalf("expected per-card total 660, got %d", review.Summary.PerCardTotal)
	}
	if review.Summary.TotalPrice != 33000 {
		t.Fatalf("expected trusted total 33000, got %d", review.Summary.TotalPrice)
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
