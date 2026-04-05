package orderpresentation

import (
	"net/http"
	"strconv"
	"strings"

	"writeandinviteco/inviteandco/card/carddomain"
	customerwriter "writeandinviteco/inviteandco/customer/customerinfrastructure/postgres/writer"
	"writeandinviteco/inviteandco/order/orderdomain"
	orderwriter "writeandinviteco/inviteandco/order/orderinfrastructure/postgres/writer"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderHandler struct {
	db             *pgxpool.Pool
	cardRepo       carddomain.CardRepo
	customerWriter *customerwriter.Queries
	orderWriter    *orderwriter.Queries
}

func NewOrderHandler(
	db *pgxpool.Pool,
	cardRepo carddomain.CardRepo,
	customerWriter *customerwriter.Queries,
	orderWriter *orderwriter.Queries,
) *OrderHandler {
	return &OrderHandler{
		db:             db,
		cardRepo:       cardRepo,
		customerWriter: customerWriter,
		orderWriter:    orderWriter,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	name := strings.TrimSpace(c.PostForm("name"))
	email := strings.TrimSpace(c.PostForm("email"))
	phone := strings.TrimSpace(c.PostForm("phone"))
	address := strings.TrimSpace(c.PostForm("address"))
	city := strings.TrimSpace(c.PostForm("city"))
	postalCode := strings.TrimSpace(c.PostForm("postal_code"))
	quantityRaw := strings.TrimSpace(c.PostForm("quantity"))
	cardIDRaw := strings.TrimSpace(c.PostForm("card_id"))
	currency := strings.ToUpper(strings.TrimSpace(c.PostForm("currency")))
	foilOption := strings.ToLower(strings.TrimSpace(c.PostForm("foil_option")))
	requestedInsertsRaw := strings.TrimSpace(c.PostForm("extra_inserts"))

	if currency == "" {
		currency = "PKR"
	}
	if currency != "NOK" {
		currency = "PKR"
	}
	if foilOption != "nofoil" {
		foilOption = "foil"
	}

	quantity, err := strconv.ParseInt(quantityRaw, 10, 64)
	if name == "" || email == "" || phone == "" || address == "" || city == "" || postalCode == "" || err != nil || quantity < 1 {
		c.String(http.StatusBadRequest, "Please provide valid checkout details.")
		return
	}

	cardID, err := strconv.ParseInt(cardIDRaw, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid card id.")
		return
	}

	requestedInserts, err := parseNonNegativeInt64(requestedInsertsRaw)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid inserts quantity.")
		return
	}

	ctx := c.Request.Context()

	card, err := h.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		c.String(http.StatusNotFound, "Card not found.")
		return
	}

	minOrder := int64(card.MinOrder)
	if minOrder < 1 {
		minOrder = 1
	}
	if quantity < minOrder {
		c.String(http.StatusBadRequest, "Minimum order quantity is %d.", minOrder)
		return
	}

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
	if foilOption == "nofoil" {
		basePrice = priceNofoil
	}

	includedInserts := int64(card.IncludedInserts)
	if includedInserts < 0 {
		includedInserts = 0
	}
	extraInserts := requestedInserts - includedInserts
	if extraInserts < 0 {
		extraInserts = 0
	}

	perCardPrice := basePrice + (extraInserts * insertPrice)
	totalPrice := perCardPrice * quantity

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create order.")
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	customerRow, err := h.customerWriter.WithTx(tx).CreateCustomer(ctx, customerwriter.CreateCustomerParams{
		Name:       name,
		Email:      &email,
		Phone:      &phone,
		Address:    &address,
		City:       &city,
		PostalCode: &postalCode,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create customer.")
		return
	}

	status := string(orderdomain.PendingOrderStatus)
	_, err = h.orderWriter.WithTx(tx).CreateOrder(ctx, orderwriter.CreateOrderParams{
		CustomerID: &customerRow.ID,
		CardID:     &cardID,
		Quantity:   quantity,
		TotalPrice: totalPrice,
		Status:     &status,
		Currency:   currency,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create order.")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.String(http.StatusInternalServerError, "Failed to finalize order.")
		return
	}

	cardName := card.Name
	if cardName == "" {
		cardName = "Invitation Card"
	}

	c.HTML(http.StatusOK, "order-confirmation.html", gin.H{
		"customerName": name,
		"quantity":     quantity,
		"totalPrice":   totalPrice,
		"currency":     currency,
		"cardName":     cardName,
	})
}

func parseNonNegativeInt64(raw string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}
