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
			Price:       row.Price,
			Image:       row.Image,
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
		Price:       row.Price,
		Image:       row.Image,
		CreatedAt:   toTimePtr(row.CreatedAt),
		UpdatedAt:   toTimePtr(row.UpdatedAt),
	}, nil
}

func (r *CardRepository) CreateCard(ctx context.Context, name string, description *string, price int64, image string) (*carddomain.Card, error) {
	row, err := r.writer.CreateCard(ctx, cardwriter.CreateCardParams{
		Name:        name,
		Description: description,
		Price:       price,
		Image:       image,
	})
	if err != nil {
		return nil, err
	}
	return &carddomain.Card{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Price:       row.Price,
		Image:       row.Image,
		CreatedAt:   toTimePtr(row.CreatedAt),
		UpdatedAt:   toTimePtr(row.UpdatedAt),
	}, nil
}

func (r *CardRepository) UpdateCard(ctx context.Context, id int64, name string, description *string, price int64, image string) error {
	return r.writer.UpdateCard(ctx, cardwriter.UpdateCardParams{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
		Image:       image,
	})
}

func (r *CardRepository) DeleteCard(ctx context.Context, id int64) error {
	return r.writer.DeleteCard(ctx, id)
}

func toTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
