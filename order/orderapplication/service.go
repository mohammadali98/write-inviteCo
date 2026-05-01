package orderapplication

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"writeandinviteco/inviteandco/card/carddomain"
	"writeandinviteco/inviteandco/customer/customerdomain"
	customerwriter "writeandinviteco/inviteandco/customer/customerinfrastructure/postgres/writer"
	"writeandinviteco/inviteandco/order/orderdomain"
	orderwriter "writeandinviteco/inviteandco/order/orderinfrastructure/postgres/writer"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCardNotFound    = errors.New("card not found")
	ErrInvalidInput    = errors.New("invalid input")
	errPricingOverflow = errors.New("pricing overflow")

	pakistanPhonePattern = regexp.MustCompile(`^(?:03\d{9}|\+923\d{9})$`)
)

const (
	maxQuantity            = 5000
	maxRequestedInserts    = 20
	maxCustomerNameLength  = 120
	maxEmailLength         = 254
	maxPhoneLength         = 20
	maxAddressLength       = 500
	maxCityLength          = 100
	maxPostalCodeLength    = 20
	maxPersonNameLength    = 120
	maxVenueNameLength     = 150
	maxVenueAddressLength  = 500
	maxNotesLength         = 2000
	maxBidBoxLabelLength   = 120
	maxBidBoxDetailsLength = 1000
)

type MinOrderError struct {
	MinOrder int64
}

func (e MinOrderError) Error() string {
	return fmt.Sprintf("minimum order quantity is %d", e.MinOrder)
}

type CustomizationInput struct {
	CardID           int64
	Quantity         int64
	Currency         string
	FoilOption       string
	RequestedInserts int64

	Name       string
	Email      string
	Phone      string
	Address    string
	City       string
	PostalCode string
}

type CustomizationSummary struct {
	CardID           int64
	CardName         string
	CardImage        string
	CardCategory     string
	Quantity         int64
	Currency         string
	FoilOption       string
	FoilLabel        string
	Side             string
	RequestedInserts int64
	IncludedInserts  int64
	ExtraInserts     int64
	UnitPrice        int64
	InsertPrice      int64
	ExtraInsertCost  int64
	PerCardTotal     int64
	TotalPrice       int64
	MinOrder         int64

	Name       string
	Email      string
	Phone      string
	Address    string
	City       string
	PostalCode string
}

type PlaceOrderInput struct {
	CardID           int64
	Quantity         int64
	Currency         string
	FoilOption       string
	RequestedInserts int64

	Name       string
	Email      string
	Phone      string
	Address    string
	City       string
	PostalCode string

	BidBoxTopLabel     string
	BidBoxCoupleName   string
	BidBoxEventDate    string
	BidBoxDetails      string
	Side               string
	BrideName          string
	GroomName          string
	BrideFatherName    string
	GroomFatherName    string
	MehndiDate         string
	MehndiDay          string
	MehndiTimeType     string
	MehndiTime         string
	MehndiDinnerTime   string
	MehndiVenueName    string
	MehndiVenueAddress string
	BaraatDate         string
	BaraatDay          string
	BaraatTimeType     string
	BaraatTime         string
	BaraatDinnerTime   string
	BaraatArrivalTime  string
	RukhsatiTime       string
	BaraatVenueName    string
	BaraatVenueAddress string
	NikkahDate         string
	NikkahDay          string
	NikkahTimeType     string
	NikkahTime         string
	NikkahDinnerTime   string
	NikkahVenueName    string
	NikkahVenueAddress string
	WalimaDate         string
	WalimaDay          string
	WalimaTimeType     string
	WalimaTime         string
	WalimaDinnerTime   string
	WalimaVenueName    string
	WalimaVenueAddress string
	ReceptionTime      string
	RsvpName           string
	RsvpPhone          string
	Notes              string
}

type PlaceOrderResult struct {
	OrderID      int64
	CustomerName string
	CardName     string
	Quantity     int64
	TotalPrice   int64
	Currency     string
}

type OrderReview struct {
	Summary  *CustomizationSummary
	Input    PlaceOrderInput
	IsBidBox bool
}

type Service struct {
	db             *pgxpool.Pool
	cardRepo       carddomain.CardRepo
	customerRepo   customerdomain.CustomerReader
	orderRepo      orderdomain.OrderRepo
	customerWriter *customerwriter.Queries
	orderWriter    *orderwriter.Queries
	emailSender    EmailSender
	adminEmail     string
}

type EmailSender interface {
	SendOrderEmail(ctx context.Context, to string, subject string, body string) error
}

func NewService(
	db *pgxpool.Pool,
	cardRepo carddomain.CardRepo,
	customerRepo customerdomain.CustomerReader,
	orderRepo orderdomain.OrderRepo,
	customerWriter *customerwriter.Queries,
	orderWriter *orderwriter.Queries,
	emailSender EmailSender,
	adminEmail string,
) *Service {
	return &Service{
		db:             db,
		cardRepo:       cardRepo,
		customerRepo:   customerRepo,
		orderRepo:      orderRepo,
		customerWriter: customerWriter,
		orderWriter:    orderWriter,
		emailSender:    emailSender,
		adminEmail:     strings.TrimSpace(adminEmail),
	}
}

type AdminOrderDetail struct {
	Order    *orderdomain.Order
	Customer *customerdomain.Customer
	Details  *orderdomain.OrderDetail
	Payment  *orderdomain.OrderPayment
}

func (s *Service) ListAdminOrders(ctx context.Context) ([]*orderdomain.AdminOrder, error) {
	return s.orderRepo.GetAdminOrders(ctx)
}

func (s *Service) GetAdminOrderDetail(ctx context.Context, orderID int64) (*AdminOrderDetail, error) {
	return s.getOrderDetail(ctx, orderID)
}

func (s *Service) GetOrderStatusDetail(ctx context.Context, orderID int64) (*AdminOrderDetail, error) {
	return s.getOrderDetail(ctx, orderID)
}

func (s *Service) getOrderDetail(ctx context.Context, orderID int64) (*AdminOrderDetail, error) {
	if orderID <= 0 {
		return nil, ErrInvalidInput
	}

	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	var customer *customerdomain.Customer
	if order.CustomerID > 0 {
		customer, err = s.customerRepo.GetCustomerByID(ctx, order.CustomerID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
		}
	}

	details, err := s.orderRepo.GetOrderDetailByOrderID(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	payment, err := s.orderRepo.GetOrderPaymentByOrderID(ctx, order.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	return &AdminOrderDetail{
		Order:    order,
		Customer: customer,
		Details:  details,
		Payment:  payment,
	}, nil
}

func (s *Service) AdminUpdateOrderStatus(ctx context.Context, orderID int64, statusRaw string) error {
	log.Printf("ADMIN STATUS FLOW START: order_id=%d requested_status=%q", orderID, statusRaw)

	if orderID <= 0 {
		log.Printf("ADMIN STATUS FLOW ERROR: invalid order id=%d", orderID)
		return ErrInvalidInput
	}

	newStatus, err := normalizeOrderStatus(statusRaw)
	if err != nil {
		log.Printf("ADMIN STATUS FLOW ERROR: order_id=%d invalid status=%q err=%v", orderID, statusRaw, err)
		return err
	}

	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		log.Printf("ADMIN STATUS FLOW ERROR: order_id=%d load order failed err=%v", orderID, err)
		return err
	}
	log.Printf(
		"ADMIN STATUS FLOW ORDER LOADED: order_id=%d current_status=%q category=%q type=%s customer_id=%d",
		orderID,
		order.Status,
		order.CardCategory,
		orderTypeLabel(order, nil),
		order.CustomerID,
	)

	details, detailsErr := s.orderRepo.GetOrderDetailByOrderID(ctx, order.ID)
	if detailsErr != nil {
		log.Printf("ADMIN STATUS FLOW ERROR: order_id=%d load details failed err=%v", orderID, detailsErr)
		return detailsErr
	}

	payment, paymentErr := s.orderRepo.GetOrderPaymentByOrderID(ctx, order.ID)
	if paymentErr != nil && !errors.Is(paymentErr, pgx.ErrNoRows) {
		log.Printf("ADMIN STATUS FLOW ERROR: order_id=%d load payment failed err=%v", orderID, paymentErr)
		return paymentErr
	}
	log.Printf(
		"ADMIN STATUS FLOW DETAILS LOADED: order_id=%d type=%s has_bid_box_fields=%t has_details=%t payment_status=%q",
		orderID,
		orderTypeLabel(order, details),
		hasBidBoxFields(details),
		details != nil,
		paymentStatusLogValue(payment),
	)

	if (newStatus == orderdomain.ConfirmedOrderStatus || newStatus == orderdomain.CompletedOrderStatus) &&
		(payment == nil || payment.PaymentStatus != orderdomain.VerifiedPaymentStatus) {
		log.Printf("ADMIN STATUS FLOW ERROR: order_id=%d status=%q payment not verified", orderID, newStatus)
		return ErrPaymentVerificationRequired
	}

	log.Printf("ADMIN STATUS FLOW BEFORE DB UPDATE: order_id=%d current_status=%q new_status=%q", orderID, order.Status, newStatus)
	if err := s.orderRepo.UpdateOrderStatus(ctx, orderID, newStatus); err != nil {
		log.Printf("ADMIN STATUS FLOW ERROR: order_id=%d update status failed err=%v", orderID, err)
		return err
	}
	log.Printf("ADMIN STATUS FLOW AFTER DB UPDATE: order_id=%d new_status=%q", orderID, newStatus)

	log.Printf(
		"ADMIN STATUS FLOW BEFORE EMAIL: order_id=%d type=%s customer_id=%d new_status=%q",
		orderID,
		orderTypeLabel(order, details),
		order.CustomerID,
		newStatus,
	)
	s.sendOrderStatusEmailAsync(order, newStatus)

	return nil
}

func (s *Service) PrepareCustomization(ctx context.Context, input CustomizationInput) (*CustomizationSummary, error) {
	input = sanitizeCustomizationInput(input)
	if err := validateCustomerFields(input.Name, input.Email, input.Phone, input.Address, input.City, input.PostalCode); err != nil {
		return nil, err
	}

	pricing, err := s.calculatePricing(ctx, input.CardID, input.Quantity, input.Currency, input.FoilOption, input.RequestedInserts)
	if err != nil {
		return nil, err
	}

	return buildCustomizationSummary(pricing, input.Name, input.Email, input.Phone, input.Address, input.City, input.PostalCode, "bride"), nil
}

func (s *Service) PrepareOrderReview(ctx context.Context, input PlaceOrderInput) (*OrderReview, error) {
	input = sanitizePlaceOrderInput(input)

	if err := validateCustomerFields(input.Name, input.Email, input.Phone, input.Address, input.City, input.PostalCode); err != nil {
		return nil, err
	}

	side, err := parseSide(input.Side)
	if err != nil {
		return nil, err
	}
	input.Side = side

	pricing, err := s.calculatePricing(ctx, input.CardID, input.Quantity, input.Currency, input.FoilOption, input.RequestedInserts)
	if err != nil {
		return nil, err
	}
	if err := validateCustomizationFields(input, pricing.card.Category); err != nil {
		return nil, err
	}

	return &OrderReview{
		Summary:  buildCustomizationSummary(pricing, input.Name, input.Email, input.Phone, input.Address, input.City, input.PostalCode, input.Side),
		Input:    input,
		IsBidBox: isBidBoxCategory(pricing.card.Category),
	}, nil
}

func (s *Service) PlaceOrder(ctx context.Context, input PlaceOrderInput) (*PlaceOrderResult, error) {
	input = sanitizePlaceOrderInput(input)

	if err := validateCustomerFields(input.Name, input.Email, input.Phone, input.Address, input.City, input.PostalCode); err != nil {
		return nil, err
	}
	side, err := parseSide(input.Side)
	if err != nil {
		return nil, err
	}
	input.Side = side

	pricing, err := s.calculatePricing(ctx, input.CardID, input.Quantity, input.Currency, input.FoilOption, input.RequestedInserts)
	if err != nil {
		return nil, err
	}
	paymentAmounts := BuildPaymentAmountSummary(pricing.totalPrice, 0)
	if err := validateCustomizationFields(input, pricing.card.Category); err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	customerRow, err := s.customerWriter.WithTx(tx).CreateCustomer(ctx, customerwriter.CreateCustomerParams{
		Name:       input.Name,
		Email:      stringPtr(input.Email),
		Phone:      stringPtr(input.Phone),
		Address:    stringPtr(input.Address),
		City:       stringPtr(input.City),
		PostalCode: stringPtr(input.PostalCode),
	})
	if err != nil {
		return nil, err
	}

	status := string(orderdomain.PendingOrderStatus)
	orderRow, err := s.orderWriter.WithTx(tx).CreateOrder(ctx, orderwriter.CreateOrderParams{
		CustomerID: &customerRow.ID,
		CardID:     &pricing.card.ID,
		Quantity:   pricing.quantity,
		TotalPrice: pricing.totalPrice,
		Status:     &status,
		Currency:   pricing.currency,
	})
	if err != nil {
		return nil, err
	}

	paymentMethod := string(orderdomain.BankTransferPaymentMethod)
	paymentStatus := string(orderdomain.PendingPaymentStatus)
	if _, err := s.orderWriter.WithTx(tx).CreateOrderPayment(ctx, orderwriter.CreateOrderPaymentParams{
		OrderID:        orderRow.ID,
		PaymentMethod:  paymentMethod,
		PaymentStatus:  paymentStatus,
		ExpectedAmount: paymentAmounts.AdvanceAmount,
	}); err != nil {
		return nil, err
	}

	_, err = s.orderWriter.WithTx(tx).CreateOrderDetail(ctx, orderwriter.CreateOrderDetailParams{
		OrderID:            orderRow.ID,
		Side:               input.Side,
		TopLabel:           nullableString(input.BidBoxTopLabel),
		CoupleName:         nullableString(input.BidBoxCoupleName),
		BidBoxEventDate:    input.BidBoxEventDate,
		BidBoxDetails:      nullableString(input.BidBoxDetails),
		BrideName:          nullableString(input.BrideName),
		GroomName:          nullableString(input.GroomName),
		BrideFatherName:    nullableString(input.BrideFatherName),
		GroomFatherName:    nullableString(input.GroomFatherName),
		MehndiDate:         input.MehndiDate,
		MehndiDay:          nullableString(input.MehndiDay),
		MehndiTimeType:     nullableString(input.MehndiTimeType),
		MehndiTime:         input.MehndiTime,
		MehndiDinnerTime:   input.MehndiDinnerTime,
		MehndiVenueName:    nullableString(input.MehndiVenueName),
		MehndiVenueAddress: nullableString(input.MehndiVenueAddress),
		BaraatDate:         input.BaraatDate,
		BaraatDay:          nullableString(input.BaraatDay),
		BaraatTimeType:     nullableString(input.BaraatTimeType),
		BaraatTime:         input.BaraatTime,
		BaraatDinnerTime:   input.BaraatDinnerTime,
		BaraatArrivalTime:  input.BaraatArrivalTime,
		RukhsatiTime:       input.RukhsatiTime,
		BaraatVenueName:    nullableString(input.BaraatVenueName),
		BaraatVenueAddress: nullableString(input.BaraatVenueAddress),
		NikkahDate:         input.NikkahDate,
		NikkahDay:          nullableString(input.NikkahDay),
		NikkahTimeType:     nullableString(input.NikkahTimeType),
		NikkahTime:         input.NikkahTime,
		NikkahDinnerTime:   input.NikkahDinnerTime,
		NikkahVenueName:    nullableString(input.NikkahVenueName),
		NikkahVenueAddress: nullableString(input.NikkahVenueAddress),
		WalimaDate:         input.WalimaDate,
		WalimaDay:          nullableString(input.WalimaDay),
		WalimaTimeType:     nullableString(input.WalimaTimeType),
		WalimaTime:         input.WalimaTime,
		WalimaDinnerTime:   input.WalimaDinnerTime,
		WalimaVenueName:    nullableString(input.WalimaVenueName),
		WalimaVenueAddress: nullableString(input.WalimaVenueAddress),
		ReceptionTime:      input.ReceptionTime,
		RsvpName:           input.RsvpName,
		RsvpPhone:          input.RsvpPhone,
		Notes:              nullableString(input.Notes),
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	cardName := pricing.card.Name
	if cardName == "" {
		cardName = "Invitation Card"
	}

	s.sendOrderCreatedEmailsAsync(orderEmailPayload{
		OrderID:       orderRow.ID,
		CustomerName:  input.Name,
		CustomerEmail: input.Email,
		CustomerPhone: input.Phone,
		ProductName:   cardName,
		Quantity:      pricing.quantity,
		TotalPrice:    pricing.totalPrice,
		AdvanceAmount: paymentAmounts.AdvanceAmount,
		Remaining:     paymentAmounts.RemainingBalance,
		Currency:      pricing.currency,
		PaymentStatus: orderdomain.PendingPaymentStatus,
		Status:        orderdomain.PendingOrderStatus,
	})

	return &PlaceOrderResult{
		OrderID:      orderRow.ID,
		CustomerName: input.Name,
		CardName:     cardName,
		Quantity:     pricing.quantity,
		TotalPrice:   pricing.totalPrice,
		Currency:     pricing.currency,
	}, nil
}

func buildCustomizationSummary(pricing *pricingResult, name string, email string, phone string, address string, city string, postalCode string, side string) *CustomizationSummary {
	return &CustomizationSummary{
		CardID:           pricing.card.ID,
		CardName:         pricing.card.Name,
		CardImage:        pricing.card.Image,
		CardCategory:     pricing.card.Category,
		Quantity:         pricing.quantity,
		Currency:         pricing.currency,
		FoilOption:       pricing.foilOption,
		FoilLabel:        pricing.foilLabel,
		Side:             side,
		RequestedInserts: pricing.requestedInserts,
		IncludedInserts:  pricing.includedInserts,
		ExtraInserts:     pricing.extraInserts,
		UnitPrice:        pricing.basePrice,
		InsertPrice:      pricing.insertPrice,
		ExtraInsertCost:  pricing.extraInsertCost,
		PerCardTotal:     pricing.perCardPrice,
		TotalPrice:       pricing.totalPrice,
		MinOrder:         pricing.minOrder,
		Name:             name,
		Email:            email,
		Phone:            phone,
		Address:          address,
		City:             city,
		PostalCode:       postalCode,
	}
}

type pricingResult struct {
	card             *carddomain.Card
	quantity         int64
	minOrder         int64
	currency         string
	foilOption       string
	foilLabel        string
	basePrice        int64
	insertPrice      int64
	requestedInserts int64
	includedInserts  int64
	extraInserts     int64
	extraInsertCost  int64
	perCardPrice     int64
	totalPrice       int64
}

func (s *Service) calculatePricing(ctx context.Context, cardID int64, quantity int64, currency string, foilOption string, requestedInserts int64) (*pricingResult, error) {
	if cardID <= 0 || quantity < 1 || quantity > maxQuantity || requestedInserts < 0 || requestedInserts > maxRequestedInserts {
		return nil, ErrInvalidInput
	}

	card, err := s.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCardNotFound
		}
		return nil, fmt.Errorf("get card: %w", err)
	}

	minOrder := int64(card.MinOrder)
	if minOrder < 1 {
		minOrder = 1
	}
	if quantity < minOrder {
		return nil, MinOrderError{MinOrder: minOrder}
	}

	// Bank transfer is currently PKR-only. Any client-submitted currency is ignored.
	currency = "PKR"
	foilOption, err = parseFoilOption(foilOption)
	if err != nil {
		return nil, err
	}

	priceFoil := card.PriceFoilPKR
	priceNofoil := card.PriceNofoilPKR
	insertPrice := card.InsertPricePKR
	if priceFoil < 0 || priceNofoil < 0 || insertPrice < 0 {
		return nil, fmt.Errorf("invalid pricing configured for card %d", card.ID)
	}
	if priceNofoil == 0 {
		priceNofoil = priceFoil
	}

	basePrice := priceFoil
	foilLabel := "With Foil"
	if foilOption == "nofoil" {
		basePrice = priceNofoil
		foilLabel = "Without Foil"
	}
	if priceFoil == priceNofoil {
		foilLabel = "Flat Rate"
	}

	includedInserts := int64(card.IncludedInserts)
	if includedInserts < 0 {
		includedInserts = 0
	}
	extraInserts := requestedInserts - includedInserts
	if extraInserts < 0 {
		extraInserts = 0
	}

	extraInsertCost, err := safeMultiplyInt64(extraInserts, insertPrice)
	if err != nil {
		return nil, fmt.Errorf("calculate extra insert cost: %w", err)
	}
	perCardPrice, err := safeAddInt64(basePrice, extraInsertCost)
	if err != nil {
		return nil, fmt.Errorf("calculate per-card price: %w", err)
	}
	totalPrice, err := safeMultiplyInt64(perCardPrice, quantity)
	if err != nil {
		return nil, fmt.Errorf("calculate total price: %w", err)
	}

	return &pricingResult{
		card:             card,
		quantity:         quantity,
		minOrder:         minOrder,
		currency:         currency,
		foilOption:       foilOption,
		foilLabel:        foilLabel,
		basePrice:        basePrice,
		insertPrice:      insertPrice,
		requestedInserts: requestedInserts,
		includedInserts:  includedInserts,
		extraInserts:     extraInserts,
		extraInsertCost:  extraInsertCost,
		perCardPrice:     perCardPrice,
		totalPrice:       totalPrice,
	}, nil
}

func sanitizeCustomizationInput(input CustomizationInput) CustomizationInput {
	input.Currency = normalizeCurrency(input.Currency)
	input.FoilOption = normalizeFoilOption(input.FoilOption)
	input.Name = sanitizeSingleLine(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Phone = normalizePhone(input.Phone)
	input.Address = sanitizeSingleLine(input.Address)
	input.City = sanitizeSingleLine(input.City)
	input.PostalCode = sanitizeSingleLine(input.PostalCode)
	return input
}

func sanitizePlaceOrderInput(input PlaceOrderInput) PlaceOrderInput {
	input.Currency = normalizeCurrency(input.Currency)
	input.FoilOption = normalizeFoilOption(input.FoilOption)
	input.Side = normalizeSide(input.Side)
	input.Name = sanitizeSingleLine(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Phone = normalizePhone(input.Phone)
	input.Address = sanitizeSingleLine(input.Address)
	input.City = sanitizeSingleLine(input.City)
	input.PostalCode = sanitizeSingleLine(input.PostalCode)
	input.BidBoxTopLabel = sanitizeSingleLine(input.BidBoxTopLabel)
	input.BidBoxCoupleName = sanitizeSingleLine(input.BidBoxCoupleName)
	input.BidBoxEventDate = strings.TrimSpace(input.BidBoxEventDate)
	input.BidBoxDetails = sanitizeMultiline(input.BidBoxDetails)
	input.BrideName = sanitizeSingleLine(input.BrideName)
	input.GroomName = sanitizeSingleLine(input.GroomName)
	input.BrideFatherName = sanitizeSingleLine(input.BrideFatherName)
	input.GroomFatherName = sanitizeSingleLine(input.GroomFatherName)
	input.MehndiDate = strings.TrimSpace(input.MehndiDate)
	input.MehndiDay = normalizeOptionalDay(input.MehndiDay)
	input.MehndiTimeType = normalizeOptionalTimeType(input.MehndiTimeType)
	input.MehndiTime = strings.TrimSpace(input.MehndiTime)
	input.MehndiDinnerTime = strings.TrimSpace(input.MehndiDinnerTime)
	input.MehndiVenueName = sanitizeSingleLine(input.MehndiVenueName)
	input.MehndiVenueAddress = sanitizeMultiline(input.MehndiVenueAddress)
	input.BaraatDate = strings.TrimSpace(input.BaraatDate)
	input.BaraatDay = normalizeOptionalDay(input.BaraatDay)
	input.BaraatTimeType = normalizeOptionalTimeType(input.BaraatTimeType)
	input.BaraatTime = strings.TrimSpace(input.BaraatTime)
	input.BaraatDinnerTime = strings.TrimSpace(input.BaraatDinnerTime)
	input.BaraatArrivalTime = strings.TrimSpace(input.BaraatArrivalTime)
	input.RukhsatiTime = strings.TrimSpace(input.RukhsatiTime)
	input.BaraatVenueName = sanitizeSingleLine(input.BaraatVenueName)
	input.BaraatVenueAddress = sanitizeMultiline(input.BaraatVenueAddress)
	input.NikkahDate = strings.TrimSpace(input.NikkahDate)
	input.NikkahDay = normalizeOptionalDay(input.NikkahDay)
	input.NikkahTimeType = normalizeOptionalTimeType(input.NikkahTimeType)
	input.NikkahTime = strings.TrimSpace(input.NikkahTime)
	input.NikkahDinnerTime = strings.TrimSpace(input.NikkahDinnerTime)
	input.NikkahVenueName = sanitizeSingleLine(input.NikkahVenueName)
	input.NikkahVenueAddress = sanitizeMultiline(input.NikkahVenueAddress)
	input.WalimaDate = strings.TrimSpace(input.WalimaDate)
	input.WalimaDay = normalizeOptionalDay(input.WalimaDay)
	input.WalimaTimeType = normalizeOptionalTimeType(input.WalimaTimeType)
	input.WalimaTime = strings.TrimSpace(input.WalimaTime)
	input.WalimaDinnerTime = strings.TrimSpace(input.WalimaDinnerTime)
	input.WalimaVenueName = sanitizeSingleLine(input.WalimaVenueName)
	input.WalimaVenueAddress = sanitizeMultiline(input.WalimaVenueAddress)
	input.ReceptionTime = strings.TrimSpace(input.ReceptionTime)
	input.RsvpName = sanitizeSingleLine(input.RsvpName)
	input.RsvpPhone = normalizePhone(input.RsvpPhone)
	input.Notes = sanitizeMultiline(input.Notes)
	return input
}

func validateCustomerFields(name string, email string, phone string, address string, city string, postalCode string) error {
	if name == "" || email == "" || phone == "" || address == "" || city == "" || postalCode == "" {
		return ErrInvalidInput
	}
	if utf8.RuneCountInString(name) > maxCustomerNameLength ||
		utf8.RuneCountInString(email) > maxEmailLength ||
		utf8.RuneCountInString(phone) > maxPhoneLength ||
		utf8.RuneCountInString(address) > maxAddressLength ||
		utf8.RuneCountInString(city) > maxCityLength ||
		utf8.RuneCountInString(postalCode) > maxPostalCodeLength {
		return ErrInvalidInput
	}
	if !containsLetterOrDigit(name) || !containsLetterOrDigit(address) || !containsLetterOrDigit(city) || !containsLetterOrDigit(postalCode) {
		return ErrInvalidInput
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidInput
	}
	if !pakistanPhonePattern.MatchString(phone) {
		return ErrInvalidInput
	}
	return nil
}

func validateCustomizationFields(input PlaceOrderInput, category string) error {
	if isBidBoxCategory(category) {
		return validateBidBoxCustomizationFields(input)
	}
	return validateWeddingCustomizationFields(input)
}

func validateBidBoxCustomizationFields(input PlaceOrderInput) error {
	for _, value := range []string{
		input.BidBoxTopLabel,
		input.BidBoxCoupleName,
	} {
		if utf8.RuneCountInString(value) > maxBidBoxLabelLength {
			return ErrInvalidInput
		}
	}
	if err := validateOptionalDate(input.BidBoxEventDate); err != nil {
		return err
	}
	if utf8.RuneCountInString(input.BidBoxDetails) > maxBidBoxDetailsLength {
		return ErrInvalidInput
	}
	return nil
}

func validateWeddingCustomizationFields(input PlaceOrderInput) error {
	if _, err := parseSide(input.Side); err != nil {
		return err
	}
	if _, err := parseOptionalDay(input.MehndiDay); err != nil {
		return err
	}
	if _, err := parseOptionalDay(input.BaraatDay); err != nil {
		return err
	}
	if _, err := parseOptionalDay(input.NikkahDay); err != nil {
		return err
	}
	if _, err := parseOptionalDay(input.WalimaDay); err != nil {
		return err
	}
	if _, err := parseOptionalTimeType(input.MehndiTimeType); err != nil {
		return err
	}
	if _, err := parseOptionalTimeType(input.BaraatTimeType); err != nil {
		return err
	}
	if _, err := parseOptionalTimeType(input.NikkahTimeType); err != nil {
		return err
	}
	if _, err := parseOptionalTimeType(input.WalimaTimeType); err != nil {
		return err
	}
	for _, value := range []string{
		input.MehndiDate,
		input.BaraatDate,
		input.NikkahDate,
		input.WalimaDate,
	} {
		if err := validateOptionalDate(value); err != nil {
			return err
		}
	}
	for _, value := range []string{
		input.MehndiTime,
		input.MehndiDinnerTime,
		input.BaraatTime,
		input.BaraatDinnerTime,
		input.BaraatArrivalTime,
		input.RukhsatiTime,
		input.NikkahTime,
		input.NikkahDinnerTime,
		input.WalimaTime,
		input.WalimaDinnerTime,
		input.ReceptionTime,
	} {
		if err := validateOptionalTime(value); err != nil {
			return err
		}
	}
	for _, value := range []string{
		input.BrideName,
		input.GroomName,
		input.BrideFatherName,
		input.GroomFatherName,
		input.RsvpName,
	} {
		if utf8.RuneCountInString(value) > maxPersonNameLength {
			return ErrInvalidInput
		}
	}
	for _, value := range []string{
		input.MehndiVenueName,
		input.BaraatVenueName,
		input.NikkahVenueName,
		input.WalimaVenueName,
	} {
		if utf8.RuneCountInString(value) > maxVenueNameLength {
			return ErrInvalidInput
		}
	}
	for _, value := range []string{
		input.MehndiVenueAddress,
		input.BaraatVenueAddress,
		input.NikkahVenueAddress,
		input.WalimaVenueAddress,
	} {
		if utf8.RuneCountInString(value) > maxVenueAddressLength {
			return ErrInvalidInput
		}
	}
	if utf8.RuneCountInString(input.Notes) > maxNotesLength {
		return ErrInvalidInput
	}
	if input.RsvpPhone != "" && !pakistanPhonePattern.MatchString(input.RsvpPhone) {
		return ErrInvalidInput
	}
	return nil
}

func isBidBoxCategory(category string) bool {
	return strings.EqualFold(strings.TrimSpace(category), "bid-boxes")
}

func normalizeCurrency(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}

func normalizeFoilOption(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func normalizeSide(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func normalizeOrderStatus(raw string) (orderdomain.OrderStatus, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(orderdomain.PendingOrderStatus):
		return orderdomain.PendingOrderStatus, nil
	case string(orderdomain.ConfirmedOrderStatus):
		return orderdomain.ConfirmedOrderStatus, nil
	case string(orderdomain.CancelledOrderStatus):
		return orderdomain.CancelledOrderStatus, nil
	case string(orderdomain.CompletedOrderStatus):
		return orderdomain.CompletedOrderStatus, nil
	default:
		return "", ErrInvalidInput
	}
}

type orderEmailPayload struct {
	OrderID       int64
	CustomerID    int64
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
	ProductName   string
	Quantity      int64
	TotalPrice    int64
	AdvanceAmount int64
	Remaining     int64
	Currency      string
	PaymentStatus orderdomain.PaymentStatus
	Status        orderdomain.OrderStatus
}

func (s *Service) sendOrderCreatedEmailsAsync(payload orderEmailPayload) {
	if s.emailSender == nil {
		log.Printf("ORDER EMAIL SKIPPED: order_id=%d sender not configured", payload.OrderID)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		adminEmail := strings.TrimSpace(s.adminEmail)
		if adminEmail == "" {
			log.Printf("ORDER EMAIL SKIPPED: order_id=%d admin email not configured", payload.OrderID)
		} else {
			if err := s.sendOrderEmail(ctx, adminEmail, newOrderAdminSubject(payload.OrderID), buildNewOrderAdminEmailBody(payload)); err != nil {
				log.Printf("ORDER EMAIL ERROR: order_id=%d recipient=admin err=%v", payload.OrderID, err)
			}
		}

		customerEmail := strings.TrimSpace(payload.CustomerEmail)
		if customerEmail == "" {
			log.Printf("ORDER EMAIL SKIPPED: order_id=%d customer email missing", payload.OrderID)
			return
		}
		if err := s.sendOrderEmail(ctx, customerEmail, newOrderCustomerSubject(payload.OrderID), buildNewOrderCustomerEmailBody(payload)); err != nil {
			log.Printf("ORDER EMAIL ERROR: order_id=%d recipient=customer err=%v", payload.OrderID, err)
		}
	}()
}

func (s *Service) sendOrderStatusEmailAsync(order *orderdomain.Order, status orderdomain.OrderStatus) {
	if order == nil {
		log.Printf("ORDER EMAIL ERROR: order detail missing for status update")
		return
	}
	if s.emailSender == nil {
		log.Printf("ORDER EMAIL SKIPPED: order_id=%d sender not configured", order.ID)
		return
	}
	if order.CustomerID <= 0 {
		log.Printf("ORDER EMAIL SKIPPED: order_id=%d customer id missing", order.ID)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		customer, err := s.customerRepo.GetCustomerByID(ctx, order.CustomerID)
		if err != nil {
			log.Printf("ORDER EMAIL ERROR: order_id=%d failed to load customer err=%v", order.ID, err)
			return
		}
		if customer.Email == nil || strings.TrimSpace(*customer.Email) == "" {
			log.Printf("ORDER EMAIL SKIPPED: order_id=%d customer email missing", order.ID)
			return
		}

		var payment *orderdomain.OrderPayment
		payment, err = s.orderRepo.GetOrderPaymentByOrderID(ctx, order.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			log.Printf("ORDER EMAIL ERROR: order_id=%d failed to load payment err=%v", order.ID, err)
			payment = nil
		}

		customerPhone := ""
		if customer.Phone != nil {
			customerPhone = strings.TrimSpace(*customer.Phone)
		}
		paymentSummary := buildOrderEmailPaymentSummary(order.TotalPrice, payment)
		paymentStatus := orderdomain.PendingPaymentStatus
		if payment != nil && payment.PaymentStatus != "" {
			paymentStatus = payment.PaymentStatus
		}

		payload := orderEmailPayload{
			OrderID:       order.ID,
			CustomerID:    order.CustomerID,
			CustomerName:  customer.Name,
			CustomerEmail: strings.TrimSpace(*customer.Email),
			CustomerPhone: customerPhone,
			ProductName:   defaultProductName(order.CardName),
			Quantity:      order.Quantity,
			TotalPrice:    order.TotalPrice,
			AdvanceAmount: paymentSummary.AdvanceAmount,
			Remaining:     paymentSummary.RemainingBalance,
			Currency:      defaultCurrency(order.Currency),
			PaymentStatus: paymentStatus,
			Status:        status,
		}

		adminEmail := strings.TrimSpace(s.adminEmail)
		if adminEmail == "" {
			log.Printf("ORDER EMAIL SKIPPED: order_id=%d admin email not configured for status update", payload.OrderID)
		} else {
			if err := s.sendOrderEmail(ctx, adminEmail, adminOrderStatusSubject(payload.OrderID, status), buildAdminOrderStatusEmailBody(payload)); err != nil {
				log.Printf("ORDER EMAIL ERROR: order_id=%d recipient=admin status=%q err=%v", payload.OrderID, status, err)
			}
		}

		if err := s.sendOrderEmail(ctx, payload.CustomerEmail, orderStatusSubject(payload.OrderID, status), buildCustomerOrderStatusEmailBody(payload)); err != nil {
			log.Printf("ORDER EMAIL ERROR: order_id=%d recipient=customer status=%q err=%v", payload.OrderID, status, err)
		}
	}()
}

func (s *Service) sendOrderEmail(ctx context.Context, to string, subject string, body string) error {
	return s.emailSender.SendOrderEmail(ctx, to, subject, body)
}

func buildNewOrderAdminEmailBody(payload orderEmailPayload) string {
	return fmt.Sprintf(
		"New order received.\n\nOrder ID: #%d\nCustomer: %s\nPhone/WhatsApp: %s\nEmail: %s\nProduct: %s\nQuantity: %d\n\nOrder Total: %s %d\nAdvance Payment Required Now: %s %d\nRemaining Balance: %s %d\n\nPayment Status: Advance Payment Pending / Final Payment Pending\n\nAdmin reminder:\nWait for the customer to transfer and submit the 50%% advance payment. Confirm the order only after the advance payment is verified.",
		payload.OrderID,
		defaultCustomerName(payload.CustomerName),
		defaultCustomerPhone(payload.CustomerPhone),
		defaultCustomerEmail(payload.CustomerEmail),
		defaultProductName(payload.ProductName),
		payload.Quantity,
		defaultCurrency(payload.Currency),
		payload.TotalPrice,
		defaultCurrency(payload.Currency),
		payload.AdvanceAmount,
		defaultCurrency(payload.Currency),
		payload.Remaining,
	)
}

func buildNewOrderCustomerEmailBody(payload orderEmailPayload) string {
	return fmt.Sprintf(
		"Dear %s,\n\nThank you for your order. We have received it.\n\nOrder ID: #%d\nProduct: %s\nQuantity: %d\n\nOrder Total: %s %d\nAdvance Payment Required Now: %s %d\nRemaining Balance: %s %d\n\nPayment Status: Advance Payment Pending / Final Payment Pending\n\nPlease transfer the 50%% advance payment and upload your payment proof from the bank transfer page. We will confirm your order for processing after the advance payment is verified.\n\nBest regards,\nWrite & InviteCo",
		defaultCustomerName(payload.CustomerName),
		payload.OrderID,
		defaultProductName(payload.ProductName),
		payload.Quantity,
		defaultCurrency(payload.Currency),
		payload.TotalPrice,
		defaultCurrency(payload.Currency),
		payload.AdvanceAmount,
		defaultCurrency(payload.Currency),
		payload.Remaining,
	)
}

func buildCustomerOrderStatusEmailBody(payload orderEmailPayload) string {
	switch payload.Status {
	case orderdomain.ConfirmedOrderStatus:
		return fmt.Sprintf(
			"Dear %s,\n\nYour order has been confirmed.\n\nOrder ID: #%d\nProduct: %s\nQuantity: %d\n\nOrder Total: %s %d\nAdvance Payment Received: %s %d\nRemaining Balance: %s %d\n\nPayment Status: Advance Payment Received / Final Payment Pending\n\nOur team will contact you on your provided WhatsApp number or email when your order is ready. We will share proof/images of your product before dispatch. Once you review and confirm the final product, please pay the remaining balance before delivery.\n\nThank you for trusting Write & InviteCo. We truly appreciate your kindness and support. 🤍\n\nBest regards,\nWrite & InviteCo",
			defaultCustomerName(payload.CustomerName),
			payload.OrderID,
			defaultProductName(payload.ProductName),
			payload.Quantity,
			defaultCurrency(payload.Currency),
			payload.TotalPrice,
			defaultCurrency(payload.Currency),
			payload.AdvanceAmount,
			defaultCurrency(payload.Currency),
			payload.Remaining,
		)
	case orderdomain.CompletedOrderStatus:
		return fmt.Sprintf(
			"Dear %s,\n\nYour order has been completed.\n\nOrder ID: #%d\nProduct: %s\nQuantity: %d\n\nOrder Total: %s %d\nAdvance Payment Received: %s %d\nRemaining Balance: %s %d\n\nPayment Status: Advance Payment Received / Final Payment Pending\n\nYour order is ready. Our team will contact you on your provided WhatsApp number or email, share proof/images of your final product if needed, and collect the remaining balance before delivery.\n\nBest regards,\nWrite & InviteCo",
			defaultCustomerName(payload.CustomerName),
			payload.OrderID,
			defaultProductName(payload.ProductName),
			payload.Quantity,
			defaultCurrency(payload.Currency),
			payload.TotalPrice,
			defaultCurrency(payload.Currency),
			payload.AdvanceAmount,
			defaultCurrency(payload.Currency),
			payload.Remaining,
		)
	case orderdomain.CancelledOrderStatus:
		return fmt.Sprintf(
			"Dear %s,\n\nYour order has been cancelled.\n\nOrder ID: #%d\nProduct: %s\nQuantity: %d\n\nOrder Total: %s %d\n%s: %s %d\nRemaining Balance: %s %d\n\nPayment Status: %s\n\nIf you have any questions about this cancellation or your payment, please reply to this email and our team will help you.\n\nBest regards,\nWrite & InviteCo",
			defaultCustomerName(payload.CustomerName),
			payload.OrderID,
			defaultProductName(payload.ProductName),
			payload.Quantity,
			defaultCurrency(payload.Currency),
			payload.TotalPrice,
			advanceEmailLabel(payload.PaymentStatus),
			defaultCurrency(payload.Currency),
			advanceAmountForStatusEmail(payload),
			defaultCurrency(payload.Currency),
			payload.Remaining,
			customerPaymentStatusLine(payload),
		)
	default:
		return fmt.Sprintf(
			"Dear %s,\n\nYour order status has been updated.\n\nOrder ID: #%d\nProduct: %s\nQuantity: %d\n\nOrder Total: %s %d\nAdvance Payment Required Now: %s %d\nRemaining Balance: %s %d\n\nPayment Status: %s\n\nPlease follow the payment instructions on your order page if your advance payment is still pending.\n\nBest regards,\nWrite & InviteCo",
			defaultCustomerName(payload.CustomerName),
			payload.OrderID,
			defaultProductName(payload.ProductName),
			payload.Quantity,
			defaultCurrency(payload.Currency),
			payload.TotalPrice,
			defaultCurrency(payload.Currency),
			payload.AdvanceAmount,
			defaultCurrency(payload.Currency),
			payload.Remaining,
			customerPaymentStatusLine(payload),
		)
	}
}

func buildAdminOrderStatusEmailBody(payload orderEmailPayload) string {
	return fmt.Sprintf(
		"%s\n\nOrder ID: #%d\nCustomer: %s\nPhone/WhatsApp: %s\nEmail: %s\nProduct: %s\nQuantity: %d\n\nOrder Total: %s %d\n%s: %s %d\nRemaining Balance: %s %d\n\nPayment Status: %s\n\nAdmin reminder:\n%s",
		adminOrderStatusIntro(payload.Status),
		payload.OrderID,
		defaultCustomerName(payload.CustomerName),
		defaultCustomerPhone(payload.CustomerPhone),
		defaultCustomerEmail(payload.CustomerEmail),
		defaultProductName(payload.ProductName),
		payload.Quantity,
		defaultCurrency(payload.Currency),
		payload.TotalPrice,
		advanceEmailLabel(payload.PaymentStatus),
		defaultCurrency(payload.Currency),
		advanceAmountForStatusEmail(payload),
		defaultCurrency(payload.Currency),
		payload.Remaining,
		adminPaymentStatusLine(payload),
		adminOrderStatusReminder(payload),
	)
}

func newOrderAdminSubject(orderID int64) string {
	return fmt.Sprintf("New Order Received (#%d)", orderID)
}

func newOrderCustomerSubject(orderID int64) string {
	return fmt.Sprintf("Order Received (#%d)", orderID)
}

func orderStatusSubject(orderID int64, status orderdomain.OrderStatus) string {
	switch status {
	case orderdomain.ConfirmedOrderStatus:
		return fmt.Sprintf("Your order #%d is confirmed", orderID)
	case orderdomain.CancelledOrderStatus:
		return fmt.Sprintf("Your order #%d has been cancelled", orderID)
	case orderdomain.CompletedOrderStatus:
		return fmt.Sprintf("Your order #%d has been completed", orderID)
	default:
		return fmt.Sprintf("Your order #%d is pending", orderID)
	}
}

func adminOrderStatusSubject(orderID int64, status orderdomain.OrderStatus) string {
	switch status {
	case orderdomain.ConfirmedOrderStatus:
		return fmt.Sprintf("Order #%d confirmed for processing", orderID)
	case orderdomain.CancelledOrderStatus:
		return fmt.Sprintf("Order #%d cancelled", orderID)
	case orderdomain.CompletedOrderStatus:
		return fmt.Sprintf("Order #%d marked completed", orderID)
	default:
		return fmt.Sprintf("Order Status Updated (#%d)", orderID)
	}
}

func buildOrderEmailPaymentSummary(totalPrice int64, payment *orderdomain.OrderPayment) PaymentAmountSummary {
	expectedAmount := int64(0)
	if payment != nil {
		expectedAmount = payment.ExpectedAmount
	}
	return BuildPaymentAmountSummary(totalPrice, expectedAmount)
}

func advanceAmountForStatusEmail(payload orderEmailPayload) int64 {
	if payload.AdvanceAmount > 0 {
		return payload.AdvanceAmount
	}
	return BuildPaymentAmountSummary(payload.TotalPrice, 0).AdvanceAmount
}

func advanceEmailLabel(status orderdomain.PaymentStatus) string {
	switch status {
	case orderdomain.VerifiedPaymentStatus:
		return "Advance Payment Received"
	case orderdomain.AwaitingVerificationPaymentStatus:
		return "Advance Payment Submitted"
	default:
		return "Advance Payment Required Now"
	}
}

func customerPaymentStatusLine(payload orderEmailPayload) string {
	switch payload.PaymentStatus {
	case orderdomain.VerifiedPaymentStatus:
		return "Advance Payment Received / Final Payment Pending"
	case orderdomain.AwaitingVerificationPaymentStatus:
		return "Advance Payment Awaiting Verification / Final Payment Pending"
	case orderdomain.RejectedPaymentStatus:
		return "Advance Payment Rejected / Final Payment Pending"
	default:
		return "Advance Payment Pending / Final Payment Pending"
	}
}

func adminPaymentStatusLine(payload orderEmailPayload) string {
	switch payload.PaymentStatus {
	case orderdomain.VerifiedPaymentStatus:
		return "Advance Verified / Final Payment Pending"
	case orderdomain.AwaitingVerificationPaymentStatus:
		return "Advance Awaiting Verification / Final Payment Pending"
	case orderdomain.RejectedPaymentStatus:
		return "Advance Rejected / Final Payment Pending"
	default:
		return "Advance Payment Pending / Final Payment Pending"
	}
}

func adminOrderStatusIntro(status orderdomain.OrderStatus) string {
	switch status {
	case orderdomain.ConfirmedOrderStatus:
		return "Order confirmed for processing."
	case orderdomain.CancelledOrderStatus:
		return "Order cancelled."
	case orderdomain.CompletedOrderStatus:
		return "Order marked completed."
	default:
		return "Order status updated."
	}
}

func adminOrderStatusReminder(payload orderEmailPayload) string {
	switch payload.PaymentStatus {
	case orderdomain.VerifiedPaymentStatus:
		return "The client has paid 50% advance only. Remaining 50% should be collected after the product is completed and approved by the customer, before delivery/dispatch."
	case orderdomain.AwaitingVerificationPaymentStatus:
		return "Review the submitted advance payment proof. Confirm the order only after the advance payment is verified. The remaining 50% should still be collected before delivery/dispatch."
	case orderdomain.RejectedPaymentStatus:
		return "The advance payment proof was rejected. Wait for the customer to re-upload proof before confirming the order. The remaining 50% will still be due before delivery/dispatch."
	default:
		return "The client has not completed the verified 50% advance step yet. Confirm the order only after the advance payment is verified. The remaining 50% should be collected before delivery/dispatch."
	}
}

func defaultCustomerName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Customer"
	}
	return name
}

func defaultProductName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Invitation Card"
	}
	return name
}

func defaultCustomerEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return "Not provided"
	}
	return email
}

func defaultCustomerPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return "Not provided"
	}
	return phone
}

func defaultCurrency(currency string) string {
	currency = strings.TrimSpace(currency)
	if currency == "" {
		return "PKR"
	}
	return currency
}

func orderTypeLabel(order *orderdomain.Order, details *orderdomain.OrderDetail) string {
	if order != nil && isBidBoxCategory(order.CardCategory) {
		return "bid-box"
	}
	if hasBidBoxFields(details) {
		return "bid-box"
	}
	return "wedding-card"
}

func hasBidBoxFields(details *orderdomain.OrderDetail) bool {
	if details == nil {
		return false
	}
	return details.BidBoxTopLabel != nil ||
		details.BidBoxCoupleName != nil ||
		details.BidBoxEventDate != nil ||
		details.BidBoxDetails != nil
}

func stringPtr(value string) *string {
	return &value
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func parseCurrency(raw string) (string, error) {
	switch normalizeCurrency(raw) {
	case "", "PKR":
		return "PKR", nil
	case "NOK":
		return "NOK", nil
	default:
		return "", ErrInvalidInput
	}
}

func parseFoilOption(raw string) (string, error) {
	switch normalizeFoilOption(raw) {
	case "", "foil":
		return "foil", nil
	case "nofoil":
		return "nofoil", nil
	default:
		return "", ErrInvalidInput
	}
}

func parseSide(raw string) (string, error) {
	switch normalizeSide(raw) {
	case "", "bride":
		return "bride", nil
	case "groom":
		return "groom", nil
	default:
		return "", ErrInvalidInput
	}
}

func parseOptionalDay(raw string) (string, error) {
	switch normalizeOptionalDay(raw) {
	case "":
		return "", nil
	case "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday":
		return normalizeOptionalDay(raw), nil
	default:
		return "", ErrInvalidInput
	}
}

func parseOptionalTimeType(raw string) (string, error) {
	switch normalizeOptionalTimeType(raw) {
	case "", "evening", "night":
		return normalizeOptionalTimeType(raw), nil
	default:
		return "", ErrInvalidInput
	}
}

func validateOptionalDate(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return ErrInvalidInput
	}
	return nil
}

func validateOptionalTime(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	if _, err := time.Parse("15:04", raw); err != nil {
		return ErrInvalidInput
	}
	return nil
}

func sanitizeSingleLine(value string) string {
	value = sanitizeText(value, false)
	return collapseWhitespace(strings.ReplaceAll(value, "\n", " "))
}

func sanitizeMultiline(value string) string {
	return sanitizeText(value, true)
}

func sanitizeText(value string, allowNewlines bool) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\r\n", "\n"))
	value = strings.Map(func(r rune) rune {
		switch {
		case r == '<' || r == '>':
			return -1
		case r == '\n' && allowNewlines:
			return r
		case r == '\t' || r == '\n' || unicode.IsPrint(r):
			return r
		default:
			return -1
		}
	}, value)
	return strings.TrimSpace(value)
}

func collapseWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func normalizePhone(value string) string {
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")
	return replacer.Replace(strings.TrimSpace(value))
}

func normalizeOptionalDay(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "monday":
		return "Monday"
	case "tuesday":
		return "Tuesday"
	case "wednesday":
		return "Wednesday"
	case "thursday":
		return "Thursday"
	case "friday":
		return "Friday"
	case "saturday":
		return "Saturday"
	case "sunday":
		return "Sunday"
	default:
		return strings.TrimSpace(raw)
	}
}

func normalizeOptionalTimeType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "evening":
		return "evening"
	case "night":
		return "night"
	default:
		return strings.TrimSpace(raw)
	}
}

func containsLetterOrDigit(value string) bool {
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func safeAddInt64(a int64, b int64) (int64, error) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, errPricingOverflow
	}
	if b < 0 && a < math.MinInt64-b {
		return 0, errPricingOverflow
	}
	return a + b, nil
}

func safeMultiplyInt64(a int64, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a < 0 || b < 0 {
		return 0, errPricingOverflow
	}
	if a > math.MaxInt64/b {
		return 0, errPricingOverflow
	}
	return a * b, nil
}
