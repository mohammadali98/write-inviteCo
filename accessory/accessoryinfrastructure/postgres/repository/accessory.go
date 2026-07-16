package accessoryrepository

import (
	"context"
	"time"

	"writeandinviteco/inviteandco/accessory/accessorydomain"
	accessoryreader "writeandinviteco/inviteandco/accessory/accessoryinfrastructure/postgres/reader"

	"github.com/jackc/pgx/v5/pgtype"
)

type AccessoryRepository struct {
	reader *accessoryreader.Queries
}

func NewAccessoryRepository(reader *accessoryreader.Queries) *AccessoryRepository {
	return &AccessoryRepository{reader: reader}
}

func (r *AccessoryRepository) ListActiveAccessories(ctx context.Context) ([]*accessorydomain.Accessory, error) {
	rows, err := r.reader.ListActiveAccessories(ctx)
	if err != nil {
		return nil, err
	}
	accessories := make([]*accessorydomain.Accessory, len(rows))
	for i, row := range rows {
		accessories[i] = &accessorydomain.Accessory{
			ID:          row.ID,
			Name:        row.Name,
			Category:    row.Category,
			Description: row.Description,
			IsActive:    row.IsActive,
			CreatedAt:   toTimePtr(row.CreatedAt),
			UpdatedAt:   toTimePtr(row.UpdatedAt),
		}
	}
	return accessories, nil
}

func (r *AccessoryRepository) GetAccessoryImagesByAccessoryID(ctx context.Context, accessoryID int64) ([]*accessorydomain.AccessoryImage, error) {
	rows, err := r.reader.GetAccessoryImagesByAccessoryID(ctx, accessoryID)
	if err != nil {
		return nil, err
	}
	images := make([]*accessorydomain.AccessoryImage, len(rows))
	for i, row := range rows {
		images[i] = &accessorydomain.AccessoryImage{
			ID:          row.ID,
			AccessoryID: row.AccessoryID,
			ImageURL:    row.ImageUrl,
			SortOrder:   row.SortOrder,
			CreatedAt:   toTimePtr(row.CreatedAt),
		}
	}
	return images, nil
}

func toTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
