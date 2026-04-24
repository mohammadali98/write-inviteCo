package productdomain

import "context"

type Product struct {
	ID          int64
	CardID      int64
	Name        string
	Category    string
	Price       int64
	ImageURL    string
	Description string
	Dimensions  string
	IsActive    bool
}

type ProductRepo interface {
	ProductWriter
	ProductReader
}

type ProductWriter interface {
	CreateProduct(ctx context.Context, cardID int64, name string, category string, price int64, imageURL string, description string, dimensions string, isActive bool) (*Product, error)
	UpdateProduct(ctx context.Context, id int64, cardID int64, name string, category string, price int64, imageURL string, description string, dimensions string, isActive bool) error
	DeleteProduct(ctx context.Context, id int64) error
	AddProductImage(ctx context.Context, productID int64, imageURL string, sortOrder int32) error
}

type ProductReader interface {
	ListProducts(ctx context.Context) ([]*Product, error)
	GetProduct(ctx context.Context, id int64) (*Product, error)
	GetProductImages(ctx context.Context, productID int64) ([]string, error)
}
