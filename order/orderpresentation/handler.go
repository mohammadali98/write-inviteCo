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
	log.Printf("Form values: name=%s, quantity=%s, card_id=%s, price=%s",
		c.PostForm("name"), c.PostForm("quantity"), c.PostForm("card_id"), c.PostForm("price"))

	name := strings.TrimSpace(c.PostForm("name"))
	quantityRaw := strings.TrimSpace(c.PostForm("quantity"))
	cardIDRaw := strings.TrimSpace(c.PostForm("card_id"))
	priceRaw := strings.TrimSpace(c.PostForm("price"))

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

	priceInt, err := strconv.Atoi(priceRaw)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid price.")
		return
	}

	quantity := int64(quantityInt)
	cardID := int64(cardIDInt)
	price := int64(priceInt)

	ctx := c.Request.Context()

	customer, err := h.customerRepo.CreateCustomer(ctx, name, nil, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create customer.")
		return
	}

	_, err = h.orderRepo.CreateOrder(ctx, customer.ID, cardID, quantity, price*quantity, orderdomain.PendingOrderStatus)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create order.")
		return
	}

	c.String(http.StatusOK,
		"Thank you, %s! Your order for %d invitation card(s) has been received.",
		name, quantity,
	)
}
