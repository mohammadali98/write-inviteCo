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
	log.Printf("Form values: name=%s, quantity=%s, card_id=%s, card_name=%s, price_pkr=%s, price_nok=%s, currency=%s",
		c.PostForm("name"),
		c.PostForm("quantity"),
		c.PostForm("card_id"),
		c.PostForm("card_name"),
		c.PostForm("price_pkr"),
		c.PostForm("price_nok"),
		c.PostForm("currency"),
	)

	name := strings.TrimSpace(c.PostForm("name"))
	quantityRaw := strings.TrimSpace(c.PostForm("quantity"))
	cardIDRaw := strings.TrimSpace(c.PostForm("card_id"))
	cardName := strings.TrimSpace(c.PostForm("card_name"))
	pricePKRRaw := strings.TrimSpace(c.PostForm("price_pkr"))
	priceNOKRaw := strings.TrimSpace(c.PostForm("price_nok"))
	currency := strings.ToUpper(strings.TrimSpace(c.PostForm("currency")))
	if currency == "" {
		currency = "PKR"
	}
	if cardName == "" {
		cardName = "Invitation Card"
	}

	quantityInt, err := strconv.Atoi(quantityRaw)
	if name == "" || err != nil || quantityInt < 1 {
		c.String(http.StatusBadRequest, "Please provide a valid name and quantity.")
		return
	}

	cardIDInt, err := strconv.Atoi(cardIDRaw)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid card id.")
		return
	}

	var priceInt int
	switch currency {
	case "NOK":
		priceInt, err = strconv.Atoi(priceNOKRaw)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid NOK price.")
			return
		}
	default:
		currency = "PKR"
		priceInt, err = strconv.Atoi(pricePKRRaw)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid PKR price.")
			return
		}
	}

	quantity := int64(quantityInt)
	cardID := int64(cardIDInt)
	price := int64(priceInt)
	totalPrice := price * quantity

	ctx := c.Request.Context()

	customer, err := h.customerRepo.CreateCustomer(ctx, name, nil, nil)
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
