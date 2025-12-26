package box

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/ingvarmattis/moving/src/infra/interceptors"
	movingrepo "github.com/ingvarmattis/moving/src/repositories/orders"
	reviewsrepo "github.com/ingvarmattis/moving/src/repositories/reviews"
	rpcvalidator "github.com/ingvarmattis/moving/src/rpctransport/validator"
	orderssvc "github.com/ingvarmattis/moving/src/services/orders"
	reviewssvc "github.com/ingvarmattis/moving/src/services/reviews"
)

type Resources struct {
	OrdersService  *orderssvc.Service
	ReviewsService *reviewssvc.Service

	Validator *validator.Validate

	UnaryGRPCServerInterceptors  []grpc.UnaryServerInterceptor
	StreamGRPCServerInterceptors []grpc.StreamServerInterceptor
}

func NewResources(envBox *Env) *Resources {
	ordersService := orderssvc.NewService(movingrepo.NewPostgres(envBox.PGXPool))
	reviewsService := reviewssvc.NewService(reviewsrepo.NewPostgres(envBox.PGXPool))

	return &Resources{
		OrdersService:  ordersService,
		ReviewsService: reviewsService,

		Validator: rpcvalidator.MustValidate(),

		UnaryGRPCServerInterceptors:  provideUnaryGRPCInterceptors(envBox),
		StreamGRPCServerInterceptors: provideStreamGRPCInterceptors(),
	}
}

func provideUnaryGRPCInterceptors(envBox *Env) []grpc.UnaryServerInterceptor {
	logger := envBox.Logger.With(zap.String("rpc", "grpc"), zap.String("type", "unary"))

	return []grpc.UnaryServerInterceptor{
		interceptors.UnaryServerMetricsInterceptor(envBox.Config.MetricsConfig.Enabled, envBox.Config.ServiceName),
		interceptors.UnaryServerTraceInterceptor(envBox.Tracer, envBox.Config.ServiceName),
		interceptors.UnaryServerLogInterceptor(logger, envBox.Config.Debug),
		interceptors.UnaryServerAuthInterceptor(envBox.Config.AuthConfig.Tokens),
		interceptors.UnaryServerPanicsInterceptor(logger, envBox.Config.ServiceName),
	}
}

func provideStreamGRPCInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{}
}
