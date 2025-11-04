package moving

import (
	"context"

	servergrpc "github.com/ingvarmattis/moving/gen/servergrpc/moving"
	"github.com/ingvarmattis/moving/src/services"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handlers struct {
	Service services.SvcLayer
}

func (s *Handlers) CreateOrder(ctx context.Context, order *servergrpc.CreateOrderRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Handlers) GetOrder(ctx context.Context, _ *emptypb.Empty) (*servergrpc.GetOrderResponse, error) {
	return &servergrpc.GetOrderResponse{}, nil
}
