package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug bool `envconfig:"MOVING_SERVICE_DEBUG" default:"false"`

	GRPCServerListenPort int `envconfig:"MOVING_SERVICE_GRPC_SERVER_LISTEN_PORT" required:"true"`
	HTTPServerListenPort int `envconfig:"MOVING_SERVICE_HTTP_SERVER_LISTEN_PORT" required:"true"`

	HostName    string `envconfig:"MOVING_SERVICE_HOST_NAME"`
	ServiceName string `envconfig:"MOVING_SERVICE_SERVICE_NAME"`

	PostgresConfig PostgresConfig
	MetricsConfig  MetricsConfig
	TracingConfig  TracingConfig
	AuthConfig     AuthConfig
}

func FromEnv() (*Config, error) {
	cfg := &Config{}

	if hostName, err := os.Hostname(); err == nil {
		cfg.HostName = hostName
	}

	if err := envconfig.Process("MOVING_SERVICE", cfg); err != nil {
		return nil, fmt.Errorf("error while parsing environment variables | %w", err)
	}

	return cfg, nil
}

type PostgresConfig struct {
	ConnectionConfig string `envconfig:"MOVING_SERVICE_POSTGRES_URL" required:"true"`
}

type MetricsConfig struct {
	Enabled bool `envconfig:"MOVING_SERVICE_METRICS_ENABLED" required:"true"`
	Port    int  `envconfig:"MOVING_SERVICE_HTTP_METRICS_SERVER_LISTEN_PORT" required:"true"`
}

type TracingConfig struct {
	Enabled bool   `envconfig:"MOVING_SERVICE_OPENTELEMETRY_ENABLED" required:"true"`
	URL     string `envconfig:"MOVING_SERVICE_OPENTELEMETRY_COLLECTOR_URL" required:"true"`
	UseTLS  bool   `envconfig:"MOVING_SERVICE_OPENTELEMETRY_USE_TLS" required:"true"`
}

type AuthConfig struct {
	ClientTokens []string `envconfig:"MOVING_SERVICE_CLIENT_AUTH_TOKENS" required:"true"`
	AdminTokens  []string `envconfig:"MOVING_SERVICE_ADMIN_AUTH_TOKENS" required:"true"`
}
