package productapplication

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"writeandinviteco/inviteandco/product/productdomain"
)

var ErrInvalidInput = errors.New("invalid input")

const (
	maxProductNameLength        = 150
	maxProductImageURLLength    = 500
	maxProductDescriptionLength = 2000
	maxProductDimensionsLength  = 200
	maxProductPrice             = 1_000_000_000
)

type ProductInput struct {
	CardID      int64
	Name        string
	Category    string
	Price       int64
	ImageURL    string
	Description string
	Dimensions  string
	IsActive    bool
}

type Service struct {
	repo productdomain.ProductRepo
}

func NewService(repo productdomain.ProductRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListProducts(ctx context.Context) ([]*productdomain.Product, error) {
	return s.repo.ListProducts(ctx)
}

func (s *Service) GetProduct(ctx context.Context, id int64) (*productdomain.Product, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.GetProduct(ctx, id)
}

func (s *Service) GetProductImages(ctx context.Context, productID int64) ([]string, error) {
	if productID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.GetProductImages(ctx, productID)
}

func (s *Service) CreateProduct(ctx context.Context, input ProductInput) (*productdomain.Product, error) {
	normalized, err := normalizeInput(input)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateProduct(ctx, normalized.CardID, normalized.Name, normalized.Category, normalized.Price, normalized.ImageURL, normalized.Description, normalized.Dimensions, normalized.IsActive)
}

func (s *Service) UpdateProduct(ctx context.Context, id int64, input ProductInput) error {
	if id <= 0 {
		return ErrInvalidInput
	}
	normalized, err := normalizeInput(input)
	if err != nil {
		return err
	}
	return s.repo.UpdateProduct(ctx, id, normalized.CardID, normalized.Name, normalized.Category, normalized.Price, normalized.ImageURL, normalized.Description, normalized.Dimensions, normalized.IsActive)
}

func (s *Service) DeleteProduct(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidInput
	}
	return s.repo.DeleteProduct(ctx, id)
}

func (s *Service) AddProductImage(ctx context.Context, productID int64, imageURL string, sortOrder int32) error {
	if productID <= 0 || strings.TrimSpace(imageURL) == "" || sortOrder < 0 {
		return ErrInvalidInput
	}
	return s.repo.AddProductImage(ctx, productID, strings.TrimSpace(imageURL), sortOrder)
}

func normalizeInput(input ProductInput) (ProductInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Category = normalizeCategory(input.Category)
	input.ImageURL = strings.TrimSpace(input.ImageURL)
	input.Description = strings.TrimSpace(input.Description)
	input.Dimensions = strings.TrimSpace(input.Dimensions)

	if input.CardID < 0 ||
		input.Name == "" ||
		input.Category == "" ||
		input.Price < 1 ||
		input.Price > maxProductPrice ||
		utf8.RuneCountInString(input.Name) > maxProductNameLength ||
		utf8.RuneCountInString(input.ImageURL) > maxProductImageURLLength ||
		utf8.RuneCountInString(input.Description) > maxProductDescriptionLength ||
		utf8.RuneCountInString(input.Dimensions) > maxProductDimensionsLength {
		return ProductInput{}, ErrInvalidInput
	}

	return input, nil
}

func normalizeCategory(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "card":
		return "card"
	case "bid_box":
		return "bid_box"
	default:
		return ""
	}
}
