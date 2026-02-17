package box

import (
	"context"

	validatorv10 "github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/ingvarmattis/moving/gen/servergrpc/server"
	"github.com/ingvarmattis/moving/src/infra/interceptors"
	movingrepo "github.com/ingvarmattis/moving/src/repositories/orders"
	reviewsrepo "github.com/ingvarmattis/moving/src/repositories/reviews"
	orderssvc "github.com/ingvarmattis/moving/src/services/orders"
	reviewssvc "github.com/ingvarmattis/moving/src/services/reviews"
	"github.com/ingvarmattis/moving/src/transport/orders"
	"github.com/ingvarmattis/moving/src/transport/reviews"
	rpcvalidator "github.com/ingvarmattis/moving/src/transport/validator"
)

// TelegramBotInterface is the contract for telegram bot (real or noop).
type TelegramBotInterface interface {
	NotifyNewOrder(order *orders.Order)
	Start()
	Close()
}

type Resources struct {
	OrdersService  *orderssvc.Service
	ReviewsService *reviewssvc.Service

	Validator *validatorv10.Validate

	UnaryGRPCServerInterceptors  []grpc.UnaryServerInterceptor
	StreamGRPCServerInterceptors []grpc.StreamServerInterceptor

	GRPCServer    *server.Server
	TelegramBot   TelegramBotInterface
	MetricsServer *server.MetricsServer
}

func NewResources(ctx context.Context, envBox *Env) (*Resources, error) {
	ordersService := orderssvc.NewService(movingrepo.NewPostgres(envBox.PGXPool))
	reviewsService := reviewssvc.NewService(reviewsrepo.NewPostgres(envBox.PGXPool))

	validator := rpcvalidator.MustValidate()
	unaryInterceptors := provideUnaryGRPCInterceptors(envBox)
	streamInterceptors := provideStreamGRPCInterceptors()

	ordersHandlers := &orders.Handlers{OrdersService: ordersService}
	reviewsHandlers := &reviews.Handlers{ReviewsService: reviewsService}

	telegramBot, err := provideTelegramBot(envBox)
	if err != nil {
		return nil, err
	}

	grpcServer := provideGRPCServer(
		ctx, envBox, ordersHandlers, reviewsHandlers,
		telegramBot, validator, unaryInterceptors, streamInterceptors,
	)

	metricsServer := provideMetricsServer(envBox)

	return &Resources{
		OrdersService:  ordersService,
		ReviewsService: reviewsService,

		Validator: validator,

		UnaryGRPCServerInterceptors:  unaryInterceptors,
		StreamGRPCServerInterceptors: streamInterceptors,

		GRPCServer:    grpcServer,
		TelegramBot:   telegramBot,
		MetricsServer: metricsServer,
	}, nil
}

func provideGRPCServer(
	ctx context.Context,
	envBox *Env,
	ordersHandlers *orders.Handlers,
	reviewsHandlers *reviews.Handlers,
	telegramBot TelegramBotInterface,
	validator *validatorv10.Validate,
	unaryInterceptors []grpc.UnaryServerInterceptor,
	streamInterceptors []grpc.StreamServerInterceptor,
) *server.Server {
	return server.NewServer(
		ctx,
		envBox.Config.GRPCServerListenPort,
		&server.NewServerOptions{
			ServiceName:         envBox.Config.ServiceName,
			OrdersGRPCHandlers:  ordersHandlers,
			ReviewsGRPCHandlers: reviewsHandlers,
			NewOrderNotifier:    orderNotifier(telegramBot),
			Validator:           validator,
			Logger:              envBox.Logger,
			UnaryInterceptors:   unaryInterceptors,
			StreamInterceptors:  streamInterceptors,
		},
	)
}

func provideUnaryGRPCInterceptors(envBox *Env) []grpc.UnaryServerInterceptor {
	logger := envBox.Logger.With(zap.String("type", "unary"))

	return []grpc.UnaryServerInterceptor{
		interceptors.UnaryServerMetricsInterceptor(envBox.Config.MetricsConfig.Enabled, envBox.Config.ServiceName),
		interceptors.UnaryServerTraceInterceptor(envBox.Tracer, envBox.Config.ServiceName),
		interceptors.UnaryServerLogInterceptor(logger, envBox.Config.Debug),
		interceptors.UnaryServerAuthInterceptor(
			envBox.Config.AuthConfig.ClientTokens, envBox.Config.AuthConfig.AdminTokens,
		),
		interceptors.UnaryServerPanicsInterceptor(logger, envBox.Config.ServiceName),
	}
}

func provideStreamGRPCInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{}
}

func provideMetricsServer(envBox *Env) *server.MetricsServer {
	return server.NewMetricsServer(
		envBox.Config.MetricsConfig.Enabled, envBox.Logger, envBox.Config.MetricsConfig.Port,
	)
}

func orderNotifier(bot TelegramBotInterface) func(*orders.Order) {
	return bot.NotifyNewOrder
}

func provideTelegramBot(envBox *Env) (TelegramBotInterface, error) {
	if !envBox.Config.TelegramConfig.Enabled {
		return server.NewNoopTelegramBot(), nil
	}

	telegramBot, err := server.NewTelegramBot(
		envBox.Logger,
		envBox.Config.TelegramConfig.Token, envBox.Config.TelegramConfig.Timeout,
		envBox.Config.TelegramConfig.AllowedChatIDs,
	)
	if err != nil {
		return nil, err
	}

	return telegramBot, nil
}
