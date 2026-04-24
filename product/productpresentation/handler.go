package productpresentation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"writeandinviteco/inviteandco/card/carddomain"
	"writeandinviteco/inviteandco/product/productapplication"
	"writeandinviteco/inviteandco/product/productdomain"
	"writeandinviteco/inviteandco/webui"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	service   *productapplication.Service
	cardRepo  carddomain.CardReader
	uploadDir string
}

func NewHandler(service *productapplication.Service, cardRepo carddomain.CardReader, uploadDir string) *Handler {
	return &Handler{
		service:   service,
		cardRepo:  cardRepo,
		uploadDir: uploadDir,
	}
}

func (h *Handler) List(c *gin.Context) {
	products, err := h.service.ListProducts(c.Request.Context())
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load products right now.")
		return
	}

	c.HTML(http.StatusOK, "admin_products.html", gin.H{
		"products":  products,
		"csrfToken": webui.EnsureCSRFToken(c),
	})
}

func (h *Handler) Detail(c *gin.Context) {
	id, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Product", "Please choose a valid product.")
		return
	}

	product, err := h.service.GetProduct(c.Request.Context(), id)
	if err != nil {
		h.handleReadError(c, err)
		return
	}
	if product.CardID > 0 {
		c.Redirect(http.StatusSeeOther, "/card/"+strconv.FormatInt(product.CardID, 10))
		return
	}

	images, err := h.service.GetProductImages(c.Request.Context(), id)
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the product gallery right now.")
		return
	}

	mainImage, gallery := galleryView(product.ImageURL, images)

	c.HTML(http.StatusOK, "product.html", gin.H{
		"Product":        product,
		"product":        product,
		"CategoryName":   publicCategoryName(product.Category),
		"categoryName":   publicCategoryName(product.Category),
		"CollectionPath": publicCollectionPath(product.Category),
		"collectionPath": publicCollectionPath(product.Category),
		"MainImage":      mainImage,
		"mainImage":      mainImage,
		"Images":         gallery,
		"images":         gallery,
	})
}

func (h *Handler) NewForm(c *gin.Context) {
	cards, err := h.loadCards(c)
	if err != nil {
		return
	}

	h.renderForm(c, http.StatusOK, productFormView{
		Title:       "Add Product",
		Subtitle:    "Create a product record for admin management.",
		FormAction:  "/admin/products",
		SubmitLabel: "Create Product",
		Product:     &productdomain.Product{IsActive: true},
		Cards:       cards,
		CSRFToken:   webui.EnsureCSRFToken(c),
	})
}

func (h *Handler) Create(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please refresh the page and try again.")
		return
	}

	input := readProductInput(c)
	cards, err := h.loadCards(c)
	if err != nil {
		return
	}
	if err := h.validateLinkedCard(c.Request.Context(), input.CardID); err != nil {
		h.handleFormError(c, err, productFormView{
			Title:       "Add Product",
			Subtitle:    "Create a product record for admin management.",
			FormAction:  "/admin/products",
			SubmitLabel: "Create Product",
			Product:     inputToProduct(input),
			Cards:       cards,
			CSRFToken:   webui.EnsureCSRFToken(c),
		})
		return
	}

	imageURLs, savedPaths, err := h.uploadProductImages(c)
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Upload Error", "We could not upload the selected images.")
		return
	}
	if input.ImageURL == "" && len(imageURLs) > 0 {
		input.ImageURL = imageURLs[0]
	}

	product, err := h.service.CreateProduct(c.Request.Context(), input)
	if err != nil {
		cleanupFiles(savedPaths)
		h.handleFormError(c, err, productFormView{
			Title:       "Add Product",
			Subtitle:    "Create a product record for admin management.",
			FormAction:  "/admin/products",
			SubmitLabel: "Create Product",
			Product:     inputToProduct(input),
			Cards:       cards,
			CSRFToken:   webui.EnsureCSRFToken(c),
		})
		return
	}

	if err := h.persistProductImages(c, product.ID, imageURLs, 0); err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "The product was created, but its gallery could not be saved.")
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/products/"+strconv.FormatInt(product.ID, 10)+"/edit")
}

func (h *Handler) EditForm(c *gin.Context) {
	id, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Product", "Please choose a valid product.")
		return
	}

	product, err := h.service.GetProduct(c.Request.Context(), id)
	if err != nil {
		h.handleReadError(c, err)
		return
	}

	cards, err := h.loadCards(c)
	if err != nil {
		return
	}

	images, err := h.service.GetProductImages(c.Request.Context(), id)
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the product gallery right now.")
		return
	}

	h.renderForm(c, http.StatusOK, productFormView{
		Title:       "Edit Product",
		Subtitle:    "Update the product details below.",
		FormAction:  "/admin/products/" + strconv.FormatInt(product.ID, 10) + "/edit",
		SubmitLabel: "Save Changes",
		Product:     product,
		Cards:       cards,
		Images:      images,
		CSRFToken:   webui.EnsureCSRFToken(c),
	})
}

func (h *Handler) Update(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please refresh the page and try again.")
		return
	}

	id, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Product", "Please choose a valid product.")
		return
	}

	input := readProductInput(c)
	cards, err := h.loadCards(c)
	if err != nil {
		return
	}
	if err := h.validateLinkedCard(c.Request.Context(), input.CardID); err != nil {
		view := productFormView{
			Title:       "Edit Product",
			Subtitle:    "Update the product details below.",
			FormAction:  "/admin/products/" + strconv.FormatInt(id, 10) + "/edit",
			SubmitLabel: "Save Changes",
			Product:     inputToProduct(input),
			Cards:       cards,
			CSRFToken:   webui.EnsureCSRFToken(c),
		}
		view.Product.ID = id
		h.handleFormError(c, err, view)
		return
	}

	existingImages, err := h.service.GetProductImages(c.Request.Context(), id)
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the product gallery right now.")
		return
	}
	imageURLs, savedPaths, err := h.uploadProductImages(c)
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Upload Error", "We could not upload the selected images.")
		return
	}
	if input.ImageURL == "" && len(imageURLs) > 0 {
		input.ImageURL = imageURLs[0]
	}
	err = h.service.UpdateProduct(c.Request.Context(), id, input)
	if err != nil {
		cleanupFiles(savedPaths)
		view := productFormView{
			Title:       "Edit Product",
			Subtitle:    "Update the product details below.",
			FormAction:  "/admin/products/" + strconv.FormatInt(id, 10) + "/edit",
			SubmitLabel: "Save Changes",
			Product:     inputToProduct(input),
			Cards:       cards,
			Images:      existingImages,
			CSRFToken:   webui.EnsureCSRFToken(c),
		}
		view.Product.ID = id
		h.handleFormError(c, err, view)
		return
	}

	if err := h.persistProductImages(c, id, imageURLs, int32(len(existingImages))); err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "The product was updated, but its gallery could not be saved.")
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/products")
}

func (h *Handler) Delete(c *gin.Context) {
	if !webui.ValidateCSRF(c) {
		webui.RenderError(c, http.StatusBadRequest, "Request Expired", "Please refresh the page and try deleting the product again.")
		return
	}

	id, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		webui.RenderError(c, http.StatusBadRequest, "Invalid Product", "Please choose a valid product.")
		return
	}

	product, err := h.service.GetProduct(c.Request.Context(), id)
	if err != nil {
		h.handleReadError(c, err)
		return
	}

	images, err := h.service.GetProductImages(c.Request.Context(), id)
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the product gallery right now.")
		return
	}

	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		h.handleReadError(c, err)
		return
	}

	removeUploadedProductFiles(h.uploadDir, append(images, product.ImageURL))
	c.Redirect(http.StatusSeeOther, "/admin/products")
}

type productFormView struct {
	Title       string
	Subtitle    string
	FormAction  string
	SubmitLabel string
	Error       string
	Product     *productdomain.Product
	Cards       []*carddomain.Card
	Images      []string
	CSRFToken   string
}

func (h *Handler) renderForm(c *gin.Context, status int, view productFormView) {
	c.HTML(status, "admin_product_form.html", gin.H{
		"title":       view.Title,
		"subtitle":    view.Subtitle,
		"formAction":  view.FormAction,
		"submitLabel": view.SubmitLabel,
		"error":       view.Error,
		"product":     view.Product,
		"cards":       view.Cards,
		"images":      view.Images,
		"csrfToken":   view.CSRFToken,
	})
}

func (h *Handler) handleReadError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, productapplication.ErrInvalidInput):
		webui.RenderError(c, http.StatusBadRequest, "Invalid Product", "Please choose a valid product.")
	case errors.Is(err, pgx.ErrNoRows):
		webui.RenderError(c, http.StatusNotFound, "Product Not Found", "The requested product was not found.")
	default:
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load the product right now.")
	}
}

func (h *Handler) handleFormError(c *gin.Context, err error, view productFormView) {
	switch {
	case errors.Is(err, productapplication.ErrInvalidInput):
		view.Error = "Please provide a valid name, price, and linked card."
		h.renderForm(c, http.StatusBadRequest, view)
	case errors.Is(err, pgx.ErrNoRows):
		view.Error = "Please choose a valid linked card."
		h.renderForm(c, http.StatusBadRequest, view)
	default:
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not save the product right now.")
	}
}

func (h *Handler) loadCards(c *gin.Context) ([]*carddomain.Card, error) {
	cards, err := h.cardRepo.GetAllCards(c.Request.Context())
	if err != nil {
		webui.RenderError(c, http.StatusInternalServerError, "Server Error", "We could not load cards right now.")
		return nil, err
	}
	return cards, nil
}

func (h *Handler) validateLinkedCard(ctx context.Context, cardID int64) error {
	if cardID <= 0 {
		return productapplication.ErrInvalidInput
	}
	_, err := h.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return err
	}
	return nil
}

func readProductInput(c *gin.Context) productapplication.ProductInput {
	return productapplication.ProductInput{
		CardID:      parseInt64Default(c.PostForm("card_id"), 0),
		Name:        c.PostForm("name"),
		Category:    c.PostForm("category"),
		Price:       parseInt64Default(c.PostForm("price"), -1),
		ImageURL:    c.PostForm("image_url"),
		Description: c.PostForm("description"),
		Dimensions:  c.PostForm("dimensions"),
		IsActive:    checkboxChecked(c.PostForm("is_active")),
	}
}

func inputToProduct(input productapplication.ProductInput) *productdomain.Product {
	return &productdomain.Product{
		CardID:      input.CardID,
		Name:        strings.TrimSpace(input.Name),
		Category:    strings.TrimSpace(input.Category),
		Price:       input.Price,
		ImageURL:    strings.TrimSpace(input.ImageURL),
		Description: strings.TrimSpace(input.Description),
		Dimensions:  strings.TrimSpace(input.Dimensions),
		IsActive:    input.IsActive,
	}
}

func parsePositiveInt64(raw string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value < 1 {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}

func parseInt64Default(raw string, fallback int64) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return fallback
	}
	return value
}

func checkboxChecked(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "on", "true", "1", "yes":
		return true
	default:
		return false
	}
}

func publicCategoryName(category string) string {
	switch category {
	case "card":
		return "Wedding Cards"
	case "bid_box":
		return "Bid Boxes"
	default:
		return "Products"
	}
}

func publicCollectionPath(category string) string {
	switch category {
	case "card":
		return "/collections/wedding-cards"
	case "bid_box":
		return "/collections/bid-boxes"
	default:
		return "/"
	}
}

func (h *Handler) uploadProductImages(c *gin.Context) ([]string, []string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, nil, err
	}
	log.Println("Multipart form received")
	if form == nil || len(form.File["images"]) == 0 {
		return nil, nil, nil
	}

	if err := os.MkdirAll(h.uploadDir, os.ModePerm); err != nil {
		return nil, nil, err
	}

	imageURLs := make([]string, 0, len(form.File["images"]))
	savedPaths := make([]string, 0, len(form.File["images"]))

	for _, file := range form.File["images"] {
		if file == nil || strings.TrimSpace(file.Filename) == "" {
			continue
		}

		filename, err := newUploadFilename(filepath.Ext(file.Filename))
		if err != nil {
			cleanupFiles(savedPaths)
			return nil, nil, err
		}

		savePath := filepath.Join(h.uploadDir, filename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			cleanupFiles(savedPaths)
			return nil, nil, err
		}

		savedPaths = append(savedPaths, savePath)
		imageURLs = append(imageURLs, "/static/uploads/"+filename)
	}

	return imageURLs, savedPaths, nil
}

func (h *Handler) persistProductImages(c *gin.Context, productID int64, imageURLs []string, sortOffset int32) error {
	for i, imageURL := range imageURLs {
		if err := h.service.AddProductImage(c.Request.Context(), productID, imageURL, sortOffset+int32(i)); err != nil {
			return err
		}
	}
	return nil
}

func galleryView(primary string, images []string) (string, []string) {
	gallery := make([]string, 0, len(images)+1)
	seen := make(map[string]struct{})

	add := func(url string) {
		url = strings.TrimSpace(url)
		if url == "" {
			return
		}
		if _, ok := seen[url]; ok {
			return
		}
		seen[url] = struct{}{}
		gallery = append(gallery, url)
	}

	add(primary)
	for _, image := range images {
		add(image)
	}
	if len(gallery) == 0 {
		gallery = append(gallery, "/static/sample.jpg")
	}

	return gallery[0], gallery
}

func newUploadFilename(ext string) (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	buffer[6] = (buffer[6] & 0x0f) | 0x40
	buffer[8] = (buffer[8] & 0x3f) | 0x80

	encoded := hex.EncodeToString(buffer)
	uuid := encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
	return uuid + strings.ToLower(ext), nil
}

func cleanupFiles(paths []string) {
	for _, path := range paths {
		_ = os.Remove(path)
	}
}

func removeUploadedProductFiles(uploadDir string, urls []string) {
	const uploadPrefix = "/static/uploads/"

	seen := make(map[string]struct{})
	for _, rawURL := range urls {
		url := strings.TrimSpace(rawURL)
		if url == "" || !strings.HasPrefix(url, uploadPrefix) {
			continue
		}

		relPath := strings.TrimPrefix(url, uploadPrefix)
		if relPath == "" || strings.Contains(relPath, "..") {
			continue
		}

		filePath := filepath.Join(uploadDir, filepath.FromSlash(relPath))
		if _, ok := seen[filePath]; ok {
			continue
		}
		seen[filePath] = struct{}{}
		_ = os.Remove(filePath)
	}
}
