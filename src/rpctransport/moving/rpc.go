package moving

import (
	"context"
	"fmt"
	"github.com/ingvarmattis/moving/src/infra/utils"
	"time"

	"github.com/ingvarmattis/moving/src/services"
	svc "github.com/ingvarmattis/moving/src/services/moving"
)

type Handlers struct {
	Service services.SvcLayer
}

func (s *Handlers) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	svcReq := &svc.CreateOrderRequest{
		PropertySize:   svc.PropertySize(req.PropertySize),
		MoveDate:       req.MoveDate,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		MoveFrom:       req.MoveFrom,
		MoveTo:         req.MoveTo,
		AdditionalInfo: req.AdditionalInfo,
	}

	order, err := s.Service.MovingService.CreateOrder(ctx, svcReq)
	if err != nil {
		return nil, fmt.Errorf("failed create order | %w", err)
	}

	return &Order{
		ID:             order.ID,
		PropertySize:   PropertySize(order.PropertySize),
		OrderStatus:    OrderStatus(order.OrderStatus),
		MoveDate:       order.MoveDate,
		Name:           order.Name,
		Email:          order.Email,
		Phone:          order.Phone,
		MoveFrom:       order.MoveFrom,
		MoveTo:         order.MoveTo,
		AdditionalInfo: order.AdditionalInfo,
	}, nil
}

func (s *Handlers) GetOrders(ctx context.Context) ([]*Order, error) {
	repoOrders, err := s.Service.MovingService.GetOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed get all order | %w", err)
	}

	orders := make([]*Order, 0, len(repoOrders))

	for _, order := range repoOrders {
		orders = append(orders, &Order{
			ID:             order.ID,
			PropertySize:   PropertySize(order.PropertySize),
			OrderStatus:    OrderStatus(order.OrderStatus),
			MoveDate:       order.MoveDate,
			Name:           order.Name,
			Email:          order.Email,
			Phone:          order.Phone,
			MoveFrom:       order.MoveFrom,
			MoveTo:         order.MoveTo,
			AdditionalInfo: order.AdditionalInfo,
		})
	}

	return orders, nil
}

func (s *Handlers) UpdateOrder(ctx context.Context, req *UpdateOrderRequest) error {
	svcReq := &svc.UpdateOrderRequest{
		ID:             req.ID,
		MoveDate:       req.MoveDate,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		MoveFrom:       req.MoveFrom,
		MoveTo:         req.MoveTo,
		AdditionalInfo: req.AdditionalInfo,
	}

	if req.PropertySize != nil {
		svcReq.PropertySize = utils.PtrIfNotZero(svc.PropertySize(*req.PropertySize))
	}

	if req.OrderStatus != nil {
		svcReq.OrderStatus = utils.PtrIfNotZero(svc.OrderStatus(*req.OrderStatus))
	}

	if err := s.Service.MovingService.UpdateOrder(ctx, svcReq); err != nil {
		return fmt.Errorf("failed update order | %w", err)
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

type OrderStatus int

const (
	OrderStatusUnknown OrderStatus = iota
	OrderStatusCreated
	OrderStatusRejected
	OrderStatusInProgress
	OrderStatusDone
)

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
