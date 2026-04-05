package orderpresentation

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"writeandinviteco/inviteandco/order/orderapplication"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service *orderapplication.Service
}

func NewOrderHandler(service *orderapplication.Service) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CustomizePage(c *gin.Context) {
	cardID, err := parsePositiveInt64(c.Query("card_id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid card id.")
		return
	}

	quantity, err := parsePositiveInt64(c.Query("quantity"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid quantity.")
		return
	}

	requestedInserts, err := parseNonNegativeInt64(c.Query("extra_inserts"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid inserts quantity.")
		return
	}

	summary, err := h.service.PrepareCustomization(c.Request.Context(), orderapplication.CustomizationInput{
		CardID:           cardID,
		Quantity:         quantity,
		Currency:         c.Query("currency"),
		FoilOption:       c.Query("foil_option"),
		RequestedInserts: requestedInserts,
		Name:             c.Query("name"),
		Email:            c.Query("email"),
		Phone:            c.Query("phone"),
		Address:          c.Query("address"),
		City:             c.Query("city"),
		PostalCode:       c.Query("postal_code"),
	})
	if err != nil {
		renderServiceError(c, err)
		return
	}

	c.HTML(http.StatusOK, "customize.html", gin.H{
		"summary": summary,
	})
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	cardID, err := parsePositiveInt64(c.PostForm("card_id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid card id.")
		return
	}

	quantity, err := parsePositiveInt64(c.PostForm("quantity"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid quantity.")
		return
	}

	requestedInserts, err := parseNonNegativeInt64(c.PostForm("extra_inserts"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid inserts quantity.")
		return
	}

	result, err := h.service.PlaceOrder(c.Request.Context(), orderapplication.PlaceOrderInput{
		CardID:            cardID,
		Quantity:          quantity,
		Currency:          c.PostForm("currency"),
		FoilOption:        c.PostForm("foil_option"),
		RequestedInserts:  requestedInserts,
		Name:              c.PostForm("name"),
		Email:             c.PostForm("email"),
		Phone:             c.PostForm("phone"),
		Address:           c.PostForm("address"),
		City:              c.PostForm("city"),
		PostalCode:        c.PostForm("postal_code"),
		Side:              c.PostForm("side"),
		BrideName:         c.PostForm("bride_name"),
		GroomName:         c.PostForm("groom_name"),
		BrideFatherName:   c.PostForm("bride_father_name"),
		GroomFatherName:   c.PostForm("groom_father_name"),
		MehndiDate:        c.PostForm("mehndi_date"),
		MehndiTimeType:    c.PostForm("mehndi_time_type"),
		MehndiTime:        c.PostForm("mehndi_time"),
		MehndiDinnerTime:  c.PostForm("mehndi_dinner_time"),
		BaraatDate:        c.PostForm("baraat_date"),
		BaraatTimeType:    c.PostForm("baraat_time_type"),
		BaraatTime:        c.PostForm("baraat_time"),
		BaraatDinnerTime:  c.PostForm("baraat_dinner_time"),
		BaraatArrivalTime: c.PostForm("baraat_arrival_time"),
		RukhsatiTime:      c.PostForm("rukhsati_time"),
		NikkahDate:        c.PostForm("nikkah_date"),
		NikkahTimeType:    c.PostForm("nikkah_time_type"),
		NikkahTime:        c.PostForm("nikkah_time"),
		NikkahDinnerTime:  c.PostForm("nikkah_dinner_time"),
		WalimaDate:        c.PostForm("walima_date"),
		WalimaTimeType:    c.PostForm("walima_time_type"),
		WalimaTime:        c.PostForm("walima_time"),
		WalimaDinnerTime:  c.PostForm("walima_dinner_time"),
		ReceptionTime:     c.PostForm("reception_time"),
		DinnerTime:        c.PostForm("dinner_time"),
		VenueName:         c.PostForm("venue_name"),
		VenueAddress:      c.PostForm("venue_address"),
		RsvpName:          c.PostForm("rsvp_name"),
		RsvpPhone:         c.PostForm("rsvp_phone"),
		Notes:             c.PostForm("notes"),
	})
	if err != nil {
		renderServiceError(c, err)
		return
	}

	c.HTML(http.StatusOK, "order-confirmation.html", gin.H{
		"customerName": result.CustomerName,
		"quantity":     result.Quantity,
		"totalPrice":   result.TotalPrice,
		"currency":     result.Currency,
		"cardName":     result.CardName,
	})
}

func renderServiceError(c *gin.Context, err error) {
	var minOrderErr orderapplication.MinOrderError
	if errors.As(err, &minOrderErr) {
		c.String(http.StatusBadRequest, "Minimum order quantity is %d.", minOrderErr.MinOrder)
		return
	}
	if errors.Is(err, orderapplication.ErrInvalidInput) {
		c.String(http.StatusBadRequest, "Please provide valid checkout details.")
		return
	}
	if errors.Is(err, orderapplication.ErrCardNotFound) {
		c.String(http.StatusNotFound, "Card not found.")
		return
	}
	log.Println("ORDER ERROR:", err)
	c.String(http.StatusInternalServerError, err.Error())
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
