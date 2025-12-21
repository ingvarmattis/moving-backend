package services

import (
	"context"

	svc "github.com/ingvarmattis/moving/src/services/moving"
)

type SvcLayer struct {
	MovingService MovingService
}

type MovingService interface {
	CreateOrder(ctx context.Context, req *svc.CreateOrderRequest) (*svc.Order, error)
	Orders(ctx context.Context) ([]*svc.Order, error)
	OrderByID(ctx context.Context, id uint64) (*svc.Order, error)
	UpdateOrder(ctx context.Context, req *svc.UpdateOrderRequest) error
}
