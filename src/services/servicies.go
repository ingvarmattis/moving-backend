package services

import (
	"context"

	"github.com/ingvarmattis/moving/src/services/orders"
	"github.com/ingvarmattis/moving/src/services/reviews"
)

type OrdersHandlers struct {
	OrdersService OrdersService
}

type ReviewsHandlers struct {
	ReviewsService ReviewsService
}

type OrdersService interface {
	CreateOrder(ctx context.Context, req *orders.CreateOrderRequest) (*orders.Order, error)
	Orders(ctx context.Context, filter *orders.Filter) ([]*orders.Order, error)
	OrderByID(ctx context.Context, id uint64) (*orders.Order, error)
	UpdateOrder(ctx context.Context, req *orders.UpdateOrderRequest) error
}

type ReviewsService interface {
	Reviews(ctx context.Context) ([]reviews.Review, error)
}
