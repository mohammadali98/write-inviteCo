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
	c.HTML(http.StatusOK, "home.html", gin.H{
		"cards": cards,
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
	c.HTML(http.StatusOK, "card.html", gin.H{
		"card": card,
	})
}
