package moving

import (
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

const packageName = "moving"

var ErrNotFound = errors.New("not found")

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}
