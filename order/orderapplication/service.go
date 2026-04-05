package orderapplication

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"writeandinviteco/inviteandco/card/carddomain"
	customerwriter "writeandinviteco/inviteandco/customer/customerinfrastructure/postgres/writer"
	"writeandinviteco/inviteandco/order/orderdomain"
	orderwriter "writeandinviteco/inviteandco/order/orderinfrastructure/postgres/writer"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCardNotFound = errors.New("card not found")
	ErrInvalidInput = errors.New("invalid input")
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

	Side              string
	BrideName         string
	GroomName         string
	BrideFatherName   string
	GroomFatherName   string
	MehndiDate        string
	MehndiTimeType    string
	MehndiTime        string
	MehndiDinnerTime  string
	BaraatDate        string
	BaraatTimeType    string
	BaraatTime        string
	BaraatDinnerTime  string
	BaraatArrivalTime string
	RukhsatiTime      string
	NikkahDate        string
	NikkahTimeType    string
	NikkahTime        string
	NikkahDinnerTime  string
	WalimaDate        string
	WalimaTimeType    string
	WalimaTime        string
	WalimaDinnerTime  string
	ReceptionTime     string
	DinnerTime        string
	VenueName         string
	VenueAddress      string
	RsvpName          string
	RsvpPhone         string
	Notes             string
}

type PlaceOrderResult struct {
	OrderID      int64
	CustomerName string
	CardName     string
	Quantity     int64
	TotalPrice   int64
	Currency     string
}

type Service struct {
	db             *pgxpool.Pool
	cardRepo       carddomain.CardRepo
	customerWriter *customerwriter.Queries
	orderWriter    *orderwriter.Queries
}

func NewService(db *pgxpool.Pool, cardRepo carddomain.CardRepo, customerWriter *customerwriter.Queries, orderWriter *orderwriter.Queries) *Service {
	return &Service{
		db:             db,
		cardRepo:       cardRepo,
		customerWriter: customerWriter,
		orderWriter:    orderWriter,
	}
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

	return &CustomizationSummary{
		CardID:           pricing.card.ID,
		CardName:         pricing.card.Name,
		CardImage:        pricing.card.Image,
		Quantity:         pricing.quantity,
		Currency:         pricing.currency,
		FoilOption:       pricing.foilOption,
		FoilLabel:        pricing.foilLabel,
		Side:             "bride",
		RequestedInserts: pricing.requestedInserts,
		IncludedInserts:  pricing.includedInserts,
		ExtraInserts:     pricing.extraInserts,
		UnitPrice:        pricing.basePrice,
		InsertPrice:      pricing.insertPrice,
		ExtraInsertCost:  pricing.extraInsertCost,
		PerCardTotal:     pricing.perCardPrice,
		TotalPrice:       pricing.totalPrice,
		MinOrder:         pricing.minOrder,
		Name:             input.Name,
		Email:            input.Email,
		Phone:            input.Phone,
		Address:          input.Address,
		City:             input.City,
		PostalCode:       input.PostalCode,
	}, nil
}

func (s *Service) PlaceOrder(ctx context.Context, input PlaceOrderInput) (*PlaceOrderResult, error) {
	input = sanitizePlaceOrderInput(input)

	if err := validateCustomerFields(input.Name, input.Email, input.Phone, input.Address, input.City, input.PostalCode); err != nil {
		return nil, err
	}
	if err := validateCustomizationFields(input); err != nil {
		return nil, err
	}

	pricing, err := s.calculatePricing(ctx, input.CardID, input.Quantity, input.Currency, input.FoilOption, input.RequestedInserts)
	if err != nil {
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

	var brideName *string
	var groomName *string
	var brideFatherName *string
	var groomFatherName *string
	if input.Side == "groom" {
		groomName = stringPtr(input.GroomName)
		groomFatherName = stringPtr(input.GroomFatherName)
	} else {
		brideName = stringPtr(input.BrideName)
		brideFatherName = stringPtr(input.BrideFatherName)
	}

	_, err = s.orderWriter.WithTx(tx).CreateOrderDetail(ctx, orderwriter.CreateOrderDetailParams{
		OrderID:           orderRow.ID,
		Side:              input.Side,
		BrideName:         brideName,
		GroomName:         groomName,
		BrideFatherName:   brideFatherName,
		GroomFatherName:   groomFatherName,
		MehndiDate:        input.MehndiDate,
		MehndiTimeType:    nullableString(input.MehndiTimeType),
		MehndiTime:        input.MehndiTime,
		MehndiDinnerTime:  input.MehndiDinnerTime,
		BaraatDate:        input.BaraatDate,
		BaraatTimeType:    nullableString(input.BaraatTimeType),
		BaraatTime:        input.BaraatTime,
		BaraatDinnerTime:  input.BaraatDinnerTime,
		BaraatArrivalTime: input.BaraatArrivalTime,
		RukhsatiTime:      input.RukhsatiTime,
		NikkahDate:        input.NikkahDate,
		NikkahTimeType:    nullableString(input.NikkahTimeType),
		NikkahTime:        input.NikkahTime,
		NikkahDinnerTime:  input.NikkahDinnerTime,
		WalimaDate:        input.WalimaDate,
		WalimaTimeType:    nullableString(input.WalimaTimeType),
		WalimaTime:        input.WalimaTime,
		WalimaDinnerTime:  input.WalimaDinnerTime,
		ReceptionTime:     input.ReceptionTime,
		DinnerTime:        input.DinnerTime,
		VenueName:         input.VenueName,
		VenueAddress:      input.VenueAddress,
		RsvpName:          input.RsvpName,
		RsvpPhone:         input.RsvpPhone,
		Notes:             nullableString(input.Notes),
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

	return &PlaceOrderResult{
		OrderID:      orderRow.ID,
		CustomerName: input.Name,
		CardName:     cardName,
		Quantity:     pricing.quantity,
		TotalPrice:   pricing.totalPrice,
		Currency:     pricing.currency,
	}, nil
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
	if cardID <= 0 || quantity < 1 || requestedInserts < 0 {
		return nil, ErrInvalidInput
	}

	card, err := s.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, ErrCardNotFound
	}

	minOrder := int64(card.MinOrder)
	if minOrder < 1 {
		minOrder = 1
	}
	if quantity < minOrder {
		return nil, MinOrderError{MinOrder: minOrder}
	}

	currency = normalizeCurrency(currency)
	foilOption = normalizeFoilOption(foilOption)

	priceFoil := card.PriceFoilPKR
	priceNofoil := card.PriceNofoilPKR
	insertPrice := card.InsertPricePKR
	if currency == "NOK" {
		priceFoil = card.PriceFoilNOK
		priceNofoil = card.PriceNofoilNOK
		insertPrice = card.InsertPriceNOK
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

	extraInsertCost := extraInserts * insertPrice
	perCardPrice := basePrice + extraInsertCost
	totalPrice := perCardPrice * quantity

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
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Address = strings.TrimSpace(input.Address)
	input.City = strings.TrimSpace(input.City)
	input.PostalCode = strings.TrimSpace(input.PostalCode)
	return input
}

func sanitizePlaceOrderInput(input PlaceOrderInput) PlaceOrderInput {
	input.Currency = normalizeCurrency(input.Currency)
	input.FoilOption = normalizeFoilOption(input.FoilOption)
	input.Side = normalizeSide(input.Side)
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Address = strings.TrimSpace(input.Address)
	input.City = strings.TrimSpace(input.City)
	input.PostalCode = strings.TrimSpace(input.PostalCode)
	input.BrideName = strings.TrimSpace(input.BrideName)
	input.GroomName = strings.TrimSpace(input.GroomName)
	input.BrideFatherName = strings.TrimSpace(input.BrideFatherName)
	input.GroomFatherName = strings.TrimSpace(input.GroomFatherName)
	input.MehndiDate = strings.TrimSpace(input.MehndiDate)
	input.MehndiTimeType = strings.TrimSpace(input.MehndiTimeType)
	input.MehndiTime = strings.TrimSpace(input.MehndiTime)
	input.MehndiDinnerTime = strings.TrimSpace(input.MehndiDinnerTime)
	input.BaraatDate = strings.TrimSpace(input.BaraatDate)
	input.BaraatTimeType = strings.TrimSpace(input.BaraatTimeType)
	input.BaraatTime = strings.TrimSpace(input.BaraatTime)
	input.BaraatDinnerTime = strings.TrimSpace(input.BaraatDinnerTime)
	input.BaraatArrivalTime = strings.TrimSpace(input.BaraatArrivalTime)
	input.RukhsatiTime = strings.TrimSpace(input.RukhsatiTime)
	input.NikkahDate = strings.TrimSpace(input.NikkahDate)
	input.NikkahTimeType = strings.TrimSpace(input.NikkahTimeType)
	input.NikkahTime = strings.TrimSpace(input.NikkahTime)
	input.NikkahDinnerTime = strings.TrimSpace(input.NikkahDinnerTime)
	input.WalimaDate = strings.TrimSpace(input.WalimaDate)
	input.WalimaTimeType = strings.TrimSpace(input.WalimaTimeType)
	input.WalimaTime = strings.TrimSpace(input.WalimaTime)
	input.WalimaDinnerTime = strings.TrimSpace(input.WalimaDinnerTime)
	input.ReceptionTime = strings.TrimSpace(input.ReceptionTime)
	input.DinnerTime = strings.TrimSpace(input.DinnerTime)
	input.VenueName = strings.TrimSpace(input.VenueName)
	input.VenueAddress = strings.TrimSpace(input.VenueAddress)
	input.RsvpName = strings.TrimSpace(input.RsvpName)
	input.RsvpPhone = strings.TrimSpace(input.RsvpPhone)
	input.Notes = strings.TrimSpace(input.Notes)
	return input
}

func validateCustomerFields(name string, email string, phone string, address string, city string, postalCode string) error {
	if name == "" || email == "" || phone == "" || address == "" || city == "" || postalCode == "" {
		return ErrInvalidInput
	}
	return nil
}

func validateCustomizationFields(input PlaceOrderInput) error {
	if input.VenueName == "" ||
		input.VenueAddress == "" ||
		input.RsvpName == "" ||
		input.RsvpPhone == "" {
		return ErrInvalidInput
	}

	switch input.Side {
	case "groom":
		if input.GroomName == "" || input.GroomFatherName == "" {
			return ErrInvalidInput
		}
	case "bride":
		if input.BrideName == "" || input.BrideFatherName == "" {
			return ErrInvalidInput
		}
	default:
		return ErrInvalidInput
	}

	return nil
}

func normalizeCurrency(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "NOK") {
		return "NOK"
	}
	return "PKR"
}

func normalizeFoilOption(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "nofoil") {
		return "nofoil"
	}
	return "foil"
}

func normalizeSide(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "groom") {
		return "groom"
	}
	return "bride"
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
