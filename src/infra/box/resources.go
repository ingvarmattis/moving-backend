package box

import (
	"github.com/go-playground/validator/v10"
	interceptors2 "github.com/ingvarmattis/moving/src/infra/interceptors"
	"github.com/ingvarmattis/moving/src/infra/log"
	validator2 "github.com/ingvarmattis/moving/src/rpctransport/validator"
	"google.golang.org/grpc"

	movingRepo "github.com/ingvarmattis/moving/src/repositories/moving"
	movingSvc "github.com/ingvarmattis/moving/src/services/moving"
)

type Resources struct {
	MovingService *movingSvc.Service

	Validator *validator.Validate

	UnaryGRPCServerInterceptors  []grpc.UnaryServerInterceptor
	StreamGRPCServerInterceptors []grpc.StreamServerInterceptor
}

func NewResources(envBox *Env) *Resources {
	movingService := movingSvc.NewService(movingRepo.NewPostgres(envBox.PGXPool))

	return &Resources{
		MovingService: movingService,

		Validator: validator2.MustValidate(),

		UnaryGRPCServerInterceptors:  provideUnaryGRPCInterceptors(envBox),
		StreamGRPCServerInterceptors: provideStreamGRPCInterceptors(),
	}
}

func provideUnaryGRPCInterceptors(envBox *Env) []grpc.UnaryServerInterceptor {
	logger := envBox.Logger.With(log.Arg{Key: "rpc", Value: "grpc"}, log.Arg{Key: "type", Value: "unary"})

	return []grpc.UnaryServerInterceptor{
		interceptors2.UnaryServerMetricsInterceptor(envBox.Config.MetricsConfig.Enabled, envBox.Config.ServiceName),
		interceptors2.UnaryServerTraceInterceptor(envBox.Tracer, envBox.Config.ServiceName),
		interceptors2.UnaryServerLogInterceptor(logger, envBox.Config.Debug),
		interceptors2.UnaryServerPanicsInterceptor(logger, envBox.Config.ServiceName),
	}
}

func provideStreamGRPCInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{}
}
