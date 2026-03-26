package carddomain

import (
	"context"
	"time"
)

type Card struct {
	ID          int64
	Name        string
	Description *string
	PricePKR    int64
	PriceNOK    int64
	Image       string
	Category    string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

type CardImage struct {
	ID        int64
	CardID    int64
	Image     string
	SortOrder int32
	CreatedAt *time.Time
}

type CardRepo interface {
	CardWriter
	CardReader
}

type CardWriter interface {
	CreateCard(ctx context.Context, name string, description *string, pricePKR int64, priceNOK int64, image string, category string) (*Card, error)
	UpdateCard(ctx context.Context, id int64, name string, description *string, pricePKR int64, priceNOK int64, image string, category string) error
	DeleteCard(ctx context.Context, id int64) error
	CreateCardImage(ctx context.Context, cardID int64, image string, sortOrder int32) (*CardImage, error)
	DeleteCardImagesByCardID(ctx context.Context, cardID int64) error
}

type CardReader interface {
	GetAllCards(ctx context.Context) ([]*Card, error)
	GetCardByID(ctx context.Context, id int64) (*Card, error)
	GetCardsByCategory(ctx context.Context, category string) ([]*Card, error)
	SearchCards(ctx context.Context, query string) ([]*Card, error)
	GetCardImagesByCardID(ctx context.Context, cardID int64) ([]*CardImage, error)
}
