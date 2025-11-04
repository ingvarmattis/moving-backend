package box

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"

	"github.com/ingvarmattis/moving/src/interceptors"
	movingRepo "github.com/ingvarmattis/moving/src/repositories/moving"
	"github.com/ingvarmattis/moving/src/rpctransport"
	movingSvc "github.com/ingvarmattis/moving/src/services/moving"
)

type Resources struct {
	MovingService *movingSvc.Service

	Validator *validator.Validate

	UnaryServerInterceptors  []grpc.UnaryServerInterceptor
	StreamServerInterceptors []grpc.StreamServerInterceptor
}

func NewResources(ctx context.Context, envBox *Env) (*Resources, error) {
	movingService, err := movingSvc.NewService(ctx, movingRepo.NewPostgres(envBox.PGXPool))
	if err != nil {
		return nil, fmt.Errorf("cannot create moving service | %w", err)
	}

	return &Resources{
		MovingService: movingService,

		Validator: rpctransport.MustValidate(),

		UnaryServerInterceptors:  provideUnaryInterceptors(envBox),
		StreamServerInterceptors: provideStreamInterceptors(),
	}, nil
}

func provideUnaryInterceptors(envBox *Env) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		interceptors.UnaryServerMetricsInterceptor(envBox.Config.MetricsConfig.Enabled, envBox.Config.ServiceName),
		interceptors.UnaryServerTraceInterceptor(envBox.Tracer, envBox.Config.ServiceName),
		interceptors.UnaryServerLogInterceptor(
			envBox.Logger.With("module", "log", "grpc"), envBox.Config.Debug,
		),
		interceptors.UnaryServerPanicsInterceptor(
			envBox.Logger.With("module", "log"), envBox.Config.ServiceName,
		),
	}
}

func provideStreamInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{}
}
