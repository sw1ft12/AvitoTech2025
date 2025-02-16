package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnection(ctx context.Context, address string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, address)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
