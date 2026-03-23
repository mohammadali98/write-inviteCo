package customerrepository

import (
	"context"
	"time"

	"writeandinviteco/inviteandco/customer/customerdomain"
	customerreader "writeandinviteco/inviteandco/customer/customerinfrastructure/postgres/reader"
	customerwriter "writeandinviteco/inviteandco/customer/customerinfrastructure/postgres/writer"

	"github.com/jackc/pgx/v5/pgtype"
)

type CustomerRepository struct {
	reader *customerreader.Queries
	writer *customerwriter.Queries
}

func NewCustomerRepository(reader *customerreader.Queries, writer *customerwriter.Queries) *CustomerRepository {
	return &CustomerRepository{reader: reader, writer: writer}
}

func (r *CustomerRepository) GetCustomerByID(ctx context.Context, id int64) (*customerdomain.Customer, error) {
	row, err := r.reader.GetCustomerByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &customerdomain.Customer{
		ID:        row.ID,
		Name:      row.Name,
		Email:     row.Email,
		Phone:     row.Phone,
		CreatedAt: toTimePtr(row.CreatedAt),
		UpdatedAt: toTimePtr(row.UpdatedAt),
	}, nil
}

func (r *CustomerRepository) CreateCustomer(ctx context.Context, name string, email *string, phone *string) (*customerdomain.Customer, error) {
	row, err := r.writer.CreateCustomer(ctx, customerwriter.CreateCustomerParams{
		Name:  name,
		Email: email,
		Phone: phone,
	})
	if err != nil {
		return nil, err
	}
	return &customerdomain.Customer{
		ID:        row.ID,
		Name:      row.Name,
		Email:     row.Email,
		Phone:     row.Phone,
		CreatedAt: toTimePtr(row.CreatedAt),
		UpdatedAt: toTimePtr(row.UpdatedAt),
	}, nil
}

func toTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
