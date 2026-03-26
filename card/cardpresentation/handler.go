package cardpresentation

import (
	"net/http"
	"strconv"

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
	cards, err := h.repo.GetAllCards(c.Request.Context())
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
