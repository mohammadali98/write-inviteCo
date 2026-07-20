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
		ID:           row.ID,
		CustomerID:   derefInt64(row.CustomerID),
		CardID:       derefInt64(row.CardID),
		CardName:     row.CardName,
		CardImage:    row.CardImage,
		CardCategory: row.CardCategory,
		Quantity:     row.Quantity,
		TotalPrice:   row.TotalPrice,
		Status:       toOrderStatus(row.Status),
		Currency:     row.Currency,
		PublicToken:  row.PublicToken,
		CreatedAt:    toTimePtr(row.CreatedAt),
		UpdatedAt:    toTimePtr(row.UpdatedAt),
	}, nil
}

func (r *OrderRepository) GetOrderByPublicToken(ctx context.Context, token string) (*orderdomain.Order, error) {
	row, err := r.reader.GetOrderByPublicToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return &orderdomain.Order{
		ID:           row.ID,
		CustomerID:   derefInt64(row.CustomerID),
		CardID:       derefInt64(row.CardID),
		CardName:     row.CardName,
		CardImage:    row.CardImage,
		CardCategory: row.CardCategory,
		Quantity:     row.Quantity,
		TotalPrice:   row.TotalPrice,
		Status:       toOrderStatus(row.Status),
		Currency:     row.Currency,
		PublicToken:  row.PublicToken,
		CreatedAt:    toTimePtr(row.CreatedAt),
		UpdatedAt:    toTimePtr(row.UpdatedAt),
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

func (r *OrderRepository) GetAdminOrders(ctx context.Context, filter orderdomain.AdminOrderFilter) ([]*orderdomain.AdminOrder, error) {
	rows, err := r.reader.GetAdminOrders(ctx, orderreader.GetAdminOrdersParams{
		OrderStatus:   stringPtrOrNil(filter.OrderStatus),
		PaymentStatus: stringPtrOrNil(filter.PaymentStatus),
		Search:        stringPtrOrNil(filter.Search),
		CreatedFrom:   toTimestamptz(filter.CreatedFrom),
		CreatedTo:     toTimestamptz(filter.CreatedTo),
	})
	if err != nil {
		return nil, err
	}

	orders := make([]*orderdomain.AdminOrder, len(rows))
	for i, row := range rows {
		orders[i] = &orderdomain.AdminOrder{
			ID:              row.ID,
			CustomerName:    row.CustomerName,
			ProductName:     row.ProductName,
			CardCategory:    row.CardCategory,
			Quantity:        row.Quantity,
			TotalPrice:      row.TotalPrice,
			Status:          toOrderStatus(row.Status),
			PaymentStatus:   toPaymentStatus(row.PaymentStatus),
			SubmittedAmount: row.SubmittedAmount,
			SubmittedAt:     toTimePtr(row.SubmittedAt),
			Currency:        row.Currency,
			CreatedAt:       toTimePtr(row.CreatedAt),
		}
	}
	return orders, nil
}

func (r *OrderRepository) GetOrderDetailByOrderID(ctx context.Context, orderID int64) (*orderdomain.OrderDetail, error) {
	rows, err := r.reader.GetLatestOrderDetailByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	row := rows[0]
	return &orderdomain.OrderDetail{
		ID:                    row.ID,
		OrderID:               row.OrderID,
		Side:                  row.Side,
		ExtraInsertsPerCard:   row.ExtraInsertsPerCard,
		BidBoxTopLabel:        emptyToNil(row.TopLabel),
		BidBoxCoupleName:      emptyToNil(row.CoupleName),
		BidBoxEventDate:       emptyToNil(row.EventDate),
		BidBoxDetails:         emptyToNil(row.BidBoxDetails),
		BrideName:             row.BrideName,
		GroomName:             row.GroomName,
		BrideFatherName:       row.BrideFatherName,
		GroomFatherName:       row.GroomFatherName,
		MehndiDate:            emptyToNil(row.MehndiDate),
		MehndiDay:             emptyToNil(row.MehndiDay),
		MehndiTimeType:        row.MehndiTimeType,
		MehndiTime:            emptyToNil(row.MehndiTime),
		MehndiDinnerTime:      emptyToNil(row.MehndiDinnerTime),
		MehndiVenueName:       emptyToNil(row.MehndiVenueName),
		MehndiVenueAddress:    emptyToNil(row.MehndiVenueAddress),
		BaraatDate:            emptyToNil(row.BaraatDate),
		BaraatDay:             emptyToNil(row.BaraatDay),
		BaraatTimeType:        row.BaraatTimeType,
		BaraatTime:            emptyToNil(row.BaraatTime),
		BaraatDinnerTime:      emptyToNil(row.BaraatDinnerTime),
		BaraatArrivalTime:     emptyToNil(row.BaraatArrivalTime),
		RukhsatiTime:          emptyToNil(row.RukhsatiTime),
		BaraatSehrabandiTime:  emptyToNil(row.BaraatSehrabandiTime),
		BaraatVenueName:       emptyToNil(row.BaraatVenueName),
		BaraatVenueAddress:    emptyToNil(row.BaraatVenueAddress),
		NikkahDate:            emptyToNil(row.NikkahDate),
		NikkahDay:             emptyToNil(row.NikkahDay),
		NikkahTimeType:        row.NikkahTimeType,
		NikkahTime:            emptyToNil(row.NikkahTime),
		NikkahDinnerTime:      emptyToNil(row.NikkahDinnerTime),
		NikkahVenueName:       emptyToNil(row.NikkahVenueName),
		NikkahVenueAddress:    emptyToNil(row.NikkahVenueAddress),
		WalimaDate:            emptyToNil(row.WalimaDate),
		WalimaDay:             emptyToNil(row.WalimaDay),
		WalimaTimeType:        row.WalimaTimeType,
		WalimaTime:            emptyToNil(row.WalimaTime),
		WalimaDinnerTime:      emptyToNil(row.WalimaDinnerTime),
		WalimaVenueName:       emptyToNil(row.WalimaVenueName),
		WalimaVenueAddress:    emptyToNil(row.WalimaVenueAddress),
		ReceptionTime:         emptyToNil(row.ReceptionTime),
		ShendiDate:            emptyToNil(row.ShendiDate),
		ShendiDay:             emptyToNil(row.ShendiDay),
		ShendiTime:            emptyToNil(row.ShendiTime),
		ShendiTimeType:        row.ShendiTimeType,
		ShendiDinnerTime:      emptyToNil(row.ShendiDinnerTime),
		ShendiVenueName:       emptyToNil(row.ShendiVenueName),
		ShendiVenueAddress:    emptyToNil(row.ShendiVenueAddress),
		ShendiArrivalTime:     emptyToNil(row.ShendiArrivalTime),
		ShendiRukhsatiTime:    emptyToNil(row.ShendiRukhsatiTime),
		ShendiSehrabandiTime:  emptyToNil(row.ShendiSehrabandiTime),
		ShalimaDate:           emptyToNil(row.ShalimaDate),
		ShalimaDay:            emptyToNil(row.ShalimaDay),
		ShalimaTime:           emptyToNil(row.ShalimaTime),
		ShalimaTimeType:       row.ShalimaTimeType,
		ShalimaDinnerTime:     emptyToNil(row.ShalimaDinnerTime),
		ShalimaVenueName:      emptyToNil(row.ShalimaVenueName),
		ShalimaVenueAddress:   emptyToNil(row.ShalimaVenueAddress),
		ShalimaArrivalTime:    emptyToNil(row.ShalimaArrivalTime),
		ShalimaRukhsatiTime:   emptyToNil(row.ShalimaRukhsatiTime),
		ShalimaSehrabandiTime: emptyToNil(row.ShalimaSehrabandiTime),
		ShalimaReceptionTime:  emptyToNil(row.ShalimaReceptionTime),
		RsvpName:              row.RsvpName,
		RsvpPhone:             row.RsvpPhone,
		Notes:                 row.Notes,
		CreatedAt:             toTimePtr(row.CreatedAt),
	}, nil
}

func (r *OrderRepository) GetOrderPaymentByOrderID(ctx context.Context, orderID int64) (*orderdomain.OrderPayment, error) {
	row, err := r.reader.GetOrderPaymentByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return &orderdomain.OrderPayment{
		ID:                   row.ID,
		OrderID:              row.OrderID,
		PaymentMethod:        toPaymentMethod(row.PaymentMethod),
		PaymentStatus:        toPaymentStatus(&row.PaymentStatus),
		ExpectedAmount:       row.ExpectedAmount,
		SubmittedAmount:      row.SubmittedAmount,
		SenderName:           row.SenderName,
		TransactionReference: row.TransactionReference,
		ProofFilePath:        row.ProofFilePath,
		CustomerNote:         row.CustomerNote,
		SubmittedAt:          toTimePtr(row.SubmittedAt),
		VerifiedAt:           toTimePtr(row.VerifiedAt),
		RejectedAt:           toTimePtr(row.RejectedAt),
		AdminNote:            row.AdminNote,
		CreatedAt:            toTimePtr(row.CreatedAt),
		UpdatedAt:            toTimePtr(row.UpdatedAt),
	}, nil
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
		ID:          row.ID,
		CustomerID:  derefInt64(row.CustomerID),
		CardID:      derefInt64(row.CardID),
		Quantity:    row.Quantity,
		TotalPrice:  row.TotalPrice,
		Status:      toOrderStatus(row.Status),
		Currency:    row.Currency,
		PublicToken: row.PublicToken,
		CreatedAt:   toTimePtr(row.CreatedAt),
		UpdatedAt:   toTimePtr(row.UpdatedAt),
	}, nil
}

func (r *OrderRepository) CreateOrderDetail(ctx context.Context, orderID int64, side string, extraInsertsPerCard int64, bidBoxTopLabel *string, bidBoxCoupleName *string, bidBoxEventDate *string, bidBoxDetails *string, brideName *string, groomName *string, brideFatherName *string, groomFatherName *string, mehndiDate *string, mehndiDay *string, mehndiTimeType *string, mehndiTime *string, mehndiDinnerTime *string, mehndiVenueName *string, mehndiVenueAddress *string, baraatDate *string, baraatDay *string, baraatTimeType *string, baraatTime *string, baraatDinnerTime *string, baraatArrivalTime *string, rukhsatiTime *string, baraatSehrabandiTime *string, baraatVenueName *string, baraatVenueAddress *string, nikkahDate *string, nikkahDay *string, nikkahTimeType *string, nikkahTime *string, nikkahDinnerTime *string, nikkahVenueName *string, nikkahVenueAddress *string, walimaDate *string, walimaDay *string, walimaTimeType *string, walimaTime *string, walimaDinnerTime *string, receptionTime *string, walimaVenueName *string, walimaVenueAddress *string, shendiDate *string, shendiDay *string, shendiTime *string, shendiTimeType *string, shendiDinnerTime *string, shendiVenueName *string, shendiVenueAddress *string, shendiArrivalTime *string, shendiRukhsatiTime *string, shendiSehrabandiTime *string, shalimaDate *string, shalimaDay *string, shalimaTime *string, shalimaTimeType *string, shalimaDinnerTime *string, shalimaVenueName *string, shalimaVenueAddress *string, shalimaArrivalTime *string, shalimaRukhsatiTime *string, shalimaSehrabandiTime *string, shalimaReceptionTime *string, rsvpName string, rsvpPhone string, notes *string) (*orderdomain.OrderDetail, error) {
	row, err := r.writer.CreateOrderDetail(ctx, orderwriter.CreateOrderDetailParams{
		OrderID:               orderID,
		Side:                  side,
		ExtraInsertsPerCard:   extraInsertsPerCard,
		TopLabel:              bidBoxTopLabel,
		CoupleName:            bidBoxCoupleName,
		BidBoxEventDate:       stringOrEmpty(bidBoxEventDate),
		BidBoxDetails:         bidBoxDetails,
		BrideName:             brideName,
		GroomName:             groomName,
		BrideFatherName:       brideFatherName,
		GroomFatherName:       groomFatherName,
		MehndiDate:            stringOrEmpty(mehndiDate),
		MehndiDay:             mehndiDay,
		MehndiTimeType:        mehndiTimeType,
		MehndiTime:            stringOrEmpty(mehndiTime),
		MehndiDinnerTime:      stringOrEmpty(mehndiDinnerTime),
		MehndiVenueName:       mehndiVenueName,
		MehndiVenueAddress:    mehndiVenueAddress,
		BaraatDate:            stringOrEmpty(baraatDate),
		BaraatDay:             baraatDay,
		BaraatTimeType:        baraatTimeType,
		BaraatTime:            stringOrEmpty(baraatTime),
		BaraatDinnerTime:      stringOrEmpty(baraatDinnerTime),
		BaraatArrivalTime:     stringOrEmpty(baraatArrivalTime),
		RukhsatiTime:          stringOrEmpty(rukhsatiTime),
		BaraatSehrabandiTime:  stringOrEmpty(baraatSehrabandiTime),
		BaraatVenueName:       baraatVenueName,
		BaraatVenueAddress:    baraatVenueAddress,
		NikkahDate:            stringOrEmpty(nikkahDate),
		NikkahDay:             nikkahDay,
		NikkahTimeType:        nikkahTimeType,
		NikkahTime:            stringOrEmpty(nikkahTime),
		NikkahDinnerTime:      stringOrEmpty(nikkahDinnerTime),
		NikkahVenueName:       nikkahVenueName,
		NikkahVenueAddress:    nikkahVenueAddress,
		WalimaDate:            stringOrEmpty(walimaDate),
		WalimaDay:             walimaDay,
		WalimaTimeType:        walimaTimeType,
		WalimaTime:            stringOrEmpty(walimaTime),
		WalimaDinnerTime:      stringOrEmpty(walimaDinnerTime),
		ReceptionTime:         stringOrEmpty(receptionTime),
		WalimaVenueName:       walimaVenueName,
		WalimaVenueAddress:    walimaVenueAddress,
		ShendiDate:            stringOrEmpty(shendiDate),
		ShendiDay:             shendiDay,
		ShendiTime:            stringOrEmpty(shendiTime),
		ShendiTimeType:        shendiTimeType,
		ShendiDinnerTime:      stringOrEmpty(shendiDinnerTime),
		ShendiVenueName:       shendiVenueName,
		ShendiVenueAddress:    shendiVenueAddress,
		ShendiArrivalTime:     stringOrEmpty(shendiArrivalTime),
		ShendiRukhsatiTime:    stringOrEmpty(shendiRukhsatiTime),
		ShendiSehrabandiTime:  stringOrEmpty(shendiSehrabandiTime),
		ShalimaDate:           stringOrEmpty(shalimaDate),
		ShalimaDay:            shalimaDay,
		ShalimaTime:           stringOrEmpty(shalimaTime),
		ShalimaTimeType:       shalimaTimeType,
		ShalimaDinnerTime:     stringOrEmpty(shalimaDinnerTime),
		ShalimaVenueName:      shalimaVenueName,
		ShalimaVenueAddress:   shalimaVenueAddress,
		ShalimaArrivalTime:    stringOrEmpty(shalimaArrivalTime),
		ShalimaRukhsatiTime:   stringOrEmpty(shalimaRukhsatiTime),
		ShalimaSehrabandiTime: stringOrEmpty(shalimaSehrabandiTime),
		ShalimaReceptionTime:  stringOrEmpty(shalimaReceptionTime),
		RsvpName:              rsvpName,
		RsvpPhone:             rsvpPhone,
		Notes:                 notes,
	})
	if err != nil {
		return nil, err
	}
	return &orderdomain.OrderDetail{
		ID:                    row.ID,
		OrderID:               row.OrderID,
		Side:                  row.Side,
		ExtraInsertsPerCard:   row.ExtraInsertsPerCard,
		BidBoxTopLabel:        emptyToNil(row.TopLabel),
		BidBoxCoupleName:      emptyToNil(row.CoupleName),
		BidBoxEventDate:       emptyToNil(row.EventDate),
		BidBoxDetails:         emptyToNil(row.BidBoxDetails),
		BrideName:             row.BrideName,
		GroomName:             row.GroomName,
		BrideFatherName:       row.BrideFatherName,
		GroomFatherName:       row.GroomFatherName,
		MehndiDate:            emptyToNil(row.MehndiDate),
		MehndiDay:             row.MehndiDay,
		MehndiTimeType:        row.MehndiTimeType,
		MehndiTime:            emptyToNil(row.MehndiTime),
		MehndiDinnerTime:      emptyToNil(row.MehndiDinnerTime),
		MehndiVenueName:       emptyToNil(row.MehndiVenueName),
		MehndiVenueAddress:    emptyToNil(row.MehndiVenueAddress),
		BaraatDate:            emptyToNil(row.BaraatDate),
		BaraatDay:             row.BaraatDay,
		BaraatTimeType:        row.BaraatTimeType,
		BaraatTime:            emptyToNil(row.BaraatTime),
		BaraatDinnerTime:      emptyToNil(row.BaraatDinnerTime),
		BaraatArrivalTime:     emptyToNil(row.BaraatArrivalTime),
		RukhsatiTime:          emptyToNil(row.RukhsatiTime),
		BaraatSehrabandiTime:  emptyToNil(row.BaraatSehrabandiTime),
		BaraatVenueName:       emptyToNil(row.BaraatVenueName),
		BaraatVenueAddress:    emptyToNil(row.BaraatVenueAddress),
		NikkahDate:            emptyToNil(row.NikkahDate),
		NikkahDay:             row.NikkahDay,
		NikkahTimeType:        row.NikkahTimeType,
		NikkahTime:            emptyToNil(row.NikkahTime),
		NikkahDinnerTime:      emptyToNil(row.NikkahDinnerTime),
		NikkahVenueName:       emptyToNil(row.NikkahVenueName),
		NikkahVenueAddress:    emptyToNil(row.NikkahVenueAddress),
		WalimaDate:            emptyToNil(row.WalimaDate),
		WalimaDay:             row.WalimaDay,
		WalimaTimeType:        row.WalimaTimeType,
		WalimaTime:            emptyToNil(row.WalimaTime),
		WalimaDinnerTime:      emptyToNil(row.WalimaDinnerTime),
		WalimaVenueName:       emptyToNil(row.WalimaVenueName),
		WalimaVenueAddress:    emptyToNil(row.WalimaVenueAddress),
		ReceptionTime:         emptyToNil(row.ReceptionTime),
		ShendiDate:            emptyToNil(row.ShendiDate),
		ShendiDay:             row.ShendiDay,
		ShendiTimeType:        row.ShendiTimeType,
		ShendiTime:            emptyToNil(row.ShendiTime),
		ShendiDinnerTime:      emptyToNil(row.ShendiDinnerTime),
		ShendiVenueName:       emptyToNil(row.ShendiVenueName),
		ShendiVenueAddress:    emptyToNil(row.ShendiVenueAddress),
		ShendiArrivalTime:     emptyToNil(row.ShendiArrivalTime),
		ShendiRukhsatiTime:    emptyToNil(row.ShendiRukhsatiTime),
		ShendiSehrabandiTime:  emptyToNil(row.ShendiSehrabandiTime),
		ShalimaDate:           emptyToNil(row.ShalimaDate),
		ShalimaDay:            row.ShalimaDay,
		ShalimaTimeType:       row.ShalimaTimeType,
		ShalimaTime:           emptyToNil(row.ShalimaTime),
		ShalimaDinnerTime:     emptyToNil(row.ShalimaDinnerTime),
		ShalimaVenueName:      emptyToNil(row.ShalimaVenueName),
		ShalimaVenueAddress:   emptyToNil(row.ShalimaVenueAddress),
		ShalimaArrivalTime:    emptyToNil(row.ShalimaArrivalTime),
		ShalimaRukhsatiTime:   emptyToNil(row.ShalimaRukhsatiTime),
		ShalimaSehrabandiTime: emptyToNil(row.ShalimaSehrabandiTime),
		ShalimaReceptionTime:  emptyToNil(row.ShalimaReceptionTime),
		RsvpName:              row.RsvpName,
		RsvpPhone:             row.RsvpPhone,
		Notes:                 row.Notes,
		CreatedAt:             toTimePtr(row.CreatedAt),
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

func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toOrderStatus(s *string) orderdomain.OrderStatus {
	if s == nil {
		return orderdomain.PendingOrderStatus
	}
	return orderdomain.OrderStatus(*s)
}

func toPaymentMethod(s string) orderdomain.PaymentMethod {
	if s == "" {
		return orderdomain.BankTransferPaymentMethod
	}
	return orderdomain.PaymentMethod(s)
}

func toPaymentStatus(s *string) orderdomain.PaymentStatus {
	if s == nil {
		return ""
	}
	return orderdomain.PaymentStatus(*s)
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
