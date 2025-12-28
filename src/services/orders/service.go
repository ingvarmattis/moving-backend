package orders

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ingvarmattis/moving/src/infra/utils"
	repo "github.com/ingvarmattis/moving/src/repositories/orders"
)

var ErrNotFound = errors.New("not found")

//go:generate bash -c "mkdir -p mocks"
//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
type ordersStorage interface {
	CreateOrder(ctx context.Context, req *repo.CreateOrderRequest) (*repo.Order, error)
	Orders(ctx context.Context) ([]*repo.Order, error)
	OrderByID(ctx context.Context, id uint64) (*repo.Order, error)
	UpdateOrder(ctx context.Context, req *repo.UpdateOrderRequest) error
}

type Service struct {
	ordersStorage ordersStorage
}

func NewService(ordersStorage ordersStorage) *Service {
	return &Service{ordersStorage: ordersStorage}
}

func (s *Service) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	repoReq := &repo.CreateOrderRequest{
		PropertySize:   repo.PropertySize(req.PropertySize),
		MoveDate:       req.MoveDate,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		MoveFrom:       req.MoveFrom,
		MoveTo:         req.MoveTo,
		AdditionalInfo: req.AdditionalInfo,
	}

	order, err := s.ordersStorage.CreateOrder(ctx, repoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create order | %w", err)
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

func (s *Service) Orders(ctx context.Context) ([]*Order, error) {
	repoOrders, err := s.ordersStorage.Orders(ctx)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get all orders | %w", err)
	}

	orders := make([]*Order, 0, len(repoOrders))

	for _, repoOrder := range repoOrders {
		orders = append(orders, &Order{
			ID:             repoOrder.ID,
			PropertySize:   PropertySize(repoOrder.PropertySize),
			OrderStatus:    OrderStatus(repoOrder.OrderStatus),
			MoveDate:       repoOrder.MoveDate,
			Name:           repoOrder.Name,
			Email:          repoOrder.Email,
			Phone:          repoOrder.Phone,
			MoveFrom:       repoOrder.MoveFrom,
			MoveTo:         repoOrder.MoveTo,
			AdditionalInfo: repoOrder.AdditionalInfo,
		})
	}

	return orders, nil
}

func (s *Service) OrderByID(ctx context.Context, id uint64) (*Order, error) {
	order, err := s.ordersStorage.OrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get order by id | %w", err)
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

func (s *Service) UpdateOrder(ctx context.Context, req *UpdateOrderRequest) error {
	repoReq := &repo.UpdateOrderRequest{
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
		repoReq.PropertySize = utils.PtrIfNotZero(repo.PropertySize(*req.PropertySize))
	}

	if req.OrderStatus != nil {
		repoReq.OrderStatus = utils.PtrIfNotZero(repo.OrderStatus(*req.OrderStatus))
	}

	if err := s.ordersStorage.UpdateOrder(ctx, repoReq); err != nil {
		return fmt.Errorf("failed to update order | %w", err)
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
