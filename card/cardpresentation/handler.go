package cardpresentation

import (
	"net/http"
	"strconv"
	"strings"

	"writeandinviteco/inviteandco/card/carddomain"

	"github.com/gin-gonic/gin"
)

type CardHandler struct {
	repo carddomain.CardRepo
}

func NewCardHandler(repo carddomain.CardRepo) *CardHandler {
	return &CardHandler{repo: repo}
}

func (h *CardHandler) ListCards(c *gin.Context) {
	cards, err := h.repo.GetCardsByCategory(c.Request.Context(), "wedding-cards")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load cards")
		return
	}
	if len(cards) > 4 {
		cards = cards[:4]
	}
	c.HTML(http.StatusOK, "home.html", gin.H{
		"cards": cards,
	})
}

func (h *CardHandler) ListCardsByCategory(c *gin.Context) {
	category := c.Param("category")
	cards, err := h.repo.GetCardsByCategory(c.Request.Context(), category)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load cards")
		return
	}

	categoryNames := map[string]string{
		"wedding-cards":       "Wedding Cards",
		"bid-boxes":           "Bid Boxes",
		"nikkah-certificate":  "Nikkah Certificate",
		"wedding-accessories": "Wedding Accessories",
	}

	displayName := categoryNames[category]
	if displayName == "" {
		displayName = category
	}

	c.HTML(http.StatusOK, "collection.html", gin.H{
		"cards":        cards,
		"category":     category,
		"categoryName": displayName,
	})
}

func (h *CardHandler) SearchCards(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.HTML(http.StatusOK, "search.html", gin.H{
			"query": "",
			"cards": []*carddomain.Card{},
			"count": 0,
		})
		return
	}

	cards, err := h.repo.SearchCards(c.Request.Context(), query)
	if err != nil {
		c.String(http.StatusInternalServerError, "Search failed")
		return
	}

	c.HTML(http.StatusOK, "search.html", gin.H{
		"query": query,
		"cards": cards,
		"count": len(cards),
	})
}

func (h *CardHandler) Checkout(c *gin.Context) {
	cardID := c.Query("card_id")
	cardName := c.Query("card_name")
	cardImage := c.Query("card_image")
	quantity := parseInt64Default(c.Query("quantity"), 1)
	priceFoil := parseInt64Default(c.Query("price_foil"), 0)
	priceNofoil := parseInt64Default(c.Query("price_nofoil"), priceFoil)
	insertPrice := parseInt64Default(c.Query("insert_price"), 0)
	extraInserts := parseInt64Default(c.Query("extra_inserts"), 0)
	minOrder := parseInt64Default(c.Query("min_order"), 1)
	includedInserts := parseInt64Default(c.Query("included_inserts"), 2)
	currency := strings.ToUpper(strings.TrimSpace(c.Query("currency")))
	foilOption := strings.ToLower(strings.TrimSpace(c.Query("foil_option")))

	if currency == "" {
		currency = "PKR"
	}
	if foilOption != "nofoil" {
		foilOption = "foil"
	}
	if minOrder < 1 {
		minOrder = 1
	}
	if quantity < minOrder {
		quantity = minOrder
	}
	if quantity < 1 {
		quantity = 1
	}
	if extraInserts < 0 {
		extraInserts = 0
	}
	if includedInserts < 0 {
		includedInserts = 0
	}
	if priceNofoil == 0 {
		priceNofoil = priceFoil
	}

	unitPrice := priceFoil
	foilLabel := "With Foil"
	if foilOption == "nofoil" {
		unitPrice = priceNofoil
		foilLabel = "Without Foil"
	}
	if priceFoil == priceNofoil {
		foilLabel = "Flat Rate"
	}
	extraInsertCostPerCard := extraInserts * insertPrice
	perCardTotal := unitPrice + extraInsertCostPerCard
	total := perCardTotal * quantity

	c.HTML(http.StatusOK, "checkout.html", gin.H{
		"cardID":                 cardID,
		"cardName":               cardName,
		"cardImage":              cardImage,
		"quantity":               quantity,
		"currency":               currency,
		"priceFoil":              priceFoil,
		"priceNofoil":            priceNofoil,
		"insertPrice":            insertPrice,
		"includedInserts":        includedInserts,
		"extraInserts":           extraInserts,
		"minOrder":               minOrder,
		"foilOption":             foilOption,
		"foilLabel":              foilLabel,
		"unitPrice":              unitPrice,
		"extraInsertCostPerCard": extraInsertCostPerCard,
		"perCardTotal":           perCardTotal,
		"totalPrice":             total,
	})
}

func (h *CardHandler) CardDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid card id")
		return
	}

	card, err := h.repo.GetCardByID(c.Request.Context(), int64(id))
	if err != nil {
		c.String(http.StatusNotFound, "Card not found")
		return
	}

	images, err := h.repo.GetCardImagesByCardID(c.Request.Context(), int64(id))
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load card images")
		return
	}

	c.HTML(http.StatusOK, "card.html", gin.H{
		"card":   card,
		"images": images,
	})
}

func parseInt64Default(raw string, fallback int64) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
