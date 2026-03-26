package customerdomain

import (
	"context"
	"time"
)

type Customer struct {
	ID         int64
	Name       string
	Email      *string
	Phone      *string
	Address    *string
	City       *string
	PostalCode *string
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

type CustomerRepo interface {
	CustomerWriter
	CustomerReader
}

type CustomerWriter interface {
	CreateCustomer(ctx context.Context, name string, email *string, phone *string, address *string, city *string, postalCode *string) (*Customer, error)
}

type CustomerReader interface {
	GetCustomerByID(ctx context.Context, id int64) (*Customer, error)
}
