package orderpresentation

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"writeandinviteco/inviteandco/order/orderapplication"
	"writeandinviteco/inviteandco/order/orderdomain"
	"writeandinviteco/inviteandco/webui"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type orderService interface {
	PrepareCustomization(ctx context.Context, input orderapplication.CustomizationInput) (*orderapplication.CustomizationSummary, error)
	PrepareOrderReview(ctx context.Context, input orderapplication.PlaceOrderInput) (*orderapplication.OrderReview, error)
	PlaceOrder(ctx context.Context, input orderapplication.PlaceOrderInput) (*orderapplication.PlaceOrderResult, error)
	GetOrderStatusDetail(ctx context.Context, orderID int64) (*orderapplication.AdminOrderDetail, error)
	ListAdminOrders(ctx context.Context) ([]*orderdomain.AdminOrder, error)
	GetAdminOrderDetail(ctx context.Context, orderID int64) (*orderapplication.AdminOrderDetail, error)
	AdminUpdateOrderStatus(ctx context.Context, orderID int64, statusRaw string) error
	SubmitBankTransferProof(ctx context.Context, orderID int64, input orderapplication.PaymentProofInput) error
	AdminProcessPayment(ctx context.Context, orderID int64, action string, adminNote string) error
}

type OrderHandler struct {
	service         orderService
	paymentProofDir string
}

func NewOrderHandler(service orderService, paymentProofDir string) *OrderHandler {
	return &OrderHandler{
		service:         service,
		paymentProofDir: paymentProofDir,
	}
}

func (h *OrderHandler) CustomizePage(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please return to checkout and submit your details again.")
		return
	}

	cardID, err := parsePositiveInt64(c.PostForm("card_id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Card", "Please choose a valid product before continuing.")
		return
	}

	quantity, err := parsePositiveInt64(c.PostForm("quantity"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Quantity", "Please choose a valid quantity before continuing.")
		return
	}

	requestedInserts, err := parseNonNegativeInt64(c.PostForm("extra_inserts"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Inserts", "Please choose a valid insert count before continuing.")
		return
	}

	summary, err := h.service.PrepareCustomization(c.Request.Context(), orderapplication.CustomizationInput{
		CardID:           cardID,
		Quantity:         quantity,
		FoilOption:       c.PostForm("foil_option"),
		RequestedInserts: requestedInserts,
		Name:             c.PostForm("name"),
		Email:            c.PostForm("email"),
		Phone:            c.PostForm("phone"),
		Address:          c.PostForm("address"),
		City:             c.PostForm("city"),
		PostalCode:       c.PostForm("postal_code"),
	})
	if err != nil {
		renderServiceError(c, err)
		return
	}

	templateName := "customize.html"
	if strings.EqualFold(summary.CardCategory, "bid-boxes") {
		templateName = "customize_bid_box.html"
	}

	c.HTML(http.StatusOK, templateName, gin.H{
		"summary":   summary,
		"csrfToken": webui.EnsureCSRFToken(c),
	})
}

func (h *OrderHandler) ReviewPage(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please refresh the page and try again.")
		return
	}

	cardID, err := parsePositiveInt64(c.PostForm("card_id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Card", "Please choose a valid product before continuing.")
		return
	}

	quantity, err := parsePositiveInt64(c.PostForm("quantity"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Quantity", "Please choose a valid quantity before continuing.")
		return
	}

	requestedInserts, err := parseNonNegativeInt64(c.PostForm("extra_inserts"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Inserts", "Please choose a valid insert count before continuing.")
		return
	}

	review, err := h.service.PrepareOrderReview(c.Request.Context(), orderapplication.PlaceOrderInput{
		CardID:             cardID,
		Quantity:           quantity,
		FoilOption:         c.PostForm("foil_option"),
		RequestedInserts:   requestedInserts,
		Name:               c.PostForm("name"),
		Email:              c.PostForm("email"),
		Phone:              c.PostForm("phone"),
		Address:            c.PostForm("address"),
		City:               c.PostForm("city"),
		PostalCode:         c.PostForm("postal_code"),
		BidBoxTopLabel:     c.PostForm("top_label"),
		BidBoxCoupleName:   c.PostForm("couple_name"),
		BidBoxEventDate:    c.PostForm("event_date"),
		BidBoxDetails:      c.PostForm("details"),
		Side:               c.PostForm("side"),
		BrideName:          c.PostForm("bride_name"),
		GroomName:          c.PostForm("groom_name"),
		BrideFatherName:    c.PostForm("bride_father_name"),
		GroomFatherName:    c.PostForm("groom_father_name"),
		MehndiDate:         c.PostForm("mehndi_date"),
		MehndiDay:          c.PostForm("mehndi_day"),
		MehndiTimeType:     c.PostForm("mehndi_time_type"),
		MehndiTime:         c.PostForm("mehndi_time"),
		MehndiDinnerTime:   c.PostForm("mehndi_dinner_time"),
		MehndiVenueName:    c.PostForm("mehndi_venue_name"),
		MehndiVenueAddress: c.PostForm("mehndi_venue_address"),
		BaraatDate:         c.PostForm("baraat_date"),
		BaraatDay:          c.PostForm("baraat_day"),
		BaraatTimeType:     c.PostForm("baraat_time_type"),
		BaraatTime:         c.PostForm("baraat_time"),
		BaraatDinnerTime:   c.PostForm("baraat_dinner_time"),
		BaraatArrivalTime:  c.PostForm("baraat_arrival_time"),
		RukhsatiTime:       c.PostForm("rukhsati_time"),
		BaraatVenueName:    c.PostForm("baraat_venue_name"),
		BaraatVenueAddress: c.PostForm("baraat_venue_address"),
		NikkahDate:         c.PostForm("nikkah_date"),
		NikkahDay:          c.PostForm("nikkah_day"),
		NikkahTimeType:     c.PostForm("nikkah_time_type"),
		NikkahTime:         c.PostForm("nikkah_time"),
		NikkahDinnerTime:   c.PostForm("nikkah_dinner_time"),
		NikkahVenueName:    c.PostForm("nikkah_venue_name"),
		NikkahVenueAddress: c.PostForm("nikkah_venue_address"),
		WalimaDate:         c.PostForm("walima_date"),
		WalimaDay:          c.PostForm("walima_day"),
		WalimaTimeType:     c.PostForm("walima_time_type"),
		WalimaTime:         c.PostForm("walima_time"),
		WalimaDinnerTime:   c.PostForm("walima_dinner_time"),
		WalimaVenueName:    c.PostForm("walima_venue_name"),
		WalimaVenueAddress: c.PostForm("walima_venue_address"),
		ReceptionTime:      c.PostForm("reception_time"),
		RsvpName:           c.PostForm("rsvp_name"),
		RsvpPhone:          c.PostForm("rsvp_phone"),
		Notes:              c.PostForm("notes"),
	})
	if err != nil {
		renderServiceError(c, err)
		return
	}

	c.HTML(http.StatusOK, "review_order.html", gin.H{
		"review":    review,
		"csrfToken": webui.EnsureCSRFToken(c),
	})
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please refresh the page and try submitting the order again.")
		return
	}

	cardID, err := parsePositiveInt64(c.PostForm("card_id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Card", "Please choose a valid product before submitting the order.")
		return
	}

	quantity, err := parsePositiveInt64(c.PostForm("quantity"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Quantity", "Please choose a valid quantity before submitting the order.")
		return
	}

	requestedInserts, err := parseNonNegativeInt64(c.PostForm("extra_inserts"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Inserts", "Please choose a valid insert count before submitting the order.")
		return
	}

	result, err := h.service.PlaceOrder(c.Request.Context(), orderapplication.PlaceOrderInput{
		CardID:             cardID,
		Quantity:           quantity,
		FoilOption:         c.PostForm("foil_option"),
		RequestedInserts:   requestedInserts,
		Name:               c.PostForm("name"),
		Email:              c.PostForm("email"),
		Phone:              c.PostForm("phone"),
		Address:            c.PostForm("address"),
		City:               c.PostForm("city"),
		PostalCode:         c.PostForm("postal_code"),
		BidBoxTopLabel:     c.PostForm("top_label"),
		BidBoxCoupleName:   c.PostForm("couple_name"),
		BidBoxEventDate:    c.PostForm("event_date"),
		BidBoxDetails:      c.PostForm("details"),
		Side:               c.PostForm("side"),
		BrideName:          c.PostForm("bride_name"),
		GroomName:          c.PostForm("groom_name"),
		BrideFatherName:    c.PostForm("bride_father_name"),
		GroomFatherName:    c.PostForm("groom_father_name"),
		MehndiDate:         c.PostForm("mehndi_date"),
		MehndiDay:          c.PostForm("mehndi_day"),
		MehndiTimeType:     c.PostForm("mehndi_time_type"),
		MehndiTime:         c.PostForm("mehndi_time"),
		MehndiDinnerTime:   c.PostForm("mehndi_dinner_time"),
		MehndiVenueName:    c.PostForm("mehndi_venue_name"),
		MehndiVenueAddress: c.PostForm("mehndi_venue_address"),
		BaraatDate:         c.PostForm("baraat_date"),
		BaraatDay:          c.PostForm("baraat_day"),
		BaraatTimeType:     c.PostForm("baraat_time_type"),
		BaraatTime:         c.PostForm("baraat_time"),
		BaraatDinnerTime:   c.PostForm("baraat_dinner_time"),
		BaraatArrivalTime:  c.PostForm("baraat_arrival_time"),
		RukhsatiTime:       c.PostForm("rukhsati_time"),
		BaraatVenueName:    c.PostForm("baraat_venue_name"),
		BaraatVenueAddress: c.PostForm("baraat_venue_address"),
		NikkahDate:         c.PostForm("nikkah_date"),
		NikkahDay:          c.PostForm("nikkah_day"),
		NikkahTimeType:     c.PostForm("nikkah_time_type"),
		NikkahTime:         c.PostForm("nikkah_time"),
		NikkahDinnerTime:   c.PostForm("nikkah_dinner_time"),
		NikkahVenueName:    c.PostForm("nikkah_venue_name"),
		NikkahVenueAddress: c.PostForm("nikkah_venue_address"),
		WalimaDate:         c.PostForm("walima_date"),
		WalimaDay:          c.PostForm("walima_day"),
		WalimaTimeType:     c.PostForm("walima_time_type"),
		WalimaTime:         c.PostForm("walima_time"),
		WalimaDinnerTime:   c.PostForm("walima_dinner_time"),
		WalimaVenueName:    c.PostForm("walima_venue_name"),
		WalimaVenueAddress: c.PostForm("walima_venue_address"),
		ReceptionTime:      c.PostForm("reception_time"),
		RsvpName:           c.PostForm("rsvp_name"),
		RsvpPhone:          c.PostForm("rsvp_phone"),
		Notes:              c.PostForm("notes"),
	})
	if err != nil {
		renderServiceError(c, err)
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/order/%d/payment", result.OrderID))
}

func (h *OrderHandler) OrderConfirmation(c *gin.Context) {
	orderID, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/order/%d/payment", orderID))
}

func (h *OrderHandler) OrderStatus(c *gin.Context) {
	orderID, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
		return
	}

	payload, err := h.service.GetOrderStatusDetail(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, orderapplication.ErrInvalidInput) {
			webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that number.")
			return
		}

		log.Println("ORDER STATUS ERROR:", err)
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the order status right now.")
		return
	}

	c.HTML(http.StatusOK, "order-status.html", gin.H{
		"order":                payload.Order,
		"customer":             payload.Customer,
		"details":              payload.Details,
		"payment":              payload.Payment,
		"isBidBox":             isBidBoxOrder(payload.Order, payload.Details),
		"statusMessage":        orderStatusMessage(payload.Order.Status),
		"paymentMessage":       orderapplication.PaymentStatusMessage(payload.Payment, payload.Order.TotalPrice),
		"amountSummary":        paymentAmountSummary(payload.Order, payload.Payment),
		"paymentStatusDisplay": paymentStatusDisplay(payload.Payment),
		"remainingStatus":      remainingBalanceStatus(payload.Order, payload.Payment),
	})
}

func renderServiceError(c *gin.Context, err error) {
	var minOrderErr orderapplication.MinOrderError
	if errors.As(err, &minOrderErr) {
		webui.RenderError(c, http.StatusBadRequest, "Minimum Order Not Met", fmt.Sprintf("The minimum order quantity for this design is %d.", minOrderErr.MinOrder))
		return
	}
	if errors.Is(err, orderapplication.ErrInvalidInput) {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Details", "Please review the submitted details and try again.")
		return
	}
	if errors.Is(err, orderapplication.ErrCardNotFound) || errors.Is(err, pgx.ErrNoRows) {
		webui.RenderError(c, http.StatusNotFound, "Card Not Found", "The selected product is no longer available.")
		return
	}
	log.Println("ORDER ERROR:", err)
	webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not process the request right now.")
}

func parsePositiveInt64(raw string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value < 1 {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}

func parseNonNegativeInt64(raw string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value < 0 {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}

func (h *OrderHandler) AdminOrders(c *gin.Context) {
	orders, err := h.service.ListAdminOrders(c.Request.Context())
	if err != nil {
		log.Println("ADMIN ORDERS ERROR:", err)
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load admin orders right now.")
		return
	}

	c.HTML(http.StatusOK, "admin_orders.html", gin.H{
		"orders": orders,
	})
}

func (h *OrderHandler) AdminOrderDetail(c *gin.Context) {
	orderID, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
		return
	}

	payload, err := h.service.GetAdminOrderDetail(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, orderapplication.ErrInvalidInput) {
			webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that number.")
			return
		}

		log.Println("ADMIN ORDER DETAIL ERROR:", err)
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the order details right now.")
		return
	}

	canConfirmOrder := canConfirmOrderFromPayment(payload.Payment)
	alertTone, alertMessage := adminOrderDetailAlert(c)
	paymentStatusLabel := paymentStatusDisplay(payload.Payment)
	paymentCanBeReviewed := false
	if payload.Payment != nil {
		paymentCanBeReviewed = payload.Payment.PaymentStatus == orderdomain.AwaitingVerificationPaymentStatus
	}
	amountSummary := paymentAmountSummary(payload.Order, payload.Payment)

	c.HTML(http.StatusOK, "admin_order_detail.html", gin.H{
		"order":                payload.Order,
		"statusStr":            string(payload.Order.Status),
		"customer":             payload.Customer,
		"details":              payload.Details,
		"payment":              payload.Payment,
		"card_name":            payload.Order.CardName,
		"card_image":           payload.Order.CardImage,
		"isBidBox":             isBidBoxOrder(payload.Order, payload.Details),
		"csrfToken":            webui.EnsureCSRFToken(c),
		"canConfirmOrder":      canConfirmOrder,
		"paymentNeedsReview":   !canConfirmOrder,
		"paymentCanBeReviewed": paymentCanBeReviewed,
		"paymentStatusDisplay": paymentStatusLabel,
		"amountSummary":        amountSummary,
		"remainingStatus":      remainingBalanceStatus(payload.Order, payload.Payment),
		"adminAlertTone":       alertTone,
		"adminAlertMessage":    alertMessage,
	})
}

func (h *OrderHandler) AdminUpdateOrderStatus(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		log.Println("ADMIN STATUS UPDATE ERROR: request expired due to CSRF validation failure")
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please refresh the page and try updating the order again.")
		return
	}

	orderID, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		log.Printf("ADMIN STATUS UPDATE ERROR: invalid order id raw=%q err=%v", c.Param("id"), err)
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
		return
	}

	requestedStatus := c.PostForm("status")
	log.Printf("ADMIN STATUS UPDATE START: order_id=%d requested_status=%q", orderID, requestedStatus)

	if err := h.service.AdminUpdateOrderStatus(c.Request.Context(), orderID, requestedStatus); err != nil {
		log.Printf("ADMIN STATUS UPDATE ERROR: order_id=%d requested_status=%q err=%v", orderID, requestedStatus, err)
		if errors.Is(err, orderapplication.ErrPaymentVerificationRequired) {
			c.Redirect(http.StatusSeeOther, adminOrderDetailRedirect(orderID, "status_notice=payment_verification_required"))
			return
		}
		if errors.Is(err, orderapplication.ErrInvalidInput) {
			webui.RenderError(c, http.StatusBadRequest, "Invalid Status", "Please choose a valid order status.")
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Order Not Found", "We could not find an order with that number.")
			return
		}
		log.Println("ADMIN STATUS UPDATE ERROR:", err)
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not update the order status right now.")
		return
	}

	log.Printf("ADMIN STATUS UPDATE SUCCESS: order_id=%d requested_status=%q redirect=/admin/orders/%d", orderID, requestedStatus, orderID)
	c.Redirect(http.StatusSeeOther, "/admin/orders/"+strconv.FormatInt(orderID, 10))
}

func (h *OrderHandler) TrackOrderPage(c *gin.Context) {
	rawOrderID := strings.TrimSpace(c.Query("order_id"))
	if rawOrderID == "" {
		c.HTML(http.StatusOK, "track-order.html", nil)
		return
	}

	orderID, err := parsePositiveInt64(rawOrderID)
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Order", "Please enter a valid order number.")
		return
	}

	c.Redirect(http.StatusSeeOther, "/order/"+strconv.FormatInt(orderID, 10))
}

func orderStatusMessage(status orderdomain.OrderStatus) string {
	switch status {
	case orderdomain.PendingOrderStatus:
		return "Your order is pending review and production confirmation."
	case orderdomain.ConfirmedOrderStatus:
		return "Your order has been confirmed and is moving through production."
	case orderdomain.CancelledOrderStatus:
		return "This order has been cancelled. Contact us if you need help."
	case orderdomain.CompletedOrderStatus:
		return "Your order is completed."
	default:
		return "Your order status has been updated."
	}
}

func isBidBoxOrder(order *orderdomain.Order, details *orderdomain.OrderDetail) bool {
	if order != nil && strings.EqualFold(strings.TrimSpace(order.CardCategory), "bid-boxes") {
		return true
	}
	if details == nil {
		return false
	}
	return details.BidBoxTopLabel != nil ||
		details.BidBoxCoupleName != nil ||
		details.BidBoxEventDate != nil ||
		details.BidBoxDetails != nil
}

func canConfirmOrderFromPayment(payment *orderdomain.OrderPayment) bool {
	return payment != nil && payment.PaymentStatus == orderdomain.VerifiedPaymentStatus
}

func paymentAmountSummary(order *orderdomain.Order, payment *orderdomain.OrderPayment) orderapplication.PaymentAmountSummary {
	totalAmount := int64(0)
	expectedAdvanceAmount := int64(0)
	if order != nil {
		totalAmount = order.TotalPrice
	}
	if payment != nil {
		expectedAdvanceAmount = payment.ExpectedAmount
	}
	return orderapplication.BuildPaymentAmountSummary(totalAmount, expectedAdvanceAmount)
}

func remainingBalanceStatus(order *orderdomain.Order, payment *orderdomain.OrderPayment) string {
	summary := paymentAmountSummary(order, payment)
	if summary.RemainingBalance <= 0 {
		return "No Remaining Balance"
	}
	return "Pending Remaining Balance"
}

func paymentStatusDisplay(payment *orderdomain.OrderPayment) string {
	if payment == nil {
		return orderapplication.PaymentStatusDisplay(orderdomain.PaymentStatus(""))
	}
	return orderapplication.PaymentStatusDisplay(payment.PaymentStatus)
}

func adminOrderDetailRedirect(orderID int64, rawQuery string) string {
	path := "/admin/orders/" + strconv.FormatInt(orderID, 10)
	if strings.TrimSpace(rawQuery) == "" {
		return path
	}
	return path + "?" + rawQuery
}

func adminOrderDetailAlert(c *gin.Context) (string, string) {
	switch strings.TrimSpace(c.Query("status_notice")) {
	case "payment_verification_required":
		return "warning", "Payment must be verified before confirming this order."
	}

	switch strings.TrimSpace(c.Query("payment_notice")) {
	case "payment_verified":
		return "success", "Advance payment verified. The order can now proceed normally."
	case "payment_rejected":
		return "warning", "Advance payment was rejected. The customer can re-upload payment proof."
	case "payment_reupload_requested":
		return "warning", "Re-upload requested. The customer can submit a new advance payment proof."
	}

	switch strings.TrimSpace(c.Query("payment_error")) {
	case "action_not_allowed":
		return "warning", "This payment is not currently awaiting verification."
	case "invalid_action":
		return "warning", "Choose a valid payment action for this order."
	case "amount_too_low":
		return "warning", "The submitted advance amount is below the trusted 50% advance due. Do not verify this payment until the customer pays the full required advance amount."
	}

	return "", ""
}
