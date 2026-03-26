package orderrepository

import (
	"context"
	"time"

	"writeandinviteco/inviteandco/order/orderdomain"
	orderreader "writeandinviteco/inviteandco/order/orderinfrastructure/postgres/reader"
	orderwriter "writeandinviteco/inviteandco/order/orderinfrastructure/postgres/writer"

	"github.com/jackc/pgx/v5/pgtype"
)

type OrderRepository struct {
	reader *orderreader.Queries
	writer *orderwriter.Queries
}

func NewOrderRepository(reader *orderreader.Queries, writer *orderwriter.Queries) *OrderRepository {
	return &OrderRepository{reader: reader, writer: writer}
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id int64) (*orderdomain.Order, error) {
	row, err := r.reader.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &orderdomain.Order{
		ID:         row.ID,
		CustomerID: derefInt64(row.CustomerID),
		CardID:     derefInt64(row.CardID),
		Quantity:   row.Quantity,
		TotalPrice: row.TotalPrice,
		Status:     toOrderStatus(row.Status),
		Currency:   row.Currency,
		CreatedAt:  toTimePtr(row.CreatedAt),
		UpdatedAt:  toTimePtr(row.UpdatedAt),
	}, nil
}

func (r *OrderRepository) GetOrdersByCustomerID(ctx context.Context, customerID int64) ([]*orderdomain.Order, error) {
	rows, err := r.reader.GetOrdersByCustomerID(ctx, &customerID)
	if err != nil {
		return nil, err
	}
	orders := make([]*orderdomain.Order, len(rows))
	for i, row := range rows {
		orders[i] = &orderdomain.Order{
			ID:         row.ID,
			CustomerID: derefInt64(row.CustomerID),
			CardID:     derefInt64(row.CardID),
			Quantity:   row.Quantity,
			TotalPrice: row.TotalPrice,
			Status:     toOrderStatus(row.Status),
			Currency:   row.Currency,
			CreatedAt:  toTimePtr(row.CreatedAt),
			UpdatedAt:  toTimePtr(row.UpdatedAt),
		}
	}
	return orders, nil
}

func (r *OrderRepository) CreateOrder(ctx context.Context, customerID int64, cardID int64, quantity int64, totalPrice int64, status orderdomain.OrderStatus, currency string) (*orderdomain.Order, error) {
	s := string(status)
	row, err := r.writer.CreateOrder(ctx, orderwriter.CreateOrderParams{
		CustomerID: &customerID,
		CardID:     &cardID,
		Quantity:   quantity,
		TotalPrice: totalPrice,
		Status:     &s,
		Currency:   currency,
	})
	if err != nil {
		return nil, err
	}
	return &orderdomain.Order{
		ID:         row.ID,
		CustomerID: derefInt64(row.CustomerID),
		CardID:     derefInt64(row.CardID),
		Quantity:   row.Quantity,
		TotalPrice: row.TotalPrice,
		Status:     toOrderStatus(row.Status),
		Currency:   row.Currency,
		CreatedAt:  toTimePtr(row.CreatedAt),
		UpdatedAt:  toTimePtr(row.UpdatedAt),
	}, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id int64, status orderdomain.OrderStatus) error {
	s := string(status)
	return r.writer.UpdateOrderStatus(ctx, orderwriter.UpdateOrderStatusParams{
		ID:     id,
		Status: &s,
	})
}

func derefInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func toTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func toOrderStatus(s *string) orderdomain.OrderStatus {
	if s == nil {
		return orderdomain.PendingOrderStatus
	}
	return orderdomain.OrderStatus(*s)
}
