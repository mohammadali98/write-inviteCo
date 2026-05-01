package cardpresentation

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"writeandinviteco/inviteandco/card/carddomain"
	"writeandinviteco/inviteandco/product/productapplication"
	"writeandinviteco/inviteandco/webui"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type CardHandler struct {
	repo           carddomain.CardRepo
	productService *productapplication.Service
}

const (
	maxCheckoutQuantity     = 5000
	maxCheckoutExtraInserts = 20
	maxSearchQueryLength    = 100
)

func NewCardHandler(repo carddomain.CardRepo, productService *productapplication.Service) *CardHandler {
	return &CardHandler{
		repo:           repo,
		productService: productService,
	}
}

func (h *CardHandler) ListCards(c *gin.Context) {
	cards, err := h.repo.GetCardsByCategory(c.Request.Context(), "wedding-cards")
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the collection right now.")
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
	if _, ok := validCategoryNames()[category]; !ok {
		webui.RenderError(c, http.StatusNotFound, "Collection Not Found", "The requested collection does not exist.")
		return
	}
	query := strings.TrimSpace(c.Query("q"))
	if len([]rune(query)) > maxSearchQueryLength {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Search", "Search terms must be shorter than 100 characters.")
		return
	}

	products, err := h.frontendProducts(c.Request.Context(), category, query)
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the collection right now.")
		return
	}

	c.HTML(http.StatusOK, "collection.html", gin.H{
		"products":     products,
		"category":     category,
		"categoryName": validCategoryNames()[category],
		"query":        query,
	})
}

func (h *CardHandler) SearchCards(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.HTML(http.StatusOK, "search.html", gin.H{
			"query": "",
			"cards": []*carddomain.Card{},
			"count": 0,
		})
		return
	}
	if len([]rune(query)) > maxSearchQueryLength {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Search", "Search terms must be shorter than 100 characters.")
		return
	}

	cards, err := h.repo.SearchCards(c.Request.Context(), query)
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "Search is unavailable right now.")
		return
	}

	c.HTML(http.StatusOK, "search.html", gin.H{
		"query": query,
		"cards": cards,
		"count": len(cards),
	})
}

func (h *CardHandler) Checkout(c *gin.Context) {
	cardID, err := parsePositiveInt64(c.Query("card_id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Card", "Please choose a valid product before continuing to checkout.")
		return
	}

	card, err := h.repo.GetCardByID(c.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Card Not Found", "The selected product is no longer available.")
			return
		}
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the selected product right now.")
		return
	}

	minOrder := int64(card.MinOrder)
	if minOrder < 1 {
		minOrder = 1
	}

	quantity := parseInt64Default(c.Query("quantity"), minOrder)
	if quantity < minOrder {
		quantity = minOrder
	}
	if quantity > maxCheckoutQuantity {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Quantity", "Please choose a quantity within the supported checkout range.")
		return
	}

	requestedInserts := parseInt64Default(c.Query("extra_inserts"), 0)
	if requestedInserts < 0 || requestedInserts > maxCheckoutExtraInserts {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Inserts", "Please enter a valid insert count per card.")
		return
	}

	currency := "PKR"

	foilOption, err := parseCheckoutFoilOption(c.Query("foil_option"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Finish", "Please choose a valid foil option.")
		return
	}

	priceFoil := card.PriceFoilPKR
	priceNofoil := card.PriceNofoilPKR
	insertPrice := card.InsertPricePKR
	if priceFoil < 0 || priceNofoil < 0 || insertPrice < 0 {
		webui.RenderError(c, http.StatusInternalServerError, "Pricing Error", "The selected product has invalid pricing data.")
		return
	}
	if priceNofoil == 0 {
		priceNofoil = priceFoil
	}

	includedInserts := int64(card.IncludedInserts)
	if includedInserts < 0 {
		includedInserts = 0
	}
	extraInserts := requestedInserts - includedInserts
	if extraInserts < 0 {
		extraInserts = 0
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

	extraInsertCostPerCard, ok := safeMultiplyInt64(extraInserts, insertPrice)
	if !ok {
		webui.RenderError(c, http.StatusInternalServerError, "Pricing Error", "The selected order exceeds the supported pricing range.")
		return
	}
	perCardTotal, ok := safeAddInt64(unitPrice, extraInsertCostPerCard)
	if !ok {
		webui.RenderError(c, http.StatusInternalServerError, "Pricing Error", "The selected order exceeds the supported pricing range.")
		return
	}
	total, ok := safeMultiplyInt64(perCardTotal, quantity)
	if !ok {
		webui.RenderError(c, http.StatusInternalServerError, "Pricing Error", "The selected order exceeds the supported pricing range.")
		return
	}

	c.HTML(http.StatusOK, "checkout.html", gin.H{
		"cardID":                 card.ID,
		"cardName":               card.Name,
		"cardImage":              card.Image,
		"cardCategory":           card.Category,
		"quantity":               quantity,
		"currency":               currency,
		"requestedInserts":       requestedInserts,
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
		"csrfToken":              webui.EnsureCSRFToken(c),
	})
}

func (h *CardHandler) CardDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Card", "Please choose a valid product.")
		return
	}

	card, err := h.repo.GetCardByID(c.Request.Context(), int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			webui.RenderError(c, http.StatusNotFound, "Card Not Found", "The selected product is no longer available.")
			return
		}
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the selected product right now.")
		return
	}

	images, err := h.repo.GetCardImagesByCardID(c.Request.Context(), int64(id))
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the product gallery right now.")
		return
	}
	if len(images) == 0 {
		images = []*carddomain.CardImage{{Image: card.Image}}
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

func parsePositiveInt64(raw string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value < 1 {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}

func parseCheckoutCurrency(raw string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "", "PKR":
		return "PKR", nil
	case "NOK":
		return "NOK", nil
	default:
		return "", strconv.ErrSyntax
	}
}

func parseCheckoutFoilOption(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "foil":
		return "foil", nil
	case "nofoil":
		return "nofoil", nil
	default:
		return "", strconv.ErrSyntax
	}
}

func validCategoryNames() map[string]string {
	return map[string]string{
		"wedding-cards":       "Wedding Cards",
		"bid-boxes":           "Bid Boxes",
		"nikkah-certificate":  "Nikkah Certificate",
		"wedding-accessories": "Wedding Accessories",
	}
}

type frontendProduct struct {
	ID        int64
	Name      string
	ImageURL  string
	Price     int64
	DetailURL string
}

func (h *CardHandler) frontendProducts(ctx context.Context, category string, query string) ([]frontendProduct, error) {
	productCategory := routeCategoryToProductCategory(category)
	if productCategory != "" && h.productService != nil {
		items, err := h.productService.ListProducts(ctx)
		if err == nil {
			filtered := make([]frontendProduct, 0, len(items))
			hasProductCollection := false
			for _, item := range items {
				if item == nil || !item.IsActive || item.Category != productCategory || item.CardID <= 0 {
					continue
				}
				hasProductCollection = true
				if !matchesTextSearch(item.Name, item.Description, query) {
					continue
				}
				filtered = append(filtered, frontendProduct{
					ID:        item.ID,
					Name:      item.Name,
					ImageURL:  imageOrFallback(item.ImageURL),
					Price:     item.Price,
					DetailURL: "/product/" + strconv.FormatInt(item.ID, 10),
				})
			}
			if hasProductCollection {
				return filtered, nil
			}
		}
	}

	var cards []*carddomain.Card
	var err error
	if strings.TrimSpace(query) != "" {
		cards, err = h.repo.SearchCards(ctx, query)
	} else {
		cards, err = h.repo.GetCardsByCategory(ctx, category)
	}
	if err != nil {
		return nil, err
	}

	filtered := make([]frontendProduct, 0, len(cards))
	for _, card := range cards {
		if card == nil || card.Category != category {
			continue
		}
		price := card.PriceNofoilPKR
		if price <= 0 {
			price = card.PriceFoilPKR
		}
		filtered = append(filtered, frontendProduct{
			ID:        card.ID,
			Name:      card.Name,
			ImageURL:  imageOrFallback(card.Image),
			Price:     price,
			DetailURL: "/card/" + strconv.FormatInt(card.ID, 10),
		})
	}

	return filtered, nil
}

func matchesTextSearch(name string, description string, query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return true
	}

	query = strings.ToLower(query)
	return strings.Contains(strings.ToLower(name), query) ||
		strings.Contains(strings.ToLower(description), query)
}

func routeCategoryToProductCategory(category string) string {
	switch category {
	case "wedding-cards":
		return "card"
	case "bid-boxes":
		return "bid_box"
	default:
		return ""
	}
}

func imageOrFallback(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "/static/sample.jpg"
	}
	return raw
}

func safeAddInt64(a int64, b int64) (int64, bool) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, false
	}
	if b < 0 && a < math.MinInt64-b {
		return 0, false
	}
	return a + b, true
}

func safeMultiplyInt64(a int64, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	if a < 0 || b < 0 {
		return 0, false
	}
	if a > math.MaxInt64/b {
		return 0, false
	}
	return a * b, true
}
