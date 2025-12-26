package interceptors

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func UnaryServerTraceInterceptor(tracer trace.Tracer, serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		span.SetAttributes(attribute.String("product", serviceName))

		resp, err := handler(ctx, req)

		setSpanStatus(span, err)

		return resp, err
	}
}

func setSpanStatus(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, "")
	} else {
		span.SetStatus(otelcodes.Ok, "")
	}
}
