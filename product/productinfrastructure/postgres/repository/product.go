package productrepository

import (
	"context"

	"writeandinviteco/inviteandco/product/productdomain"
	productreader "writeandinviteco/inviteandco/product/productinfrastructure/postgres/reader"
	productwriter "writeandinviteco/inviteandco/product/productinfrastructure/postgres/writer"
)

type ProductRepository struct {
	reader *productreader.Queries
	writer *productwriter.Queries
}

func NewProductRepository(reader *productreader.Queries, writer *productwriter.Queries) *ProductRepository {
	return &ProductRepository{reader: reader, writer: writer}
}

func (r *ProductRepository) ListProducts(ctx context.Context) ([]*productdomain.Product, error) {
	rows, err := r.reader.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	products := make([]*productdomain.Product, len(rows))
	for i, row := range rows {
		products[i] = mapProductRow(
			row.ID,
			row.CardID,
			row.Name,
			row.Category,
			row.Price,
			row.ImageUrl,
			row.Description,
			row.Dimensions,
			row.IsActive,
		)
	}

	return products, nil
}

func (r *ProductRepository) GetProduct(ctx context.Context, id int64) (*productdomain.Product, error) {
	row, err := r.reader.GetProduct(ctx, int32(id))
	if err != nil {
		return nil, err
	}

	return mapProductRow(
		row.ID,
		row.CardID,
		row.Name,
		row.Category,
		row.Price,
		row.ImageUrl,
		row.Description,
		row.Dimensions,
		row.IsActive,
	), nil
}

func (r *ProductRepository) GetProductImages(ctx context.Context, productID int64) ([]string, error) {
	images, err := r.reader.GetProductImages(ctx, int32(productID))
	if err != nil {
		return nil, err
	}

	return images, nil
}

func (r *ProductRepository) CreateProduct(ctx context.Context, cardID int64, name string, category string, price int64, imageURL string, description string, dimensions string, isActive bool) (*productdomain.Product, error) {
	row, err := r.writer.CreateProduct(ctx, productwriter.CreateProductParams{
		CardID:      nullableInt64(cardID),
		Name:        name,
		Category:    category,
		Price:       int32(price),
		ImageUrl:    nullableString(imageURL),
		Description: nullableString(description),
		Dimensions:  nullableString(dimensions),
		IsActive:    nullableBool(isActive),
	})
	if err != nil {
		return nil, err
	}

	return mapProductRow(
		row.ID,
		row.CardID,
		row.Name,
		row.Category,
		row.Price,
		row.ImageUrl,
		row.Description,
		row.Dimensions,
		row.IsActive,
	), nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, id int64, cardID int64, name string, category string, price int64, imageURL string, description string, dimensions string, isActive bool) error {
	return r.writer.UpdateProduct(ctx, productwriter.UpdateProductParams{
		ID:          int32(id),
		CardID:      nullableInt64(cardID),
		Name:        name,
		Category:    category,
		Price:       int32(price),
		ImageUrl:    nullableString(imageURL),
		Description: nullableString(description),
		Dimensions:  nullableString(dimensions),
		IsActive:    nullableBool(isActive),
	})
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id int64) error {
	return r.writer.DeleteProduct(ctx, int32(id))
}

func (r *ProductRepository) AddProductImage(ctx context.Context, productID int64, imageURL string, sortOrder int32) error {
	return r.writer.AddProductImage(ctx, productwriter.AddProductImageParams{
		ProductID: int32(productID),
		ImageUrl:  imageURL,
		SortOrder: nullableInt32(sortOrder),
	})
}

func mapProductRow(id int32, cardID *int64, name string, category string, price int32, imageURL *string, description *string, dimensions *string, isActive *bool) *productdomain.Product {
	return &productdomain.Product{
		ID:          int64(id),
		CardID:      int64Value(cardID),
		Name:        name,
		Category:    category,
		Price:       int64(price),
		ImageURL:    stringValue(imageURL),
		Description: stringValue(description),
		Dimensions:  stringValue(dimensions),
		IsActive:    boolValue(isActive, true),
	}
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func nullableBool(value bool) *bool {
	return &value
}

func nullableInt32(value int32) *int32 {
	return &value
}

func nullableInt64(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
