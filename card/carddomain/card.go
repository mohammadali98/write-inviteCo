package carddomain

import (
	"context"
	"time"
)

type Card struct {
	ID          int64
	Name        string
	Description *string
	Price       int64
	Image       string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

type CardRepo interface {
	CardWriter
	CardReader
}

type CardWriter interface {
	CreateCard(ctx context.Context, name string, description *string, price int64, image string) (*Card, error)
	UpdateCard(ctx context.Context, id int64, name string, description *string, price int64, image string) error
	DeleteCard(ctx context.Context, id int64) error
}

type CardReader interface {
	GetAllCards(ctx context.Context) ([]*Card, error)
	GetCardByID(ctx context.Context, id int64) (*Card, error)
}
