package moving

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

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

func (p *Postgres) AllOrders(ctx context.Context) ([]*Order, error) {
	query := `
select
	id, name, email, phone, move_date, move_from, move_to,
	property_size, status, additional_info, created_at, updated_at
from moving.orders
order by created_at desc
`

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		log.Fatal(err)
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

	return orders, nil
}

func (p *Postgres) UpdateOrder(ctx context.Context, req *UpdateOrderRequest) error {
	if req.ID == 0 {
		return fmt.Errorf("invalid id")
	}

	setClauses := make([]string, 0)
	args := make([]interface{}, 0)
	argPos := 1

	if req.PropertySize != nil {
		setClauses = append(setClauses, fmt.Sprintf("property_size = $%d", argPos))
		args = append(args, *req.PropertySize)
		argPos++
	}
	if req.OrderStatus != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *req.OrderStatus)
		argPos++
	}
	if req.MoveDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("move_date = $%d", argPos))
		args = append(args, *req.MoveDate)
		argPos++
	}
	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *req.Name)
		argPos++
	}
	if req.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argPos))
		args = append(args, *req.Email)
		argPos++
	}
	if req.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argPos))
		args = append(args, *req.Phone)
		argPos++
	}
	if req.MoveFrom != nil {
		setClauses = append(setClauses, fmt.Sprintf("move_from = $%d", argPos))
		args = append(args, *req.MoveFrom)
		argPos++
	}
	if req.MoveTo != nil {
		setClauses = append(setClauses, fmt.Sprintf("move_to = $%d", argPos))
		args = append(args, *req.MoveTo)
		argPos++
	}
	if req.AdditionalInfo != nil {
		setClauses = append(setClauses, fmt.Sprintf("additional_info = $%d", argPos))
		args = append(args, *req.AdditionalInfo)
		argPos++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no set clauses found")
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = now()"))

	query := fmt.Sprintf(`
update moving.orders
set %s
where id = $%d
`, strings.Join(setClauses, ", "), argPos)

	args = append(args, req.ID)

	if _, err := p.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("failed update row | %w", err)
	}

	return nil
}

type PropertySize int

const (
	PropertySizeUnknown PropertySize = iota
	PropertySizeStudio
	PropertySize1Bedroom
	PropertySize2Bedrooms
	PropertySize3Bedrooms
	PropertySize4PlusBedrooms
	PropertySizeCommercial
)

func (p PropertySize) String() string {
	switch p {
	case PropertySizeStudio:
		return "studio"
	case PropertySize1Bedroom:
		return "1_bedroom"
	case PropertySize2Bedrooms:
		return "2_bedrooms"
	case PropertySize3Bedrooms:
		return "3_bedrooms"
	case PropertySize4PlusBedrooms:
		return "4_plus_bedrooms"
	case PropertySizeCommercial:
		return "commercial"
	default:
		return "unknown"
	}
}

func NewPropertySize(s string) PropertySize {
	switch s {
	case "studio":
		return PropertySizeStudio
	case "1_bedroom":
		return PropertySize1Bedroom
	case "2_bedrooms":
		return PropertySize2Bedrooms
	case "3_bedrooms":
		return PropertySize3Bedrooms
	case "4_plus_bedrooms":
		return PropertySize4PlusBedrooms
	case "commercial":
		return PropertySizeCommercial
	default:
		return PropertySizeUnknown
	}
}

type OrderStatus int

const (
	OrderStatusUnknown OrderStatus = iota
	OrderStatusCreated
	OrderStatusRejected
	OrderStatusInProgress
	OrderStatusDone
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusCreated:
		return "created"
	case OrderStatusRejected:
		return "rejected"
	case OrderStatusInProgress:
		return "in_progress"
	case OrderStatusDone:
		return "done"
	default:
		return "unknown"
	}
}

func NewOrderStatus(s string) OrderStatus {
	switch s {
	case "created":
		return OrderStatusCreated
	case "rejected":
		return OrderStatusRejected
	case "in_progress":
		return OrderStatusInProgress
	case "done":
		return OrderStatusDone
	default:
		return OrderStatusUnknown
	}
}

type CreateOrderRequest struct {
	PropertySize   PropertySize
	MoveDate       time.Time
	Name           string
	Email          string
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
	Email          string
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
