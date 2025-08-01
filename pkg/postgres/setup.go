package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type API struct {
	Pool   *pgxpool.Pool
	Config Config
}

type Config struct {
	Host         string        `env:"DB_HOST"`
	Port         string        `env:"DB_PORT"`
	DBName       string        `env:"DB_NAME"`
	User         string        `env:"DB_USER"`
	Password     string        `env:"DB_PASSWORD"`
	MaxOpenConns int32         `env:"POSTGRES_MAX_OPEN_CONN" envDefault:"25"`
	MaxIdleTime  time.Duration `env:"POSTGRES_MAX_IDLE_TIME" envDefault:"15m"`
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName,
	)
}

func New(ctx context.Context, cfg Config) (*API, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxOpenConns
	poolConfig.MaxConnIdleTime = cfg.MaxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &API{
		Pool:   pool,
		Config: cfg,
	}, nil
}
