package orders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (p *Postgres) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	query := `
insert into moving.orders (
	name, email, phone, move_date, move_from, move_to,
	property_size, status, additional_info, created_at, updated_at
) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
returning id, name, email, phone, move_date, move_from, move_to,
	property_size, status, additional_info, created_at, updated_at;
`

	row := p.pool.QueryRow(ctx, query,
		req.Name, req.Email, req.Phone, req.MoveDate, req.MoveFrom,
		req.MoveTo, req.PropertySize, OrderStatusCreated, req.AdditionalInfo,
	)

	var (
		order                     Order
		propertySize, orderStatus string
		additionInfo              sql.NullString
	)
	err := row.Scan(
		&order.ID, &order.Name, &order.Email, &order.Phone,
		&order.MoveDate, &order.MoveFrom, &order.MoveTo, &propertySize,
		&orderStatus, &additionInfo, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan inserted order: %w", err)
	}

	order.PropertySize = NewPropertySize(propertySize)
	order.OrderStatus = NewOrderStatus(orderStatus)

	if additionInfo.Valid {
		order.AdditionalInfo = &additionInfo.String
	}

	return &order, nil
}

func (p *Postgres) Orders(ctx context.Context) ([]*Order, error) {
	query := `
select
	id, name, email, phone, move_date, move_from, move_to,
	property_size, status, additional_info, created_at, updated_at
from moving.orders
order by created_at desc
`

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders | %w", err)
	}
	defer rows.Close()

	var orders []*Order

	for rows.Next() {
		var (
			order                     Order
			propertySize, orderStatus string
		)
		if err = rows.Scan(
			&order.ID, &order.Name, &order.Email, &order.Phone, &order.MoveDate, &order.MoveFrom, &order.MoveTo,
			&propertySize, &orderStatus, &order.AdditionalInfo, &order.CreatedAt, &order.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed scan order | %w", err)
		}

		order.PropertySize = NewPropertySize(propertySize)
		order.OrderStatus = NewOrderStatus(orderStatus)

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed get all orders | %w", err)
	}

	if len(orders) == 0 {
		return nil, ErrNotFound
	}

	return orders, nil
}

func (p *Postgres) OrderByID(ctx context.Context, id uint64) (*Order, error) {
	query := `
select
	id, name, email, phone, move_date, move_from, move_to,
	property_size, status, additional_info, created_at, updated_at
from moving.orders
where id = $1
`

	row := p.pool.QueryRow(ctx, query, id)

	var (
		order                     Order
		propertySize, orderStatus string
	)

	if err := row.Scan(
		&order.ID, &order.Name, &order.Email, &order.Phone, &order.MoveDate, &order.MoveFrom, &order.MoveTo,
		&propertySize, &orderStatus, &order.AdditionalInfo, &order.CreatedAt, &order.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed scan order | %w", err)
	}

	order.PropertySize = NewPropertySize(propertySize)
	order.OrderStatus = NewOrderStatus(orderStatus)

	return &order, nil
}

func (p *Postgres) UpdateOrder(ctx context.Context, req *UpdateOrderRequest) error {
	if req.ID == 0 {
		return fmt.Errorf("invalid id")
	}

	// Prepare arguments with proper types
	var propertySizeStr, orderStatusStr interface{}
	if req.PropertySize != nil {
		propertySizeStr = req.PropertySize.String()
	}
	if req.OrderStatus != nil {
		orderStatusStr = req.OrderStatus.String()
	}

	query := `
update moving.orders
set
	property_size = coalesce($1, property_size),
	status = coalesce($2, status),
	move_date = coalesce($3, move_date),
	name = coalesce(nullif($4, ''), name),
	email = coalesce(nullif($5, ''), email),
	phone = coalesce(nullif($6, ''), phone),
	move_from = coalesce(nullif($7, ''), move_from),
	move_to = coalesce(nullif($8, ''), move_to),
	additional_info = coalesce($9, additional_info),
	updated_at = now()
where id = $10
`

	args := []interface{}{
		propertySizeStr, orderStatusStr, req.MoveDate, req.Name,
		req.Email, req.Phone, req.MoveFrom, req.MoveTo,
		req.AdditionalInfo, req.ID,
	}

	if _, err := p.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("failed update row | %w", err)
	}

	return nil
}

type CreateOrderRequest struct {
	PropertySize   PropertySize
	MoveDate       time.Time
	Name           string
	Email          *string
	Phone          string
	MoveFrom       string
	MoveTo         string
	AdditionalInfo *string
}

type Order struct {
	ID             uint64
	PropertySize   PropertySize
	OrderStatus    OrderStatus
	MoveDate       time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Name           string
	Email          *string
	Phone          string
	MoveFrom       string
	MoveTo         string
	AdditionalInfo *string
}

type UpdateOrderRequest struct {
	ID             uint64
	PropertySize   *PropertySize
	OrderStatus    *OrderStatus
	MoveDate       *time.Time
	Name           *string
	Email          *string
	Phone          *string
	MoveFrom       *string
	MoveTo         *string
	AdditionalInfo *string
}
