package orderdomain

import (
	"context"
	"time"
)

type OrderStatus string

const (
	PendingOrderStatus   OrderStatus = "pending"
	ConfirmedOrderStatus OrderStatus = "confirmed"
	CancelledOrderStatus OrderStatus = "cancelled"
	CompletedOrderStatus OrderStatus = "completed"
)

type Order struct {
	ID         int64
	CustomerID int64
	CardID     int64
	Quantity   int64
	TotalPrice int64
	Status     OrderStatus
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

type OrderRepo interface {
	OrderWriter
	OrderReader
}

type OrderWriter interface {
	CreateOrder(ctx context.Context, customerID int64, cardID int64, quantity int64, totalPrice int64, status OrderStatus) (*Order, error)
	UpdateOrderStatus(ctx context.Context, id int64, status OrderStatus) error
}

type OrderReader interface {
	GetOrderByID(ctx context.Context, id int64) (*Order, error)
	GetOrdersByCustomerID(ctx context.Context, customerID int64) ([]*Order, error)
}
