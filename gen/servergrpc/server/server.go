package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/go-playground/validator/v10"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	rpc "github.com/ingvarmattis/moving/gen/servergrpc/moving"
	"github.com/ingvarmattis/moving/src/infra/utils"
	"github.com/ingvarmattis/moving/src/rpctransport"
)

const domain = "moving"

var (
	ErrPortNotSpecified = errors.New("port not specified")
	ErrValidationFailed = errors.New("validation failed")
)

type OrdersGRPCHandlers interface {
	CreateOrder(ctx context.Context, req *rpctransport.CreateOrderRequest) (*rpctransport.Order, error)
	Orders(ctx context.Context) ([]*rpctransport.Order, error)
	OrderByID(ctx context.Context, id uint64) (*rpctransport.Order, error)
	UpdateOrder(ctx context.Context, req *rpctransport.UpdateOrderRequest) error
}

type ReviewsGRPCHandlers interface {
	Reviews(ctx context.Context) ([]rpctransport.Review, error)
}

type GRPCErrors interface {
	Error() string
}

type Server struct {
	rpc.UnimplementedOrdersServiceServer
	rpc.UnimplementedReviewsServiceServer

	OrdersGRPCHandlers  OrdersGRPCHandlers
	ReviewsGRPCHandlers ReviewsGRPCHandlers

	Validator *validator.Validate
	Logger    *zap.Logger

	grpcServer *grpc.Server
	httpServer *runtime.ServeMux
}

func (s *Server) Serve(serviceName string, port *int) error {
	if port == nil {
		return ErrPortNotSpecified
	}

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		return err
	}

	s.serveHealthCheck(serviceName)

	s.Logger.Info("starting grpc server", zap.Int("port", *port))

	if err = s.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("error while serve grpc | %w", err)
	}

	return nil
}

func (s *Server) ServeHTTP(port *int) error {
	if port == nil {
		return ErrPortNotSpecified
	}

	s.Logger.Info("starting http server", zap.Int("port", *port))

	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *port),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle CORS preflight requests
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "3600")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Add CORS headers to all responses
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.ProtoMajor == 2 && r.Header.Get("Content-Type") == "application/grpc" {
				s.grpcServer.ServeHTTP(w, r)
				return
			}

			if len(r.URL.Path) > 1 && r.URL.Path[len(r.URL.Path)-1] == '/' {
				http.Redirect(w, r, r.URL.Path[:len(r.URL.Path)-1], http.StatusPermanentRedirect)
				return
			}

			s.httpServer.ServeHTTP(w, r)
		})); err != nil {
		return fmt.Errorf("error while serve http | %w", err)
	}

	return nil
}

func (s *Server) serveHealthCheck(serviceName string) {
	healthCheckServer := health.NewServer()
	grpchealth.RegisterHealthServer(s.grpcServer, healthCheckServer)
	healthCheckServer.SetServingStatus(serviceName, grpchealth.HealthCheckResponse_SERVING)
}

func (s *Server) ServeWithCustomListener(l net.Listener) error {
	s.Logger.Info("starting grpc server with custom listener", zap.Int("port", l.Addr().(*net.TCPAddr).Port))

	if err := s.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("error while Serve grpc | %w", err)
	}

	return nil
}

func (s *Server) Close() {
	s.grpcServer.GracefulStop()
}

type NewServerOptions struct {
	ServiceName string

	OrdersGRPCHandlers  OrdersGRPCHandlers
	ReviewsGRPCHandlers ReviewsGRPCHandlers

	Logger    *zap.Logger
	Validator *validator.Validate

	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor

	ServerOptions []grpc.ServerOption
}

func NewServer(ctx context.Context, grpcPort int, opts *NewServerOptions) *Server {
	srvOpts := make([]grpc.ServerOption, 0)

	srvOpts = append(
		srvOpts,
		grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(opts.UnaryInterceptors...)),
		grpc.StreamInterceptor(grpcmiddleware.ChainStreamServer(opts.StreamInterceptors...)),
	)

	grpcServer := grpc.NewServer(srvOpts...)

	httpServer := runtime.NewServeMux()

	if opts.Validator == nil {
		opts.Validator = validator.New()
	}

	s := Server{
		UnimplementedOrdersServiceServer:  rpc.UnimplementedOrdersServiceServer{},
		UnimplementedReviewsServiceServer: rpc.UnimplementedReviewsServiceServer{},

		OrdersGRPCHandlers:  opts.OrdersGRPCHandlers,
		ReviewsGRPCHandlers: opts.ReviewsGRPCHandlers,

		Validator: opts.Validator,
		Logger:    opts.Logger,

		grpcServer: grpcServer,
		httpServer: httpServer,
	}
	rpc.RegisterOrdersServiceServer(grpcServer, &s)
	rpc.RegisterReviewsServiceServer(grpcServer, &s)

	httpOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := rpc.RegisterOrdersServiceHandlerFromEndpoint(
		ctx, httpServer, fmt.Sprintf("0.0.0.0:%v", grpcPort), httpOpts,
	); err != nil {
		panic(err)
	}

	if err := rpc.RegisterReviewsServiceHandlerFromEndpoint(
		ctx, httpServer, fmt.Sprintf("0.0.0.0:%v", grpcPort), httpOpts,
	); err != nil {
		panic(err)
	}

	reflection.Register(grpcServer)

	return &s
}

func (s *Server) CreateOrder(ctx context.Context, req *rpc.CreateOrderRequest) (*rpc.CreateOrderResponse, error) {
	rpcReq := &rpctransport.CreateOrderRequest{
		PropertySize:   rpctransport.PropertySize(req.PropertySize),
		MoveDate:       req.MoveDate.AsTime(),
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		MoveFrom:       req.MoveFrom,
		MoveTo:         req.MoveTo,
		AdditionalInfo: req.AdditionalInfo,
	}

	if err := validate(s.Validator, rpcReq, ErrValidationFailed); err != nil {
		return nil, err
	}

	order, err := s.OrdersGRPCHandlers.CreateOrder(ctx, rpcReq)
	if err != nil {
		return nil, GRPCUnknownError(err, nil)
	}

	return &rpc.CreateOrderResponse{Order: &rpc.Order{
		ID:             order.ID,
		PropertySize:   utils.PtrIfNotZero(rpc.PropertySize(order.PropertySize)),
		OrderStatus:    utils.PtrIfNotZero(rpc.OrderStatus(order.OrderStatus)),
		MoveDate:       timestamppb.New(order.MoveDate),
		Name:           &order.Name,
		Email:          &order.Email,
		Phone:          &order.Phone,
		MoveFrom:       &order.MoveFrom,
		MoveTo:         &order.MoveTo,
		AdditionalInfo: order.AdditionalInfo,
	}}, nil
}

func (s *Server) Orders(ctx context.Context, _ *emptypb.Empty) (*rpc.OrdersResponse, error) {
	rpcOrders, err := s.OrdersGRPCHandlers.Orders(ctx)
	if err != nil {
		if errors.Is(err, rpctransport.ErrNotFound) {
			return nil, GRPCNotFoundError(err, nil)
		}

		return nil, GRPCUnknownError(err, nil)
	}

	orders := make([]*rpc.Order, 0, len(rpcOrders))
	for _, order := range rpcOrders {
		propertySize := rpc.PropertySize(order.PropertySize)
		orderStatus := rpc.OrderStatus(order.OrderStatus)

		orders = append(orders, &rpc.Order{
			ID:             order.ID,
			PropertySize:   &propertySize,
			OrderStatus:    &orderStatus,
			MoveDate:       timestamppb.New(order.MoveDate),
			Name:           &order.Name,
			Email:          &order.Email,
			Phone:          &order.Phone,
			MoveFrom:       &order.MoveFrom,
			MoveTo:         &order.MoveTo,
			AdditionalInfo: order.AdditionalInfo,
		})
	}

	return &rpc.OrdersResponse{Orders: orders}, nil
}

func (s *Server) Order(ctx context.Context, req *rpc.OrderRequest) (*rpc.OrderResponse, error) {
	rpcOrder, err := s.OrdersGRPCHandlers.OrderByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, rpctransport.ErrNotFound) {
			return nil, GRPCNotFoundError(err, nil)
		}

		return nil, GRPCUnknownError(err, nil)
	}

	propertySize := rpc.PropertySize(rpcOrder.PropertySize)
	orderStatus := rpc.OrderStatus(rpcOrder.OrderStatus)

	return &rpc.OrderResponse{Order: &rpc.Order{
		ID:             rpcOrder.ID,
		PropertySize:   &propertySize,
		OrderStatus:    &orderStatus,
		MoveDate:       timestamppb.New(rpcOrder.MoveDate),
		Name:           &rpcOrder.Name,
		Email:          &rpcOrder.Email,
		Phone:          &rpcOrder.Phone,
		MoveFrom:       &rpcOrder.MoveFrom,
		MoveTo:         &rpcOrder.MoveTo,
		AdditionalInfo: rpcOrder.AdditionalInfo,
	}}, nil
}

func (s *Server) UpdateOrder(ctx context.Context, req *rpc.UpdateOrderRequest) (*emptypb.Empty, error) {
	rpcReq := &rpctransport.UpdateOrderRequest{
		ID:             req.GetID(),
		PropertySize:   utils.PtrIfNotZero(rpctransport.PropertySize(req.GetPropertySize())),
		OrderStatus:    utils.PtrIfNotZero(rpctransport.OrderStatus(req.GetOrderStatus())),
		Name:           utils.PtrIfNotZero(req.GetName()),
		Email:          utils.PtrIfNotZero(req.GetEmail()),
		Phone:          utils.PtrIfNotZero(req.GetPhone()),
		MoveFrom:       utils.PtrIfNotZero(req.GetMoveFrom()),
		MoveTo:         utils.PtrIfNotZero(req.GetMoveTo()),
		AdditionalInfo: utils.PtrIfNotZero(req.GetAdditionalInfo()),
	}

	if req.GetMoveDate() != nil {
		rpcReq.MoveDate = utils.PtrIfNotZero(req.GetMoveDate().AsTime())
	}

	if err := validate(s.Validator, rpcReq, ErrValidationFailed); err != nil {
		return nil, err
	}

	if err := s.OrdersGRPCHandlers.UpdateOrder(ctx, rpcReq); err != nil {
		return nil, GRPCUnknownError(err, nil)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Reviews(ctx context.Context, _ *emptypb.Empty) (*rpc.ReviewsResponse, error) {
	rpcReviews, err := s.ReviewsGRPCHandlers.Reviews(ctx)
	if err != nil {
		if errors.Is(err, rpctransport.ErrNotFound) {
			return nil, GRPCNotFoundError(err, nil)
		}

		return nil, GRPCUnknownError(err, nil)
	}

	reviews := make([]*rpc.Review, 0, len(rpcReviews))
	for _, review := range rpcReviews {
		reviews = append(reviews, &rpc.Review{
			Text:     review.Text,
			Name:     review.Name,
			Rate:     review.Rate,
			PhotoURL: review.PhotoURL,
		})
	}

	return &rpc.ReviewsResponse{Reviews: reviews}, nil
}

func GRPCValidationError[T GRPCErrors](reason T, err error) error {
	return gRPCError(codes.InvalidArgument, reason, err)
}

func GRPCUnknownError[T GRPCErrors](reason T, err error) error {
	return gRPCError(codes.Unknown, reason, err)
}

func GRPCNotFoundError[T GRPCErrors](reason T, err error) error {
	return gRPCError(codes.NotFound, reason, err)
}

func gRPCError[T GRPCErrors](code codes.Code, reason T, serviceErr error) error {
	if serviceErr == nil {
		serviceErr = errors.New("error not set")
	}

	st, err := status.Newf(code, "error: %v", serviceErr.Error()).WithDetails(
		&errdetails.ErrorInfo{
			Reason:   reason.Error(),
			Domain:   domain,
			Metadata: nil,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("unexpected error attaching metadata: %v", err))
	}

	return st.Err()
}

func validate[T GRPCErrors](v *validator.Validate, req any, reason T) error {
	if err := v.Struct(req); err != nil {
		return GRPCValidationError(reason, err)
	}

	return nil
}
