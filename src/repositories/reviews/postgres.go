package reviews

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (p *Postgres) Reviews(ctx context.Context) ([]*Review, error) {
	query := `
select
	id, name, rate, photo_url, text, review_url, created_at, updated_at
from moving.reviews
order by id
`

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviews | %w", err)
	}
	defer rows.Close()

	var reviews []*Review

	for rows.Next() {
		var review Review

		if err = rows.Scan(
			&review.ID, &review.Name, &review.Rate, &review.PhotoURL,
			&review.Text, &review.URL, &review.CreatedAt, &review.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed scan review | %w", err)
		}

		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed get all reviews | %w", err)
	}

	if len(reviews) == 0 {
		return nil, ErrNotFound
	}

	return reviews, nil
}

type Review struct {
	ID        uint64
	Rate      int32
	Text      string
	Name      string
	PhotoURL  string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}
