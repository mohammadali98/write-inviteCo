package orderpresentation

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"writeandinviteco/inviteandco/customer/customerdomain"
	"writeandinviteco/inviteandco/order/orderdomain"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderRepo    orderdomain.OrderRepo
	customerRepo customerdomain.CustomerRepo
}

func NewOrderHandler(orderRepo orderdomain.OrderRepo, customerRepo customerdomain.CustomerRepo) *OrderHandler {
	return &OrderHandler{
		orderRepo:    orderRepo,
		customerRepo: customerRepo,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	log.Printf("Form values: name=%s, quantity=%s, card_id=%s, card_name=%s, price=%s, currency=%s, email=%s, phone=%s, address=%s, city=%s, postal_code=%s",
		c.PostForm("name"),
		c.PostForm("quantity"),
		c.PostForm("card_id"),
		c.PostForm("card_name"),
		c.PostForm("price"),
		c.PostForm("currency"),
		c.PostForm("email"),
		c.PostForm("phone"),
		c.PostForm("address"),
		c.PostForm("city"),
		c.PostForm("postal_code"),
	)

	name := strings.TrimSpace(c.PostForm("name"))
	email := strings.TrimSpace(c.PostForm("email"))
	phone := strings.TrimSpace(c.PostForm("phone"))
	address := strings.TrimSpace(c.PostForm("address"))
	city := strings.TrimSpace(c.PostForm("city"))
	postalCode := strings.TrimSpace(c.PostForm("postal_code"))
	quantityRaw := strings.TrimSpace(c.PostForm("quantity"))
	cardIDRaw := strings.TrimSpace(c.PostForm("card_id"))
	cardName := strings.TrimSpace(c.PostForm("card_name"))
	priceRaw := strings.TrimSpace(c.PostForm("price"))
	currency := strings.ToUpper(strings.TrimSpace(c.PostForm("currency")))
	if currency == "" {
		currency = "PKR"
	}
	if cardName == "" {
		cardName = "Invitation Card"
	}

	quantityInt, err := strconv.Atoi(quantityRaw)
	if name == "" || email == "" || phone == "" || address == "" || city == "" || postalCode == "" || err != nil || quantityInt < 1 {
		c.String(http.StatusBadRequest, "Please provide valid checkout details.")
		return
	}

	cardIDInt, err := strconv.Atoi(cardIDRaw)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid card id.")
		return
	}

	priceInt, err := strconv.Atoi(priceRaw)
	if err != nil || priceInt < 0 {
		c.String(http.StatusBadRequest, "Invalid price.")
		return
	}

	quantity := int64(quantityInt)
	cardID := int64(cardIDInt)
	price := int64(priceInt)
	totalPrice := price * quantity

	ctx := c.Request.Context()

	customer, err := h.customerRepo.CreateCustomer(ctx, name, &email, &phone, &address, &city, &postalCode)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create customer.")
		return
	}

	_, err = h.orderRepo.CreateOrder(
		ctx,
		customer.ID,
		cardID,
		quantity,
		totalPrice,
		orderdomain.PendingOrderStatus,
		currency,
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create order.")
		return
	}

	c.HTML(http.StatusOK, "order-confirmation.html", gin.H{
		"customerName": name,
		"quantity":     quantity,
		"totalPrice":   totalPrice,
		"currency":     currency,
		"cardName":     cardName,
	})
}
