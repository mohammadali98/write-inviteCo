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

func (r *OrderRepository) CreateOrderDetail(ctx context.Context, orderID int64, side string, brideName *string, groomName *string, brideFatherName *string, groomFatherName *string, mehndiDate *string, mehndiTimeType *string, mehndiTime *string, mehndiDinnerTime *string, baraatDate *string, baraatTimeType *string, baraatTime *string, baraatDinnerTime *string, baraatArrivalTime *string, rukhsatiTime *string, nikkahDate *string, nikkahTimeType *string, nikkahTime *string, nikkahDinnerTime *string, walimaDate *string, walimaTimeType *string, walimaTime *string, walimaDinnerTime *string, receptionTime *string, dinnerTime *string, venueName string, venueAddress string, rsvpName string, rsvpPhone string, notes *string) (*orderdomain.OrderDetail, error) {
	row, err := r.writer.CreateOrderDetail(ctx, orderwriter.CreateOrderDetailParams{
		OrderID:           orderID,
		Side:              side,
		BrideName:         brideName,
		GroomName:         groomName,
		BrideFatherName:   brideFatherName,
		GroomFatherName:   groomFatherName,
		MehndiDate:        stringOrEmpty(mehndiDate),
		MehndiTimeType:    mehndiTimeType,
		MehndiTime:        stringOrEmpty(mehndiTime),
		MehndiDinnerTime:  stringOrEmpty(mehndiDinnerTime),
		BaraatDate:        stringOrEmpty(baraatDate),
		BaraatTimeType:    baraatTimeType,
		BaraatTime:        stringOrEmpty(baraatTime),
		BaraatDinnerTime:  stringOrEmpty(baraatDinnerTime),
		BaraatArrivalTime: stringOrEmpty(baraatArrivalTime),
		RukhsatiTime:      stringOrEmpty(rukhsatiTime),
		NikkahDate:        stringOrEmpty(nikkahDate),
		NikkahTimeType:    nikkahTimeType,
		NikkahTime:        stringOrEmpty(nikkahTime),
		NikkahDinnerTime:  stringOrEmpty(nikkahDinnerTime),
		WalimaDate:        stringOrEmpty(walimaDate),
		WalimaTimeType:    walimaTimeType,
		WalimaTime:        stringOrEmpty(walimaTime),
		WalimaDinnerTime:  stringOrEmpty(walimaDinnerTime),
		ReceptionTime:     stringOrEmpty(receptionTime),
		DinnerTime:        stringOrEmpty(dinnerTime),
		VenueName:         venueName,
		VenueAddress:      venueAddress,
		RsvpName:          rsvpName,
		RsvpPhone:         rsvpPhone,
		Notes:             notes,
	})
	if err != nil {
		return nil, err
	}
	return &orderdomain.OrderDetail{
		ID:                row.ID,
		OrderID:           row.OrderID,
		Side:              row.Side,
		BrideName:         row.BrideName,
		GroomName:         row.GroomName,
		BrideFatherName:   row.BrideFatherName,
		GroomFatherName:   row.GroomFatherName,
		MehndiDate:        emptyToNil(row.MehndiDate),
		MehndiTimeType:    row.MehndiTimeType,
		MehndiTime:        emptyToNil(row.MehndiTime),
		MehndiDinnerTime:  emptyToNil(row.MehndiDinnerTime),
		BaraatDate:        emptyToNil(row.BaraatDate),
		BaraatTimeType:    row.BaraatTimeType,
		BaraatTime:        emptyToNil(row.BaraatTime),
		BaraatDinnerTime:  emptyToNil(row.BaraatDinnerTime),
		BaraatArrivalTime: emptyToNil(row.BaraatArrivalTime),
		RukhsatiTime:      emptyToNil(row.RukhsatiTime),
		NikkahDate:        emptyToNil(row.NikkahDate),
		NikkahTimeType:    row.NikkahTimeType,
		NikkahTime:        emptyToNil(row.NikkahTime),
		NikkahDinnerTime:  emptyToNil(row.NikkahDinnerTime),
		WalimaDate:        emptyToNil(row.WalimaDate),
		WalimaTimeType:    row.WalimaTimeType,
		WalimaTime:        emptyToNil(row.WalimaTime),
		WalimaDinnerTime:  emptyToNil(row.WalimaDinnerTime),
		ReceptionTime:     emptyToNil(row.ReceptionTime),
		DinnerTime:        emptyToNil(row.DinnerTime),
		VenueName:         row.VenueName,
		VenueAddress:      row.VenueAddress,
		RsvpName:          row.RsvpName,
		RsvpPhone:         row.RsvpPhone,
		Notes:             row.Notes,
		CreatedAt:         toTimePtr(row.CreatedAt),
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

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func emptyToNil(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}
