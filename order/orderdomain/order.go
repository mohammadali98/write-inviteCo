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
	Currency   string
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

type OrderDetail struct {
	ID                int64
	OrderID           int64
	Side              string
	BrideName         *string
	GroomName         *string
	BrideFatherName   *string
	GroomFatherName   *string
	MehndiDate        *string
	MehndiTimeType    *string
	MehndiTime        *string
	MehndiDinnerTime  *string
	BaraatDate        *string
	BaraatTimeType    *string
	BaraatTime        *string
	BaraatDinnerTime  *string
	BaraatArrivalTime *string
	RukhsatiTime      *string
	NikkahDate        *string
	NikkahTimeType    *string
	NikkahTime        *string
	NikkahDinnerTime  *string
	WalimaDate        *string
	WalimaTimeType    *string
	WalimaTime        *string
	WalimaDinnerTime  *string
	ReceptionTime     *string
	DinnerTime        *string
	VenueName         string
	VenueAddress      string
	RsvpName          string
	RsvpPhone         string
	Notes             *string
	CreatedAt         *time.Time
}

type OrderRepo interface {
	OrderWriter
	OrderReader
}

type OrderWriter interface {
	CreateOrder(ctx context.Context, customerID int64, cardID int64, quantity int64, totalPrice int64, status OrderStatus, currency string) (*Order, error)
	CreateOrderDetail(ctx context.Context, orderID int64, side string, brideName *string, groomName *string, brideFatherName *string, groomFatherName *string, mehndiDate *string, mehndiTimeType *string, mehndiTime *string, mehndiDinnerTime *string, baraatDate *string, baraatTimeType *string, baraatTime *string, baraatDinnerTime *string, baraatArrivalTime *string, rukhsatiTime *string, nikkahDate *string, nikkahTimeType *string, nikkahTime *string, nikkahDinnerTime *string, walimaDate *string, walimaTimeType *string, walimaTime *string, walimaDinnerTime *string, receptionTime *string, dinnerTime *string, venueName string, venueAddress string, rsvpName string, rsvpPhone string, notes *string) (*OrderDetail, error)
	UpdateOrderStatus(ctx context.Context, id int64, status OrderStatus) error
}

type OrderReader interface {
	GetOrderByID(ctx context.Context, id int64) (*Order, error)
	GetOrdersByCustomerID(ctx context.Context, customerID int64) ([]*Order, error)
}
