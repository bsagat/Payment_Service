package config

import (
	"log"
	"payment/pkg/postgres"
	"time"

	"github.com/bsagat/envzilla/v2"
)

type (
	Config struct {
		Postgres postgres.Config
		Server   Server
		Broker   Broker
		DevLevel string `env:"LEVEL"`
	}

	Server struct {
		GRPCServer GRPCServer
	}

	GRPCServer struct {
		Port                  string        `env:"GRPC_PORT" default:"50002"`
		MaxRecvMsgSizeMiB     int           `env:"GRPC_MAX_MESSAGE_SIZE_MIB" default:"12"`
		MaxConnectionAge      time.Duration `env:"GRPC_MAX_CONNECTION_AGE" default:"30s"`
		MaxConnectionAgeGrace time.Duration `env:"GRPC_MAX_CONNECTION_AGE_GRACE" default:"10s"`
	}

	Broker struct {
		Login    string `env:"BEREKE_MERCHANT_LOGIN"`
		Password string `env:"BEREKE_MERCHANT_PASSWORD"`
		Mode     string `env:"BEREKE_MERCHANT_MODE"`
	}
)

func MustLoad() Config {
	if err := envzilla.Loader(".env"); err != nil {
		log.Fatal("Failed to load configuration: ", err)
	}

	var cfg Config
	if err := envzilla.Parse(&cfg); err != nil {
		log.Fatal("Failed to parse configuration: ", err)
	}

	return cfg
}
