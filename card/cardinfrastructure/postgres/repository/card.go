package cardrepository

import (
	"context"
	"time"

	"writeandinviteco/inviteandco/card/carddomain"
	cardreader "writeandinviteco/inviteandco/card/cardinfrastructure/postgres/reader"
	cardwriter "writeandinviteco/inviteandco/card/cardinfrastructure/postgres/writer"

	"github.com/jackc/pgx/v5/pgtype"
)

type CardRepository struct {
	reader *cardreader.Queries
	writer *cardwriter.Queries
}

func NewCardRepository(reader *cardreader.Queries, writer *cardwriter.Queries) *CardRepository {
	return &CardRepository{reader: reader, writer: writer}
}

func (r *CardRepository) GetAllCards(ctx context.Context) ([]*carddomain.Card, error) {
	rows, err := r.reader.GetAllCards(ctx)
	if err != nil {
		return nil, err
	}
	cards := make([]*carddomain.Card, len(rows))
	for i, row := range rows {
		cards[i] = &carddomain.Card{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			PricePKR:    row.PricePkr,
			PriceNOK:    row.PriceNok,
			Image:       row.Image,
			Category:    row.Category,
			CreatedAt:   toTimePtr(row.CreatedAt),
			UpdatedAt:   toTimePtr(row.UpdatedAt),
		}
	}
	return cards, nil
}

func (r *CardRepository) GetCardByID(ctx context.Context, id int64) (*carddomain.Card, error) {
	row, err := r.reader.GetCardByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &carddomain.Card{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		PricePKR:    row.PricePkr,
		PriceNOK:    row.PriceNok,
		Image:       row.Image,
		Category:    row.Category,
		CreatedAt:   toTimePtr(row.CreatedAt),
		UpdatedAt:   toTimePtr(row.UpdatedAt),
	}, nil
}

func (r *CardRepository) GetCardsByCategory(ctx context.Context, category string) ([]*carddomain.Card, error) {
	rows, err := r.reader.GetCardsByCategory(ctx, category)
	if err != nil {
		return nil, err
	}
	cards := make([]*carddomain.Card, len(rows))
	for i, row := range rows {
		cards[i] = &carddomain.Card{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			PricePKR:    row.PricePkr,
			PriceNOK:    row.PriceNok,
			Image:       row.Image,
			Category:    row.Category,
			CreatedAt:   toTimePtr(row.CreatedAt),
			UpdatedAt:   toTimePtr(row.UpdatedAt),
		}
	}
	return cards, nil
}

func (r *CardRepository) GetCardImagesByCardID(ctx context.Context, cardID int64) ([]*carddomain.CardImage, error) {
	rows, err := r.reader.GetCardImagesByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	images := make([]*carddomain.CardImage, len(rows))
	for i, row := range rows {
		images[i] = &carddomain.CardImage{
			ID:        row.ID,
			CardID:    row.CardID,
			Image:     row.Image,
			SortOrder: row.SortOrder,
			CreatedAt: toTimePtr(row.CreatedAt),
		}
	}
	return images, nil
}

func (r *CardRepository) CreateCard(ctx context.Context, name string, description *string, pricePKR int64, priceNOK int64, image string, category string) (*carddomain.Card, error) {
	row, err := r.writer.CreateCard(ctx, cardwriter.CreateCardParams{
		Name:        name,
		Description: description,
		PricePkr:    pricePKR,
		PriceNok:    priceNOK,
		Image:       image,
		Category:    category,
	})
	if err != nil {
		return nil, err
	}
	return &carddomain.Card{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		PricePKR:    row.PricePkr,
		PriceNOK:    row.PriceNok,
		Image:       row.Image,
		Category:    row.Category,
		CreatedAt:   toTimePtr(row.CreatedAt),
		UpdatedAt:   toTimePtr(row.UpdatedAt),
	}, nil
}

func (r *CardRepository) UpdateCard(ctx context.Context, id int64, name string, description *string, pricePKR int64, priceNOK int64, image string, category string) error {
	return r.writer.UpdateCard(ctx, cardwriter.UpdateCardParams{
		ID:          id,
		Name:        name,
		Description: description,
		PricePkr:    pricePKR,
		PriceNok:    priceNOK,
		Image:       image,
		Category:    category,
	})
}

func (r *CardRepository) DeleteCard(ctx context.Context, id int64) error {
	return r.writer.DeleteCard(ctx, id)
}

func (r *CardRepository) CreateCardImage(ctx context.Context, cardID int64, image string, sortOrder int32) (*carddomain.CardImage, error) {
	row, err := r.writer.CreateCardImage(ctx, cardwriter.CreateCardImageParams{
		CardID:    cardID,
		Image:     image,
		SortOrder: sortOrder,
	})
	if err != nil {
		return nil, err
	}
	return &carddomain.CardImage{
		ID:        row.ID,
		CardID:    row.CardID,
		Image:     row.Image,
		SortOrder: row.SortOrder,
		CreatedAt: toTimePtr(row.CreatedAt),
	}, nil
}

func (r *CardRepository) DeleteCardImagesByCardID(ctx context.Context, cardID int64) error {
	return r.writer.DeleteCardImagesByCardID(ctx, cardID)
}

func toTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
