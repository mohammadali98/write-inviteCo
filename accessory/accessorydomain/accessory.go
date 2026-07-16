package accessorydomain

import (
	"context"
	"time"
)

// Accessory is a non-purchasable showcase item (wax seals, ribbons, etc.)
// shown for reference alongside the real, purchasable cards/bid boxes.
// Deliberately has no price, no card_id, no foil/insert fields.
type Accessory struct {
	ID          int64
	Name        string
	Category    string
	Description *string
	IsActive    bool
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

type AccessoryImage struct {
	ID          int64
	AccessoryID int64
	ImageURL    string
	SortOrder   int32
	CreatedAt   *time.Time
}

type AccessoryReader interface {
	ListActiveAccessories(ctx context.Context) ([]*Accessory, error)
	GetAccessoryImagesByAccessoryID(ctx context.Context, accessoryID int64) ([]*AccessoryImage, error)
}
