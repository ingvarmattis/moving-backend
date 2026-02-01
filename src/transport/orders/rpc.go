package orders

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ingvarmattis/moving/src/infra/utils"
	"github.com/ingvarmattis/moving/src/services"
	orderssvc "github.com/ingvarmattis/moving/src/services/orders"
)

var ErrNotFound = errors.New("not found")

type Handlers struct {
	OrdersService services.OrdersService
}

func (s *Handlers) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	svcReq := &orderssvc.CreateOrderRequest{
		PropertySize:   orderssvc.PropertySize(req.PropertySize),
		MoveDate:       req.MoveDate,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		MoveFrom:       req.MoveFrom,
		MoveTo:         req.MoveTo,
		AdditionalInfo: req.AdditionalInfo,
	}

	order, err := s.OrdersService.CreateOrder(ctx, svcReq)
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
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}, nil
}

func normalizeFilter(filter *Filter) *orderssvc.Filter {
	if filter == nil {
		return nil
	}

	var createdFrom, createdTo, moveDateFrom, moveDateTo *time.Time
	if filter.CreatedFrom != nil && !filter.CreatedFrom.IsZero() {
		createdFrom = filter.CreatedFrom
	}
	if filter.CreatedTo != nil && !filter.CreatedTo.IsZero() {
		createdTo = filter.CreatedTo
	}
	if filter.MoveDateFrom != nil && !filter.MoveDateFrom.IsZero() {
		moveDateFrom = filter.MoveDateFrom
	}
	if filter.MoveDateTo != nil && !filter.MoveDateTo.IsZero() {
		moveDateTo = filter.MoveDateTo
	}

	var orderStatus *orderssvc.OrderStatus
	if filter.OrderStatus != nil {
		os := orderssvc.OrderStatus(*filter.OrderStatus)
		orderStatus = &os
	}

	var propertySize *orderssvc.PropertySize
	if filter.PropertySize != nil {
		ps := orderssvc.PropertySize(*filter.PropertySize)
		propertySize = &ps
	}

	// Check if all fields are empty
	if orderStatus == nil && propertySize == nil && createdFrom == nil &&
		createdTo == nil && moveDateFrom == nil && moveDateTo == nil {
		return nil
	}

	return &orderssvc.Filter{
		OrderStatus:  orderStatus,
		PropertySize: propertySize,
		CreatedFrom:  createdFrom,
		CreatedTo:    createdTo,
		MoveDateFrom: moveDateFrom,
		MoveDateTo:   moveDateTo,
	}
}

func (s *Handlers) Orders(ctx context.Context, filter *Filter) ([]*Order, error) {
	svcFilter := normalizeFilter(filter)

	repoOrders, err := s.OrdersService.Orders(ctx, svcFilter)
	if err != nil {
		if errors.Is(err, orderssvc.ErrNotFound) {
			return nil, ErrNotFound
		}

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
			CreatedAt:      order.CreatedAt,
			UpdatedAt:      order.UpdatedAt,
		})
	}

	if len(orders) == 0 {
		return nil, ErrNotFound
	}

	return orders, nil
}

func (s *Handlers) OrderByID(ctx context.Context, id uint64) (*Order, error) {
	order, err := s.OrdersService.OrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, orderssvc.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed get order | %w", err)
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
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}, nil
}

func (s *Handlers) UpdateOrder(ctx context.Context, req *UpdateOrderRequest) error {
	svcReq := &orderssvc.UpdateOrderRequest{
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
		svcReq.PropertySize = utils.PtrIfNotZero(orderssvc.PropertySize(*req.PropertySize))
	}

	if req.OrderStatus != nil {
		svcReq.OrderStatus = utils.PtrIfNotZero(orderssvc.OrderStatus(*req.OrderStatus))
	}

	if err := s.OrdersService.UpdateOrder(ctx, svcReq); err != nil {
		return fmt.Errorf("failed update order | %w", err)
	}

	return nil
}

type PropertySize int8

const (
	PropertySizeUnknown PropertySize = iota
	PropertySizeStudio
	PropertySize1Bedroom
	PropertySize2Bedrooms
	PropertySize3Bedrooms
	PropertySize4PlusBedrooms
	PropertySizeCommercial
)

type OrderStatus int8

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
	Name           string
	Email          *string
	Phone          string
	MoveFrom       string
	MoveTo         string
	AdditionalInfo *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
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

type Filter struct {
	OrderStatus  *OrderStatus
	PropertySize *PropertySize
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	MoveDateFrom *time.Time
	MoveDateTo   *time.Time
}
