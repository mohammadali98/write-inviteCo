package orderpresentation

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"writeandinviteco/inviteandco/order/orderapplication"
	"writeandinviteco/inviteandco/order/orderdomain"

	"github.com/gin-gonic/gin"
)

type fakeOrderService struct {
	placeOrderInput               orderapplication.PlaceOrderInput
	adminUpdateOrderStatusOrderID int64
	adminUpdateOrderStatusStatus  string
	adminUpdateOrderStatusErr     error
	adminProcessPaymentOrderID    int64
	adminProcessPaymentAction     string
	adminProcessPaymentNote       string
	adminProcessPaymentErr        error
}

func (f *fakeOrderService) PrepareCustomization(ctx context.Context, input orderapplication.CustomizationInput) (*orderapplication.CustomizationSummary, error) {
	return nil, nil
}

func (f *fakeOrderService) PrepareOrderReview(ctx context.Context, input orderapplication.PlaceOrderInput) (*orderapplication.OrderReview, error) {
	return nil, nil
}

func (f *fakeOrderService) PlaceOrder(ctx context.Context, input orderapplication.PlaceOrderInput) (*orderapplication.PlaceOrderResult, error) {
	f.placeOrderInput = input
	return &orderapplication.PlaceOrderResult{OrderID: 55}, nil
}

func (f *fakeOrderService) GetOrderStatusDetail(ctx context.Context, orderID int64) (*orderapplication.AdminOrderDetail, error) {
	return nil, nil
}

func (f *fakeOrderService) ListAdminOrders(ctx context.Context) ([]*orderdomain.AdminOrder, error) {
	return nil, nil
}

func (f *fakeOrderService) GetAdminOrderDetail(ctx context.Context, orderID int64) (*orderapplication.AdminOrderDetail, error) {
	return nil, nil
}

func (f *fakeOrderService) AdminUpdateOrderStatus(ctx context.Context, orderID int64, statusRaw string) error {
	f.adminUpdateOrderStatusOrderID = orderID
	f.adminUpdateOrderStatusStatus = statusRaw
	return f.adminUpdateOrderStatusErr
}

func (f *fakeOrderService) SubmitBankTransferProof(ctx context.Context, orderID int64, input orderapplication.PaymentProofInput) error {
	return nil
}

func (f *fakeOrderService) AdminProcessPayment(ctx context.Context, orderID int64, action string, adminNote string) error {
	f.adminProcessPaymentOrderID = orderID
	f.adminProcessPaymentAction = action
	f.adminProcessPaymentNote = adminNote
	return f.adminProcessPaymentErr
}

func TestCreateOrderIgnoresTamperedDisplayFields(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	form := url.Values{
		"csrf_token":            {"trusted-csrf"},
		"card_id":               {"7"},
		"quantity":              {"250"},
		"foil_option":           {"foil"},
		"extra_inserts":         {"3"},
		"name":                  {"Aimen"},
		"email":                 {"aimen@example.com"},
		"phone":                 {"03001234567"},
		"address":               {"123 Karim Block"},
		"city":                  {"Lahore"},
		"postal_code":           {"54000"},
		"price":                 {"1"},
		"total":                 {"2"},
		"currency":              {"NOK"},
		"card_name":             {"Tampered Card"},
		"product_name":          {"Tampered Product"},
		"advance_amount":        {"1"},
		"remaining_balance":     {"1"},
		"payment_amount":        {"999999"},
		"bank_transfer_amount":  {"999999"},
		"transaction_reference": {"fake-ref"},
	}

	req := httptest.NewRequest(http.MethodPost, "/order", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "trusted-csrf"})

	recorder := httptest.NewRecorder()
	service := &fakeOrderService{}
	handler := &OrderHandler{service: service, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("error.html").Parse("{{ .message }}")))
	router.POST("/order", handler.CreateOrder)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status %d, got %d", http.StatusSeeOther, recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "/order/55/payment" {
		t.Fatalf("expected redirect to payment page, got %q", location)
	}

	if service.placeOrderInput.CardID != 7 {
		t.Fatalf("expected card id 7, got %d", service.placeOrderInput.CardID)
	}
	if service.placeOrderInput.Quantity != 250 {
		t.Fatalf("expected quantity 250, got %d", service.placeOrderInput.Quantity)
	}
	if service.placeOrderInput.Currency != "" {
		t.Fatalf("expected handler to ignore client currency, got %q", service.placeOrderInput.Currency)
	}

	inputType := reflect.TypeOf(orderapplication.PlaceOrderInput{})
	for _, fieldName := range []string{"Price", "Total", "ProductName", "CardName", "AdvanceAmount", "RemainingBalance", "PaymentAmount", "BankTransferAmount"} {
		if _, ok := inputType.FieldByName(fieldName); ok {
			t.Fatalf("trusted order input should not accept %s", fieldName)
		}
	}
}

func TestAdminUpdateOrderStatusRedirectsWhenPaymentVerificationRequired(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	form := url.Values{
		"csrf_token": {"trusted-csrf"},
		"status":     {"confirmed"},
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/orders/42/status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "trusted-csrf"})

	recorder := httptest.NewRecorder()
	service := &fakeOrderService{adminUpdateOrderStatusErr: orderapplication.ErrPaymentVerificationRequired}
	handler := &OrderHandler{service: service, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("error.html").Parse("{{ .message }}")))
	router.POST("/admin/orders/:id/status", handler.AdminUpdateOrderStatus)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status %d, got %d", http.StatusSeeOther, recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "/admin/orders/42?status_notice=payment_verification_required" {
		t.Fatalf("expected redirect back to admin order detail with inline notice, got %q", location)
	}
	if service.adminUpdateOrderStatusOrderID != 42 {
		t.Fatalf("expected order id 42, got %d", service.adminUpdateOrderStatusOrderID)
	}
	if service.adminUpdateOrderStatusStatus != "confirmed" {
		t.Fatalf("expected requested status confirmed, got %q", service.adminUpdateOrderStatusStatus)
	}
}

func TestAdminPaymentActionRequestReuploadRedirectsWithNotice(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	form := url.Values{
		"csrf_token": {"trusted-csrf"},
		"action":     {"request_reupload"},
		"admin_note": {"Please send a clearer receipt."},
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/orders/42/payment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "trusted-csrf"})

	recorder := httptest.NewRecorder()
	service := &fakeOrderService{}
	handler := &OrderHandler{service: service, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("error.html").Parse("{{ .message }}")))
	router.POST("/admin/orders/:id/payment", handler.AdminPaymentAction)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status %d, got %d", http.StatusSeeOther, recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "/admin/orders/42?payment_notice=payment_reupload_requested" {
		t.Fatalf("expected redirect with re-upload notice, got %q", location)
	}
	if service.adminProcessPaymentOrderID != 42 {
		t.Fatalf("expected order id 42, got %d", service.adminProcessPaymentOrderID)
	}
	if service.adminProcessPaymentAction != "request_reupload" {
		t.Fatalf("expected request_reupload action, got %q", service.adminProcessPaymentAction)
	}
}

func TestTrackOrderPageRedirectsToOrderStatus(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/track-order?order_id=75", nil)
	recorder := httptest.NewRecorder()
	handler := &OrderHandler{service: &fakeOrderService{}, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("error.html").Parse("{{ .message }}")))
	router.GET("/track-order", handler.TrackOrderPage)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status %d, got %d", http.StatusSeeOther, recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "/order/75" {
		t.Fatalf("expected redirect to order status page, got %q", location)
	}
}
