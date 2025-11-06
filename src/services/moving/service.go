package moving

import (
	"context"
	"errors"
	"fmt"
	"github.com/ingvarmattis/moving/src/infra/utils"
	"time"

	repo "github.com/ingvarmattis/moving/src/repositories/moving"
)

const serviceName = "moving-service"

var ErrNotFound = errors.New("not found")

//go:generate bash -c "mkdir -p mocks"
//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
type movingStorage interface {
	CreateOrder(ctx context.Context, req *repo.CreateOrderRequest) (*repo.Order, error)
	GetOrders(ctx context.Context) ([]*repo.Order, error)
	UpdateOrder(ctx context.Context, req *repo.UpdateOrderRequest) error
}

type Service struct {
	movingStorage movingStorage
}

func NewService(movingStorage movingStorage) *Service {
	return &Service{movingStorage: movingStorage}
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

	order, err := s.movingStorage.CreateOrder(ctx, repoReq)
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

func (s *Service) GetOrders(ctx context.Context) ([]*Order, error) {
	repoOrders, err := s.movingStorage.GetOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all order | %w", err)
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

	if err := s.movingStorage.UpdateOrder(ctx, repoReq); err != nil {
		return fmt.Errorf("failed to update order | %w", err)
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
