package interceptors

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryServerLogInterceptor(logger *zap.Logger, debugMode bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startTime := time.Now()

		traceID := trace.SpanFromContext(ctx).SpanContext().TraceID()

		resp, err := handler(ctx, req)

		executionDuration := time.Since(startTime)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", executionDuration),
			zap.String("status", status.Code(err).String()),
		}

		if traceID.IsValid() {
			fields = append(fields, zap.String("traceID", traceID.String()))
		}

		if debugMode {
			fields = append(fields, zap.Any("request", req), zap.Any("response", resp))
		}

		logger.Info("incoming grpc request", fields...)

		return resp, err
	}
}
