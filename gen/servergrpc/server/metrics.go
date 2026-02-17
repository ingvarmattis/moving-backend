package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// NotOperational is the name used for disabled/noop metrics server.
const NotOperational = "noop"

type MetricsServer struct {
	*http.Server
	name string
	port int

	logger *zap.Logger
}

func (m *MetricsServer) Addr() string {
	return m.Server.Addr
}

func (m *MetricsServer) Name() string {
	return m.name
}

func NewMetricsServer(enabled bool, logger *zap.Logger, port int) *MetricsServer {
	if !enabled {
		return &MetricsServer{
			name:   NotOperational,
			Server: nil,
			port:   port,
			logger: logger,
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &MetricsServer{
		name: "prometheus",
		Server: &http.Server{
			ReadHeaderTimeout: time.Minute,
			Handler:           mux,
			Addr:              ":" + strconv.Itoa(port),
		},
		port:   port,
		logger: logger,
	}
}

func (m *MetricsServer) ListenAndServe() error {
	m.logger.Info("starting http metrics server", zap.Int("port", m.port))

	if err := m.Server.ListenAndServe(); err != nil {
		return fmt.Errorf("cannot start http metrics server | %w", err)
	}

	return nil
}
