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
		DevLevel string `env:"LEVEL"`
	}

	Server struct {
		GRPCServer GRPCServer
	}

	GRPCServer struct {
		Port                  string        `env:"GRPC_PORT" default:"50002"`
		MaxRecvMsgSizeMiB     int           `env:"GRPC_MAX_MESSAGE_SIZE_MIB" envDefault:"12"`
		MaxConnectionAge      time.Duration `env:"GRPC_MAX_CONNECTION_AGE" envDefault:"30s"`
		MaxConnectionAgeGrace time.Duration `env:"GRPC_MAX_CONNECTION_AGE_GRACE" envDefault:"10s"`
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
