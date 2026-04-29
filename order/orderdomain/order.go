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
	ID           int64
	CustomerID   int64
	CardID       int64
	CardName     string
	CardImage    string
	CardCategory string
	Quantity     int64
	TotalPrice   int64
	Status       OrderStatus
	Currency     string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

type AdminOrder struct {
	ID              int64
	CustomerName    string
	TotalPrice      int64
	Status          OrderStatus
	PaymentStatus   PaymentStatus
	SubmittedAmount *int64
	SubmittedAt     *time.Time
	Currency        string
	CreatedAt       *time.Time
}

type OrderDetail struct {
	ID                 int64
	OrderID            int64
	Side               string
	BidBoxTopLabel     *string
	BidBoxCoupleName   *string
	BidBoxEventDate    *string
	BidBoxDetails      *string
	BrideName          *string
	GroomName          *string
	BrideFatherName    *string
	GroomFatherName    *string
	MehndiDate         *string
	MehndiDay          *string
	MehndiTimeType     *string
	MehndiTime         *string
	MehndiDinnerTime   *string
	BaraatDate         *string
	BaraatDay          *string
	BaraatTimeType     *string
	BaraatTime         *string
	BaraatDinnerTime   *string
	BaraatArrivalTime  *string
	RukhsatiTime       *string
	NikkahDate         *string
	NikkahDay          *string
	NikkahTimeType     *string
	NikkahTime         *string
	NikkahDinnerTime   *string
	NikkahVenueName    *string
	NikkahVenueAddress *string
	WalimaDate         *string
	WalimaDay          *string
	WalimaTimeType     *string
	WalimaTime         *string
	WalimaDinnerTime   *string
	WalimaVenueName    *string
	WalimaVenueAddress *string
	ReceptionTime      *string
	MehndiVenueName    *string
	MehndiVenueAddress *string
	BaraatVenueName    *string
	BaraatVenueAddress *string
	RsvpName           string
	RsvpPhone          string
	Notes              *string
	CreatedAt          *time.Time
}

type OrderRepo interface {
	OrderWriter
	OrderReader
}

type OrderWriter interface {
	CreateOrder(ctx context.Context, customerID int64, cardID int64, quantity int64, totalPrice int64, status OrderStatus, currency string) (*Order, error)
	CreateOrderDetail(ctx context.Context, orderID int64, side string, bidBoxTopLabel *string, bidBoxCoupleName *string, bidBoxEventDate *string, bidBoxDetails *string, brideName *string, groomName *string, brideFatherName *string, groomFatherName *string, mehndiDate *string, mehndiDay *string, mehndiTimeType *string, mehndiTime *string, mehndiDinnerTime *string, mehndiVenueName *string, mehndiVenueAddress *string, baraatDate *string, baraatDay *string, baraatTimeType *string, baraatTime *string, baraatDinnerTime *string, baraatArrivalTime *string, rukhsatiTime *string, baraatVenueName *string, baraatVenueAddress *string, nikkahDate *string, nikkahDay *string, nikkahTimeType *string, nikkahTime *string, nikkahDinnerTime *string, nikkahVenueName *string, nikkahVenueAddress *string, walimaDate *string, walimaDay *string, walimaTimeType *string, walimaTime *string, walimaDinnerTime *string, receptionTime *string, walimaVenueName *string, walimaVenueAddress *string, rsvpName string, rsvpPhone string, notes *string) (*OrderDetail, error)
	UpdateOrderStatus(ctx context.Context, id int64, status OrderStatus) error
}

type OrderReader interface {
	GetOrderByID(ctx context.Context, id int64) (*Order, error)
	GetOrdersByCustomerID(ctx context.Context, customerID int64) ([]*Order, error)
	GetAdminOrders(ctx context.Context) ([]*AdminOrder, error)
	GetOrderDetailByOrderID(ctx context.Context, orderID int64) (*OrderDetail, error)
	GetOrderPaymentByOrderID(ctx context.Context, orderID int64) (*OrderPayment, error)
}
