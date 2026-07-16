package cardpresentation

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"writeandinviteco/inviteandco/card/carddomain"

	"github.com/gin-gonic/gin"
)

type fakeCardRepo struct {
	card *carddomain.Card
}

func (f fakeCardRepo) CreateCard(ctx context.Context, name string, description *string, priceFoilPKR int64, priceNofoilPKR int64, priceFoilNOK int64, priceNofoilNOK int64, insertPricePKR int64, insertPriceNOK int64, minOrder int32, includedInserts int32, image string, category string) (*carddomain.Card, error) {
	return nil, nil
}

func (f fakeCardRepo) UpdateCard(ctx context.Context, id int64, name string, description *string, priceFoilPKR int64, priceNofoilPKR int64, priceFoilNOK int64, priceNofoilNOK int64, insertPricePKR int64, insertPriceNOK int64, minOrder int32, includedInserts int32, image string, category string) error {
	return nil
}

func (f fakeCardRepo) DeleteCard(ctx context.Context, id int64) error {
	return nil
}

func (f fakeCardRepo) CreateCardImage(ctx context.Context, cardID int64, image string, sortOrder int32) (*carddomain.CardImage, error) {
	return nil, nil
}

func (f fakeCardRepo) DeleteCardImagesByCardID(ctx context.Context, cardID int64) error {
	return nil
}

func (f fakeCardRepo) GetAllCards(ctx context.Context) ([]*carddomain.Card, error) {
	return nil, nil
}

func (f fakeCardRepo) GetCardByID(ctx context.Context, id int64) (*carddomain.Card, error) {
	return f.card, nil
}

func (f fakeCardRepo) GetCardsByCategory(ctx context.Context, category string) ([]*carddomain.Card, error) {
	return nil, nil
}

func (f fakeCardRepo) SearchCards(ctx context.Context, query string) ([]*carddomain.Card, error) {
	return nil, nil
}

func (f fakeCardRepo) GetCardImagesByCardID(ctx context.Context, cardID int64) ([]*carddomain.CardImage, error) {
	return nil, nil
}

func TestJoinedCheckoutValuePreservesRepeatedRSVPValues(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	form := url.Values{
		"rsvp_name":  {"Ali", "Sara"},
		"rsvp_phone": {"03001234567", "03007654321"},
	}
	req := httptest.NewRequest(http.MethodPost, "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	if got := joinedCheckoutValue(c, "rsvp_name"); got != "Ali\nSara" {
		t.Fatalf("expected joined RSVP names, got %q", got)
	}
	if got := joinedCheckoutValue(c, "rsvp_phone"); got != "03001234567\n03007654321" {
		t.Fatalf("expected joined RSVP phones, got %q", got)
	}
}

func TestCheckoutPostAfterPersonalizationRendersCustomerInfo(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	form := url.Values{
		"csrf_token":        {"trusted-csrf"},
		"card_id":           {"7"},
		"quantity":          {"50"},
		"foil_option":       {"foil"},
		"extra_inserts":     {"1"},
		"side":              {"bride"},
		"bride_name":        {"Aimen"},
		"groom_name":        {"Sohail"},
		"bride_father_name": {"Bride Parents"},
		"groom_father_name": {"Groom Parents"},
		"mehndi_venue_name": {"Mehndi Hall"},
		"baraat_venue_name": {"Baraat Hall"},
		"nikkah_venue_name": {"Nikkah Hall"},
		"walima_venue_name": {"Walima Hall"},
		"rsvp_name":         {"Ali", "Sara"},
		"rsvp_phone":        {"03001234567", "03007654321"},
	}

	req := httptest.NewRequest(http.MethodPost, "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "trusted-csrf"})

	recorder := httptest.NewRecorder()
	handler := NewCardHandler(fakeCardRepo{card: &carddomain.Card{
		ID:              7,
		Name:            "Trusted Card",
		PriceFoilPKR:    500,
		PriceNofoilPKR:  450,
		InsertPricePKR:  60,
		MinOrder:        50,
		IncludedInserts: 2,
		Category:        "wedding-cards",
	}}, nil, nil)

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("checkout.html").Parse(`{{ .cardName }}|extra={{ .extraInserts }}|bride={{ .personalization.BrideName }}|rsvp={{ .personalization.RsvpName }}|<form action="/review" method="post">`)))
	router.POST("/checkout", handler.Checkout)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Trusted Card",
		"extra=1",
		"bride=Aimen",
		"rsvp=Ali\nSara",
		`<form action="/review" method="post">`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected checkout body to contain %q, got %q", want, body)
		}
	}
}

func TestCheckoutPostCalculatesSelectedExtraInsertsAsAdditionalCharge(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	form := url.Values{
		"csrf_token":    {"trusted-csrf"},
		"card_id":       {"5"},
		"quantity":      {"50"},
		"foil_option":   {"foil"},
		"extra_inserts": {"1"},
		"side":          {"bride"},
	}

	req := httptest.NewRequest(http.MethodPost, "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "trusted-csrf"})

	recorder := httptest.NewRecorder()
	handler := NewCardHandler(fakeCardRepo{card: &carddomain.Card{
		ID:              5,
		Name:            "Sage & Gold Nikkah Suite",
		PriceFoilPKR:    580,
		PriceNofoilPKR:  580,
		InsertPricePKR:  80,
		MinOrder:        50,
		IncludedInserts: 2,
		Category:        "wedding-cards",
	}}, nil, nil)

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("checkout.html").Parse(`included={{ .includedInserts }}|extra={{ .extraInserts }}|extraCost={{ .extraInsertCostPerCard }}|perCard={{ .perCardTotal }}|total={{ .totalPrice }}|hidden=<input name="extra_inserts" value="{{ .requestedInserts }}">`)))
	router.POST("/checkout", handler.Checkout)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"included=2",
		"extra=1",
		"extraCost=80",
		"perCard=660",
		"total=33000",
		`hidden=<input name="extra_inserts" value="1">`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected checkout body to contain %q, got %q", want, body)
		}
	}
}

func TestCardTemplateExplainsExtraInsertsAsAdditionalPrintedCards(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("../../templates/card.html")
	if err != nil {
		t.Fatalf("read card template: %v", err)
	}
	text := string(body)

	for _, want := range []string{
		"An insert is an additional printed card placed inside the invitation",
		"includes {{ .card.IncludedInserts }} inserts per card",
		"Choose extra inserts only if you need more printed cards",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected card template to contain %q", want)
		}
	}
}
