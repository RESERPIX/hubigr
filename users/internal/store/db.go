package store

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustOpen(ctx context.Context) *pgxpool.Pool {
	url := os.Getenv("DATABASE_URL") // postgres://user:pass@host:port/db?sslmode=disable
	if url == "" {
		panic("DATABASE_URL is empty")
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil { panic(err) }
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnIdleTime = 60 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil { panic(err) }
	if err := pool.Ping(ctx); err != nil { panic(err) }
	return pool
}
