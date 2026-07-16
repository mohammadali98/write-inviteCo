package cardpresentation

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"writeandinviteco/inviteandco/accessory/accessorydomain"
	"writeandinviteco/inviteandco/card/carddomain"
	"writeandinviteco/inviteandco/product/productapplication"
	"writeandinviteco/inviteandco/webui"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type CardHandler struct {
	repo           carddomain.CardRepo
	accessoryRepo  accessorydomain.AccessoryReader
	productService *productapplication.Service
}

type CheckoutPersonalization struct {
	BidBoxTopLabel     string
	BidBoxCoupleName   string
	BidBoxEventDate    string
	BidBoxDetails      string
	Side               string
	BrideName          string
	GroomName          string
	BrideFatherName    string
	GroomFatherName    string
	MehndiDate         string
	MehndiDay          string
	MehndiTimeType     string
	MehndiTime         string
	MehndiDinnerTime   string
	MehndiVenueName    string
	MehndiVenueAddress string
	BaraatDate         string
	BaraatDay          string
	BaraatTimeType     string
	BaraatTime         string
	BaraatDinnerTime   string
	BaraatArrivalTime  string
	RukhsatiTime       string
	BaraatVenueName    string
	BaraatVenueAddress string
	NikkahDate         string
	NikkahDay          string
	NikkahTimeType     string
	NikkahTime         string
	NikkahDinnerTime   string
	NikkahVenueName    string
	NikkahVenueAddress string
	WalimaDate         string
	WalimaDay          string
	WalimaTimeType     string
	WalimaTime         string
	WalimaDinnerTime   string
	WalimaVenueName    string
	WalimaVenueAddress string
	ReceptionTime      string
	RsvpName           string
	RsvpPhone          string
	Notes              string
}

const (
	maxCheckoutQuantity     = 5000
	maxCheckoutExtraInserts = 20
	maxSearchQueryLength    = 100
	// Must match bulkDiscountMinQty/bulkDiscountPercent in order/orderapplication/service.go
	// so this preview always agrees with the server-side charge.
	checkoutBulkDiscountMinQty  = 70
	checkoutBulkDiscountPercent = 15
)

func NewCardHandler(repo carddomain.CardRepo, accessoryRepo accessorydomain.AccessoryReader, productService *productapplication.Service) *CardHandler {
	return &CardHandler{
		repo:           repo,
		accessoryRepo:  accessoryRepo,
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

type accessoryGalleryItem struct {
	ID          int64
	Name        string
	Description string
	Images      []string
}

// ListAccessories renders the wedding-accessories showcase. This is a
// deliberately separate route/handler from ListCardsByCategory: accessories
// are never purchasable, have no price, and never link into the card or
// checkout flow, so they're kept fully out of frontendProducts.
func (h *CardHandler) ListAccessories(c *gin.Context) {
	if h.accessoryRepo == nil {
		c.HTML(http.StatusOK, "accessories.html", gin.H{
			"categoryName": validCategoryNames()["wedding-accessories"],
			"accessories":  []accessoryGalleryItem{},
		})
		return
	}

	accessories, err := h.accessoryRepo.ListActiveAccessories(c.Request.Context())
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the collection right now.")
		return
	}

	items := make([]accessoryGalleryItem, 0, len(accessories))
	for _, a := range accessories {
		if a == nil {
			continue
		}
		images, err := h.accessoryRepo.GetAccessoryImagesByAccessoryID(c.Request.Context(), a.ID)
		if err != nil {
			webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the collection right now.")
			return
		}
		imageURLs := make([]string, 0, len(images))
		for _, img := range images {
			if img == nil {
				continue
			}
			imageURLs = append(imageURLs, imageOrFallback(img.ImageURL))
		}

		description := ""
		if a.Description != nil {
			description = *a.Description
		}

		items = append(items, accessoryGalleryItem{
			ID:          a.ID,
			Name:        a.Name,
			Description: description,
			Images:      imageURLs,
		})
	}

	c.HTML(http.StatusOK, "accessories.html", gin.H{
		"categoryName": validCategoryNames()["wedding-accessories"],
		"accessories":  items,
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
	if c.Request.Method == http.MethodPost && !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please return to personalization and submit your details again.")
		return
	}

	cardID, err := parsePositiveInt64(checkoutValue(c, "card_id"))
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

	quantity := parseInt64Default(checkoutValue(c, "quantity"), minOrder)
	if quantity < minOrder {
		quantity = minOrder
	}
	if quantity > maxCheckoutQuantity {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Quantity", "Please choose a quantity within the supported checkout range.")
		return
	}

	requestedInserts := parseInt64Default(checkoutValue(c, "extra_inserts"), 0)
	if requestedInserts < 0 || requestedInserts > maxCheckoutExtraInserts {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Inserts", "Please enter a valid insert count per card.")
		return
	}

	currency := "PKR"

	foilOption, err := parseCheckoutFoilOption(checkoutValue(c, "foil_option"))
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
	extraInserts := requestedInserts

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

	cardSubtotal, ok := safeMultiplyInt64(unitPrice, quantity)
	if !ok {
		webui.RenderError(c, http.StatusInternalServerError, "Pricing Error", "The selected order exceeds the supported pricing range.")
		return
	}
	insertSubtotal, ok := safeMultiplyInt64(extraInsertCostPerCard, quantity)
	if !ok {
		webui.RenderError(c, http.StatusInternalServerError, "Pricing Error", "The selected order exceeds the supported pricing range.")
		return
	}

	discountApplied := quantity > checkoutBulkDiscountMinQty
	discountAmount := int64(0)
	if discountApplied {
		discounted, ok := safeMultiplyInt64(cardSubtotal, checkoutBulkDiscountPercent)
		if !ok {
			webui.RenderError(c, http.StatusInternalServerError, "Pricing Error", "The selected order exceeds the supported pricing range.")
			return
		}
		discountAmount = discounted / 100
	}

	total, ok := safeAddInt64(cardSubtotal-discountAmount, insertSubtotal)
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
		"discountApplied":        discountApplied,
		"discountAmount":         discountAmount,
		"totalPrice":             total,
		"personalization":        readCheckoutPersonalization(c),
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
		"card":      card,
		"images":    images,
		"csrfToken": webui.EnsureCSRFToken(c),
	})
}

func checkoutValue(c *gin.Context, key string) string {
	if c.Request.Method == http.MethodPost {
		return c.PostForm(key)
	}
	return c.Query(key)
}

func readCheckoutPersonalization(c *gin.Context) CheckoutPersonalization {
	return CheckoutPersonalization{
		BidBoxTopLabel:     checkoutValue(c, "top_label"),
		BidBoxCoupleName:   checkoutValue(c, "couple_name"),
		BidBoxEventDate:    checkoutValue(c, "event_date"),
		BidBoxDetails:      checkoutValue(c, "details"),
		Side:               checkoutValue(c, "side"),
		BrideName:          checkoutValue(c, "bride_name"),
		GroomName:          checkoutValue(c, "groom_name"),
		BrideFatherName:    checkoutValue(c, "bride_father_name"),
		GroomFatherName:    checkoutValue(c, "groom_father_name"),
		MehndiDate:         checkoutValue(c, "mehndi_date"),
		MehndiDay:          checkoutValue(c, "mehndi_day"),
		MehndiTimeType:     checkoutValue(c, "mehndi_time_type"),
		MehndiTime:         checkoutValue(c, "mehndi_time"),
		MehndiDinnerTime:   checkoutValue(c, "mehndi_dinner_time"),
		MehndiVenueName:    checkoutValue(c, "mehndi_venue_name"),
		MehndiVenueAddress: checkoutValue(c, "mehndi_venue_address"),
		BaraatDate:         checkoutValue(c, "baraat_date"),
		BaraatDay:          checkoutValue(c, "baraat_day"),
		BaraatTimeType:     checkoutValue(c, "baraat_time_type"),
		BaraatTime:         checkoutValue(c, "baraat_time"),
		BaraatDinnerTime:   checkoutValue(c, "baraat_dinner_time"),
		BaraatArrivalTime:  checkoutValue(c, "baraat_arrival_time"),
		RukhsatiTime:       checkoutValue(c, "rukhsati_time"),
		BaraatVenueName:    checkoutValue(c, "baraat_venue_name"),
		BaraatVenueAddress: checkoutValue(c, "baraat_venue_address"),
		NikkahDate:         checkoutValue(c, "nikkah_date"),
		NikkahDay:          checkoutValue(c, "nikkah_day"),
		NikkahTimeType:     checkoutValue(c, "nikkah_time_type"),
		NikkahTime:         checkoutValue(c, "nikkah_time"),
		NikkahDinnerTime:   checkoutValue(c, "nikkah_dinner_time"),
		NikkahVenueName:    checkoutValue(c, "nikkah_venue_name"),
		NikkahVenueAddress: checkoutValue(c, "nikkah_venue_address"),
		WalimaDate:         checkoutValue(c, "walima_date"),
		WalimaDay:          checkoutValue(c, "walima_day"),
		WalimaTimeType:     checkoutValue(c, "walima_time_type"),
		WalimaTime:         checkoutValue(c, "walima_time"),
		WalimaDinnerTime:   checkoutValue(c, "walima_dinner_time"),
		WalimaVenueName:    checkoutValue(c, "walima_venue_name"),
		WalimaVenueAddress: checkoutValue(c, "walima_venue_address"),
		ReceptionTime:      checkoutValue(c, "reception_time"),
		RsvpName:           joinedCheckoutValue(c, "rsvp_name"),
		RsvpPhone:          joinedCheckoutValue(c, "rsvp_phone"),
		Notes:              checkoutValue(c, "notes"),
	}
}

func joinedCheckoutValue(c *gin.Context, key string) string {
	var values []string
	if c.Request.Method == http.MethodPost {
		values = c.PostFormArray(key)
	} else {
		values = c.QueryArray(key)
	}
	if len(values) == 0 {
		return checkoutValue(c, key)
	}

	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}
	return strings.Join(cleaned, "\n")
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
