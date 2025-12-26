package interceptors

import (
	"context"
	"errors"
	"runtime/debug"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var ErrPanicHandled = errors.New("panic handled")

func UnaryServerPanicsInterceptor(logger *zap.Logger, serviceName string) grpc.UnaryServerInterceptor {
	serviceName = strings.ReplaceAll(serviceName, "-", "_")

	panicsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: serviceName,
		Subsystem: "grpc",
		Name:      "panics_count",
		Help:      "Panics count by method.",
	}, []string{"method"})

	prometheus.MustRegister(panicsCounter)

	return func(
		ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp any, err error) {
		method := extractShortMethodName(info.FullMethod)

		defer func() {
			if r := recover(); r != nil {
				logger.Warn("panic: " + string(debug.Stack()))
				panicsCounter.WithLabelValues(method).Inc()

				err = ErrPanicHandled
				resp = nil
			}
		}()

		return handler(ctx, req)
	}
}
