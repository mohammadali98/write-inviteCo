package orderpresentation

import (
	"bytes"
	"context"
	"html/template"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"writeandinviteco/inviteandco/order/orderapplication"
	"writeandinviteco/inviteandco/order/orderdomain"

	"github.com/gin-gonic/gin"
)

type fakeOrderService struct {
	prepareCustomizationInput     orderapplication.CustomizationInput
	prepareCustomizationSummary   *orderapplication.CustomizationSummary
	prepareCustomizationErr       error
	prepareOrderReviewInput       orderapplication.PlaceOrderInput
	placeOrderInput               orderapplication.PlaceOrderInput
	adminUpdateOrderStatusOrderID int64
	adminUpdateOrderStatusStatus  string
	adminUpdateOrderStatusErr     error
	adminProcessPaymentOrderID    int64
	adminProcessPaymentAction     string
	adminProcessPaymentNote       string
	adminProcessPaymentErr        error
	statusDetail                  *orderapplication.AdminOrderDetail
	submitPaymentProofOrderID     int64
	submitPaymentProofInput       orderapplication.PaymentProofInput
	submitPaymentProofErr         error
	listAdminOrdersInput          orderapplication.AdminOrderListInput
}

func (f *fakeOrderService) PrepareCustomization(ctx context.Context, input orderapplication.CustomizationInput) (*orderapplication.CustomizationSummary, error) {
	f.prepareCustomizationInput = input
	if f.prepareCustomizationSummary != nil || f.prepareCustomizationErr != nil {
		return f.prepareCustomizationSummary, f.prepareCustomizationErr
	}
	return &orderapplication.CustomizationSummary{CardCategory: "wedding-cards"}, nil
}

func (f *fakeOrderService) PrepareOrderReview(ctx context.Context, input orderapplication.PlaceOrderInput) (*orderapplication.OrderReview, error) {
	f.prepareOrderReviewInput = input
	return &orderapplication.OrderReview{
		Summary: &orderapplication.CustomizationSummary{
			CardID:       input.CardID,
			CardName:     "Trusted Card",
			CardCategory: "wedding-cards",
			Quantity:     input.Quantity,
			Currency:     "PKR",
			Side:         input.Side,
		},
		Input: input,
	}, nil
}

const fakeOrderToken = "11111111-1111-4111-8111-111111111111"

func (f *fakeOrderService) PlaceOrder(ctx context.Context, input orderapplication.PlaceOrderInput) (*orderapplication.PlaceOrderResult, error) {
	f.placeOrderInput = input
	return &orderapplication.PlaceOrderResult{OrderID: 55, OrderToken: fakeOrderToken}, nil
}

func (f *fakeOrderService) GetOrderStatusDetail(ctx context.Context, orderID int64) (*orderapplication.AdminOrderDetail, error) {
	if f.statusDetail != nil {
		return f.statusDetail, nil
	}
	return nil, nil
}

func (f *fakeOrderService) GetOrderStatusDetailByToken(ctx context.Context, token string) (*orderapplication.AdminOrderDetail, error) {
	if f.statusDetail != nil {
		return f.statusDetail, nil
	}
	return nil, nil
}

func (f *fakeOrderService) ListAdminOrders(ctx context.Context, input orderapplication.AdminOrderListInput) ([]*orderdomain.AdminOrder, error) {
	f.listAdminOrdersInput = input
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
	f.submitPaymentProofOrderID = orderID
	f.submitPaymentProofInput = input
	return f.submitPaymentProofErr
}

func (f *fakeOrderService) AdminProcessPayment(ctx context.Context, orderID int64, action string, adminNote string) error {
	f.adminProcessPaymentOrderID = orderID
	f.adminProcessPaymentAction = action
	f.adminProcessPaymentNote = adminNote
	return f.adminProcessPaymentErr
}

func TestCustomizePageReadsProductOptionsFromPostBodyAndIgnoresQueryPII(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	form := url.Values{
		"csrf_token":    {"trusted-csrf"},
		"card_id":       {"7"},
		"quantity":      {"250"},
		"foil_option":   {"foil"},
		"extra_inserts": {"3"},
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/customize?name=Query+Name&email=query@example.com&phone=09999999999&address=Query+Street&city=Query+City&postal_code=99999",
		strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "trusted-csrf"})

	recorder := httptest.NewRecorder()
	service := &fakeOrderService{}
	handler := &OrderHandler{service: service, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("customize.html").Parse("ok")))
	router.POST("/customize", handler.CustomizePage)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "ok") {
		t.Fatalf("expected customize template to render, got %q", body)
	}

	if service.prepareCustomizationInput.Name != "" {
		t.Fatalf("expected customer name to be ignored before checkout, got %q", service.prepareCustomizationInput.Name)
	}
	if service.prepareCustomizationInput.Email != "" {
		t.Fatalf("expected customer email to be ignored before checkout, got %q", service.prepareCustomizationInput.Email)
	}
	if service.prepareCustomizationInput.Phone != "" {
		t.Fatalf("expected customer phone to be ignored before checkout, got %q", service.prepareCustomizationInput.Phone)
	}
	if service.prepareCustomizationInput.Address != "" {
		t.Fatalf("expected customer address to be ignored before checkout, got %q", service.prepareCustomizationInput.Address)
	}
	if service.prepareCustomizationInput.City != "" {
		t.Fatalf("expected customer city to be ignored before checkout, got %q", service.prepareCustomizationInput.City)
	}
	if service.prepareCustomizationInput.PostalCode != "" {
		t.Fatalf("expected customer postal code to be ignored before checkout, got %q", service.prepareCustomizationInput.PostalCode)
	}
	if service.prepareCustomizationInput.CardID != 7 {
		t.Fatalf("expected card id 7, got %d", service.prepareCustomizationInput.CardID)
	}
	if service.prepareCustomizationInput.Quantity != 250 {
		t.Fatalf("expected quantity 250, got %d", service.prepareCustomizationInput.Quantity)
	}
	if service.prepareCustomizationInput.RequestedInserts != 3 {
		t.Fatalf("expected extra inserts 3, got %d", service.prepareCustomizationInput.RequestedInserts)
	}
}

func TestReviewPagePreservesMultipleRSVPValues(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	form := url.Values{
		"csrf_token":    {"trusted-csrf"},
		"card_id":       {"7"},
		"quantity":      {"250"},
		"foil_option":   {"foil"},
		"extra_inserts": {"1"},
		"name":          {"Aimen"},
		"email":         {"aimen@example.com"},
		"phone":         {"03001234567"},
		"address":       {"123 Karim Block"},
		"city":          {"Lahore"},
		"postal_code":   {"54000"},
		"side":          {"bride"},
		"rsvp_name":     {"Ali", "Sara"},
		"rsvp_phone":    {"03001234567", "03007654321"},
	}

	req := httptest.NewRequest(http.MethodPost, "/review", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "trusted-csrf"})

	recorder := httptest.NewRecorder()
	service := &fakeOrderService{}
	handler := &OrderHandler{service: service, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("review_order.html").Parse("ok")))
	router.POST("/review", handler.ReviewPage)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if service.prepareOrderReviewInput.RsvpName != "Ali\nSara" {
		t.Fatalf("expected multiple RSVP names to be preserved, got %q", service.prepareOrderReviewInput.RsvpName)
	}
	if service.prepareOrderReviewInput.RsvpPhone != "03001234567\n03007654321" {
		t.Fatalf("expected multiple RSVP phones to be preserved, got %q", service.prepareOrderReviewInput.RsvpPhone)
	}
}

func TestReviewPageAllowsOptionalBlankRSVPRows(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	form := url.Values{
		"csrf_token":    {"trusted-csrf"},
		"card_id":       {"7"},
		"quantity":      {"250"},
		"foil_option":   {"foil"},
		"extra_inserts": {"1"},
		"name":          {"Aimen"},
		"email":         {"aimen@example.com"},
		"phone":         {"03001234567"},
		"address":       {"123 Karim Block"},
		"city":          {"Lahore"},
		"postal_code":   {"54000"},
		"side":          {"bride"},
		"rsvp_name":     {"Ali", "", "Sara"},
		"rsvp_phone":    {"", "03084549268", "", "+923084549268"},
	}

	req := httptest.NewRequest(http.MethodPost, "/review", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "trusted-csrf"})

	recorder := httptest.NewRecorder()
	service := &fakeOrderService{}
	handler := &OrderHandler{service: service, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("review_order.html").Parse("ok")))
	router.POST("/review", handler.ReviewPage)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if service.prepareOrderReviewInput.RsvpName != "Ali\nSara" {
		t.Fatalf("expected blank RSVP names skipped, got %q", service.prepareOrderReviewInput.RsvpName)
	}
	if service.prepareOrderReviewInput.RsvpPhone != "03084549268\n+923084549268" {
		t.Fatalf("expected blank RSVP phones skipped, got %q", service.prepareOrderReviewInput.RsvpPhone)
	}
}

func TestReviewPageRendersAfterCheckoutWithFullHiddenFields(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	form := url.Values{
		"csrf_token":           {"trusted-csrf"},
		"card_id":              {"7"},
		"quantity":             {"50"},
		"foil_option":          {"foil"},
		"extra_inserts":        {"1"},
		"side":                 {"bride"},
		"name":                 {"Aimen"},
		"email":                {"aimen@example.com"},
		"phone":                {"03001234567"},
		"address":              {"123 Karim Block"},
		"city":                 {"Lahore"},
		"postal_code":          {"54000"},
		"bride_name":           {"Bride"},
		"groom_name":           {"Groom"},
		"bride_father_name":    {"Bride Parents"},
		"groom_father_name":    {"Groom Parents"},
		"mehndi_date":          {"2026-06-01"},
		"mehndi_day":           {"Monday"},
		"mehndi_time_type":     {"evening"},
		"mehndi_time":          {"18:00"},
		"mehndi_dinner_time":   {"21:00"},
		"mehndi_venue_name":    {"Mehndi Hall"},
		"mehndi_venue_address": {"Mehndi Address"},
		"baraat_date":          {"2026-06-02"},
		"baraat_day":           {"Tuesday"},
		"baraat_time_type":     {"night"},
		"baraat_time":          {"20:00"},
		"baraat_dinner_time":   {"22:00"},
		"baraat_arrival_time":  {"19:00"},
		"rukhsati_time":        {"23:00"},
		"baraat_venue_name":    {"Baraat Hall"},
		"baraat_venue_address": {"Baraat Address"},
		"nikkah_date":          {"2026-06-03"},
		"nikkah_day":           {"Wednesday"},
		"nikkah_time_type":     {"evening"},
		"nikkah_time":          {"17:00"},
		"nikkah_dinner_time":   {"20:00"},
		"nikkah_venue_name":    {"Nikkah Hall"},
		"nikkah_venue_address": {"Nikkah Address"},
		"walima_date":          {"2026-06-04"},
		"walima_day":           {"Thursday"},
		"walima_time_type":     {"night"},
		"walima_time":          {"19:00"},
		"walima_dinner_time":   {"21:00"},
		"walima_venue_name":    {"Walima Hall"},
		"walima_venue_address": {"Walima Address"},
		"reception_time":       {"18:30"},
		"rsvp_name":            {"Ali", "Sara"},
		"rsvp_phone":           {"03001234567", "03007654321"},
		"notes":                {"Please use formal wording."},
	}

	req := httptest.NewRequest(http.MethodPost, "/review", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "trusted-csrf"})

	recorder := httptest.NewRecorder()
	service := &fakeOrderService{}
	handler := &OrderHandler{service: service, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("review_order.html").Parse("ok")))
	router.POST("/review", handler.ReviewPage)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if service.prepareOrderReviewInput.RequestedInserts != 1 {
		t.Fatalf("expected extra inserts 1, got %d", service.prepareOrderReviewInput.RequestedInserts)
	}
	if service.prepareOrderReviewInput.BrideFatherName != "Bride Parents" {
		t.Fatalf("expected bride parents field to be preserved, got %q", service.prepareOrderReviewInput.BrideFatherName)
	}
	if service.prepareOrderReviewInput.NikkahVenueAddress != "Nikkah Address" {
		t.Fatalf("expected nikkah venue address to be preserved, got %q", service.prepareOrderReviewInput.NikkahVenueAddress)
	}
	if service.prepareOrderReviewInput.RsvpName != "Ali\nSara" {
		t.Fatalf("expected RSVP names to be preserved, got %q", service.prepareOrderReviewInput.RsvpName)
	}
}

func TestReviewPageMissingRequiredProductFieldsReturnsBadRequest(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	form := url.Values{
		"csrf_token": {"trusted-csrf"},
		"quantity":   {"50"},
	}

	req := httptest.NewRequest(http.MethodPost, "/review", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "trusted-csrf"})

	recorder := httptest.NewRecorder()
	handler := &OrderHandler{service: &fakeOrderService{}, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("error.html").Parse("{{ .title }}")))
	router.POST("/review", handler.ReviewPage)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "Invalid Card") {
		t.Fatalf("expected invalid card error, got %q", body)
	}
}

func TestPersonalizationLabelsAreVisualOnlyParentsWording(t *testing.T) {
	t.Parallel()

	customize, err := os.ReadFile("../../templates/customize.html")
	if err != nil {
		t.Fatalf("read customize template: %v", err)
	}
	customizeBody := string(customize)

	for _, unwanted := range []string{
		"(Optional)",
		"Father Name",
		"Bride's Father Name",
		"Groom's Father Name",
	} {
		if strings.Contains(customizeBody, unwanted) {
			t.Fatalf("expected customize labels to avoid %q", unwanted)
		}
	}
	for _, want := range []string{
		"Bride Name",
		"Groom Name",
		"Bride's Parents Name",
		"Groom's Parents Name",
		`name="bride_father_name"`,
		`name="groom_father_name"`,
	} {
		if !strings.Contains(customizeBody, want) {
			t.Fatalf("expected customize template to contain %q", want)
		}
	}

	for _, path := range []string{
		"../../templates/review_order.html",
		"../../templates/admin_order_detail.html",
		"../../templates/order-status.html",
	} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read template %s: %v", path, err)
		}
		text := string(body)
		if strings.Contains(text, "Bride's Father") || strings.Contains(text, "Groom's Father") {
			t.Fatalf("expected %s to use parents wording", path)
		}
	}
}

func TestCheckoutCustomerFormPostsPIIToReviewOnly(t *testing.T) {
	t.Parallel()

	checkout, err := os.ReadFile("../../templates/checkout.html")
	if err != nil {
		t.Fatalf("read checkout template: %v", err)
	}
	body := string(checkout)

	if !strings.Contains(body, `form action="/review" method="post"`) {
		t.Fatalf("expected checkout customer form to post to /review")
	}
	if strings.Contains(body, `method="get"`) || strings.Contains(body, `method="GET"`) {
		t.Fatalf("checkout template must not submit customer information with GET")
	}
	for _, field := range []string{"name", "email", "phone", "address", "city", "postal_code"} {
		if !strings.Contains(body, `name="`+field+`"`) {
			t.Fatalf("expected checkout form to keep customer field %q in POST body", field)
		}
	}
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
	if location := recorder.Header().Get("Location"); location != "/order/"+fakeOrderToken+"/payment" {
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

func TestAdminOrdersThreadsQueryFiltersToService(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/admin/orders?status=confirmed&payment_status=awaiting_verification&search=ali&date_range=week", nil)
	recorder := httptest.NewRecorder()
	service := &fakeOrderService{}
	handler := &OrderHandler{service: service, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("admin_orders.html").Parse("ok")))
	router.GET("/admin/orders", handler.AdminOrders)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%q", recorder.Code, recorder.Body.String())
	}
	if service.listAdminOrdersInput.OrderStatus != "confirmed" {
		t.Fatalf("expected status filter 'confirmed', got %q", service.listAdminOrdersInput.OrderStatus)
	}
	if service.listAdminOrdersInput.PaymentStatus != "awaiting_verification" {
		t.Fatalf("expected payment_status filter 'awaiting_verification', got %q", service.listAdminOrdersInput.PaymentStatus)
	}
	if service.listAdminOrdersInput.Search != "ali" {
		t.Fatalf("expected search filter 'ali', got %q", service.listAdminOrdersInput.Search)
	}
	if service.listAdminOrdersInput.DateRange != "week" {
		t.Fatalf("expected date_range filter 'week', got %q", service.listAdminOrdersInput.DateRange)
	}
}

func TestAdminOrdersWithNoQueryParamsPassesEmptyFilter(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/admin/orders", nil)
	recorder := httptest.NewRecorder()
	service := &fakeOrderService{}
	handler := &OrderHandler{service: service, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("admin_orders.html").Parse("ok")))
	router.GET("/admin/orders", handler.AdminOrders)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%q", recorder.Code, recorder.Body.String())
	}
	empty := orderapplication.AdminOrderListInput{}
	if service.listAdminOrdersInput != empty {
		t.Fatalf("expected an empty filter for the default view, got %+v", service.listAdminOrdersInput)
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

func TestPaymentSubmittedAmountBelowExpectedDetectsUnderpaidProof(t *testing.T) {
	t.Parallel()

	submitted := int64(13000)
	payment := &orderdomain.OrderPayment{SubmittedAmount: &submitted}
	summary := orderapplication.PaymentAmountSummary{AdvanceAmount: 16500}

	if !paymentSubmittedAmountBelowExpected(payment, summary) {
		t.Fatalf("expected submitted amount below required advance to be detected")
	}

	submitted = 16500
	if paymentSubmittedAmountBelowExpected(payment, summary) {
		t.Fatalf("expected exact submitted amount to be accepted")
	}
}

func TestTrackOrderPageRedirectsToOrderStatus(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/track-order?token="+fakeOrderToken, nil)
	recorder := httptest.NewRecorder()
	handler := &OrderHandler{service: &fakeOrderService{}, paymentProofDir: t.TempDir()}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("error.html").Parse("{{ .message }}")))
	router.GET("/track-order", handler.TrackOrderPage)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status %d, got %d", http.StatusSeeOther, recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "/order/"+fakeOrderToken {
		t.Fatalf("expected redirect to order status page, got %q", location)
	}
}

func TestSubmitPaymentProofDoesNotTrustSubmittedAmountField(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	mustWriteMultipartField(t, writer, "csrf_token", "trusted-csrf")
	mustWriteMultipartField(t, writer, "sender_name", "Ali Sender")
	mustWriteMultipartField(t, writer, "transaction_reference", "TXN-123")
	mustWriteMultipartField(t, writer, "submitted_amount", "13000")
	fileWriter, err := writer.CreateFormFile("payment_proof", "receipt.pdf")
	if err != nil {
		t.Fatalf("create payment proof part: %v", err)
	}
	if _, err := fileWriter.Write([]byte("%PDF-1.4\n% test receipt\n")); err != nil {
		t.Fatalf("write payment proof: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/order/"+fakeOrderToken+"/payment-proof", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "trusted-csrf"})

	recorder := httptest.NewRecorder()
	service := &fakeOrderService{
		statusDetail: &orderapplication.AdminOrderDetail{
			Order: &orderdomain.Order{
				ID:          42,
				TotalPrice:  33000,
				Currency:    "PKR",
				PublicToken: fakeOrderToken,
			},
			Payment: &orderdomain.OrderPayment{
				OrderID:        42,
				PaymentMethod:  orderdomain.BankTransferPaymentMethod,
				PaymentStatus:  orderdomain.PendingPaymentStatus,
				ExpectedAmount: 16500,
			},
		},
	}
	mockCloudinary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"secure_url":"https://res.cloudinary.com/test-cloud/image/upload/v1/payment-proofs/42/abc123.pdf","public_id":"payment-proofs/42/abc123"}`))
	}))
	defer mockCloudinary.Close()

	handler := &OrderHandler{
		service:         service,
		paymentProofDir: t.TempDir(),
		proofUploader: &cloudinaryProofUploader{
			cloudName:  "test-cloud",
			apiKey:     "test-key",
			apiSecret:  "test-secret",
			httpClient: mockCloudinary.Client(),
			apiBaseURL: mockCloudinary.URL,
		},
	}

	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("error.html").Parse("{{ .message }}")))
	router.POST("/order/:token/payment-proof", handler.SubmitPaymentProof)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status %d, got %d body=%q", http.StatusSeeOther, recorder.Code, recorder.Body.String())
	}
	if service.submitPaymentProofOrderID != 42 {
		t.Fatalf("expected proof submission for order 42, got %d", service.submitPaymentProofOrderID)
	}
	if service.submitPaymentProofInput.SubmittedAmount != 0 {
		t.Fatalf("expected presentation layer to ignore tampered submitted_amount, got %d", service.submitPaymentProofInput.SubmittedAmount)
	}
	if !strings.HasPrefix(service.submitPaymentProofInput.ProofFilePath, "https://res.cloudinary.com/") {
		t.Fatalf("expected uploaded proof path to be a Cloudinary URL, got %q", service.submitPaymentProofInput.ProofFilePath)
	}
}

func TestReviewPageDoesNotExposeFrontendTrustDisclaimer(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("../../templates/review_order.html")
	if err != nil {
		t.Fatalf("read review template: %v", err)
	}

	if strings.Contains(string(body), "frontend display is only for review") ||
		strings.Contains(string(body), "recalculated and saved by the server") {
		t.Fatalf("review page should not show internal pricing trust disclaimer to customers")
	}
}

func TestOrderStatusTemplateShowsCustomizationAndRSVPDetails(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("../../templates/order-status.html")
	if err != nil {
		t.Fatalf("read order status template: %v", err)
	}
	text := string(body)

	for _, want := range []string{
		"Customization Details",
		"Bride's Parents Name",
		"Groom's Parents Name",
		"RSVP Name(s)",
		"RSVP Phone Number(s)",
		"Extra Inserts / Card",
		"Remaining Balance",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected order status template to contain %q", want)
		}
	}
}

func mustWriteMultipartField(t *testing.T, writer *multipart.Writer, field string, value string) {
	t.Helper()
	if err := writer.WriteField(field, value); err != nil {
		t.Fatalf("write multipart field %s: %v", field, err)
	}
}
